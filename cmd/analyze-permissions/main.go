package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Report はレポートの最上位構造｡
type Report struct {
	Metadata        ReportMetadata   `json:"metadata"`
	CurrentAllow    []string         `json:"current_allow"`
	CurrentDeny     []string         `json:"current_deny"`
	CurrentAsk      []string         `json:"current_ask"`
	Recommendations Recommendations  `json:"recommendations"`
	AllPatterns     []PatternSummary `json:"all_patterns"`
}

// ReportMetadata は分析の概要統計を保持する｡
type ReportMetadata struct {
	AnalysisDate   string `json:"analysis_date"`
	DaysAnalyzed   int    `json:"days_analyzed"`
	FilesScanned   int    `json:"files_scanned"`
	TotalToolCalls int    `json:"total_tool_calls"`
}

// Recommendations はカテゴリ別の推奨事項を含む｡
type Recommendations struct {
	Add               []PatternRecommendation `json:"add"`
	Review            []PatternRecommendation `json:"review"`
	Unused            []UnusedEntry           `json:"unused"`
	BareEntryWarnings []string                `json:"bare_entry_warnings,omitempty"`
}

// PatternRecommendation は追加または確認が推奨されるパターン｡
type PatternRecommendation struct {
	ToolName string   `json:"tool_name"`
	Pattern  string   `json:"pattern"`
	Count    int      `json:"count"`
	Category Category `json:"category"`
	Reason   string   `json:"reason"`
}

// UnusedEntry はパーミッションリストにあるが使用されていないエントリ｡
type UnusedEntry struct {
	Entry string `json:"entry"`
	List  string `json:"list"` // "allow", "deny", "ask"
	Note  string `json:"note"`
}

// PatternSummary はパターンの使用状況と分類の概要｡
type PatternSummary struct {
	ToolName    string   `json:"tool_name"`
	Pattern     string   `json:"pattern"`
	Count       int      `json:"count"`
	Category    Category `json:"category"`
	InAllowlist bool     `json:"in_allowlist"`
	InDenylist  bool     `json:"in_denylist"`
	InAsklist   bool     `json:"in_asklist"`
}

