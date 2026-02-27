package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSettingsPath(t *testing.T) {
	t.Run("worktree環境ではgitルートのdotclaude/settings.jsonを返す", func(t *testing.T) {
		tmpDir := t.TempDir()

		// worktree 環境をシミュレート: .git はファイル
		gitFile := filepath.Join(tmpDir, ".git")
		if err := os.WriteFile(gitFile, []byte("gitdir: /some/path/.git/worktrees/test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// dotclaude/settings.json を配置
		dotclaudeDir := filepath.Join(tmpDir, "dotclaude")
		if err := os.MkdirAll(dotclaudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settingsFile := filepath.Join(dotclaudeDir, "settings.json")
		if err := os.WriteFile(settingsFile, []byte(`{"permissions":{}}`), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := resolveSettingsPath(tmpDir, "/dummy/home")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != settingsFile {
			t.Errorf("got %s, want %s", got, settingsFile)
		}
	})

	t.Run("通常リポジトリでもgitルートのdotclaude/settings.jsonを返す", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 通常リポジトリ: .git はディレクトリ
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatal(err)
		}

		// dotclaude/settings.json を配置
		dotclaudeDir := filepath.Join(tmpDir, "dotclaude")
		if err := os.MkdirAll(dotclaudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settingsFile := filepath.Join(dotclaudeDir, "settings.json")
		if err := os.WriteFile(settingsFile, []byte(`{"permissions":{}}`), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := resolveSettingsPath(tmpDir, "/dummy/home")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != settingsFile {
			t.Errorf("got %s, want %s", got, settingsFile)
		}
	})

	t.Run("gitルートにdotclaude/settings.jsonがなければデフォルトを返す", func(t *testing.T) {
		tmpDir := t.TempDir()

		// .git ディレクトリのみ (dotclaude/settings.json なし)
		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatal(err)
		}

		home := t.TempDir()
		want := filepath.Join(home, ".claude", "settings.json")

		got, err := resolveSettingsPath(tmpDir, home)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("gitリポジトリ外ではデフォルトを返す", func(t *testing.T) {
		tmpDir := t.TempDir()
		home := t.TempDir()
		want := filepath.Join(home, ".claude", "settings.json")

		got, err := resolveSettingsPath(tmpDir, home)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("サブディレクトリからも親のgitルートを検出する", func(t *testing.T) {
		tmpDir := t.TempDir()

		// .git ファイル (worktree)
		gitFile := filepath.Join(tmpDir, ".git")
		if err := os.WriteFile(gitFile, []byte("gitdir: /some/path\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// dotclaude/settings.json を配置
		dotclaudeDir := filepath.Join(tmpDir, "dotclaude")
		if err := os.MkdirAll(dotclaudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settingsFile := filepath.Join(dotclaudeDir, "settings.json")
		if err := os.WriteFile(settingsFile, []byte(`{"permissions":{}}`), 0644); err != nil {
			t.Fatal(err)
		}

		// サブディレクトリ作成
		subDir := filepath.Join(tmpDir, "cmd", "analyze-permissions")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		got, err := resolveSettingsPath(subDir, "/dummy/home")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != settingsFile {
			t.Errorf("got %s, want %s", got, settingsFile)
		}
	})
}
