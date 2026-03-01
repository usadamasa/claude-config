package main

import (
	"testing"
)

func TestCheckPaths(t *testing.T) {
	home := "/Users/masaru_uchida"
	cwd := "/Users/masaru_uchida/src/github.com/usadamasa/claude-config"

	tests := []struct {
		name    string
		paths   []string
		wantMsg string // 空なら deny なし
	}{
		{
			name:    "許可サブディレクトリ: .claude",
			paths:   []string{home + "/.claude/settings.json"},
			wantMsg: "",
		},
		{
			name:    "許可サブディレクトリ: src",
			paths:   []string{home + "/src/project/main.go"},
			wantMsg: "",
		},
		{
			name:    "許可サブディレクトリ: obsidian",
			paths:   []string{home + "/obsidian/notes.md"},
			wantMsg: "",
		},
		{
			name:    "許可サブディレクトリ: tmp",
			paths:   []string{home + "/tmp/scratch.txt"},
			wantMsg: "",
		},
		{
			name:    "許可サブディレクトリ: workspace",
			paths:   []string{home + "/workspace/project"},
			wantMsg: "",
		},
		{
			name:    "cwd は許可",
			paths:   []string{cwd + "/main.go"},
			wantMsg: "",
		},
		{
			name:    "ホーム外は通過",
			paths:   []string{"/tmp/foo.txt"},
			wantMsg: "",
		},
		{
			name:    "禁止パス: Downloads",
			paths:   []string{home + "/Downloads/secret.pdf"},
			wantMsg: home + "/Downloads/secret.pdf",
		},
		{
			name:    "禁止パス: Documents",
			paths:   []string{home + "/Documents/private.docx"},
			wantMsg: home + "/Documents/private.docx",
		},
		{
			name:    "禁止パス: .ssh",
			paths:   []string{home + "/.ssh/id_rsa"},
			wantMsg: home + "/.ssh/id_rsa",
		},
		{
			name:    "ホームディレクトリ自体",
			paths:   []string{home},
			wantMsg: home,
		},
		{
			name:    "複数パス: 全て許可",
			paths:   []string{home + "/src/a", home + "/tmp/b"},
			wantMsg: "",
		},
		{
			name:    "複数パス: 1つが禁止",
			paths:   []string{home + "/src/a", home + "/Downloads/b"},
			wantMsg: home + "/Downloads/b",
		},
		{
			name:    "空パスリスト",
			paths:   nil,
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPaths(tt.paths, home, cwd)
			if tt.wantMsg == "" && got != "" {
				t.Errorf("checkPaths(%v) = %q, want empty (通過)", tt.paths, got)
			}
			if tt.wantMsg != "" && got == "" {
				t.Errorf("checkPaths(%v) = empty, want deny containing %q", tt.paths, tt.wantMsg)
			}
			if tt.wantMsg != "" && got != "" && got != tt.wantMsg {
				t.Errorf("checkPaths(%v) = %q, want %q", tt.paths, got, tt.wantMsg)
			}
		})
	}
}

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]any
		want     string
	}{
		{
			name:     "Read: file_path",
			toolName: "Read",
			input:    map[string]any{"file_path": "/tmp/test.txt"},
			want:     "/tmp/test.txt",
		},
		{
			name:     "Edit: file_path",
			toolName: "Edit",
			input:    map[string]any{"file_path": "/tmp/test.txt"},
			want:     "/tmp/test.txt",
		},
		{
			name:     "Write: file_path",
			toolName: "Write",
			input:    map[string]any{"file_path": "/tmp/test.txt"},
			want:     "/tmp/test.txt",
		},
		{
			name:     "NotebookEdit: notebook_path",
			toolName: "NotebookEdit",
			input:    map[string]any{"notebook_path": "/tmp/notebook.ipynb"},
			want:     "/tmp/notebook.ipynb",
		},
		{
			name:     "Glob: path",
			toolName: "Glob",
			input:    map[string]any{"path": "/tmp"},
			want:     "/tmp",
		},
		{
			name:     "Grep: path",
			toolName: "Grep",
			input:    map[string]any{"path": "/tmp"},
			want:     "/tmp",
		},
		{
			name:     "Glob: path が空 (cwd 使用)",
			toolName: "Glob",
			input:    map[string]any{},
			want:     "",
		},
		{
			name:     "未知のツール",
			toolName: "Agent",
			input:    map[string]any{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFilePath(tt.toolName, tt.input)
			if got != tt.want {
				t.Errorf("extractFilePath(%q, %v) = %q, want %q", tt.toolName, tt.input, got, tt.want)
			}
		})
	}
}
