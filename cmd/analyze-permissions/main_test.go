package main

import (
	"encoding/json"
	"testing"
)

func TestGenerateReport(t *testing.T) {
	t.Run("基本的なレポート生成", func(t *testing.T) {
		scanResults := []ScanResult{
			{ToolName: "Bash", Pattern: "git status", FilePath: "a.jsonl"},
			{ToolName: "Bash", Pattern: "git status", FilePath: "a.jsonl"},
			{ToolName: "Bash", Pattern: "go test", FilePath: "a.jsonl"},
			{ToolName: "Bash", Pattern: "curl", FilePath: "a.jsonl"},
			{ToolName: "Read", Pattern: "CLAUDE.md", FilePath: "b.jsonl"},
			{ToolName: "Read", Pattern: "~/.ssh/**", FilePath: "b.jsonl"},
			{ToolName: "Write", Pattern: "src/**", FilePath: "b.jsonl"},
		}

		allow := []string{
			"Bash(git status:*)",
			"Read(CLAUDE.md)",
		}
		deny := []string{
			"Bash(curl:*)",
			"Read(~/.ssh/**)",
		}
		ask := []string{
			"Bash(git commit:*)",
		}

		report := GenerateReport(scanResults, allow, deny, ask, 30, 2)

		// メタデータ検証
		if report.Metadata.DaysAnalyzed != 30 {
			t.Errorf("DaysAnalyzed: got %d, want 30", report.Metadata.DaysAnalyzed)
		}
		if report.Metadata.FilesScanned != 2 {
			t.Errorf("FilesScanned: got %d, want 2", report.Metadata.FilesScanned)
		}
		if report.Metadata.TotalToolCalls != 7 {
			t.Errorf("TotalToolCalls: got %d, want 7", report.Metadata.TotalToolCalls)
		}

		// 推奨事項検証
		if len(report.Recommendations.Add) == 0 {
			t.Error("追加推奨が空")
		}

		// go test は allow にないが safe なので追加推奨に含まれるべき
		found := false
		for _, rec := range report.Recommendations.Add {
			if rec.ToolName == "Bash" && rec.Pattern == "go test" {
				found = true
				break
			}
		}
		if !found {
			t.Error("go test が追加推奨に含まれていない")
		}

		// git commit は ask にあるが使用なし → unused に含まれるべき
		foundUnused := false
		for _, u := range report.Recommendations.Unused {
			if u.Entry == "Bash(git commit:*)" {
				foundUnused = true
				break
			}
		}
		if !foundUnused {
			t.Error("Bash(git commit:*) が未使用リストに含まれていない")
		}
	})

	t.Run("空のスキャン結果", func(t *testing.T) {
		report := GenerateReport(nil, nil, nil, nil, 30, 0)

		if report.Metadata.TotalToolCalls != 0 {
			t.Errorf("TotalToolCalls: got %d, want 0", report.Metadata.TotalToolCalls)
		}
		if len(report.AllPatterns) != 0 {
			t.Errorf("AllPatterns: got %d, want 0", len(report.AllPatterns))
		}
	})

	t.Run("JSON 出力が有効", func(t *testing.T) {
		report := GenerateReport(
			[]ScanResult{{ToolName: "Bash", Pattern: "git status", FilePath: "a.jsonl"}},
			[]string{"Bash(git status:*)"},
			nil, nil, 30, 1,
		)

		data, err := json.Marshal(report)
		if err != nil {
			t.Fatalf("JSON マーシャルに失敗: %v", err)
		}
		if len(data) == 0 {
			t.Error("JSON 出力が空")
		}
	})

	t.Run("ベアエントリ警告", func(t *testing.T) {
		allow := []string{"Bash"}
		report := GenerateReport(nil, allow, nil, nil, 30, 0)

		if len(report.Recommendations.BareEntryWarnings) == 0 {
			t.Error("ベアエントリ警告が含まれていない")
		}
		found := false
		for _, w := range report.Recommendations.BareEntryWarnings {
			if w == "Bash" {
				found = true
			}
		}
		if !found {
			t.Error("Bash ベアエントリ警告が見つからない")
		}
	})

	t.Run("ask 内のベアエントリ警告", func(t *testing.T) {
		ask := []string{"Read"}
		report := GenerateReport(nil, nil, nil, ask, 30, 0)

		found := false
		for _, w := range report.Recommendations.BareEntryWarnings {
			if w == "Read" {
				found = true
			}
		}
		if !found {
			t.Error("Read ベアエントリ警告が見つからない")
		}
	})
}

func TestCountUniqueFiles(t *testing.T) {
	results := []ScanResult{
		{FilePath: "a.jsonl"},
		{FilePath: "a.jsonl"},
		{FilePath: "b.jsonl"},
	}
	got := countUniqueFiles(results)
	if got != 2 {
		t.Errorf("countUniqueFiles: got %d, want 2", got)
	}
}
