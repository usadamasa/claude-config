package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/usadamasa/claude-config/internal/jsonlscan"
)

// ModelTokens はモデル別のtoken使用量を表す｡
type ModelTokens struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	CallCount    int   `json:"call_count"`
}

// SessionResult はセッション1つ分のtoken使用量集計結果を表す｡
type SessionResult struct {
	SessionID                string                 `json:"session_id"`
	Project                  string                 `json:"project"`
	Model                    string                 `json:"model"`
	TotalInputTokens         int64                  `json:"total_input_tokens"`
	TotalOutputTokens        int64                  `json:"total_output_tokens"`
	TotalCacheCreationTokens int64                  `json:"total_cache_creation_tokens"`
	TotalCacheReadTokens     int64                  `json:"total_cache_read_tokens"`
	APICallCount             int                    `json:"api_call_count"`
	UserMessageCount         int                    `json:"user_message_count"`
	ModelUsage               map[string]ModelTokens `json:"model_usage"`
	ToolUsage                map[string]int         `json:"tool_usage"`
	FilePath                 string                 `json:"file_path"`
}

// AverageInputTokensPerCall は1APIコールあたりの平均input tokensを返す｡
func (r *SessionResult) AverageInputTokensPerCall() int64 {
	if r.APICallCount == 0 {
		return 0
	}
	return r.TotalInputTokens / int64(r.APICallCount)
}

// jsonlEntry はセッションJSONLファイルの1行を表す｡
type jsonlEntry struct {
	Type     string          `json:"type"`
	CWD      string          `json:"cwd"`
	Session  string          `json:"sessionId"`
	UserType string          `json:"userType"`
	Message  json.RawMessage `json:"message"`
	Data     json.RawMessage `json:"data"`
}

// assistantMessage はassistantエントリのmessageフィールド｡
type assistantMessage struct {
	Model   string                  `json:"model"`
	Content []jsonlscan.ContentBlock `json:"content"`
	Usage   tokenUsage              `json:"usage"`
}

// tokenUsage はAPIレスポンスのusageフィールド｡
type tokenUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
}

// progressData はprogressエントリのdataフィールド｡
type progressData struct {
	Message *progressMessage `json:"message"`
}

// progressMessage はprogress.data.messageフィールド｡
type progressMessage struct {
	Type    string           `json:"type"`
	Message *assistantMessage `json:"message"`
}

// ScanSessionFile は1つのセッションJSONLファイルからtoken使用量を集計する｡
func ScanSessionFile(path string) (*SessionResult, error) {
	f, err := os.Open(path) // #nosec G304 -- CLIツール: パスはWalkDir由来
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	result := &SessionResult{
		ModelUsage: make(map[string]ModelTokens),
		ToolUsage:  make(map[string]int),
		FilePath:   path,
	}

	scanner := jsonlscan.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if result.SessionID == "" && entry.Session != "" {
			result.SessionID = entry.Session
		}
		if result.Project == "" && entry.CWD != "" {
			result.Project = ExtractProjectName(entry.CWD)
		}

		switch entry.Type {
		case "assistant":
			processAssistantEntry(result, entry.Message)
		case "progress":
			processProgressEntry(result, entry.Data)
		case "user":
			if entry.UserType == "external" {
				result.UserMessageCount++
			}
		}
	}

	return result, scanner.Err()
}

func processAssistantEntry(result *SessionResult, raw json.RawMessage) {
	if raw == nil {
		return
	}
	var msg assistantMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}

	addUsage(result, msg.Model, msg.Usage)

	if result.Model == "" && msg.Model != "" {
		result.Model = msg.Model
	}

	for _, block := range msg.Content {
		if block.Type == "tool_use" && block.Name != "" {
			result.ToolUsage[block.Name]++
		}
	}

	result.APICallCount++
}

func processProgressEntry(result *SessionResult, raw json.RawMessage) {
	if raw == nil {
		return
	}
	var data progressData
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}
	if data.Message == nil || data.Message.Message == nil {
		return
	}

	msg := data.Message.Message
	if msg.Usage.InputTokens == 0 && msg.Usage.OutputTokens == 0 {
		return
	}

	addUsage(result, msg.Model, msg.Usage)

	for _, block := range msg.Content {
		if block.Type == "tool_use" && block.Name != "" {
			result.ToolUsage[block.Name]++
		}
	}

	result.APICallCount++
}

func addUsage(result *SessionResult, model string, usage tokenUsage) {
	result.TotalInputTokens += usage.InputTokens
	result.TotalOutputTokens += usage.OutputTokens
	result.TotalCacheCreationTokens += usage.CacheCreationInputTokens
	result.TotalCacheReadTokens += usage.CacheReadInputTokens

	if model != "" {
		mt := result.ModelUsage[model]
		mt.InputTokens += usage.InputTokens
		mt.OutputTokens += usage.OutputTokens
		mt.CallCount++
		result.ModelUsage[model] = mt
	}
}

// ScanProjectsDir は指定ディレクトリ以下の全JONLファイルを走査してtoken使用量を集計する｡
func ScanProjectsDir(projectsDir string, days int) ([]SessionResult, error) {
	var results []SessionResult

	err := jsonlscan.WalkJSONLFiles(projectsDir, jsonlscan.WalkOptions{Days: days}, func(path string) error {
		result, err := ScanSessionFile(path)
		if err != nil {
			return nil
		}
		if result.APICallCount > 0 {
			results = append(results, *result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ExtractProjectName はcwdからプロジェクト名を抽出する｡
func ExtractProjectName(cwd string) string {
	if cwd == "" {
		return ""
	}

	parts := strings.Split(cwd, "/")

	for i, p := range parts {
		if p == "worktree" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}
