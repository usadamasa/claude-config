package pathutil

import (
	"os"
	"path/filepath"
)

// ResolveRealpath はパスを正規化する｡
// GNU realpath -m と同等: パスが存在しなくてもエラーにならず、
// 存在する部分までシンボリックリンクを解決し、残りを正規化する｡
func ResolveRealpath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = abs
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}

	dir := path
	var tail []string
	for {
		parent := filepath.Dir(dir)
		tail = append([]string{filepath.Base(dir)}, tail...)
		dir = parent

		if dir == "/" || dir == "." {
			break
		}

		resolved, err := filepath.EvalSymlinks(dir)
		if err == nil {
			result := resolved
			for _, part := range tail {
				result = filepath.Join(result, part)
			}
			return filepath.Clean(result), nil
		}
	}

	return filepath.Clean(path), nil
}

// ResolveProjectsDir はプロジェクトディレクトリのパスを解決する｡
// フラグが指定されていればそれを使い、なければ home/.claude/projects を返す｡
func ResolveProjectsDir(flag, home string) string {
	if flag != "" {
		return flag
	}
	return filepath.Join(home, ".claude", "projects")
}

// ResolveSettingsPath は settings.json のパスを自動検出する｡
// cwd から親方向に git ルートを探索し､そこに settings.json があればそのパスを返す｡
// 見つからなければ home/.claude/settings.json をデフォルトとして返す｡
func ResolveSettingsPath(cwd, home string) (string, error) {
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Lstat(gitPath); err == nil {
			settingsPath := filepath.Join(dir, "dotclaude", "settings.json")
			if _, err := os.Stat(settingsPath); err == nil {
				return settingsPath, nil
			}
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return filepath.Join(home, ".claude", "settings.json"), nil
}
