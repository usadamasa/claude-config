package main

import (
	"path/filepath"
)

// resolveRealpath はパスを正規化する｡
// GNU realpath -m と同等: パスが存在しなくてもエラーにならず、
// 存在する部分までシンボリックリンクを解決し、残りを正規化する｡
// cmd/realpath/main.go からのコピー
func resolveRealpath(path string) (string, error) {
	// 絶対パスに変換
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = abs
	}

	// まず完全な解決を試みる
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}

	// 存在する部分まで遡って解決する
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

	// どの部分も解決できない場合は Clean のみ
	return filepath.Clean(path), nil
}
