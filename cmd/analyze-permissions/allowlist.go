package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// settingsJSON は settings.json のパーミッション関連部分を表す｡
type settingsJSON struct {
	Permissions struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
		Ask   []string `json:"ask"`
	} `json:"permissions"`
}

// targetToolNames はパーミッション分析対象のツール名セット｡
var targetToolNames = map[string]bool{
	"Bash":  true,
	"Read":  true,
	"Write": true,
	"Edit":  true,
}

// LoadPermissions は settings.json から allow, deny, ask リストを読み込む｡
func LoadPermissions(settingsPath string) (allow, deny, ask []string, err error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("settings ファイルの読み込みに失敗: %w", err)
	}

	var settings settingsJSON
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, nil, nil, fmt.Errorf("settings JSON のパースに失敗: %w", err)
	}

	return settings.Permissions.Allow, settings.Permissions.Deny, settings.Permissions.Ask, nil
}

// ParsePermissionEntry はパーミッション文字列をツール名とパターンに分解する｡
// 対象ツール(Bash, Read, Write, Edit)のエントリのみ ok=true を返す｡
// 例: "Bash(git status:*)" → tool="Bash", pattern="git status", ok=true
// 例: "Bash" → tool="Bash", pattern="", ok=true (ベアエントリ)
// 例: "WebFetch(domain:github.com)" → ok=false
func ParsePermissionEntry(entry string) (tool, pattern string, ok bool) {
	// 括弧なしのベアエントリ
	if !strings.Contains(entry, "(") {
		if targetToolNames[entry] {
			return entry, "", true
		}
		return "", "", false
	}

	// 括弧付きエントリ: Tool(pattern) or Tool(pattern:*)
	parenIdx := strings.Index(entry, "(")
	if parenIdx < 0 || !strings.HasSuffix(entry, ")") {
		return "", "", false
	}

	tool = entry[:parenIdx]
	if !targetToolNames[tool] {
		return "", "", false
	}

	inner := entry[parenIdx+1 : len(entry)-1]

	// ":*" サフィックスを除去(例: "git status:*" → "git status")
	inner = strings.TrimSuffix(inner, ":*")

	return tool, inner, true
}

// MatchesPermission はツール名とパターンが既存のパーミッションリストにマッチするか判定する｡
func MatchesPermission(toolName, pattern string, permissions []string) bool {
	for _, perm := range permissions {
		permTool, permPattern, ok := ParsePermissionEntry(perm)
		if !ok {
			continue
		}
		if permTool != toolName {
			continue
		}

		// ベアエントリは全パターンにマッチ
		if permPattern == "" {
			return true
		}

		// 完全一致
		if permPattern == pattern {
			return true
		}

		// ワイルドカードマッチ: "~/.claude/**" は "~/.claude/skills/foo" にマッチ
		if strings.HasSuffix(permPattern, "/**") {
			prefix := permPattern[:len(permPattern)-3]
			if strings.HasPrefix(pattern, prefix+"/") || pattern == prefix {
				return true
			}
		}

		// パターン末尾の ** マッチ: "src/**" は "src/main.go" にマッチ
		if strings.HasSuffix(permPattern, "**") {
			prefix := permPattern[:len(permPattern)-2]
			if strings.HasPrefix(pattern, prefix) {
				return true
			}
		}
	}
	return false
}
