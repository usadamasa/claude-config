package main

import (
	"encoding/json"
	"testing"
)

func TestGenerateReport(t *testing.T) {
	t.Run("generates report with recommendations", func(t *testing.T) {
		scanResults := []ScanResult{
			{URL: "https://github.com/foo", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
			{URL: "https://github.com/bar", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
			{URL: "https://docs.new-lib.io/api", Domain: "docs.new-lib.io", Tool: "WebFetch", FilePath: "a.jsonl"},
			{URL: "https://stackoverflow.com/q/123", Domain: "stackoverflow.com", Tool: "WebFetch", FilePath: "b.jsonl"},
			{URL: "https://random-site.xyz/page", Domain: "random-site.xyz", Tool: "WebFetch", FilePath: "b.jsonl"},
		}
		allowlist := []AllowlistEntry{
			{Tool: "WebFetch", Domain: "github.com"},
			{Tool: "WebFetch", Domain: "docs.unused.com"},
		}

		report := GenerateReport(scanResults, allowlist, nil, 30, 2)

		// Verify metadata
		if report.Metadata.DaysAnalyzed != 30 {
			t.Errorf("expected days_analyzed=30, got %d", report.Metadata.DaysAnalyzed)
		}
		if report.Metadata.FilesScanned != 2 {
			t.Errorf("expected files_scanned=2, got %d", report.Metadata.FilesScanned)
		}
		if report.Metadata.WebFetchCalls != 5 {
			t.Errorf("expected webfetch_calls=5, got %d", report.Metadata.WebFetchCalls)
		}

		// Verify current allowlist
		if len(report.CurrentAllowlist) != 2 {
			t.Fatalf("expected 2 allowlist entries, got %d", len(report.CurrentAllowlist))
		}

		// Verify add recommendations (safe domains not in allowlist)
		foundNewLib := false
		for _, rec := range report.Recommendations.Add {
			if rec.Domain == "docs.new-lib.io" {
				foundNewLib = true
				if rec.Category != CategorySafe {
					t.Errorf("expected docs.new-lib.io to be safe, got %s", rec.Category)
				}
			}
		}
		if !foundNewLib {
			t.Error("expected docs.new-lib.io in add recommendations")
		}

		// Verify review recommendations (medium/review domains not in allowlist)
		foundSO := false
		for _, rec := range report.Recommendations.Review {
			if rec.Domain == "stackoverflow.com" {
				foundSO = true
			}
		}
		if !foundSO {
			t.Error("expected stackoverflow.com in review recommendations")
		}

		// Verify unused domains
		foundUnused := false
		for _, rec := range report.Recommendations.Unused {
			if rec.Domain == "docs.unused.com" {
				foundUnused = true
			}
		}
		if !foundUnused {
			t.Error("expected docs.unused.com in unused recommendations")
		}

		// Verify all_domains list
		if len(report.AllDomains) != 4 {
			t.Errorf("expected 4 domains in all_domains, got %d", len(report.AllDomains))
		}
	})

	t.Run("report is valid JSON", func(t *testing.T) {
		report := GenerateReport(nil, nil, nil, 30, 0)

		data, err := json.Marshal(report)
		if err != nil {
			t.Fatalf("report is not valid JSON: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty JSON output")
		}
	})

	t.Run("includes sandbox domains in report", func(t *testing.T) {
		scanResults := []ScanResult{
			{URL: "https://github.com/foo", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
		}
		allowlist := []AllowlistEntry{
			{Tool: "WebFetch", Domain: "github.com"},
		}
		sandboxDomains := []string{"*.github.com", "go.dev"}

		report := GenerateReport(scanResults, allowlist, sandboxDomains, 30, 1)

		if len(report.CurrentSandbox) != 2 {
			t.Fatalf("expected 2 sandbox domains, got %d", len(report.CurrentSandbox))
		}
		if report.CurrentSandbox[0] != "*.github.com" {
			t.Errorf("expected first sandbox domain *.github.com, got %s", report.CurrentSandbox[0])
		}
	})

	t.Run("recommends adding to sandbox when domain in permissions but not sandbox", func(t *testing.T) {
		scanResults := []ScanResult{
			{URL: "https://github.com/foo", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
		}
		allowlist := []AllowlistEntry{
			{Tool: "WebFetch", Domain: "github.com"},
			{Tool: "WebFetch", Domain: "docs.example.com"},
		}
		// sandbox に github.com はワイルドカードで含まれるが docs.example.com は含まれない
		sandboxDomains := []string{"*.github.com", "go.dev"}

		report := GenerateReport(scanResults, allowlist, sandboxDomains, 30, 1)

		foundDocsExample := false
		for _, rec := range report.Recommendations.AddToSandbox {
			if rec.Domain == "docs.example.com" {
				foundDocsExample = true
			}
		}
		if !foundDocsExample {
			t.Error("expected docs.example.com in add_to_sandbox recommendations")
		}
	})

	t.Run("counts WebFetch and Fetch calls separately", func(t *testing.T) {
		scanResults := []ScanResult{
			{URL: "https://github.com/foo", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
			{URL: "https://github.com/bar", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
			{URL: "https://api.example.com/v1", Domain: "api.example.com", Tool: "Fetch", FilePath: "a.jsonl"},
		}

		report := GenerateReport(scanResults, nil, nil, 30, 1)

		if report.Metadata.WebFetchCalls != 2 {
			t.Errorf("expected webfetch_calls=2, got %d", report.Metadata.WebFetchCalls)
		}
		if report.Metadata.FetchCalls != 1 {
			t.Errorf("expected fetch_calls=1, got %d", report.Metadata.FetchCalls)
		}
	})

	t.Run("works without sandbox config (backward compatible)", func(t *testing.T) {
		scanResults := []ScanResult{
			{URL: "https://github.com/foo", Domain: "github.com", Tool: "WebFetch", FilePath: "a.jsonl"},
		}
		allowlist := []AllowlistEntry{
			{Tool: "WebFetch", Domain: "github.com"},
		}

		report := GenerateReport(scanResults, allowlist, nil, 30, 1)

		if report.CurrentSandbox != nil {
			t.Errorf("expected nil CurrentSandbox, got %v", report.CurrentSandbox)
		}
		if report.Recommendations.AddToSandbox != nil {
			t.Errorf("expected nil AddToSandbox, got %v", report.Recommendations.AddToSandbox)
		}
	})
}
