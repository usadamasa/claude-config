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

func TestResolveProjectsDir(t *testing.T) {
	t.Run("--projects-dir 指定時はそのパスを使う", func(t *testing.T) {
		got := resolveProjectsDir("/custom/projects", "/home/user")
		want := "/custom/projects"
		if got != want {
			t.Errorf("resolveProjectsDir: got %q, want %q", got, want)
		}
	})

	t.Run("未指定時は ~/.claude/projects を使う", func(t *testing.T) {
		got := resolveProjectsDir("", "/home/user")
		want := "/home/user/.claude/projects"
		if got != want {
			t.Errorf("resolveProjectsDir: got %q, want %q", got, want)
		}
	})
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name        string
		scanPattern string
		permPattern string
		want        bool
	}{
		{"完全一致", "git status", "git status", true},
		{"不一致", "git status", "go test", false},
		{"/** ワイルドカード", "~/.claude/skills/foo", "~/.claude/**", true},
		{"プレフィックスマッチ (スペース)", "gh pr", "gh", true},
		{"プレフィックスマッチ (スラッシュ)", "src/main.go", "src", true},
		{"プレフィックスが部分一致しない", "ghost", "gh", false},
		{"/** が別プレフィックスに誤マッチしない", "src2/foo", "src/**", false},
		{"/** が正しいプレフィックスにマッチ", "src/foo", "src/**", true},
		{"空パターン", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.scanPattern, tt.permPattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.scanPattern, tt.permPattern, got, tt.want)
			}
		})
	}
}

func TestGenerateReportDenyBypassWarnings(t *testing.T) {
	t.Run("Bash cat が Read deny をバイパスする警告", func(t *testing.T) {
		scanResults := []ScanResult{
			{ToolName: "Bash", Pattern: "cat", FilePath: "a.jsonl"},
		}
		allow := []string{"Bash(cat:*)"}
		deny := []string{"Read(~/.ssh/**)"}

		report := GenerateReport(scanResults, allow, deny, nil, 30, 1)

		if len(report.Recommendations.DenyBypassWarnings) == 0 {
			t.Error("DenyBypassWarnings が空")
		}
		found := false
		for _, w := range report.Recommendations.DenyBypassWarnings {
			if w.AllowEntry == "Bash(cat:*)" && w.BypassedDeny == "Read(~/.ssh/**)" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("cat → Read(~/.ssh/**) バイパス警告が見つからない: %+v", report.Recommendations.DenyBypassWarnings)
		}
	})

	t.Run("echo が Write deny をバイパスする警告", func(t *testing.T) {
		scanResults := []ScanResult{
			{ToolName: "Bash", Pattern: "echo", FilePath: "a.jsonl"},
		}
		allow := []string{"Bash(echo:*)"}
		deny := []string{"Write(.env)"}

		report := GenerateReport(scanResults, allow, deny, nil, 30, 1)

		found := false
		for _, w := range report.Recommendations.DenyBypassWarnings {
			if w.AllowEntry == "Bash(echo:*)" {
				found = true
				break
			}
		}
		if !found {
			t.Error("echo → Write(.env) バイパス警告が見つからない")
		}
	})

	t.Run("安全なコマンドにはバイパス警告なし", func(t *testing.T) {
		scanResults := []ScanResult{
			{ToolName: "Bash", Pattern: "git status", FilePath: "a.jsonl"},
		}
		allow := []string{"Bash(git status:*)"}
		deny := []string{"Read(~/.ssh/**)"}

		report := GenerateReport(scanResults, allow, deny, nil, 30, 1)

		if len(report.Recommendations.DenyBypassWarnings) != 0 {
			t.Errorf("安全なコマンドに対してバイパス警告がある: %+v", report.Recommendations.DenyBypassWarnings)
		}
	})
}

func TestDetectAllowEncompassesDeny(t *testing.T) {
	t.Run("Bash(gh:*) が Bash(gh auth:*) を包含する", func(t *testing.T) {
		allow := []string{"Bash(gh:*)"}
		deny := []string{"Bash(gh auth:*)"}

		report := GenerateReport(nil, allow, deny, nil, 30, 0)

		if len(report.Recommendations.AllowEncompassesDeny) == 0 {
			t.Error("AllowEncompassesDeny が空")
		}
		found := false
		for _, w := range report.Recommendations.AllowEncompassesDeny {
			if w.AllowEntry == "Bash(gh:*)" && w.DenyEntry == "Bash(gh auth:*)" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Bash(gh:*) → Bash(gh auth:*) 包含警告が見つからない: %+v", report.Recommendations.AllowEncompassesDeny)
		}
	})

	t.Run("同一ツール同一パターンでは警告しない", func(t *testing.T) {
		allow := []string{"Bash(git status:*)"}
		deny := []string{"Bash(git status:*)"}

		report := GenerateReport(nil, allow, deny, nil, 30, 0)

		if len(report.Recommendations.AllowEncompassesDeny) != 0 {
			t.Errorf("同一パターンで警告が出ている: %+v", report.Recommendations.AllowEncompassesDeny)
		}
	})

	t.Run("異なるツールでは警告しない", func(t *testing.T) {
		allow := []string{"Bash(gh:*)"}
		deny := []string{"Read(~/.ssh/**)"}

		report := GenerateReport(nil, allow, deny, nil, 30, 0)

		if len(report.Recommendations.AllowEncompassesDeny) != 0 {
			t.Errorf("異なるツールで警告が出ている: %+v", report.Recommendations.AllowEncompassesDeny)
		}
	})

	t.Run("Read ワイルドカードで包含検出", func(t *testing.T) {
		allow := []string{"Read(~/src/**)"}
		deny := []string{"Read(~/src/secret/**)"}

		report := GenerateReport(nil, allow, deny, nil, 30, 0)

		if len(report.Recommendations.AllowEncompassesDeny) == 0 {
			t.Error("Read ワイルドカード包含が検出されない")
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
