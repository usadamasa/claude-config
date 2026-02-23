package main

import (
	"fmt"
	"net/url"
	"strings"
)

// Category represents the safety classification of a domain.
type Category string

const (
	CategorySafe   Category = "safe"
	CategoryMedium Category = "medium"
	CategoryReview Category = "review"
)

// CategoryResult holds a domain's classification and the reason for it.
type CategoryResult struct {
	Category Category `json:"category"`
	Reason   string   `json:"reason"`
}

// safePatterns defines domain patterns classified as safe.
var safePatterns = []struct {
	match  func(domain string) bool
	reason string
}{
	{func(d string) bool { return strings.HasPrefix(d, "docs.") }, "公式ドキュメントサイト"},
	{func(d string) bool { return strings.HasPrefix(d, "developer.") }, "開発者向け公式サイト"},
	{func(d string) bool { return strings.HasSuffix(d, ".readthedocs.io") }, "ReadTheDocs ドキュメント"},
	{func(d string) bool { return d == "github.com" || d == "raw.githubusercontent.com" }, "コードホスティング"},
	{func(d string) bool { return d == "gitlab.com" }, "コードホスティング"},
	{func(d string) bool {
		return d == "pkg.go.dev" || d == "go.dev"
	}, "Go パッケージレジストリ"},
	{func(d string) bool { return d == "www.npmjs.com" || d == "npmjs.com" }, "npm パッケージレジストリ"},
	{func(d string) bool { return d == "registry.terraform.io" }, "Terraform レジストリ"},
	{func(d string) bool { return d == "cloud.google.com" }, "Google Cloud 公式"},
	{func(d string) bool { return strings.HasSuffix(d, ".databricks.com") }, "Databricks 公式"},
	{func(d string) bool { return strings.HasSuffix(d, ".anthropic.com") }, "Anthropic 公式"},
}

// mediumPatterns defines domain patterns classified as medium risk.
var mediumPatterns = []struct {
	match  func(domain string) bool
	reason string
}{
	{func(d string) bool { return d == "stackoverflow.com" || d == "www.stackoverflow.com" }, "Q&A コミュニティ"},
	{func(d string) bool { return d == "medium.com" }, "ブログプラットフォーム"},
	{func(d string) bool { return d == "dev.to" }, "開発者コミュニティ"},
	{func(d string) bool { return d == "learning.oreilly.com" }, "学習プラットフォーム"},
}

// CategorizeDomain classifies a domain into a safety category.
func CategorizeDomain(domain string) CategoryResult {
	for _, p := range safePatterns {
		if p.match(domain) {
			return CategoryResult{Category: CategorySafe, Reason: p.reason}
		}
	}
	for _, p := range mediumPatterns {
		if p.match(domain) {
			return CategoryResult{Category: CategoryMedium, Reason: p.reason}
		}
	}
	return CategoryResult{Category: CategoryReview, Reason: "手動確認が必要"}
}

// ExtractDomain extracts the hostname from a raw URL string.
func ExtractDomain(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("empty URL")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("no host in URL: %s", rawURL)
	}
	return host, nil
}
