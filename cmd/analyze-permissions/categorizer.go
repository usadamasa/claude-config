package main

import (
	"strings"
)

// Category はパーミッションパターンの安全性分類を表す｡
type Category string

const (
	CategorySafe   Category = "safe"
	CategoryAsk    Category = "ask"
	CategoryDeny   Category = "deny"
	CategoryReview Category = "review"
)

// CategoryResult はパターンの分類結果を保持する｡
type CategoryResult struct {
	Category Category `json:"category"`
	Reason   string   `json:"reason"`
}

// bashSafePatterns は安全な Bash コマンドのプレフィックスパターン｡
var bashSafePatterns = []struct {
	match  func(pattern string) bool
	reason string
}{
	{func(p string) bool { return matchBashPrefix(p, "git status") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git log") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git diff") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git branch") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git fetch") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git ls-tree") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git rev-parse") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git rev-list") }, "git 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "git add") }, "git ステージング"},
	{func(p string) bool { return matchBashPrefix(p, "git mv") }, "git ファイル操作"},
	{func(p string) bool { return matchBashPrefix(p, "git rm") }, "git ファイル操作"},
	{func(p string) bool { return matchBashPrefix(p, "git checkout") }, "git ブランチ操作"},
	{func(p string) bool { return matchBashPrefix(p, "git pull") }, "git 取得系"},
	{func(p string) bool { return strings.HasPrefix(p, "go ") }, "Go ツールチェイン"},
	{func(p string) bool { return strings.HasPrefix(p, "task ") }, "タスクランナー"},
	{func(p string) bool { return strings.HasPrefix(p, "make ") }, "ビルドツール"},
	{func(p string) bool { return matchBashPrefix(p, "gh pr") }, "GitHub CLI 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "gh run") }, "GitHub CLI 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "gh repo") }, "GitHub CLI 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "gh api") }, "GitHub API"},
	{func(p string) bool { return matchBashPrefix(p, "gh issues") }, "GitHub CLI 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "brew list") }, "Homebrew 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "brew info") }, "Homebrew 読取系"},
	{func(p string) bool { return matchBashPrefix(p, "brew install") }, "Homebrew インストール"},
	{func(p string) bool { return p == "ls" }, "ファイル一覧"},
	{func(p string) bool { return strings.HasPrefix(p, "golangci-lint") }, "リンター"},
	{func(p string) bool { return strings.HasPrefix(p, "docker ") }, "Docker"},
	{func(p string) bool { return strings.HasPrefix(p, "cargo ") }, "Cargo"},
}

// bashAskPatterns は確認が必要な Bash コマンドのプレフィックスパターン｡
var bashAskPatterns = []struct {
	match  func(pattern string) bool
	reason string
}{
	{func(p string) bool { return matchBashPrefix(p, "git commit") }, "git 変更操作"},
	{func(p string) bool { return matchBashPrefix(p, "git push") }, "git リモート操作"},
	{func(p string) bool { return matchBashPrefix(p, "git rebase") }, "git 履歴変更"},
	{func(p string) bool { return matchBashPrefix(p, "git reset") }, "git 履歴変更"},
	{func(p string) bool { return strings.HasPrefix(p, "rm -rf") }, "再帰的削除"},
	{func(p string) bool { return strings.HasPrefix(p, "rm -r") }, "再帰的削除"},
}

// bashDenyPatterns は拒否すべき Bash コマンドのプレフィックスパターン｡
var bashDenyPatterns = []struct {
	match  func(pattern string) bool
	reason string
}{
	{func(p string) bool { return p == "curl" || strings.HasPrefix(p, "curl ") }, "外部通信"},
	{func(p string) bool { return p == "wget" || strings.HasPrefix(p, "wget ") }, "外部通信"},
	{func(p string) bool { return p == "sudo" || strings.HasPrefix(p, "sudo ") }, "特権昇格"},
	{func(p string) bool { return p == "ssh" || strings.HasPrefix(p, "ssh ") }, "リモートアクセス"},
	{func(p string) bool { return p == "scp" || strings.HasPrefix(p, "scp ") }, "リモートコピー"},
	{func(p string) bool { return p == "eval" || strings.HasPrefix(p, "eval ") }, "任意コード実行"},
	{func(p string) bool { return matchBashPrefix(p, "gh auth") }, "認証操作"},
}

