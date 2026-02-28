package main

import (
	"os"
	"path/filepath"
	"testing"
)

// resolvedTempDir は t.TempDir() のパスを symlink 解決済みで返す｡
// macOS では /var → /private/var の symlink があるため必要｡
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
		got, err := resolveRealpath(f)
		if err != nil {
			t.Fatal(err)
		}
		if got != f {
			t.Errorf("got %q, want %q", got, f)
		}
	})

	t.Run("存在しないパスを正規化する", func(t *testing.T) {
		got, err := resolveRealpath("/nonexistent/path/to/file.txt")
		if err != nil {
			t.Fatal(err)
		}
		want := "/nonexistent/path/to/file.txt"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run(".. を含むパスを正規化する", func(t *testing.T) {
		got, err := resolveRealpath("/foo/bar/../baz")
		if err != nil {
			t.Fatal(err)
		}
		want := "/foo/baz"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run(". を含むパスを正規化する", func(t *testing.T) {
		got, err := resolveRealpath("/foo/./bar")
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
		got, err := resolveRealpath(filepath.Join(link, "file.txt"))
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
		got, err := resolveRealpath(filepath.Join(link, "subdir", "file.txt"))
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(real, "subdir", "file.txt")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("相対パスを絶対パスに変換する", func(t *testing.T) {
		got, err := resolveRealpath("relative/path")
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(got) {
			t.Errorf("got relative path %q, want absolute", got)
		}
	})
}
