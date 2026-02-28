package main

import (
	"fmt"
	"strings"
)

// FormatSummary はレポートをコンパクトなテキストサマリに変換する｡
func FormatSummary(r Report, jsonOutputPath string) string {
	var b strings.Builder

	// ヘッダ
	fmt.Fprintf(&b, "=== Permission Optimizer Report ===\n")
	fmt.Fprintf(&b, "Period: %d days | Files: %d | Tool Calls: %d\n",
		r.Metadata.DaysAnalyzed, r.Metadata.FilesScanned, r.Metadata.TotalToolCalls)

	// 追加推奨
	if len(r.Recommendations.Add) > 0 {
		total := len(r.Recommendations.Add)
		limit := min(total, 10)
		fmt.Fprintf(&b, "\n[ADD to allow] %d patterns (showing %d/%d):\n", total, limit, total)
		for _, rec := range r.Recommendations.Add[:limit] {
			fmt.Fprintf(&b, "  %-30s %4d uses  %s\n",
				formatPermission(rec.ToolName, rec.Pattern), rec.Count, rec.Reason)
		}
		if total > limit {
			fmt.Fprintf(&b, "  ... and %d more (use --format json for full list)\n", total-limit)
		}
	}

	// 要確認
	if len(r.Recommendations.Review) > 0 {
		total := len(r.Recommendations.Review)
		limit := min(total, 10)
		fmt.Fprintf(&b, "\n[REVIEW] %d patterns (showing %d/%d):\n", total, limit, total)
		for _, rec := range r.Recommendations.Review[:limit] {
			fmt.Fprintf(&b, "  %-30s %4d uses  %s\n",
				formatPermission(rec.ToolName, rec.Pattern), rec.Count, rec.Reason)
		}
		if total > limit {
			fmt.Fprintf(&b, "  ... and %d more (use --format json for full list)\n", total-limit)
		}
	}

	// 未使用
	if len(r.Recommendations.Unused) > 0 {
		total := len(r.Recommendations.Unused)
		limit := min(total, 10)
		fmt.Fprintf(&b, "\n[UNUSED] %d entries (showing %d/%d):\n", total, limit, total)
		for _, u := range r.Recommendations.Unused[:limit] {
			fmt.Fprintf(&b, "  %s: %s  %s\n", u.List, u.Entry, u.Note)
		}
		if total > limit {
			fmt.Fprintf(&b, "  ... and %d more (use --format json for full list)\n", total-limit)
		}
	}

	// 警告
	hasWarnings := len(r.Recommendations.BareEntryWarnings) > 0 ||
		len(r.Recommendations.DenyBypassWarnings) > 0 ||
		len(r.Recommendations.AllowEncompassesDeny) > 0
	if hasWarnings {
		fmt.Fprintf(&b, "\n[WARNINGS]\n")
		if len(r.Recommendations.BareEntryWarnings) > 0 {
			fmt.Fprintf(&b, "  Bare entries: %s\n", strings.Join(r.Recommendations.BareEntryWarnings, ", "))
		}
		for _, w := range r.Recommendations.DenyBypassWarnings {
			fmt.Fprintf(&b, "  Deny bypass: %s -> %s\n", w.AllowEntry, w.BypassedDeny)
		}
		for _, w := range r.Recommendations.AllowEncompassesDeny {
			fmt.Fprintf(&b, "  Allow encompasses deny: %s covers %s\n", w.AllowEntry, w.DenyEntry)
		}
	}

	// サマリ行
	fmt.Fprintf(&b, "\nSummary: %d allow / %d deny / %d ask",
		len(r.CurrentAllow), len(r.CurrentDeny), len(r.CurrentAsk))

	addCount := len(r.Recommendations.Add)
	unusedCount := len(r.Recommendations.Unused)
	reviewCount := len(r.Recommendations.Review)
	if addCount > 0 || unusedCount > 0 || reviewCount > 0 {
		fmt.Fprintf(&b, " | +%d add, %d unused, %d review", addCount, unusedCount, reviewCount)
	}
	fmt.Fprintln(&b)

	if jsonOutputPath != "" {
		fmt.Fprintf(&b, "Full JSON: %s\n", jsonOutputPath)
	}

	return b.String()
}

// formatPermission はツール名とパターンからパーミッション形式の文字列を生成する｡
func formatPermission(toolName, pattern string) string {
	if toolName == "Bash" {
		return fmt.Sprintf("Bash(%s:*)", pattern)
	}
	return fmt.Sprintf("%s(%s)", toolName, pattern)
}
