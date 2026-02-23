package main

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTestJSONL はテスト用のJSONLファイルを作成するヘルパー｡
func writeTestJSONL(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("テストファイル作成失敗: %v", err)
	}
	return path
}

func TestScanSessionFile(t *testing.T) {
	t.Run("assistantメッセージからusageを抽出", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"user","cwd":"/Users/test/project","sessionId":"sess-1","message":{"content":[{"type":"text","text":"hello"}]}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-1","message":{"model":"claude-opus-4-6","type":"message","role":"assistant","content":[{"type":"text","text":"hi"}],"usage":{"input_tokens":50000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":100}}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-1","message":{"model":"claude-opus-4-6","type":"message","role":"assistant","content":[{"type":"tool_use","name":"Read","input":{}}],"usage":{"input_tokens":55000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":200}}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}

		if result.SessionID != "sess-1" {
			t.Errorf("SessionID = %q, want %q", result.SessionID, "sess-1")
		}
		if result.APICallCount != 2 {
			t.Errorf("APICallCount = %d, want %d", result.APICallCount, 2)
		}
		if result.TotalInputTokens != 105000 {
			t.Errorf("TotalInputTokens = %d, want %d", result.TotalInputTokens, 105000)
		}
		if result.TotalOutputTokens != 300 {
			t.Errorf("TotalOutputTokens = %d, want %d", result.TotalOutputTokens, 300)
		}
		if result.Model != "claude-opus-4-6" {
			t.Errorf("Model = %q, want %q", result.Model, "claude-opus-4-6")
		}
	})

	t.Run("progressメッセージ(subagent)からusageを抽出", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"user","cwd":"/Users/test/project","sessionId":"sess-2","userType":"external","message":{"content":[{"type":"text","text":"explore"}]}}
{"type":"progress","cwd":"/Users/test/project","sessionId":"sess-2","data":{"message":{"type":"assistant","message":{"model":"claude-haiku-4-5-20251001","type":"message","role":"assistant","content":[{"type":"tool_use","name":"Glob","input":{}}],"usage":{"input_tokens":30000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":50}}},"type":"message"}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-2","message":{"model":"claude-opus-4-6","type":"message","role":"assistant","content":[{"type":"text","text":"done"}],"usage":{"input_tokens":60000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":150}}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}

		// subagentの使用量も含める
		if result.TotalInputTokens != 90000 {
			t.Errorf("TotalInputTokens = %d, want %d", result.TotalInputTokens, 90000)
		}
		if result.TotalOutputTokens != 200 {
			t.Errorf("TotalOutputTokens = %d, want %d", result.TotalOutputTokens, 200)
		}
		if result.APICallCount != 2 {
			t.Errorf("APICallCount = %d, want %d (assistant=1 + progress/subagent=1)", result.APICallCount, 2)
		}
		// モデル別の集計
		if result.ModelUsage["claude-opus-4-6"].InputTokens != 60000 {
			t.Errorf("opus InputTokens = %d, want %d", result.ModelUsage["claude-opus-4-6"].InputTokens, 60000)
		}
		if result.ModelUsage["claude-haiku-4-5-20251001"].InputTokens != 30000 {
			t.Errorf("haiku InputTokens = %d, want %d", result.ModelUsage["claude-haiku-4-5-20251001"].InputTokens, 30000)
		}
	})

	t.Run("userメッセージ数をカウント", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"user","cwd":"/Users/test/project","sessionId":"sess-3","userType":"external","message":{"content":[{"type":"text","text":"msg1"}]}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-3","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":1000,"output_tokens":10}}}
{"type":"user","cwd":"/Users/test/project","sessionId":"sess-3","userType":"external","message":{"content":[{"type":"text","text":"msg2"}]}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-3","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":2000,"output_tokens":20}}}
{"type":"user","cwd":"/Users/test/project","sessionId":"sess-3","userType":"internal","message":{"content":[{"type":"text","text":"system"}]}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}

		// externalのuserメッセージのみカウント
		if result.UserMessageCount != 2 {
			t.Errorf("UserMessageCount = %d, want %d", result.UserMessageCount, 2)
		}
	})

	t.Run("空ファイル", func(t *testing.T) {
		dir := t.TempDir()
		writeTestJSONL(t, dir, "empty.jsonl", "")

		result, err := ScanSessionFile(filepath.Join(dir, "empty.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}
		if result.APICallCount != 0 {
			t.Errorf("APICallCount = %d, want 0", result.APICallCount)
		}
	})

	t.Run("不正なJSONを含む行はスキップ", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-4","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":1000,"output_tokens":10}}}
not valid json
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-4","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":2000,"output_tokens":20}}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}
		if result.APICallCount != 2 {
			t.Errorf("APICallCount = %d, want %d", result.APICallCount, 2)
		}
		if result.TotalInputTokens != 3000 {
			t.Errorf("TotalInputTokens = %d, want %d", result.TotalInputTokens, 3000)
		}
	})

	t.Run("cwdからプロジェクト名を抽出", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"user","cwd":"/Users/test/src/github.com/org/my-project","sessionId":"sess-5","userType":"external","message":{"content":[]}}
{"type":"assistant","cwd":"/Users/test/src/github.com/org/my-project","sessionId":"sess-5","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":1000,"output_tokens":10}}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}
		if result.Project != "my-project" {
			t.Errorf("Project = %q, want %q", result.Project, "my-project")
		}
	})

	t.Run("cache tokenの集計", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-6","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":50000,"cache_creation_input_tokens":10000,"cache_read_input_tokens":5000,"output_tokens":100}}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-6","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":40000,"cache_creation_input_tokens":0,"cache_read_input_tokens":20000,"output_tokens":200}}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}
		if result.TotalCacheCreationTokens != 10000 {
			t.Errorf("TotalCacheCreationTokens = %d, want %d", result.TotalCacheCreationTokens, 10000)
		}
		if result.TotalCacheReadTokens != 25000 {
			t.Errorf("TotalCacheReadTokens = %d, want %d", result.TotalCacheReadTokens, 25000)
		}
	})

	t.Run("ツール使用をカウント", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-7","message":{"model":"claude-opus-4-6","content":[{"type":"tool_use","name":"Read","input":{}},{"type":"tool_use","name":"Bash","input":{}}],"usage":{"input_tokens":1000,"output_tokens":10}}}
{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess-7","message":{"model":"claude-opus-4-6","content":[{"type":"tool_use","name":"Read","input":{}}],"usage":{"input_tokens":2000,"output_tokens":20}}}
`
		writeTestJSONL(t, dir, "test.jsonl", content)

		result, err := ScanSessionFile(filepath.Join(dir, "test.jsonl"))
		if err != nil {
			t.Fatalf("ScanSessionFile失敗: %v", err)
		}
		if result.ToolUsage["Read"] != 2 {
			t.Errorf("ToolUsage[Read] = %d, want %d", result.ToolUsage["Read"], 2)
		}
		if result.ToolUsage["Bash"] != 1 {
			t.Errorf("ToolUsage[Bash] = %d, want %d", result.ToolUsage["Bash"], 1)
		}
	})
}

func TestScanProjectsDir(t *testing.T) {
	t.Run("複数プロジェクトの走査", func(t *testing.T) {
		dir := t.TempDir()

		// プロジェクト1
		proj1 := filepath.Join(dir, "-Users-test-src-github-com-org-project-a")
		os.MkdirAll(proj1, 0755)
		writeTestJSONL(t, proj1, "sess-a.jsonl", `{"type":"user","cwd":"/Users/test/src/github.com/org/project-a","sessionId":"sess-a","userType":"external","message":{"content":[]}}
{"type":"assistant","cwd":"/Users/test/src/github.com/org/project-a","sessionId":"sess-a","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":50000,"output_tokens":100}}}
`)
		// プロジェクト2
		proj2 := filepath.Join(dir, "-Users-test-src-github-com-org-project-b")
		os.MkdirAll(proj2, 0755)
		writeTestJSONL(t, proj2, "sess-b.jsonl", `{"type":"user","cwd":"/Users/test/src/github.com/org/project-b","sessionId":"sess-b","userType":"external","message":{"content":[]}}
{"type":"assistant","cwd":"/Users/test/src/github.com/org/project-b","sessionId":"sess-b","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":30000,"output_tokens":50}}}
`)

		results, err := ScanProjectsDir(dir, 30)
		if err != nil {
			t.Fatalf("ScanProjectsDir失敗: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want %d", len(results), 2)
		}
	})

	t.Run("存在しないディレクトリはエラーなし", func(t *testing.T) {
		results, err := ScanProjectsDir("/nonexistent/path", 30)
		if err != nil {
			t.Fatalf("ScanProjectsDir失敗: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("len(results) = %d, want 0", len(results))
		}
	})

	t.Run("非JSONLファイルはスキップ", func(t *testing.T) {
		dir := t.TempDir()
		proj := filepath.Join(dir, "-Users-test-project")
		os.MkdirAll(proj, 0755)
		// ディレクトリ(セッションID名)はスキップ
		os.MkdirAll(filepath.Join(proj, "some-uuid"), 0755)
		// 通常のファイルはスキップ
		os.WriteFile(filepath.Join(proj, "notes.txt"), []byte("test"), 0644)
		// JSONLファイルのみ処理
		writeTestJSONL(t, proj, "sess.jsonl", `{"type":"assistant","cwd":"/Users/test/project","sessionId":"sess","message":{"model":"claude-opus-4-6","content":[],"usage":{"input_tokens":1000,"output_tokens":10}}}
