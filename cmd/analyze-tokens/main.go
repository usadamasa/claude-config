package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Report はtoken使用量分析レポートの全体構造｡
type Report struct {
	Summary        ReportSummary    `json:"summary"`
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

	return report
}

func main() {
	days := flag.Int("days", 30, "分析対象期間(日数)")
	topN := flag.Int("top", 10, "表示するTop Nセッション数")
	projectsDir := flag.String("dir", "", "セッションディレクトリ (デフォルト: ~/.claude/projects)")
	flag.Parse()

	if *projectsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ホームディレクトリ取得失敗: %v\n", err)
			os.Exit(1)
		}
		*projectsDir = filepath.Join(home, ".claude", "projects")
	}

	results, err := ScanProjectsDir(*projectsDir, *days)
	if err != nil {
		fmt.Fprintf(os.Stderr, "スキャン失敗: %v\n", err)
		os.Exit(1)
	}

	report := GenerateReport(results, *topN)
	report.Summary.Days = *days

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "JSON出力失敗: %v\n", err)
		os.Exit(1)
	}
}
