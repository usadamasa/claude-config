package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/usadamasa/claude-config/internal/jsonlscan"
)

// ScanResult はセッションログから抽出されたツール使用情報を表す｡
type ScanResult struct {
	ToolName string
	Pattern  string
	FilePath string
}

// bashInput は Bash tool_use の入力フィールド｡
type bashInput struct {
	Command string `json:"command"`
}

// fileInput は Read/Write/Edit tool_use の入力フィールド｡
type fileInput struct {
	FilePath string `json:"file_path"`
}

// ScanJSONLFiles は指定ディレクトリの JSONL ファイルから Bash/Read/Write/Edit
// の tool_use エントリを抽出する｡
func ScanJSONLFiles(projectsDir string, days int) ([]ScanResult, error) {
	var results []ScanResult

	err := jsonlscan.WalkJSONLFiles(projectsDir, jsonlscan.WalkOptions{Days: days}, func(path string) error {
		fileResults, err := scanSingleFile(path)
		if err != nil {
			return nil
		}
		results = append(results, fileResults...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// scanSingleFile は JSONL ファイルを1行ずつ読み取り、対象ツールのエントリを抽出する｡
func scanSingleFile(path string) ([]ScanResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var results []ScanResult
	scanner := jsonlscan.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlscan.JSONLLine
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		for _, block := range entry.Message.Content {
			if block.Type != "tool_use" {
				continue
			}
			if !targetTools[block.Name] {
				continue
			}

			result := ScanResult{
				ToolName: block.Name,
				FilePath: path,
			}

			switch block.Name {
			case "Bash":
				var input bashInput
				if err := json.Unmarshal(block.Input, &input); err != nil {
					continue
				}
				if input.Command == "" {
					continue
				}
				result.Pattern = ExtractBashPrefix(input.Command)
			case "Read", "Write", "Edit":
				var input fileInput
				if err := json.Unmarshal(block.Input, &input); err != nil {
					continue
				}
				if input.FilePath == "" {
					continue
				}
				result.Pattern = NormalizePath(input.FilePath)
			}

			results = append(results, result)
		}
	}
	return results, scanner.Err()
}

// targetTools はスキャン対象のツール名セット｡
var targetTools = map[string]bool{
	"Bash":  true,
	"Read":  true,
	"Write": true,
	"Edit":  true,
}

// subcommandTools は2語目までプレフィックスとして取得するコマンド群｡
var subcommandTools = map[string]bool{
	"git":    true,
	"go":     true,
	"gh":     true,
	"docker": true,
	"task":   true,
	"brew":   true,
	"make":   true,
}

// ExtractBashPrefix はコマンド文字列から先頭1〜2語のプレフィックスを抽出する｡
func ExtractBashPrefix(command string) string {
	if command == "" {
		return ""
	}

	for _, sep := range []string{"|", ">>", ">", "&&", ";"} {
		if idx := strings.Index(command, sep); idx >= 0 {
			command = command[:idx]
		}
	}
	command = strings.TrimSpace(command)

	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}

	if fields[0] == "rm" && len(fields) > 1 && strings.HasPrefix(fields[1], "-") && strings.Contains(fields[1], "r") {
		return "rm " + fields[1]
	}

	if subcommandTools[fields[0]] && len(fields) > 1 {
		return fields[0] + " " + fields[1]
	}

	return fields[0]
}

// NormalizePath はファイルパスを settings.json のパーミッションパターンに正規化する｡
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}

	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}

	if strings.HasPrefix(path, "/Users/") {
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 4 {
			path = "~/" + parts[3]
		}
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return filepath.Base(path)
	}

	if dir == "~" {
		return path
	}

	parts := strings.Split(dir, "/")
	depth := 2
	if strings.HasPrefix(dir, "~") {
		depth = 3
	}

	if len(parts) > depth {
		return strings.Join(parts[:depth], "/") + "/**"
	}

	return dir + "/**"
}
