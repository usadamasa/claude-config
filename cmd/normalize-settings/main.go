package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/usadamasa/claude-config/internal/pathutil"
)

func main() {
	settingsPath := flag.String("settings", "", "settings.json パス (デフォルト: 自動検出)")
	check := flag.Bool("check", false, "読み取り専用モード。差分があれば exit 1")
	pinnedModel := flag.String("pinned-model", "claude-opus-4-6", "model の期待値")
	stripFieldsCSV := flag.String("strip-fields", "effortLevel,teammateMode", "除去するランタイムフィールド (カンマ区切り)")
	flag.Parse()

	var stripFields []string
	if *stripFieldsCSV != "" {
		stripFields = strings.Split(*stripFieldsCSV, ",")
	}

	path, err := resolveSettingsPath(*settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "settings.json パスの解決に失敗: %v\n", err)
		os.Exit(1)
	}

	if *check {
		runCheck(path, *pinnedModel, stripFields)
	} else {
		runNormalize(path, *pinnedModel, stripFields)
	}
}

func resolveSettingsPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("カレントディレクトリの取得に失敗: %w", err)
	}

	return pathutil.ResolveSettingsPath(cwd, home)
}

func runCheck(path, pinnedModel string, stripFields []string) {
	data, err := os.ReadFile(path) // #nosec G304 -- CLIツール: パスはフラグ引数由来
	if err != nil {
		fmt.Fprintf(os.Stderr, "ファイル読み込みに失敗: %v\n", err)
		os.Exit(1)
	}

	normalized, warns, err := Normalize(data, pinnedModel, stripFields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "正規化に失敗: %v\n", err)
		os.Exit(1)
	}

	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "警告: %s (期待値: %s)\n", w, pinnedModel)
	}

	if string(data) != string(normalized) {
		fmt.Fprintf(os.Stderr, "settings.json が正規化されていません: %s\n", path)
		fmt.Fprintf(os.Stderr, "`task settings:normalize` を実行してください\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "settings.json は正規化済みです: %s\n", path)
}

func runNormalize(path, pinnedModel string, stripFields []string) {
	changed, warns, err := NormalizeFile(path, pinnedModel, stripFields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "正規化に失敗: %v\n", err)
		os.Exit(1)
	}

	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "警告: %s (期待値: %s)\n", w, pinnedModel)
	}

	if changed {
		fmt.Fprintf(os.Stderr, "settings.json を正規化しました: %s\n", path)
	} else {
		fmt.Fprintf(os.Stderr, "settings.json は既に正規化済みです: %s\n", path)
	}
}