// GenerateReport はスキャン結果と現在のパーミッション設定からレポートを生成する｡
func GenerateReport(scanResults []ScanResult, allow, deny, ask []string, days, filesScanned int) Report {
	// パターンごとのカウントを集計
	type patternKey struct {
		toolName string
		pattern  string
	}
	counts := make(map[patternKey]int)
	for _, r := range scanResults {
		counts[patternKey{r.ToolName, r.Pattern}]++
	}

	// ベアエントリ警告を検出
	var bareWarnings []string
	for _, lists := range [][]string{allow, deny, ask} {
		for _, entry := range lists {
			tool, pattern, ok := ParsePermissionEntry(entry)
			if ok && pattern == "" && tool != "" {
				bareWarnings = append(bareWarnings, tool)
			}
		}
	}

	// 全パターンを分類
	var allPatterns []PatternSummary
	var addRecs []PatternRecommendation
	var reviewRecs []PatternRecommendation

	for key, count := range counts {
		cat := CategorizePermission(key.toolName, key.pattern)
		inAllow := MatchesPermission(key.toolName, key.pattern, allow)
		inDeny := MatchesPermission(key.toolName, key.pattern, deny)
		inAsk := MatchesPermission(key.toolName, key.pattern, ask)

		allPatterns = append(allPatterns, PatternSummary{
			ToolName:    key.toolName,
			Pattern:     key.pattern,
			Count:       count,
			Category:    cat.Category,
			InAllowlist: inAllow,
			InDenylist:  inDeny,
			InAsklist:   inAsk,
		})

		// 既存のパーミッションに含まれていないパターンを推奨
		if !inAllow && !inDeny && !inAsk {
			rec := PatternRecommendation{
				ToolName: key.toolName,
				Pattern:  key.pattern,
				Count:    count,
				Category: cat.Category,
				Reason:   cat.Reason,
			}
			switch cat.Category {
			case CategorySafe:
				addRecs = append(addRecs, rec)
			case CategoryReview, CategoryAsk:
				reviewRecs = append(reviewRecs, rec)
			case CategoryDeny:
				// deny カテゴリは deny リストに追加推奨
				rec.Reason = cat.Reason + " (deny リストへの追加を推奨)"
				reviewRecs = append(reviewRecs, rec)
			}
		}
	}

	// 未使用のパーミッションエントリを検出
	var unusedRecs []UnusedEntry
	checkUnused := func(entries []string, listName string) {
		for _, entry := range entries {
			tool, pattern, ok := ParsePermissionEntry(entry)
			if !ok {
				continue
			}
			// ベアエントリは未使用チェック対象外
			if pattern == "" {
				continue
			}

			used := false
			for key := range counts {
				if key.toolName == tool && matchPattern(key.pattern, pattern) {
					used = true
					break
				}
			}
			if !used {
				unusedRecs = append(unusedRecs, UnusedEntry{
					Entry: entry,
					List:  listName,
					Note:  fmt.Sprintf("過去%d日間使用なし", days),
				})
			}
		}
	}
	checkUnused(allow, "allow")
	checkUnused(deny, "deny")
	checkUnused(ask, "ask")

	// ソート
	sort.Slice(allPatterns, func(i, j int) bool { return allPatterns[i].Count > allPatterns[j].Count })
	sort.Slice(addRecs, func(i, j int) bool { return addRecs[i].Count > addRecs[j].Count })
	sort.Slice(reviewRecs, func(i, j int) bool { return reviewRecs[i].Count > reviewRecs[j].Count })

	return Report{
		Metadata: ReportMetadata{
			AnalysisDate:   time.Now().Format("2006-01-02"),
			DaysAnalyzed:   days,
			FilesScanned:   filesScanned,
			TotalToolCalls: len(scanResults),
		},
		CurrentAllow: allow,
		CurrentDeny:  deny,
		CurrentAsk:   ask,
		Recommendations: Recommendations{
			Add:               addRecs,
			Review:            reviewRecs,
			Unused:            unusedRecs,
			BareEntryWarnings: bareWarnings,
		},
		AllPatterns: allPatterns,
	}
}

// matchPattern はスキャンパターンがパーミッションパターンにマッチするか判定する｡
func matchPattern(scanPattern, permPattern string) bool {
	if scanPattern == permPattern {
		return true
	}
	// ワイルドカードマッチ
	if len(permPattern) > 3 && permPattern[len(permPattern)-3:] == "/**" {
		prefix := permPattern[:len(permPattern)-3]
		return len(scanPattern) > len(prefix) && scanPattern[:len(prefix)] == prefix
	}
	return false
}

// countUniqueFiles はスキャン結果のユニークなファイル数をカウントする｡
func countUniqueFiles(results []ScanResult) int {
	seen := make(map[string]bool)
	for _, r := range results {
		seen[r.FilePath] = true
	}
	return len(seen)
}

func main() {
	days := flag.Int("days", 30, "集計期間(日数)")
	settingsPath := flag.String("settings", "", "settings.json パス (デフォルト: ~/.claude/settings.json)")
	flag.Parse()

	if *settingsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ホームディレクトリの取得に失敗: %v\n", err)
			os.Exit(1)
		}
		*settingsPath = filepath.Join(home, ".claude", "settings.json")
	}

	projectsDir := filepath.Join(filepath.Dir(*settingsPath), "projects")

	// パーミッション読み込み
	allow, deny, ask, err := LoadPermissions(*settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "settings.json の読み込みに失敗: %v\n", err)
		os.Exit(1)
	}

	// JSONL ファイルスキャン
	scanResults, err := ScanJSONLFiles(projectsDir, *days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSONL ファイルの走査に失敗: %v\n", err)
		os.Exit(1)
	}

	filesScanned := countUniqueFiles(scanResults)

	// レポート生成・出力
	report := GenerateReport(scanResults, allow, deny, ask, *days, filesScanned)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "レポートの出力に失敗: %v\n", err)
		os.Exit(1)
	}
}
