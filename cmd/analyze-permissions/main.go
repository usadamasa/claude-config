package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	Add                    []PatternRecommendation `json:"add"`
	Review                 []PatternRecommendation `json:"review"`
	Unused                 []UnusedEntry           `json:"unused"`
	BareEntryWarnings      []string                `json:"bare_entry_warnings,omitempty"`
	DenyBypassWarnings     []DenyBypassWarning     `json:"deny_bypass_warnings,omitempty"`
	AllowEncompassesDeny   []AllowEncompassesDeny   `json:"allow_encompasses_deny,omitempty"`
}

// AllowEncompassesDeny は allow エントリが deny エントリを包含するパターン｡
type AllowEncompassesDeny struct {
	AllowEntry string `json:"allow_entry"`
	DenyEntry  string `json:"deny_entry"`
	Note       string `json:"note"`
}

// DenyBypassWarning は allow の Bash コマンドが deny の Read/Write をバイパスするリスク｡
type DenyBypassWarning struct {
	AllowEntry   string `json:"allow_entry"`
	BypassedDeny string `json:"bypassed_deny"`
	Risk         string `json:"risk"`
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

	// deny バイパスリスクのクロスチェック
	denyBypassWarnings := detectDenyBypassWarnings(allow, deny)

	// allow が deny を包含するパターンの検出
	allowEncompassesDeny := detectAllowEncompassesDeny(allow, deny)

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
			Add:                  addRecs,
			Review:               reviewRecs,
			Unused:               unusedRecs,
			BareEntryWarnings:    bareWarnings,
			DenyBypassWarnings:   denyBypassWarnings,
			AllowEncompassesDeny: allowEncompassesDeny,
		},
		AllPatterns: allPatterns,
	}
}

// getDenyBypassType は Bash コマンドパターンの deny バイパスタイプを返す｡
func getDenyBypassType(pattern string) string {
	for _, p := range bashDenyBypassPatterns {
		if p.match(pattern) {
			return p.bypass
		}
	}
	return ""
}

// detectDenyBypassWarnings は allow の Bash コマンドが deny の Read/Write をバイパスできるか検出する｡
func detectDenyBypassWarnings(allow, deny []string) []DenyBypassWarning {
	var warnings []DenyBypassWarning
	for _, allowEntry := range allow {
		tool, pattern, ok := ParsePermissionEntry(allowEntry)
		if !ok || tool != "Bash" {
			continue
		}
		cat := CategorizePermission("Bash", pattern)
		if !cat.DenyBypassRisk {
			continue
		}
		// deny リストの Read/Write エントリとクロスチェック
		bypassType := getDenyBypassType(pattern)
		for _, denyEntry := range deny {
			denyTool, _, denyOk := ParsePermissionEntry(denyEntry)
			if !denyOk {
				continue
			}
			if matchesBypassType(bypassType, denyTool) {
				warnings = append(warnings, DenyBypassWarning{
					AllowEntry:   allowEntry,
					BypassedDeny: denyEntry,
					Risk:         cat.Reason,
				})
			}
		}
	}
	return warnings
}

// detectAllowEncompassesDeny は allow エントリが deny エントリを包含するパターンを検出する｡
// 例: allow に Bash(gh:*) があり deny に Bash(gh auth:*) がある場合｡
func detectAllowEncompassesDeny(allow, deny []string) []AllowEncompassesDeny {
	var warnings []AllowEncompassesDeny
	for _, allowEntry := range allow {
		allowTool, allowPattern, allowOk := ParsePermissionEntry(allowEntry)
		if !allowOk || allowPattern == "" {
			continue
		}
		for _, denyEntry := range deny {
			denyTool, denyPattern, denyOk := ParsePermissionEntry(denyEntry)
			if !denyOk || denyPattern == "" {
				continue
			}
			if allowTool != denyTool {
				continue
			}
			// allow パターンが deny パターンのプレフィックスか判定
			if matchPattern(denyPattern, allowPattern) && allowPattern != denyPattern {
				warnings = append(warnings, AllowEncompassesDeny{
					AllowEntry: allowEntry,
					DenyEntry:  denyEntry,
					Note:       fmt.Sprintf("allow の %s が deny の %s を包含 (deny が優先されるが意図の確認を推奨)", allowEntry, denyEntry),
				})
			}
		}
	}
	return warnings
}

