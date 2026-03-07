package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/usadamasa/claude-config/internal/pathutil"
	"github.com/usadamasa/claude-config/internal/settings"
)

// Warning は閾値超過の警告を表す｡
type Warning struct {
	Type           string `json:"type"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
	Project        string `json:"project,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	Value          int64  `json:"value"`
	Threshold      int64  `json:"threshold"`
}

// ConfigHealth はグローバル設定の健全性情報を表す｡
type ConfigHealth struct {
	EnabledPlugins     int      `json:"enabled_plugins"`
	GlobalSkills       int      `json:"global_skills"`
	EnabledPluginNames []string `json:"enabled_plugin_names"`
	GlobalSkillNames   []string `json:"global_skill_names"`
}

// Report はtoken使用量分析レポートの全体構造｡
type Report struct {
	Summary        ReportSummary    `json:"summary"`
	ConfigHealth   *ConfigHealth    `json:"config_health,omitempty"`
	Warnings       []Warning        `json:"warnings"`
	TopSessions    []SessionResult  `json:"top_sessions"`
	ProjectSummary []ProjectSummary `json:"project_summary"`
	ModelSummary   []ModelSummary   `json:"model_summary"`
}

// ReportSummary は全体統計｡
type ReportSummary struct {
	TotalSessions      int   `json:"total_sessions"`
	TotalInputTokens   int64 `json:"total_input_tokens"`
	TotalOutputTokens  int64 `json:"total_output_tokens"`
	TotalAPICalls      int   `json:"total_api_calls"`
	AverageInputPerCall int64 `json:"average_input_per_call"`
	Days               int   `json:"days"`
}

// ProjectSummary はプロジェクト別の集計｡
type ProjectSummary struct {
	Project            string `json:"project"`
	TotalInputTokens   int64  `json:"total_input_tokens"`
	TotalOutputTokens  int64  `json:"total_output_tokens"`
	SessionCount       int    `json:"session_count"`
	TotalAPICalls      int    `json:"total_api_calls"`
	AverageInputPerCall int64 `json:"average_input_per_call"`
}

