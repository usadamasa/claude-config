package main

import (
	"testing"
)

func TestLoadAllowlist(t *testing.T) {
	t.Run("parses WebFetch domain entries", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [
      "WebFetch(domain:github.com)",
      "WebFetch(domain:docs.example.com)",
      "Bash(git status)"
    ]
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		entries, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Tool != "WebFetch" || entries[0].Domain != "github.com" {
			t.Errorf("unexpected first entry: %+v", entries[0])
		}
		if entries[1].Tool != "WebFetch" || entries[1].Domain != "docs.example.com" {
			t.Errorf("unexpected second entry: %+v", entries[1])
		}
	})

	t.Run("parses Fetch domain entries", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [
      "Fetch(domain:api.example.com)"
    ]
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		entries, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Tool != "Fetch" || entries[0].Domain != "api.example.com" {
			t.Errorf("unexpected entry: %+v", entries[0])
		}
	})

	t.Run("handles wildcard domains", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [
      "WebFetch(domain:*.databricks.com)"
    ]
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		entries, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Domain != "*.databricks.com" {
			t.Errorf("expected wildcard domain *.databricks.com, got %s", entries[0].Domain)
		}
	})

	t.Run("handles mixed WebFetch and Fetch entries", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [
      "WebFetch(domain:github.com)",
      "Fetch(domain:api.github.com)",
      "Bash(git log)",
      "WebFetch(domain:*.anthropic.com)"
    ]
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		entries, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}
	})

	t.Run("handles empty allow list", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": []
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		entries, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("handles missing permissions key", func(t *testing.T) {
		settingsJSON := `{
  "someOtherKey": true
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		entries, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		path := writeTestFile(t, t.TempDir(), "settings.json", "not json")

		_, err := LoadAllowlist(path)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := LoadAllowlist("/nonexistent/settings.json")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
	})
}

func TestLoadSandboxDomains(t *testing.T) {
	t.Run("loads sandbox domains from settings", func(t *testing.T) {
		settingsJSON := `{
  "sandbox": {
    "network": {
      "allowedDomains": ["*.github.com", "go.dev", "registry.terraform.io"]
    }
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		domains, err := LoadSandboxDomains(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(domains) != 3 {
			t.Fatalf("expected 3 domains, got %d", len(domains))
		}
		if domains[0] != "*.github.com" {
			t.Errorf("expected first domain *.github.com, got %s", domains[0])
		}
		if domains[1] != "go.dev" {
			t.Errorf("expected second domain go.dev, got %s", domains[1])
		}
		if domains[2] != "registry.terraform.io" {
			t.Errorf("expected third domain registry.terraform.io, got %s", domains[2])
		}
	})

	t.Run("handles missing sandbox key gracefully", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": ["WebFetch(domain:github.com)"]
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		domains, err := LoadSandboxDomains(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(domains) != 0 {
			t.Fatalf("expected 0 domains, got %d", len(domains))
		}
	})

	t.Run("handles empty allowedDomains array", func(t *testing.T) {
		settingsJSON := `{
  "sandbox": {
    "network": {
      "allowedDomains": []
    }
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		domains, err := LoadSandboxDomains(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(domains) != 0 {
			t.Fatalf("expected 0 domains, got %d", len(domains))
		}
	})

	t.Run("loads both permissions and sandbox domains", func(t *testing.T) {
		settingsJSON := `{
  "permissions": {
    "allow": [
      "WebFetch(domain:github.com)",
      "Bash(git status)"
    ]
  },
  "sandbox": {
    "network": {
      "allowedDomains": ["*.github.com", "go.dev"]
    }
  }
}`
		path := writeTestFile(t, t.TempDir(), "settings.json", settingsJSON)

		allowlist, err := LoadAllowlist(path)
		if err != nil {
			t.Fatalf("unexpected error loading allowlist: %v", err)
		}
		if len(allowlist) != 1 {
			t.Errorf("expected 1 allowlist entry, got %d", len(allowlist))
		}

		domains, err := LoadSandboxDomains(path)
		if err != nil {
			t.Fatalf("unexpected error loading sandbox: %v", err)
		}
		if len(domains) != 2 {
			t.Errorf("expected 2 sandbox domains, got %d", len(domains))
		}
	})
}
