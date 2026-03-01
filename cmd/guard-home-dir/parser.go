package main

import (
	"path/filepath"
	"strings"
)

// extractScanTargets はコマンド文字列からファイルシステムスキャン対象パスを抽出する｡
// ガード対象コマンド: find, du, tree, ls -R
// 戻り値: チェック対象パスのスライス (スキャンコマンドなしなら nil → 通過)
func extractScanTargets(command string, home string) []string {
	if command == "" {
		return nil
	}

	segments := splitCommands(command)
	var targets []string

	for _, seg := range segments {
		tokens := tokenize(strings.TrimSpace(seg))
		if len(tokens) == 0 {
			continue
		}

		// 先頭トークンからコマンド名を取得 (パス付きの場合はベース名)
		cmd := filepath.Base(tokens[0])

		switch cmd {
		case "find":
			paths := parseFindPaths(tokens, home)
			targets = append(targets, paths...)
		case "du", "tree":
			paths := parseGenericPaths(tokens, home)
			targets = append(targets, paths...)
		case "ls":
			if hasRecursiveFlag(tokens) {
				paths := parseGenericPaths(tokens, home)
				targets = append(targets, paths...)
			}
		}
	}

	if len(targets) == 0 {
		return nil
	}
	return targets
}

// splitCommands は &&, ||, ;, | で分割して個別コマンドセグメントを返す｡
// クォート内の区切り文字は無視する｡
func splitCommands(command string) []string {
	var segments []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		// クォート状態の追跡
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			current.WriteByte(ch)
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			current.WriteByte(ch)
			continue
		}

		// クォート内なら区切り文字を無視
		if inSingle || inDouble {
			current.WriteByte(ch)
			continue
		}

		// && チェック
		if ch == '&' && i+1 < len(command) && command[i+1] == '&' {
			segments = append(segments, current.String())
			current.Reset()
			i++ // 2文字目をスキップ
			continue
		}

		// || チェック
		if ch == '|' && i+1 < len(command) && command[i+1] == '|' {
			segments = append(segments, current.String())
			current.Reset()
			i++ // 2文字目をスキップ
			continue
		}

		// 単一 | (パイプ)
		if ch == '|' {
			segments = append(segments, current.String())
			current.Reset()
			continue
		}

		// ; (セミコロン)
		if ch == ';' {
			segments = append(segments, current.String())
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	// 最後のセグメント
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

// tokenize はクォートを考慮してトークン分割する｡
// クォート文字自体は除去される｡
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if ch == ' ' && !inSingle && !inDouble {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// expandHome は ~ と $HOME をホームディレクトリに展開する｡
func expandHome(token string, home string) string {
	if token == "~" {
		return home
	}
	if strings.HasPrefix(token, "~/") {
		return home + token[1:]
	}
	if token == "$HOME" {
		return home
	}
	if strings.HasPrefix(token, "$HOME/") {
		return home + token[5:]
	}
	return token
}

// parseFindPaths は find コマンドからパス引数を抽出する｡
// find [path...] [expression] の path 部分 (最初の - や ( や ! の前まで)
func parseFindPaths(tokens []string, home string) []string {
	if len(tokens) < 2 {
		return nil
	}

	var paths []string
	// tokens[0] は "find" なので tokens[1] から開始
	for _, tok := range tokens[1:] {
		// リダイレクト (2>/dev/null 等) はスキップ
		if isRedirect(tok) {
			continue
		}
		// - で始まるオプション、(、!、\( はパスの終わり
		if strings.HasPrefix(tok, "-") || tok == "(" || tok == "!" || tok == "\\(" {
			break
		}
		paths = append(paths, expandHome(tok, home))
	}

	if len(paths) == 0 {
		return nil
	}
	return paths
}

// parseGenericPaths は du/tree/ls のようなコマンドからパス引数を抽出する｡
// オプション (- で始まるトークン) 以外の引数をパスとして扱う｡
func parseGenericPaths(tokens []string, home string) []string {
	if len(tokens) < 2 {
		return nil
	}

	var paths []string
	// tokens[0] はコマンド名なので tokens[1] から開始
	for _, tok := range tokens[1:] {
		// リダイレクト はスキップ
		if isRedirect(tok) {
			continue
		}
		// - で始まるオプションはスキップ
		if strings.HasPrefix(tok, "-") {
			continue
		}
		paths = append(paths, expandHome(tok, home))
	}

	if len(paths) == 0 {
		return nil
	}
	return paths
}

// hasRecursiveFlag は ls コマンドに -R または --recursive フラグがあるか確認する｡
func hasRecursiveFlag(tokens []string) bool {
	for _, tok := range tokens[1:] {
		if tok == "--recursive" {
			return true
		}
		// -R を含む短縮オプション (例: -laR, -Rl)
		if strings.HasPrefix(tok, "-") && !strings.HasPrefix(tok, "--") && strings.Contains(tok, "R") {
			return true
		}
	}
	return false
}

// isRedirect はトークンがリダイレクト (例: 2>/dev/null, >/dev/null) かどうか判定する｡
func isRedirect(tok string) bool {
	return strings.Contains(tok, ">/") || strings.Contains(tok, ">&")
}