`)

		results, err := ScanProjectsDir(dir, 30)
		if err != nil {
			t.Fatalf("ScanProjectsDir失敗: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("len(results) = %d, want 1", len(results))
		}
	})
}

func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		name string
		cwd  string
		want string
	}{
		{
			name: "GitHub構成のパス",
			cwd:  "/Users/test/src/github.com/org/my-project",
			want: "my-project",
		},
		{
			name: "worktreeパス",
			cwd:  "/Users/test/src/github.com/org/worktree/my-project/feature-branch",
			want: "my-project",
		},
		{
			name: "ホームディレクトリ直下",
			cwd:  "/Users/test",
			want: "test",
		},
		{
			name: "一般パス",
			cwd:  "/tmp/workspace/project",
			want: "project",
		},
		{
			name: "空文字列",
			cwd:  "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractProjectName(tt.cwd)
			if got != tt.want {
				t.Errorf("ExtractProjectName(%q) = %q, want %q", tt.cwd, got, tt.want)
			}
		})
	}
}

func TestAverageInputTokensPerCall(t *testing.T) {
	t.Run("通常計算", func(t *testing.T) {
		r := &SessionResult{
			TotalInputTokens: 100000,
			APICallCount:     10,
		}
		if avg := r.AverageInputTokensPerCall(); avg != 10000 {
			t.Errorf("AverageInputTokensPerCall() = %d, want %d", avg, 10000)
		}
	})
	t.Run("APIコール0の場合", func(t *testing.T) {
		r := &SessionResult{
			TotalInputTokens: 100000,
			APICallCount:     0,
		}
		if avg := r.AverageInputTokensPerCall(); avg != 0 {
			t.Errorf("AverageInputTokensPerCall() = %d, want 0", avg)
		}
	})
}
