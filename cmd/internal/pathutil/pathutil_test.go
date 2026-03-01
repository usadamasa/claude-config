package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func resolvedTempDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	resolved, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func TestResolveRealpath(t *testing.T) {
	t.Run("存在するファイルの絶対パスをそのまま返す", func(t *testing.T) {
		tmp := resolvedTempDir(t)
		f := filepath.Join(tmp, "file.txt")
		if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveRealpath(f)
		if err != nil {
			t.Fatal(err)
		}
		if got != f {
			t.Errorf("got %q, want %q", got, f)
		}
	})

	t.Run("存在しないパスを正規化する", func(t *testing.T) {
		got, err := ResolveRealpath("/nonexistent/path/to/file.txt")
		if err != nil {
			t.Fatal(err)
		}
		want := "/nonexistent/path/to/file.txt"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("..を含むパスを正規化する", func(t *testing.T) {
		got, err := ResolveRealpath("/foo/bar/../baz")
		if err != nil {
			t.Fatal(err)
		}
		want := "/foo/baz"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run(".を含むパスを正規化する", func(t *testing.T) {
		got, err := ResolveRealpath("/foo/./bar")
		if err != nil {
			t.Fatal(err)
		}
		want := "/foo/bar"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("シンボリックリンクを解決する", func(t *testing.T) {
		tmp := resolvedTempDir(t)
		real := filepath.Join(tmp, "real")
		if err := os.Mkdir(real, 0755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(tmp, "link")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveRealpath(filepath.Join(link, "file.txt"))
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(real, "file.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("シンボリックリンク配下の存在しないファイルを解決する", func(t *testing.T) {
		tmp := resolvedTempDir(t)
		real := filepath.Join(tmp, "real")
		if err := os.Mkdir(real, 0755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(tmp, "link")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveRealpath(filepath.Join(link, "subdir", "file.txt"))
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(real, "subdir", "file.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("相対パスを絶対パスに変換する", func(t *testing.T) {
		got, err := ResolveRealpath("relative/path")
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(got) {
			t.Errorf("got relative path %q, want absolute", got)
		}
	})
}

func TestResolveProjectsDir(t *testing.T) {
	t.Run("フラグ指定時はその値を返す", func(t *testing.T) {
		got := ResolveProjectsDir("/custom/dir", "/home/user")
		if got != "/custom/dir" {
			t.Errorf("got %q, want %q", got, "/custom/dir")
		}
	})

	t.Run("フラグ未指定時はデフォルトを返す", func(t *testing.T) {
		got := ResolveProjectsDir("", "/home/user")
		want := "/home/user/.claude/projects"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestResolveSettingsPath(t *testing.T) {
	t.Run("worktree環境ではgitルートのdotclaude/settings.jsonを返す", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitFile := filepath.Join(tmpDir, ".git")
		if err := os.WriteFile(gitFile, []byte("gitdir: /some/path/.git/worktrees/test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		dotclaudeDir := filepath.Join(tmpDir, "dotclaude")
		if err := os.MkdirAll(dotclaudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settingsFile := filepath.Join(dotclaudeDir, "settings.json")
		if err := os.WriteFile(settingsFile, []byte(`{"permissions":{}}`), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ResolveSettingsPath(tmpDir, "/dummy/home")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != settingsFile {
			t.Errorf("got %s, want %s", got, settingsFile)
		}
	})

	t.Run("通常リポジトリでもgitルートのdotclaude/settings.jsonを返す", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatal(err)
		}

		dotclaudeDir := filepath.Join(tmpDir, "dotclaude")
		if err := os.MkdirAll(dotclaudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settingsFile := filepath.Join(dotclaudeDir, "settings.json")
		if err := os.WriteFile(settingsFile, []byte(`{"permissions":{}}`), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ResolveSettingsPath(tmpDir, "/dummy/home")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != settingsFile {
			t.Errorf("got %s, want %s", got, settingsFile)
		}
	})

	t.Run("gitルートにdotclaude/settings.jsonがなければデフォルトを返す", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitDir := filepath.Join(tmpDir, ".git")
		if err := os.Mkdir(gitDir, 0755); err != nil {
			t.Fatal(err)
		}

		home := t.TempDir()
		want := filepath.Join(home, ".claude", "settings.json")

		got, err := ResolveSettingsPath(tmpDir, home)
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

		got, err := ResolveSettingsPath(tmpDir, home)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("サブディレクトリからも親のgitルートを検出する", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitFile := filepath.Join(tmpDir, ".git")
		if err := os.WriteFile(gitFile, []byte("gitdir: /some/path\n"), 0644); err != nil {
			t.Fatal(err)
		}

		dotclaudeDir := filepath.Join(tmpDir, "dotclaude")
		if err := os.MkdirAll(dotclaudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settingsFile := filepath.Join(dotclaudeDir, "settings.json")
		if err := os.WriteFile(settingsFile, []byte(`{"permissions":{}}`), 0644); err != nil {
			t.Fatal(err)
		}

		subDir := filepath.Join(tmpDir, "cmd", "analyze-permissions")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		got, err := ResolveSettingsPath(subDir, "/dummy/home")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != settingsFile {
			t.Errorf("got %s, want %s", got, settingsFile)
		}
	})
}
