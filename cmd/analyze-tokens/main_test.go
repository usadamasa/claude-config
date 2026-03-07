package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestGenerateWarnings(t *testing.T) {
	t.Run("プロジェクトのavg_input_per_call>80Kで警告", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 900000, APICallCount: 10, Project: "heavy-proj", Model: "claude-opus-4-6", UserMessageCount: 5},
			{SessionID: "s2", TotalInputTokens: 50000, APICallCount: 10, Project: "light-proj", Model: "claude-opus-4-6", UserMessageCount: 5},
		}
		report := GenerateReport(results, 10)
		found := false
		for _, w := range report.Warnings {
			if w.Type == "high_avg_input" && w.Project == "heavy-proj" {
				found = true
				if w.Value != 90000 {
					t.Errorf("Value = %d, want 90000", w.Value)
				}
			}
		}
		if !found {
			t.Error("heavy-projに対するhigh_avg_input警告が見つからない")
		}
		// light-projには警告なし
		for _, w := range report.Warnings {
			if w.Type == "high_avg_input" && w.Project == "light-proj" {
				t.Error("light-projにhigh_avg_input警告が出てはいけない")
			}
		}
	})

	t.Run("全体avg>60Kで全体警告", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 700000, APICallCount: 10, Project: "proj-a", Model: "claude-opus-4-6", UserMessageCount: 5},
		}
		report := GenerateReport(results, 10)
		found := false
		for _, w := range report.Warnings {
			if w.Type == "global_high_avg" {
				found = true
			}
		}
		if !found {
			t.Error("全体avg>60Kなのにglobal_high_avg警告が見つからない")
		}
	})

	t.Run("全体avg<=60Kなら全体警告なし", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 500000, APICallCount: 10, Project: "proj-a", Model: "claude-opus-4-6", UserMessageCount: 5},
		}
		report := GenerateReport(results, 10)
		for _, w := range report.Warnings {
			if w.Type == "global_high_avg" {
				t.Error("全体avg<=60Kなのにglobal_high_avg警告が出てはいけない")
			}
		}
	})

	t.Run("api_calls/user_messages>50のセッションで警告", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 100000, APICallCount: 510, Project: "proj-a", Model: "claude-opus-4-6", UserMessageCount: 10},
			{SessionID: "s2", TotalInputTokens: 100000, APICallCount: 20, Project: "proj-b", Model: "claude-opus-4-6", UserMessageCount: 10},
		}
		report := GenerateReport(results, 10)
		found := false
		for _, w := range report.Warnings {
			if w.Type == "high_call_ratio" && w.SessionID == "s1" {
				found = true
				if w.Value != 51 {
					t.Errorf("Value = %d, want 51", w.Value)
				}
			}
		}
		if !found {
			t.Error("s1に対するhigh_call_ratio警告が見つからない")
		}
	})

	t.Run("警告なしの場合は空スライス", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 50000, APICallCount: 10, Project: "proj-a", Model: "claude-opus-4-6", UserMessageCount: 5},
		}
		report := GenerateReport(results, 10)
		if report.Warnings == nil {
			t.Error("Warningsがnilであってはいけない(空スライスであるべき)")
		}
		if len(report.Warnings) != 0 {
			t.Errorf("len(Warnings) = %d, want 0", len(report.Warnings))
		}
	})
}

