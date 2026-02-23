package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper to build a JSONL line in the actual Claude session log format.
// The real format nests tool_use inside message.content[].
func makeWebFetchLine(url string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"WebFetch","input":{"url":"` + url + `","prompt":"test"}}]}}`
}

func makeFetchLine(url string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Fetch","input":{"url":"` + url + `","prompt":"test"}}]}}`
}

func makeOtherToolLine(name string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"` + name + `","input":{"file_path":"/tmp/test"}}]}}`
}

func TestScanJSONLFiles(t *testing.T) {
	t.Run("extracts URL from WebFetch tool_use", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeWebFetchLine("https://docs.example.com/api") + "\n" +
			makeOtherToolLine("Read") + "\n" +
			makeWebFetchLine("https://github.com/foo/bar") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0].URL != "https://docs.example.com/api" {
			t.Errorf("expected first URL to be https://docs.example.com/api, got %s", results[0].URL)
		}
		if results[0].Domain != "docs.example.com" {
			t.Errorf("expected first domain to be docs.example.com, got %s", results[0].Domain)
		}
		if results[1].URL != "https://github.com/foo/bar" {
			t.Errorf("expected second URL to be https://github.com/foo/bar, got %s", results[1].URL)
		}
		if results[1].Domain != "github.com" {
			t.Errorf("expected second domain to be github.com, got %s", results[1].Domain)
		}
	})

	t.Run("extracts multiple WebFetch from single content array", func(t *testing.T) {
		dir := t.TempDir()
		// A single message can have multiple tool_use entries in content[]
		jsonlContent := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"WebFetch","input":{"url":"https://a.com","prompt":"t"}},{"type":"tool_use","name":"WebFetch","input":{"url":"https://b.com","prompt":"t"}}]}}` + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("skips non-WebFetch tool_use", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeOtherToolLine("Read") + "\n" +
			makeOtherToolLine("Bash") + "\n" +
			`{"type":"user","message":{"role":"user","content":[{"type":"tool_result","content":"ok"}]}}` + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("skips invalid JSON lines", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := "not valid json\n" +
			makeWebFetchLine("https://example.com") + "\n" +
			"{broken json here\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Domain != "example.com" {
			t.Errorf("expected domain example.com, got %s", results[0].Domain)
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		dir := t.TempDir()
		writeTestFile(t, dir, "empty.jsonl", "")

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("scans nested directories", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "project-a", "subdir")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		jsonlContent := makeWebFetchLine("https://nested.example.com") + "\n"
		writeTestFile(t, subDir, "deep.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Domain != "nested.example.com" {
			t.Errorf("expected domain nested.example.com, got %s", results[0].Domain)
		}
	})

	t.Run("filters by file modification time", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeWebFetchLine("https://old.example.com") + "\n"
		filePath := writeTestFile(t, dir, "old.jsonl", jsonlContent)
		// Set mtime to 60 days ago
		oldTime := time.Now().Add(-60 * 24 * time.Hour)
		if err := os.Chtimes(filePath, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected 0 results (old file filtered out), got %d", len(results))
		}
	})

	t.Run("handles non-existent directory", func(t *testing.T) {
		results, err := ScanJSONLFiles("/nonexistent/path", 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("extracts URL from Fetch tool_use", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeFetchLine("https://api.example.com/v1") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].URL != "https://api.example.com/v1" {
			t.Errorf("expected URL https://api.example.com/v1, got %s", results[0].URL)
		}
		if results[0].Domain != "api.example.com" {
			t.Errorf("expected domain api.example.com, got %s", results[0].Domain)
		}
		if results[0].Tool != "Fetch" {
			t.Errorf("expected tool Fetch, got %s", results[0].Tool)
		}
	})

	t.Run("extracts both WebFetch and Fetch from mixed content", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeWebFetchLine("https://docs.example.com/api") + "\n" +
			makeFetchLine("https://api.example.com/v1") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if results[0].Tool != "WebFetch" {
			t.Errorf("expected first tool WebFetch, got %s", results[0].Tool)
		}
		if results[1].Tool != "Fetch" {
			t.Errorf("expected second tool Fetch, got %s", results[1].Tool)
		}
	})

	t.Run("records tool name in ScanResult", func(t *testing.T) {
		dir := t.TempDir()
		jsonlContent := makeWebFetchLine("https://github.com/foo") + "\n"
		writeTestFile(t, dir, "session.jsonl", jsonlContent)

		results, err := ScanJSONLFiles(dir, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Tool != "WebFetch" {
			t.Errorf("expected tool WebFetch, got %s", results[0].Tool)
		}
	})
}

// writeTestFile creates a file in the given directory and returns its path.
func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
