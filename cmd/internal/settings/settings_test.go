package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("有効なsettings.jsonを正しくパースする", func(t *testing.T) {
		tmp := t.TempDir()
		f := filepath.Join(tmp, "settings.json")
		content := `{
			"permissions": {
				"allow": ["Bash(git status:*)"],
				"deny": ["Bash(curl:*)"],
				"ask": ["Bash(git push:*)"]
			},
			"sandbox": {
				"network": {
					"allowedDomains": ["github.com", "*.anthropic.com"]
				}
			}
		}`
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		s, err := Load(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(s.Permissions.Allow) != 1 {
			t.Errorf("Allow: got %d entries, want 1", len(s.Permissions.Allow))
		}
		if len(s.Permissions.Deny) != 1 {
			t.Errorf("Deny: got %d entries, want 1", len(s.Permissions.Deny))
		}
		if len(s.Permissions.Ask) != 1 {
			t.Errorf("Ask: got %d entries, want 1", len(s.Permissions.Ask))
		}
		if len(s.Sandbox.Network.AllowedDomains) != 2 {
			t.Errorf("AllowedDomains: got %d entries, want 2", len(s.Sandbox.Network.AllowedDomains))
		}
	})

	t.Run("存在しないファイルはエラーを返す", func(t *testing.T) {
		_, err := Load("/nonexistent/settings.json")
		if err == nil {
			t.Error("エラーが返されるべき")
		}
	})

	t.Run("不正なJSONはエラーを返す", func(t *testing.T) {
		tmp := t.TempDir()
		f := filepath.Join(tmp, "settings.json")
		if err := os.WriteFile(f, []byte("{invalid"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := Load(f)
		if err == nil {
			t.Error("エラーが返されるべき")
		}
	})

	t.Run("フィールドが省略されても正常にパースする", func(t *testing.T) {
		tmp := t.TempDir()
		f := filepath.Join(tmp, "settings.json")
		if err := os.WriteFile(f, []byte(`{}`), 0644); err != nil {
			t.Fatal(err)
		}

		s, err := Load(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Permissions.Allow != nil {
			t.Errorf("Allow should be nil, got %v", s.Permissions.Allow)
		}
	})
}