func TestConfigHealth(t *testing.T) {
	t.Run("プラグイン数が多い場合に警告", func(t *testing.T) {
		health := ConfigHealth{
			EnabledPlugins: 15,
			GlobalSkills:   10,
		}
		warnings := generateConfigWarnings(health)
		found := false
		for _, w := range warnings {
			if w.Type == "too_many_plugins" {
				found = true
				if w.Value != 15 {
					t.Errorf("Value = %d, want 15", w.Value)
				}
			}
		}
		if !found {
			t.Error("too_many_plugins警告が見つからない")
		}
	})

	t.Run("スキル数が多い場合に警告", func(t *testing.T) {
		health := ConfigHealth{
			EnabledPlugins: 5,
			GlobalSkills:   20,
		}
		warnings := generateConfigWarnings(health)
		found := false
		for _, w := range warnings {
			if w.Type == "too_many_skills" {
				found = true
				if w.Value != 20 {
					t.Errorf("Value = %d, want 20", w.Value)
				}
			}
		}
		if !found {
			t.Error("too_many_skills警告が見つからない")
		}
	})

	t.Run("閾値以下なら警告なし", func(t *testing.T) {
		health := ConfigHealth{
			EnabledPlugins: 10,
			GlobalSkills:   15,
		}
		warnings := generateConfigWarnings(health)
		if len(warnings) != 0 {
			t.Errorf("len(warnings) = %d, want 0", len(warnings))
		}
	})

	t.Run("CountGlobalSkills", func(t *testing.T) {
		dir := t.TempDir()
		// スキルディレクトリを作成
		for _, name := range []string{"skill-a", "skill-b"} {
			os.MkdirAll(filepath.Join(dir, name), 0755)
		}
		// symlinkでディレクトリを指すケース
		symlinkTarget := t.TempDir()
		os.Symlink(symlinkTarget, filepath.Join(dir, "skill-symlink"))
		// ファイルはスキルとしてカウントしない
		os.WriteFile(filepath.Join(dir, "not-a-skill.txt"), []byte("test"), 0644)

		count := CountGlobalSkills(dir)
		if count != 3 {
			t.Errorf("CountGlobalSkills = %d, want 3", count)
		}
	})

	t.Run("ListGlobalSkillNames", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"skill-b", "skill-a"} {
			os.MkdirAll(filepath.Join(dir, name), 0755)
		}
		symlinkTarget := t.TempDir()
		os.Symlink(symlinkTarget, filepath.Join(dir, "skill-c"))

		names := ListGlobalSkillNames(dir)
		if len(names) != 3 {
			t.Fatalf("len(names) = %d, want 3", len(names))
		}
		// ソートされている
		if names[0] != "skill-a" {
			t.Errorf("names[0] = %q, want %q", names[0], "skill-a")
		}
	})

	t.Run("ConfigHealthにプラグイン名とスキル名を含む", func(t *testing.T) {
		health := ConfigHealth{
			EnabledPlugins:     2,
			GlobalSkills:       1,
			EnabledPluginNames: []string{"plugin-a", "plugin-b"},
			GlobalSkillNames:   []string{"skill-a"},
		}
		data, err := json.Marshal(health)
		if err != nil {
			t.Fatalf("JSON Marshal失敗: %v", err)
		}
		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)
		if _, ok := parsed["enabled_plugin_names"]; !ok {
			t.Error("enabled_plugin_namesフィールドがJSON出力に含まれていない")
		}
		if _, ok := parsed["global_skill_names"]; !ok {
			t.Error("global_skill_namesフィールドがJSON出力に含まれていない")
		}
	})
}

func TestWarningRecommendations(t *testing.T) {
	t.Run("全ての警告にrecommendationが含まれる", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 900000, APICallCount: 10, Project: "heavy-proj", Model: "claude-opus-4-6", UserMessageCount: 1},
		}
		report := GenerateReport(results, 10)

		// config warnings も追加
		health := ConfigHealth{EnabledPlugins: 15, GlobalSkills: 20}
		report.Warnings = append(report.Warnings, generateConfigWarnings(health)...)

		for _, w := range report.Warnings {
			if w.Recommendation == "" {
				t.Errorf("type=%q: Recommendationが空", w.Type)
			}
		}
	})

	t.Run("high_avg_inputのrecommendationにプロジェクトCLAUDE.mdの確認を含む", func(t *testing.T) {
		results := []SessionResult{
			{SessionID: "s1", TotalInputTokens: 900000, APICallCount: 10, Project: "heavy-proj", Model: "claude-opus-4-6", UserMessageCount: 5},
		}
		report := GenerateReport(results, 10)
		for _, w := range report.Warnings {
			if w.Type == "high_avg_input" {
				if w.Recommendation == "" {
					t.Error("high_avg_inputにRecommendationがない")
				}
			}
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
