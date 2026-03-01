package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/usadamasa/claude-config/internal/jsonlscan"
)

// ScanResult represents a single WebFetch/Fetch invocation found in a JSONL file.
type ScanResult struct {
	URL       string
	Domain    string
	Tool      string
	Timestamp time.Time
	FilePath  string
}

// webFetchInput represents the input fields of a WebFetch tool_use.
type webFetchInput struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// ScanJSONLFiles walks the given directory for .jsonl files modified within
// the specified number of days and extracts WebFetch tool_use entries.
func ScanJSONLFiles(projectsDir string, days int) ([]ScanResult, error) {
	var results []ScanResult

	err := jsonlscan.WalkJSONLFiles(projectsDir, jsonlscan.WalkOptions{Days: days}, func(path string) error {
		fileResults, err := scanSingleFile(path)
		if err != nil {
			return nil
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
	scanner := jsonlscan.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlscan.JSONLLine
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
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
