package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScanResult represents a single WebFetch/Fetch invocation found in a JSONL file.
type ScanResult struct {
	URL       string
	Domain    string
	Tool      string // "WebFetch" or "Fetch"
	Timestamp time.Time
	FilePath  string
}

// jsonlLine represents a single line in a Claude session JSONL file.
// The actual format is: {"type":"assistant","message":{"content":[{"type":"tool_use","name":"WebFetch","input":{...}}]}}
type jsonlLine struct {
	Type    string `json:"type"`
	Message struct {
		Content []contentBlock `json:"content"`
	} `json:"message"`
}

// contentBlock represents an element in message.content[].
type contentBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// webFetchInput represents the input fields of a WebFetch tool_use.
type webFetchInput struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// ScanJSONLFiles walks the given directory for .jsonl files modified within
// the specified number of days and extracts WebFetch tool_use entries.
func ScanJSONLFiles(projectsDir string, days int) ([]ScanResult, error) {
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	var results []ScanResult

	// If directory doesn't exist, return empty results.
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return results, nil
	}

	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}

		// Filter by modification time.
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			return nil
		}

		fileResults, err := scanSingleFile(path)
		if err != nil {
			return nil // skip files that can't be processed
		}
		results = append(results, fileResults...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// scanSingleFile reads a JSONL file line by line and extracts WebFetch entries.
func scanSingleFile(path string) ([]ScanResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var results []ScanResult
	scanner := bufio.NewScanner(f)
	// Increase buffer size for potentially long lines.
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlLine
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip invalid JSON lines
		}

		for _, block := range entry.Message.Content {
			if block.Type != "tool_use" {
				continue
			}
			if block.Name != "WebFetch" && block.Name != "Fetch" {
				continue
			}

			var input webFetchInput
			if err := json.Unmarshal(block.Input, &input); err != nil {
				continue
			}
			if input.URL == "" {
				continue
			}

			domain, err := ExtractDomain(input.URL)
			if err != nil {
				continue
			}

			results = append(results, ScanResult{
				URL:      input.URL,
				Domain:   domain,
				Tool:     block.Name,
				FilePath: path,
			})
		}
	}
	return results, scanner.Err()
}
