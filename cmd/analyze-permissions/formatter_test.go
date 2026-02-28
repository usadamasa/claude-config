package main

import (
	"strings"
	"testing"
)

func TestFormatSummary(t *testing.T) {
	t.Run("全セクションを含むレポート", func(t *testing.T) {
		report := Report{
			Metadata: ReportMetadata{
				AnalysisDate:   "2026-02-28",
				DaysAnalyzed:   30,
				FilesScanned:   42,
				TotalToolCalls: 1523,
			},
			CurrentAllow: make([]string, 109),
			CurrentDeny:  make([]string, 26),
			CurrentAsk:   make([]string, 8),
			Recommendations: Recommendations{
				Add: []PatternRecommendation{
					{ToolName: "Bash", Pattern: "go vet", Count: 12, Category: CategorySafe, Reason: "Go ツールチェイン"},
					{ToolName: "Bash", Pattern: "task lint", Count: 8, Category: CategorySafe, Reason: "タスクランナー"},
				},
				Review: []PatternRecommendation{
					{ToolName: "Bash", Pattern: "cat", Count: 3, Category: CategoryReview, Reason: "Read deny バイパスリスク"},
				},
				Unused: []UnusedEntry{
					{Entry: "Bash(brew upgrade:*)", List: "allow", Note: "過去30日間使用なし"},
				},
				BareEntryWarnings: []string{"Bash"},
				DenyBypassWarnings: []DenyBypassWarning{
					{AllowEntry: "Bash(cat:*)", BypassedDeny: "Read(~/.ssh/**)", Risk: "Read deny バイパス"},
				},
			},
		}

		output := FormatSummary(report, "/tmp/report.json")

		// ヘッダ
		if !strings.Contains(output, "=== Permission Optimizer Report ===") {
			t.Error("ヘッダが含まれていない")
		}
		if !strings.Contains(output, "Period: 30 days | Files: 42 | Tool Calls: 1523") {
			t.Error("メタデータが含まれていない")
		}

		// 追加推奨
		if !strings.Contains(output, "[ADD to allow] 2 patterns (showing 2/2):") {
			t.Error("追加推奨セクションが含まれていない")
		}
		if !strings.Contains(output, "Bash(go vet:*)") {
			t.Error("go vet 推奨が含まれていない")
		}

		// 要確認
		if !strings.Contains(output, "[REVIEW] 1 patterns (showing 1/1):") {
			t.Error("要確認セクションが含まれていない")
		}
		if !strings.Contains(output, "Bash(cat:*)") {
			t.Error("cat 要確認が含まれていない")
		}

		// 未使用
		if !strings.Contains(output, "[UNUSED] 1 entries (showing 1/1):") {
			t.Error("未使用セクションが含まれていない")
		}

		// 警告
		if !strings.Contains(output, "[WARNINGS]") {
			t.Error("警告セクションが含まれていない")
		}
		if !strings.Contains(output, "Bare entries: Bash") {
			t.Error("ベアエントリ警告が含まれていない")
		}
		if !strings.Contains(output, "Deny bypass: Bash(cat:*) -> Read(~/.ssh/**)") {
			t.Error("deny バイパス警告が含まれていない")
		}
		// サマリ行
		if !strings.Contains(output, "Summary: 109 allow / 26 deny / 8 ask") {
			t.Error("サマリ行が含まれていない")
		}
		if !strings.Contains(output, "+2 add, 1 unused, 1 review") {
			t.Error("変更サマリが含まれていない")
		}

		// JSON パス
		if !strings.Contains(output, "Full JSON: /tmp/report.json") {
			t.Error("JSON パスが含まれていない")
		}
	})

	t.Run("allow-encompasses-deny 警告", func(t *testing.T) {
		report := Report{
			Recommendations: Recommendations{
				AllowEncompassesDeny: []AllowEncompassesDeny{
					{AllowEntry: "Bash(gh:*)", DenyEntry: "Bash(gh auth:*)", Note: "allow が deny を包含"},
				},
			},
		}

		output := FormatSummary(report, "")

		if !strings.Contains(output, "[WARNINGS]") {
			t.Error("警告セクションが含まれていない")
		}
		if !strings.Contains(output, "Allow encompasses deny: Bash(gh:*) covers Bash(gh auth:*)") {
			t.Error("allow-encompasses-deny 警告が含まれていない")
		}
	})

	t.Run("空のレポート", func(t *testing.T) {
		report := Report{
			Metadata: ReportMetadata{
				DaysAnalyzed: 30,
			},
		}

		output := FormatSummary(report, "")

		if !strings.Contains(output, "=== Permission Optimizer Report ===") {
			t.Error("ヘッダが含まれていない")
		}
		if strings.Contains(output, "[ADD") {
			t.Error("空なのに追加推奨セクションが含まれている")
		}
		if strings.Contains(output, "[REVIEW]") {
			t.Error("空なのに要確認セクションが含まれている")
		}
		if strings.Contains(output, "[UNUSED]") {
			t.Error("空なのに未使用セクションが含まれている")
		}
		if strings.Contains(output, "[WARNINGS]") {
			t.Error("空なのに警告セクションが含まれている")
		}
		if strings.Contains(output, "Full JSON:") {
			t.Error("空なのに JSON パスが含まれている")
		}
	})

	t.Run("10件超で省略表示", func(t *testing.T) {
		recs := make([]PatternRecommendation, 15)
		for i := range recs {
			recs[i] = PatternRecommendation{
				ToolName: "Bash", Pattern: "cmd" + string(rune('a'+i)),
				Count: 100 - i, Category: CategorySafe, Reason: "テスト",
			}
		}
		report := Report{
			Recommendations: Recommendations{Add: recs},
		}

		output := FormatSummary(report, "")

		if !strings.Contains(output, "[ADD to allow] 15 patterns (showing 10/15):") {
			t.Error("件数が正しくない")
		}
		if !strings.Contains(output, "... and 5 more (use --format json for full list)") {
			t.Error("省略表示が含まれていない")
		}
	})
}

func TestFormatPermission(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		pattern  string
		want     string
	}{
		{"Bash コマンド", "Bash", "git status", "Bash(git status:*)"},
		{"Read パス", "Read", "~/.ssh/**", "Read(~/.ssh/**)"},
		{"Write パス", "Write", "src/**", "Write(src/**)"},
		{"Edit パス", "Edit", "CLAUDE.md", "Edit(CLAUDE.md)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPermission(tt.toolName, tt.pattern)
			if got != tt.want {
				t.Errorf("formatPermission(%q, %q) = %q, want %q", tt.toolName, tt.pattern, got, tt.want)
			}
		})
	}
}
