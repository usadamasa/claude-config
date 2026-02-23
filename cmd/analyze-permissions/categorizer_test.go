package main

import (
	"testing"
)

func TestCategorizePermission(t *testing.T) {
	// Bash: safe カテゴリ
	bashSafeTests := []struct {
		name    string
		pattern string
	}{
		{"git status", "git status"},
		{"git log", "git log"},
		{"git diff", "git diff"},
		{"git branch", "git branch"},
		{"git fetch", "git fetch"},
		{"go test", "go test"},
		{"go build", "go build"},
		{"go mod", "go mod"},
		{"go vet", "go vet"},
		{"task build", "task build"},
		{"make build", "make build"},
		{"gh pr list", "gh pr"},
		{"gh run", "gh run"},
		{"brew list", "brew list"},
		{"brew info", "brew info"},
	}
	for _, tt := range bashSafeTests {
		t.Run("Bash/safe/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Bash", tt.pattern)
			if result.Category != CategorySafe {
				t.Errorf("CategorizePermission(Bash, %q) = %s, want safe", tt.pattern, result.Category)
			}
		})
	}

	// Bash: ask カテゴリ
	bashAskTests := []struct {
		name    string
		pattern string
	}{
		{"git commit", "git commit"},
		{"git push", "git push"},
		{"git rebase", "git rebase"},
		{"git reset", "git reset"},
		{"rm -rf", "rm -rf"},
	}
	for _, tt := range bashAskTests {
		t.Run("Bash/ask/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Bash", tt.pattern)
			if result.Category != CategoryAsk {
				t.Errorf("CategorizePermission(Bash, %q) = %s, want ask", tt.pattern, result.Category)
			}
		})
	}

	// Bash: deny カテゴリ
	bashDenyTests := []struct {
		name    string
		pattern string
	}{
		{"curl", "curl"},
		{"wget", "wget"},
		{"sudo", "sudo"},
		{"ssh", "ssh"},
		{"scp", "scp"},
		{"eval", "eval"},
	}
	for _, tt := range bashDenyTests {
		t.Run("Bash/deny/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Bash", tt.pattern)
			if result.Category != CategoryDeny {
				t.Errorf("CategorizePermission(Bash, %q) = %s, want deny", tt.pattern, result.Category)
			}
		})
	}

	// Read: safe カテゴリ
	readSafeTests := []struct {
		name    string
		pattern string
	}{
		{"プロジェクトファイル", "src/**"},
		{"CLAUDE.md", "CLAUDE.md"},
		{".claude 配下", ".claude/skills/**"},
	}
	for _, tt := range readSafeTests {
		t.Run("Read/safe/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Read", tt.pattern)
			if result.Category != CategorySafe {
				t.Errorf("CategorizePermission(Read, %q) = %s, want safe", tt.pattern, result.Category)
			}
		})
	}

	// Read: deny カテゴリ
	readDenyTests := []struct {
		name    string
		pattern string
	}{
		{".ssh", "~/.ssh/**"},
		{".aws", "~/.aws/**"},
		{".env", ".env"},
		{"credentials", "credentials"},
		{"history", "~/.zsh_history"},
		{"bash_history", "~/.bash_history"},
	}
	for _, tt := range readDenyTests {
		t.Run("Read/deny/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Read", tt.pattern)
			if result.Category != CategoryDeny {
				t.Errorf("CategorizePermission(Read, %q) = %s, want deny", tt.pattern, result.Category)
			}
		})
	}

	// Write: safe カテゴリ
	writeSafeTests := []struct {
		name    string
		pattern string
	}{
		{"src 配下", "src/**"},
		{"docs 配下", "docs/**"},
		{".claude/skills 配下", ".claude/skills/**"},
	}
	for _, tt := range writeSafeTests {
		t.Run("Write/safe/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Write", tt.pattern)
			if result.Category != CategorySafe {
				t.Errorf("CategorizePermission(Write, %q) = %s, want safe", tt.pattern, result.Category)
			}
		})
	}

	// Write: deny カテゴリ
	writeDenyTests := []struct {
		name    string
		pattern string
	}{
		{".env", ".env"},
		{"credentials", "credentials"},
	}
	for _, tt := range writeDenyTests {
		t.Run("Write/deny/"+tt.name, func(t *testing.T) {
			result := CategorizePermission("Write", tt.pattern)
			if result.Category != CategoryDeny {
				t.Errorf("CategorizePermission(Write, %q) = %s, want deny", tt.pattern, result.Category)
			}
		})
	}

	// Edit: Write と同様の分類
	t.Run("Edit/safe/src配下", func(t *testing.T) {
		result := CategorizePermission("Edit", "src/**")
		if result.Category != CategorySafe {
			t.Errorf("CategorizePermission(Edit, src/**) = %s, want safe", result.Category)
		}
	})

	t.Run("Edit/deny/.env", func(t *testing.T) {
		result := CategorizePermission("Edit", ".env")
		if result.Category != CategoryDeny {
			t.Errorf("CategorizePermission(Edit, .env) = %s, want deny", result.Category)
		}
	})

	// 未知のパターンは review
	t.Run("Bash/review/未知のコマンド", func(t *testing.T) {
		result := CategorizePermission("Bash", "unknown-cmd")
		if result.Category != CategoryReview {
			t.Errorf("CategorizePermission(Bash, unknown-cmd) = %s, want review", result.Category)
		}
	})

	t.Run("Read/review/未分類パス", func(t *testing.T) {
		result := CategorizePermission("Read", "~/.config/unknown/**")
		if result.Category != CategoryReview {
			t.Errorf("CategorizePermission(Read, ~/.config/unknown/**) = %s, want review", result.Category)
		}
	})
}
