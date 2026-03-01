package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 許可するホームディレクトリ配下のサブディレクトリ
var allowedSubdirs = []string{".claude", "obsidian", "src", "tmp", "workspace"}

// hookInput は PreToolUse フックの入力 JSON 構造
type hookInput struct {
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
	CWD       string                 `json:"cwd"`
}

// denyResponse は deny 時の出力 JSON 構造
type denyResponse struct {
	HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
}

type hookOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

func main() {
	// --help フラグ対応 (Docker verify.sh での実行可能性チェック用)
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Fprintf(os.Stderr, "使い方: guard-home-dir < JSON\nPreToolUse フック: ホームディレクトリ走査を防止する\n")
		os.Exit(0)
	}

	// stdin から JSON 入力を読み取り
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stdin 読み取りエラー: %v\n", err)
		os.Exit(1)
	}

	var input hookInput
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Fprintf(os.Stderr, "JSON パースエラー: %v\n", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ホームディレクトリ取得エラー: %v\n", err)
		os.Exit(1)
	}
	// macOS の /var → /private/var 等の symlink を解決
	if resolved, err := resolveRealpath(home); err == nil {
		home = resolved
	}

	cwd := input.CWD
	if cwd != "" {
		if resolved, err := resolveRealpath(cwd); err == nil {
			cwd = resolved
		}
	}

	var paths []string

	switch input.ToolName {
	case "Read", "Edit", "Write", "NotebookEdit":
		p := extractFilePath(input.ToolName, input.ToolInput)
		if p != "" {
			paths = append(paths, p)
		}
	case "Glob", "Grep":
		p := extractFilePath(input.ToolName, input.ToolInput)
		if p == "" {
			// パスが空なら cwd → 通過
			os.Exit(0)
		}
		paths = append(paths, p)
	case "Bash":
		command, ok := input.ToolInput["command"].(string)
		if !ok || command == "" {
			os.Exit(0)
		}
		targets := extractScanTargets(command, home)
		if targets == nil {
			// スキャンコマンドでない → 通過
			os.Exit(0)
		}
		paths = targets
	default:
		// 未知のツール → 通過
		os.Exit(0)
	}

	// パスが空なら通過
	if len(paths) == 0 {
		os.Exit(0)
	}

	// 各パスを正規化してチェック
	var resolvedPaths []string
	for _, p := range paths {
		// 相対パスを絶対パスに変換
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		// シンボリックリンク解決
		resolved, err := resolveRealpath(p)
		if err != nil {
			resolved = filepath.Clean(p)
		}
		resolvedPaths = append(resolvedPaths, resolved)
	}

	// 許可パスチェック
	deniedPath := checkPaths(resolvedPaths, home, cwd)
	if deniedPath == "" {
		os.Exit(0)
	}

	// deny JSON 出力
	resp := denyResponse{
		HookSpecificOutput: hookOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: fmt.Sprintf("ホームディレクトリ走査防止: %s はプロジェクトディレクトリおよび許可パスの外にあるためアクセスできません", deniedPath),
		},
	}
	out, _ := json.Marshal(resp)
	out = append(out, '\n')
	os.Stdout.Write(out)
}

// extractFilePath はツール名に応じたファイルパスを tool_input から抽出する｡
func extractFilePath(toolName string, input map[string]any) string {
	var key string
	switch toolName {
	case "Read", "Edit", "Write":
		key = "file_path"
	case "NotebookEdit":
		key = "notebook_path"
	case "Glob", "Grep":
		key = "path"
	default:
		return ""
	}

	v, ok := input[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// checkPaths はパスリストを検証し、禁止パスがあればそのパスを返す｡
// 全て許可なら空文字列を返す｡
func checkPaths(paths []string, home string, cwd string) string {
	if len(paths) == 0 {
		return ""
	}

	// 許可パスリストを構築
	allowed := make([]string, 0, len(allowedSubdirs)+1)
	for _, sub := range allowedSubdirs {
		allowed = append(allowed, filepath.Join(home, sub))
	}
	if cwd != "" {
		allowed = append(allowed, cwd)
	}

	for _, p := range paths {
		// ホームディレクトリ配下でなければ通過
		if !strings.HasPrefix(p, home+"/") && p != home {
			continue
		}

		// 許可パスに前方一致するか
		isAllowed := false
		for _, a := range allowed {
			if p == a || strings.HasPrefix(p, a+"/") {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return p
		}
	}

	return ""
}
