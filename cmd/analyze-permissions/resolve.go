package main

import (
	"os"
	"path/filepath"
)

// resolveSettingsPath は settings.json のパスを自動検出する｡
// cwd から親方向に git ルートを探索し､そこに settings.json があればそのパスを返す｡
// 見つからなければ home/.claude/settings.json をデフォルトとして返す｡
func resolveSettingsPath(cwd, home string) (string, error) {
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Lstat(gitPath); err == nil {
			// git ルートを検出
			settingsPath := filepath.Join(dir, "dotclaude", "settings.json")
			if _, err := os.Stat(settingsPath); err == nil {
				return settingsPath, nil
			}
			// git ルートだが settings.json がない → デフォルトへ
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