// fileSafePatterns は安全なファイルパスパターン(Read/Write/Edit 共通)｡
var fileSafePatterns = []struct {
	match  func(pattern string) bool
	reason string
}{
	{func(p string) bool { return p == "CLAUDE.md" }, "Claude 設定ファイル"},
	{func(p string) bool { return strings.HasPrefix(p, ".claude/") }, "Claude 設定ディレクトリ"},
	{func(p string) bool { return strings.HasPrefix(p, "~/.claude/") }, "Claude 設定ディレクトリ"},
	{func(p string) bool { return strings.HasPrefix(p, "src/") }, "ソースコード"},
	{func(p string) bool { return strings.HasPrefix(p, "docs/") }, "ドキュメント"},
	{func(p string) bool { return strings.HasPrefix(p, "cmd/") }, "コマンドソース"},
	{func(p string) bool { return strings.HasPrefix(p, "config/") }, "設定ファイル"},
	{func(p string) bool { return strings.HasPrefix(p, "test/") || strings.HasPrefix(p, "tests/") }, "テストファイル"},
	{func(p string) bool { return strings.HasPrefix(p, "classes/") }, "クラスファイル"},
	{func(p string) bool { return p == ".env.sample" }, "サンプル環境ファイル"},
}

// fileDenyPatterns は拒否すべきファイルパスパターン(Read/Write/Edit 共通)｡
var fileDenyPatterns = []struct {
	match  func(pattern string) bool
	reason string
}{
	{func(p string) bool { return strings.HasPrefix(p, "~/.ssh/") }, "SSH 鍵"},
	{func(p string) bool { return strings.HasPrefix(p, "~/.aws/") }, "AWS 認証情報"},
	{func(p string) bool { return strings.HasPrefix(p, "~/.gnupg/") }, "GPG 鍵"},
	{func(p string) bool { return strings.HasPrefix(p, "~/.kube/") }, "Kubernetes 設定"},
	{func(p string) bool {
		return p == ".env" || p == ".env.local" || p == ".env.development" || p == ".env.production"
	}, "環境変数ファイル"},
	{func(p string) bool {
		base := p
		if idx := strings.LastIndex(p, "/"); idx >= 0 {
			base = p[idx+1:]
		}
		return base == "credentials" || base == "credentials.json"
	}, "認証情報ファイル"},
	{func(p string) bool { return p == "~/.docker/config.json" }, "Docker 認証設定"},
	{func(p string) bool {
		return p == "~/.zsh_history" || p == "~/.bash_history"
	}, "シェル履歴"},
	{func(p string) bool { return p == "~/.netrc" }, "ネットワーク認証情報"},
	{func(p string) bool {
		base := p
		if idx := strings.LastIndex(p, "/"); idx >= 0 {
			base = p[idx+1:]
		}
		return base == "id_rsa" || base == "id_ed25519"
	}, "秘密鍵ファイル"},
}

// CategorizePermission はツール名とパターンから安全性カテゴリを判定する｡
func CategorizePermission(toolName, pattern string) CategoryResult {
	switch toolName {
	case "Bash":
		return categorizeBash(pattern)
	case "Read", "Write", "Edit":
		return categorizeFile(pattern)
	default:
		return CategoryResult{Category: CategoryReview, Reason: "未知のツール"}
	}
}

func categorizeBash(pattern string) CategoryResult {
	// deny を最初にチェック
	for _, p := range bashDenyPatterns {
		if p.match(pattern) {
			return CategoryResult{Category: CategoryDeny, Reason: p.reason}
		}
	}
	// 次に ask
	for _, p := range bashAskPatterns {
		if p.match(pattern) {
			return CategoryResult{Category: CategoryAsk, Reason: p.reason}
		}
	}
	// safe
	for _, p := range bashSafePatterns {
		if p.match(pattern) {
			return CategoryResult{Category: CategorySafe, Reason: p.reason}
		}
	}
	return CategoryResult{Category: CategoryReview, Reason: "手動確認が必要"}
}

func categorizeFile(pattern string) CategoryResult {
	// deny を最初にチェック
	for _, p := range fileDenyPatterns {
		if p.match(pattern) {
			return CategoryResult{Category: CategoryDeny, Reason: p.reason}
		}
	}
	// safe
	for _, p := range fileSafePatterns {
		if p.match(pattern) {
			return CategoryResult{Category: CategorySafe, Reason: p.reason}
		}
	}
	return CategoryResult{Category: CategoryReview, Reason: "手動確認が必要"}
}

// matchBashPrefix はコマンドプレフィックスの完全一致または前方一致を判定する｡
func matchBashPrefix(pattern, prefix string) bool {
	return pattern == prefix || strings.HasPrefix(pattern, prefix+" ")
}