// matchesBypassType は bypass タイプと deny ツール名の組み合わせがマッチするか判定する｡
func matchesBypassType(bypassType, denyTool string) bool {
	switch bypassType {
	case "read":
		return denyTool == "Read"
	case "write":
		return denyTool == "Write"
	case "both":
		return denyTool == "Read" || denyTool == "Write"
	default:
		return false
	}
}

// matchPattern はスキャンパターンがパーミッションパターンにマッチするか判定する｡
func matchPattern(scanPattern, permPattern string) bool {
	if scanPattern == permPattern {
		return true
	}
	// プレフィックスマッチ (Bash の :* 形式に対応)
	// "gh" は "gh pr" にマッチ、"src" は "src/main.go" にマッチ
	if strings.HasPrefix(scanPattern, permPattern+" ") || strings.HasPrefix(scanPattern, permPattern+"/") {
		return true
	}
	// ワイルドカードマッチ
	if len(permPattern) > 3 && permPattern[len(permPattern)-3:] == "/**" {
		prefix := permPattern[:len(permPattern)-3]
		return strings.HasPrefix(scanPattern, prefix+"/") || scanPattern == prefix
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

// resolveProjectsDir は JSONL スキャン対象の projects ディレクトリを決定する｡
func resolveProjectsDir(projectsDirFlag, home string) string {
	if projectsDirFlag != "" {
		return projectsDirFlag
	}
	return filepath.Join(home, ".claude", "projects")
}

func main() {
	days := flag.Int("days", 30, "集計期間(日数)")
	settingsPath := flag.String("settings", "", "settings.json パス (デフォルト: git ルートの settings.json または ~/.claude/settings.json)")
	projectsDirFlag := flag.String("projects-dir", "", "projects ディレクトリパス (デフォルト: ~/.claude/projects)")
	format := flag.String("format", "summary", "出力形式: summary (テキストサマリ) または json (フル JSON)")
	outputPath := flag.String("output", "", "フル JSON の出力先ファイルパス (summary 形式と併用可)")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ホームディレクトリの取得に失敗: %v\n", err)
		os.Exit(1)
	}

	if *settingsPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "カレントディレクトリの取得に失敗: %v\n", err)
			os.Exit(1)
		}
		resolved, err := resolveSettingsPath(cwd, home)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		*settingsPath = resolved
	}

	projectsDir := resolveProjectsDir(*projectsDirFlag, home)

	// パーミッション読み込み
	allow, deny, ask, err := LoadPermissions(*settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "settings.json の読み込みに失敗 (%s): %v\n", *settingsPath, err)
		os.Exit(1)
	}

	// JSONL ファイルスキャン
	scanResults, err := ScanJSONLFiles(projectsDir, *days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSONL ファイルの走査に失敗: %v\n", err)
		os.Exit(1)
	}

	filesScanned := countUniqueFiles(scanResults)

	// レポート生成
	report := GenerateReport(scanResults, allow, deny, ask, *days, filesScanned)

	// --output 指定時はフル JSON をファイルに書き出し
	if *outputPath != "" {
		f, err := os.Create(*outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "出力ファイルの作成に失敗: %v\n", err)
			os.Exit(1)
		}
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			f.Close()
			fmt.Fprintf(os.Stderr, "JSON の書き出しに失敗: %v\n", err)
			os.Exit(1)
		}
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "出力ファイルのクローズに失敗: %v\n", err)
			os.Exit(1)
		}
	}

	// stdout への出力
	switch *format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "レポートの出力に失敗: %v\n", err)
			os.Exit(1)
		}
	case "summary":
		fmt.Print(FormatSummary(report, *outputPath))
	default:
		fmt.Fprintf(os.Stderr, "不明な出力形式: %s (summary または json を指定)\n", *format)
		os.Exit(1)
	}
}
