package main

import (
	"strings"

	"github.com/usadamasa/claude-config/internal/settings"
)

// targetToolNames はパーミッション分析対象のツール名セット｡
var targetToolNames = map[string]bool{
	"Bash":  true,
	"Read":  true,
	"Write": true,
	"Edit":  true,
}

// LoadPermissions は settings.json から allow, deny, ask リストを読み込む｡
func LoadPermissions(settingsPath string) (allow, deny, ask []string, err error) {
	s, err := settings.Load(settingsPath)
	if err != nil {
		return nil, nil, nil, err
	}
	return s.Permissions.Allow, s.Permissions.Deny, s.Permissions.Ask, nil
}

// ParsePermissionEntry はパーミッション文字列をツール名とパターンに分解する｡
// 対象ツール(Bash, Read, Write, Edit)のエントリのみ ok=true を返す｡
func ParsePermissionEntry(entry string) (tool, pattern string, ok bool) {
	if !strings.Contains(entry, "(") {
		if targetToolNames[entry] {
			return entry, "", true
		}
		return "", "", false
	}

	parenIdx := strings.Index(entry, "(")
	if parenIdx < 0 || !strings.HasSuffix(entry, ")") {
		return "", "", false
	}

	tool = entry[:parenIdx]
	if !targetToolNames[tool] {
		return "", "", false
	}

	inner := entry[parenIdx+1 : len(entry)-1]
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

		if permPattern == "" {
			return true
		}

		if permPattern == pattern {
			return true
		}

		if strings.HasPrefix(pattern, permPattern+" ") || strings.HasPrefix(pattern, permPattern+"/") {
			return true
		}

		if strings.HasSuffix(permPattern, "/**") {
			prefix := permPattern[:len(permPattern)-3]
			if strings.HasPrefix(pattern, prefix+"/") || pattern == prefix {
				return true
			}
		}

		if strings.HasSuffix(permPattern, "**") {
			prefix := permPattern[:len(permPattern)-2]
			if strings.HasPrefix(pattern, prefix) {
				return true
			}
		}
	}
	return false
}