// ModelSummary はモデル別の集計｡
type ModelSummary struct {
	Model        string `json:"model"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	CallCount    int    `json:"call_count"`
}

// GenerateReport はセッション結果からレポートを生成する｡
func GenerateReport(results []SessionResult, topN int) Report {
	report := Report{}

	if len(results) == 0 {
		return report
	}

	// 全体統計
	var totalInput, totalOutput int64
	var totalCalls int
	for _, r := range results {
		totalInput += r.TotalInputTokens
		totalOutput += r.TotalOutputTokens
		totalCalls += r.APICallCount
	}

	var avgPerCall int64
	if totalCalls > 0 {
		avgPerCall = totalInput / int64(totalCalls)
	}

	report.Summary = ReportSummary{
		TotalSessions:      len(results),
		TotalInputTokens:   totalInput,
		TotalOutputTokens:  totalOutput,
		TotalAPICalls:      totalCalls,
		AverageInputPerCall: avgPerCall,
	}

	// Top N セッション(input tokens降順)
	sorted := make([]SessionResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalInputTokens > sorted[j].TotalInputTokens
	})

	if topN > 0 && topN < len(sorted) {
		report.TopSessions = sorted[:topN]
	} else if topN > 0 {
		report.TopSessions = sorted
	} else {
		report.TopSessions = []SessionResult{}
	}

	// プロジェクト別集計
	projMap := make(map[string]*ProjectSummary)
	for _, r := range results {
		proj := r.Project
		if proj == "" {
			proj = "(unknown)"
		}
		ps, ok := projMap[proj]
		if !ok {
			ps = &ProjectSummary{Project: proj}
			projMap[proj] = ps
		}
		ps.TotalInputTokens += r.TotalInputTokens
		ps.TotalOutputTokens += r.TotalOutputTokens
		ps.SessionCount++
		ps.TotalAPICalls += r.APICallCount
	}
	for _, ps := range projMap {
		if ps.TotalAPICalls > 0 {
			ps.AverageInputPerCall = ps.TotalInputTokens / int64(ps.TotalAPICalls)
		}
		report.ProjectSummary = append(report.ProjectSummary, *ps)
	}
	// input tokens降順でソート
	sort.Slice(report.ProjectSummary, func(i, j int) bool {
		return report.ProjectSummary[i].TotalInputTokens > report.ProjectSummary[j].TotalInputTokens
	})

	// モデル別集計
	modelMap := make(map[string]*ModelSummary)
	for _, r := range results {
		for model, mt := range r.ModelUsage {
			ms, ok := modelMap[model]
			if !ok {
				ms = &ModelSummary{Model: model}
				modelMap[model] = ms
			}
			ms.InputTokens += mt.InputTokens
			ms.OutputTokens += mt.OutputTokens
			ms.CallCount += mt.CallCount
		}
	}
	for _, ms := range modelMap {
		report.ModelSummary = append(report.ModelSummary, *ms)
	}
	sort.Slice(report.ModelSummary, func(i, j int) bool {
		return report.ModelSummary[i].InputTokens > report.ModelSummary[j].InputTokens
	})

	// 警告生成
	report.Warnings = generateWarnings(report, results)

	return report
}

const (
	thresholdProjectAvgInput  int64 = 80000
	thresholdGlobalAvgInput   int64 = 60000
	thresholdCallMessageRatio int64 = 50
)

func generateWarnings(report Report, results []SessionResult) []Warning {
	var warnings []Warning

	// 全体avg > 60K
	if report.Summary.AverageInputPerCall > thresholdGlobalAvgInput {
		warnings = append(warnings, Warning{
			Type:           "global_high_avg",
			Message:        "全体のaverage_input_per_callが閾値を超えています",
			Recommendation: "settings.jsonのenabledPluginsで不要なプラグインをfalseに設定し、~/.claude/skills/から未使用スキルを移動してください",
			Value:          report.Summary.AverageInputPerCall,
			Threshold:      thresholdGlobalAvgInput,
		})
	}

	// プロジェクト別 avg > 80K
	for _, ps := range report.ProjectSummary {
		if ps.AverageInputPerCall > thresholdProjectAvgInput {
			warnings = append(warnings, Warning{
				Type:           "high_avg_input",
				Message:        "プロジェクトのaverage_input_per_callが閾値を超えています",
				Recommendation: "プロジェクトのCLAUDE.mdが肥大化していないか確認し、プロジェクト固有のMCP設定やスキルを見直してください",
				Project:        ps.Project,
				Value:          ps.AverageInputPerCall,
				Threshold:      thresholdProjectAvgInput,
			})
		}
	}

	// セッション別 api_calls/user_messages > 50
	for _, r := range results {
		if r.UserMessageCount > 0 {
			ratio := int64(r.APICallCount) / int64(r.UserMessageCount)
			if ratio > thresholdCallMessageRatio {
				warnings = append(warnings, Warning{
					Type:           "high_call_ratio",
					Message:        "api_calls/user_messages比率が高すぎます｡subagent多段呼び出しの可能性があります",
					Recommendation: "subagentにmodel:haiku指定があるか確認し、不要なExploreエージェント呼び出しを減らしてください",
					SessionID:      r.SessionID,
					Project:        r.Project,
					Value:          ratio,
					Threshold:      thresholdCallMessageRatio,
				})
			}
		}
	}

	if warnings == nil {
		warnings = []Warning{}
	}

	return warnings
}

const (
	thresholdEnabledPlugins int64 = 10
	thresholdGlobalSkills   int64 = 15
)

func generateConfigWarnings(health ConfigHealth) []Warning {
	var warnings []Warning

	if int64(health.EnabledPlugins) > thresholdEnabledPlugins {
		warnings = append(warnings, Warning{
			Type:           "too_many_plugins",
			Message:        "有効プラグイン数が多すぎます",
			Recommendation: "settings.jsonのenabledPluginsを確認し、使用頻度の低いプラグインをfalseに設定してください｡LSP系(gopls,jdtls)は必要な言語のみ有効にしてください",
			Value:          int64(health.EnabledPlugins),
			Threshold:      thresholdEnabledPlugins,
		})
	}

	if int64(health.GlobalSkills) > thresholdGlobalSkills {
		warnings = append(warnings, Warning{
			Type:           "too_many_skills",
			Message:        "グローバルスキル数が多すぎます",
			Recommendation: "~/.claude/skills/を確認し、プロジェクト固有のスキルは各リポジトリの.claude/skills/に移動してください",
			Value:          int64(health.GlobalSkills),
			Threshold:      thresholdGlobalSkills,
		})
	}

	return warnings
}

// ListGlobalSkillNames は指定ディレクトリ内のスキル名一覧をソート済みで返す｡
// symlink先がディレクトリの場合もカウントする｡
func ListGlobalSkillNames(skillsDir string) []string {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			target, err := os.Stat(filepath.Join(skillsDir, e.Name()))
			if err == nil && target.IsDir() {
				names = append(names, e.Name())
			}
		}
	}
	sort.Strings(names)
	return names
}

// CountGlobalSkills は指定ディレクトリ内のスキル数(ディレクトリ数)を返す｡
func CountGlobalSkills(skillsDir string) int {
	return len(ListGlobalSkillNames(skillsDir))
}

func main() {
	days := flag.Int("days", 30, "分析対象期間(日数)")
	topN := flag.Int("top", 10, "表示するTop Nセッション数")
	projectsDir := flag.String("dir", "", "セッションディレクトリ (デフォルト: ~/.claude/projects)")
	settingsPath := flag.String("settings", "", "settings.jsonのパス (デフォルト: ~/.claude/settings.json)")
	warningsOnly := flag.Bool("warnings-only", false, "警告とconfig_healthのみ出力")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ホームディレクトリ取得失敗: %v\n", err)
		os.Exit(1)
	}

	dir := pathutil.ResolveProjectsDir(*projectsDir, home)

	results, err := ScanProjectsDir(dir, *days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "スキャン失敗: %v\n", err)
		os.Exit(1)
	}

	report := GenerateReport(results, *topN)
	report.Summary.Days = *days

	// ConfigHealth: グローバル設定の健全性チェック
	sPath := *settingsPath
	if sPath == "" {
		sPath = filepath.Join(home, ".claude", "settings.json")
	}
	skillsDir := filepath.Join(home, ".claude", "skills")
	health := ConfigHealth{
		GlobalSkills:   CountGlobalSkills(skillsDir),
		GlobalSkillNames: ListGlobalSkillNames(skillsDir),
	}
	if s, err := settings.Load(sPath); err == nil {
		health.EnabledPlugins = s.CountEnabledPlugins()
		health.EnabledPluginNames = s.ListEnabledPluginNames()
	}
	report.ConfigHealth = &health
	report.Warnings = append(report.Warnings, generateConfigWarnings(health)...)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if *warningsOnly {
		compact := struct {
			Summary      ReportSummary `json:"summary"`
			ConfigHealth *ConfigHealth `json:"config_health,omitempty"`
			Warnings     []Warning     `json:"warnings"`
		}{
			Summary:      report.Summary,
			ConfigHealth: report.ConfigHealth,
			Warnings:     report.Warnings,
		}
		if err := encoder.Encode(compact); err != nil {
			fmt.Fprintf(os.Stderr, "JSON出力失敗: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "JSON出力失敗: %v\n", err)
		os.Exit(1)
	}
}
