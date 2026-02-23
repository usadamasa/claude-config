package main

import (
	"testing"
)

func TestLoadPermissions(t *testing.T) {
	t.Run("allow/deny/ask を正しくパースする", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [
      "Bash(git status:*)",
      "Bash(go test:*)",
      "Read(CLAUDE.md)",
      "Write(src/**)"
    ],
    "deny": [
      "Bash(curl:*)",
      "Read(~/.ssh/**)"
    ],
    "ask": [
      "Bash(git commit:*)",
      "Bash(git push:*)"
    ]
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		allow, deny, ask, err := LoadPermissions(path)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(allow) != 4 {
			t.Errorf("allow 数: got %d, want 4", len(allow))
		}
		if len(deny) != 2 {
			t.Errorf("deny 数: got %d, want 2", len(deny))
		}
		if len(ask) != 2 {
			t.Errorf("ask 数: got %d, want 2", len(ask))
		}
	})

	t.Run("空のパーミッションを処理する", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		allow, deny, ask, err := LoadPermissions(path)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(allow) != 0 || len(deny) != 0 || len(ask) != 0 {
			t.Errorf("空でない結果: allow=%d, deny=%d, ask=%d", len(allow), len(deny), len(ask))
		}
	})

	t.Run("permissions キーがない場合", func(t *testing.T) {
		settingsJSON := `{"model": "claude-opus-4-6"}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		allow, deny, ask, err := LoadPermissions(path)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(allow) != 0 || len(deny) != 0 || len(ask) != 0 {
			t.Errorf("空でない結果: allow=%d, deny=%d, ask=%d", len(allow), len(deny), len(ask))
		}
	})

	t.Run("不正な JSON でエラーを返す", func(t *testing.T) {
		path := writeTestFile(t, t.TempDir(), "settings.json", "not json")
		_, _, _, err := LoadPermissions(path)
		if err == nil {
			t.Fatal("エラーが期待されるがnilが返された")
		}
	})

	t.Run("存在しないファイルでエラーを返す", func(t *testing.T) {
		_, _, _, err := LoadPermissions("/nonexistent/settings.json")
		if err == nil {
			t.Fatal("エラーが期待されるがnilが返された")
		}
	})
}

func TestParsePermissionEntry(t *testing.T) {
	tests := []struct {
		name        string
		entry       string
		wantTool    string
		wantPattern string
		wantOk      bool
	}{
		{"Bash with glob", "Bash(git status:*)", "Bash", "git status", true},
		{"Bash without glob", "Bash(git status)", "Bash", "git status", true},
		{"Read path", "Read(~/.ssh/**)", "Read", "~/.ssh/**", true},
		{"Write path", "Write(src/**)", "Write", "src/**", true},
		{"Edit path", "Edit(~/.claude/**)", "Edit", "~/.claude/**", true},
		{"WebFetch (対象外)", "WebFetch(domain:github.com)", "", "", false},
		{"ベアエントリ", "Bash", "Bash", "", true},
		{"ベアエントリ WebSearch", "WebSearch", "", "", false},
		{"MCP ツール", "mcp__obsidian__*", "", "", false},
		{"Skill", "Skill(commit-commands:commit-push-pr)", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, pattern, ok := ParsePermissionEntry(tt.entry)
			if ok != tt.wantOk {
				t.Errorf("ok: got %v, want %v", ok, tt.wantOk)
			}
			if ok {
				if tool != tt.wantTool {
					t.Errorf("tool: got %q, want %q", tool, tt.wantTool)
				}
				if pattern != tt.wantPattern {
					t.Errorf("pattern: got %q, want %q", pattern, tt.wantPattern)
				}
			}
		})
	}
}

func TestMatchesPermission(t *testing.T) {
	permissions := []string{
		"Bash(git status:*)",
		"Bash(go test:*)",
		"Read(CLAUDE.md)",
		"Read(~/.claude/**)",
		"Write(src/**)",
	}

	tests := []struct {
		name     string
		toolName string
		pattern  string
		want     bool
	}{
		{"完全一致", "Bash", "git status", true},
		{"一致しない", "Bash", "curl", false},
		{"Read 一致", "Read", "CLAUDE.md", true},
		{"Read ワイルドカード", "Read", "~/.claude/skills/foo", true},
		{"Write ワイルドカード", "Write", "src/main.go", true},
		{"Write 不一致", "Write", "docs/README.md", false},
		{"ツール名不一致", "Write", "git status", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesPermission(tt.toolName, tt.pattern, permissions)
			if got != tt.want {
				t.Errorf("MatchesPermission(%q, %q) = %v, want %v", tt.toolName, tt.pattern, got, tt.want)
			}
		})
	}
}
