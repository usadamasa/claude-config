package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// テスト用ヘルパー: JSONL ファイルを作成してパスを返す
func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// テスト用ヘルパー: Bash tool_use の JSONL 行を生成
func makeBashLine(command string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"` + command + `"}}]}}`
}

// テスト用ヘルパー: Read tool_use の JSONL 行を生成
func makeReadLine(filePath string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Read","input":{"file_path":"` + filePath + `"}}]}}`
}

// テスト用ヘルパー: Write tool_use の JSONL 行を生成
func makeWriteLine(filePath string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":{"file_path":"` + filePath + `","content":"test"}}]}}`
}

// テスト用ヘルパー: WebFetch tool_use の JSONL 行を生成(無視されるべき)
func makeWebFetchLine(url string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"WebFetch","input":{"url":"` + url + `","prompt":"test"}}]}}`
}

func TestScanJSONLFiles(t *testing.T) {
	t.Run("Bash tool_use を抽出する", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeBashLine("git status") + "\n" +
			makeBashLine("go test ./...") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("結果数: got %d, want 2", len(results))
		}
		if results[0].ToolName != "Bash" {
			t.Errorf("ToolName: got %s, want Bash", results[0].ToolName)
		}
		if results[0].Pattern != "git status" {
			t.Errorf("Pattern: got %s, want 'git status'", results[0].Pattern)
		}
		if results[1].Pattern != "go test" {
			t.Errorf("Pattern: got %s, want 'go test'", results[1].Pattern)
		}
	})

	t.Run("Read tool_use を抽出する", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeReadLine("/Users/testuser/project/src/main.go") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("結果数: got %d, want 1", len(results))
		}
		if results[0].ToolName != "Read" {
			t.Errorf("ToolName: got %s, want Read", results[0].ToolName)
		}
	})

	t.Run("Write tool_use を抽出する", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeWriteLine("/Users/testuser/project/src/main.go") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("結果数: got %d, want 1", len(results))
		}
		if results[0].ToolName != "Write" {
			t.Errorf("ToolName: got %s, want Write", results[0].ToolName)
		}
	})

	t.Run("WebFetch 等の対象外ツールを無視する", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeWebFetchLine("https://example.com") + "\n" +
			makeBashLine("git status") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("結果数: got %d, want 1", len(results))
		}
		if results[0].ToolName != "Bash" {
			t.Errorf("ToolName: got %s, want Bash", results[0].ToolName)
		}
	})

	t.Run("不正な JSON 行をスキップする", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := "not valid json\n" +
			makeBashLine("git log") + "\n" +
			"{broken json\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("結果数: got %d, want 1", len(results))
		}
	})

	t.Run("空ファイルを処理する", func(t *testing.T) {
		dir := t.TempDir()
		writeTestFile(t, dir, "empty.jsonl", "")

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("結果数: got %d, want 0", len(results))
		}
	})

	t.Run("ネストされたディレクトリを走査する", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "project-a", "subdir")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, subDir, "deep.jsonl", makeBashLine("task build")+"\n")

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("結果数: got %d, want 1", len(results))
		}
	})

	t.Run("古いファイルをフィルタする", func(t *testing.T) {
		dir := t.TempDir()
		filePath := writeTestFile(t, dir, "old.jsonl", makeBashLine("git status")+"\n")
		oldTime := time.Now().Add(-60 * 24 * time.Hour)
		if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("結果数: got %d, want 0", len(results))
		}
	})

	t.Run("存在しないディレクトリを処理する", func(t *testing.T) {
		results, err := ScanJSONLFiles("/nonexistent/path", 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("結果数: got %d, want 0", len(results))
		}
	})

	t.Run("Edit tool_use を抽出する", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Edit","input":{"file_path":"/tmp/test.go","old_string":"foo","new_string":"bar"}}]}}` + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("結果数: got %d, want 1", len(results))
		}
		// Edit は Write と同じパーミッションカテゴリ
		if results[0].ToolName != "Edit" {
			t.Errorf("ToolName: got %s, want Edit", results[0].ToolName)
		}
	})
}

func TestExtractBashPrefix(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{"単純コマンド", "ls", "ls"},
		{"引数付きコマンド", "ls -la /tmp", "ls"},
		{"git サブコマンド", "git status", "git status"},
		{"git add ファイル", "git add foo.go bar.go", "git add"},
		{"go test", "go test ./...", "go test"},
		{"go mod tidy", "go mod tidy", "go mod"},
		{"gh pr create", "gh pr create --title foo", "gh pr"},
		{"docker compose up", "docker compose up -d", "docker compose"},
		{"task build", "task build", "task build"},
		{"brew install", "brew install git", "brew install"},
		{"パイプ付き", "git log | head -5", "git log"},
		{"リダイレクト付き", "go test ./... > result.txt", "go test"},
		{"&& 付き", "git add . && git commit -m msg", "git add"},
		{"; 付き", "echo hello; git status", "echo"},
		{"空コマンド", "", ""},
		{"rm -rf", "rm -rf /tmp/test", "rm -rf"},
		{"make ターゲット", "make build", "make build"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBashPrefix(tt.command)
			if got != tt.want {
				t.Errorf("ExtractBashPrefix(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"ホームディレクトリ配下", "/Users/testuser/.ssh/id_rsa", "~/.ssh/**"},
		{"ホームディレクトリ直下ファイル", "/Users/testuser/.zshrc", "~/.zshrc"},
		{"深いパス", "/Users/testuser/.config/git/config", "~/.config/git/**"},
		{"プロジェクトの相対パス", "src/main.go", "src/**"},
		{"CLAUDE.md", "CLAUDE.md", "CLAUDE.md"},
		{".claude 配下", ".claude/skills/foo/SKILL.md", ".claude/skills/**"},
		{"空パス", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.path)
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
