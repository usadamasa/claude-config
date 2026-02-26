package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScanResult はセッションログから抽出されたツール使用情報を表す｡
type ScanResult struct {
	ToolName string // "Bash", "Read", "Write", "Edit"
	Pattern  string // コマンドプレフィックス or パスパターン
	FilePath string // JSONL ファイルパス
}

// jsonlLine はセッション JSONL ファイルの1行を表す｡
type jsonlLine struct {
	Type    string `json:"type"`
	Message struct {
		Content []contentBlock `json:"content"`
	} `json:"message"`
}

// contentBlock は message.content[] の要素を表す｡
type contentBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// bashInput は Bash tool_use の入力フィールド｡
type bashInput struct {
	Command string `json:"command"`
}

// fileInput は Read/Write/Edit tool_use の入力フィールド｡
type fileInput struct {
	FilePath string `json:"file_path"`
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

// ScanJSONLFiles は指定ディレクトリの JSONL ファイルから Bash/Read/Write/Edit
// の tool_use エントリを抽出する｡
func ScanJSONLFiles(projectsDir string, days int) ([]ScanResult, error) {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	var results []ScanResult

	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return results, nil
	}

	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			return nil
		}

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
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlLine
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

// ExtractBashPrefix はコマンド文字列から先頭1〜2語のプレフィックスを抽出する｡
// サブコマンド付きツール(git, go, gh 等)は2語目まで取得する｡
// パイプ・リダイレクト・連結演算子より前だけを対象にする｡
func ExtractBashPrefix(command string) string {
	if command == "" {
		return ""
	}

	// パイプ・リダイレクト・連結演算子で切る
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

	// rm -rf は特殊ケース: フラグも含める
	if fields[0] == "rm" && len(fields) > 1 && strings.HasPrefix(fields[1], "-") && strings.Contains(fields[1], "r") {
		return "rm " + fields[1]
	}

	// サブコマンド付きツールは2語目まで取得
	if subcommandTools[fields[0]] && len(fields) > 1 {
		return fields[0] + " " + fields[1]
	}

	return fields[0]
}

// NormalizePath はファイルパスを settings.json のパーミッションパターンに正規化する｡
// - /Users/<user>/... → ~/... に変換
// - ディレクトリ付きパスは親ディレクトリ/** にパターン化
// - ファイル名のみやルート直下のファイルはそのまま
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}

	// ホームディレクトリのプレフィックスを ~ に変換
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		path = "~" + path[len(home):]
	}

	// /Users/<任意のユーザー>/... の場合も ~ に変換
	if strings.HasPrefix(path, "/Users/") {
		parts := strings.SplitN(path, "/", 4) // ["", "Users", "user", "rest"]
		if len(parts) >= 4 {
			path = "~/" + parts[3]
		}
	}

	// ディレクトリ構造に基づいてパターン化
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		// ルートレベルのファイルはそのまま返す
		return filepath.Base(path)
	}

	// ~ 直下のファイル(例: ~/.zshrc)はそのまま
	if dir == "~" {
		return path
	}

	// ディレクトリが2階層以上ある場合、最初の2階層/** にする
	// 例: ~/.ssh/id_rsa → ~/.ssh/**
	// 例: src/main.go → src/**
	// 例: .claude/skills/foo/SKILL.md → .claude/skills/**
	parts := strings.Split(dir, "/")
	depth := 2
	if strings.HasPrefix(dir, "~") {
		// ~ をプレフィックスとして1階層分追加
		// ~/.ssh → ["~", ".ssh"] → depth 3 だが len=2 なのでそのまま
		// ~/.config/git → ["~", ".config", "git"] → depth 3 でちょうど良い
		depth = 3
	}

	if len(parts) > depth {
		return strings.Join(parts[:depth], "/") + "/**"
	}

	return dir + "/**"
}
