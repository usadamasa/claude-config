package main

import (
	"encoding/json"
	"testing"
)

func TestGenerateReport(t *testing.T) {
	t.Run("Top Nセッションを抽出", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000, APICallCount: 10, Project: "proj-a", Model: "claude-opus-4-6"},
			{SessionID: "s2", TotalInputTokens: 300000, APICallCount: 30, Project: "proj-b", Model: "claude-opus-4-6"},
			{SessionID: "s3", TotalInputTokens: 200000, APICallCount: 20, Project: "proj-a", Model: "claude-opus-4-6"},
		}

		report := GenerateReport(results, 2)

		if len(report.TopSessions) != 2 {
			t.Fatalf("len(TopSessions) = %d, want 2", len(report.TopSessions))
		}
		// input_tokens降順でソートされている
		if report.TopSessions[0].SessionID != "s2" {
			t.Errorf("TopSessions[0].SessionID = %q, want %q", report.TopSessions[0].SessionID, "s2")
		}
		if report.TopSessions[1].SessionID != "s3" {
			t.Errorf("TopSessions[1].SessionID = %q, want %q", report.TopSessions[1].SessionID, "s3")
		}
	})

	t.Run("プロジェクト別サマリー", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000, TotalOutputTokens: 1000, APICallCount: 10, Project: "proj-a", Model: "claude-opus-4-6", UserMessageCount: 5},
			{SessionID: "s2", TotalInputTokens: 300000, TotalOutputTokens: 3000, APICallCount: 30, Project: "proj-b", Model: "claude-opus-4-6", UserMessageCount: 3},
			{SessionID: "s3", TotalInputTokens: 200000, TotalOutputTokens: 2000, APICallCount: 20, Project: "proj-a", Model: "claude-opus-4-6", UserMessageCount: 7},
		}

		report := GenerateReport(results, 10)

		if len(report.ProjectSummary) != 2 {
			t.Fatalf("len(ProjectSummary) = %d, want 2", len(report.ProjectSummary))
		}

		// proj-aのサマリーを確認
		var projA *ProjectSummary
		for i := range report.ProjectSummary {
			if report.ProjectSummary[i].Project == "proj-a" {
				projA = &report.ProjectSummary[i]
				break
			}
		}
		if projA == nil {
			t.Fatal("proj-aのサマリーが見つからない")
		}
		if projA.TotalInputTokens != 300000 {
			t.Errorf("proj-a TotalInputTokens = %d, want %d", projA.TotalInputTokens, 300000)
		}
		if projA.SessionCount != 2 {
			t.Errorf("proj-a SessionCount = %d, want %d", projA.SessionCount, 2)
		}
		if projA.AverageInputPerCall != 10000 {
			t.Errorf("proj-a AverageInputPerCall = %d, want %d", projA.AverageInputPerCall, 10000)
		}
	})

	t.Run("モデル別サマリー", func(t *testing.T) {
		results := []SessionResult{
			{
				SessionID:        "s1",
				TotalInputTokens: 100000,
				APICallCount:     10,
				Project:          "proj-a",
				Model:            "claude-opus-4-6",
				ModelUsage: map[string]ModelTokens{
					"claude-opus-4-6":           {InputTokens: 80000, OutputTokens: 800, CallCount: 8},
					"claude-haiku-4-5-20251001": {InputTokens: 20000, OutputTokens: 200, CallCount: 2},
				},
			},
		}

		report := GenerateReport(results, 10)

		if len(report.ModelSummary) != 2 {
			t.Fatalf("len(ModelSummary) = %d, want 2", len(report.ModelSummary))
		}
	})

	t.Run("全体統計", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000, TotalOutputTokens: 1000, APICallCount: 10, Project: "proj-a", UserMessageCount: 5},
			{SessionID: "s2", TotalInputTokens: 200000, TotalOutputTokens: 2000, APICallCount: 20, Project: "proj-b", UserMessageCount: 3},
		}

		report := GenerateReport(results, 10)

		if report.Summary.TotalSessions != 2 {
			t.Errorf("TotalSessions = %d, want 2", report.Summary.TotalSessions)
		}
		if report.Summary.TotalInputTokens != 300000 {
			t.Errorf("TotalInputTokens = %d, want %d", report.Summary.TotalInputTokens, 300000)
		}
		if report.Summary.TotalOutputTokens != 3000 {
			t.Errorf("TotalOutputTokens = %d, want %d", report.Summary.TotalOutputTokens, 3000)
		}
		if report.Summary.TotalAPICalls != 30 {
			t.Errorf("TotalAPICalls = %d, want %d", report.Summary.TotalAPICalls, 30)
		}
		if report.Summary.AverageInputPerCall != 10000 {
			t.Errorf("AverageInputPerCall = %d, want %d", report.Summary.AverageInputPerCall, 10000)
		}
	})

	t.Run("結果0件", func(t *testing.T) {
		report := GenerateReport(nil, 10)
		if report.Summary.TotalSessions != 0 {
			t.Errorf("TotalSessions = %d, want 0", report.Summary.TotalSessions)
		}
		if len(report.TopSessions) != 0 {
			t.Errorf("len(TopSessions) = %d, want 0", len(report.TopSessions))
		}
	})

	t.Run("JSON出力可能", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000, APICallCount: 10, Project: "proj-a", Model: "claude-opus-4-6"},
		}
		report := GenerateReport(results, 10)
		data, err := json.Marshal(report)
		if err != nil {
			t.Fatalf("JSON Marshal失敗: %v", err)
		}
		if len(data) == 0 {
			t.Error("JSON出力が空")
		}
	})
}

func TestTopNSelection(t *testing.T) {
	t.Run("NがセッションDe数より大きい場合は全件返す", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000},
		}
		report := GenerateReport(results, 10)
		if len(report.TopSessions) != 1 {
			t.Errorf("len(TopSessions) = %d, want 1", len(report.TopSessions))
		}
	})

	t.Run("N=0の場合は空", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000},
		}
		report := GenerateReport(results, 0)
		if len(report.TopSessions) != 0 {
			t.Errorf("len(TopSessions) = %d, want 0", len(report.TopSessions))
		}
	})
}
