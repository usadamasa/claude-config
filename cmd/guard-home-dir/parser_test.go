package main

import (
	"slices"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "単純なコマンド",
			input: "find /tmp -name foo",
			want:  []string{"find", "/tmp", "-name", "foo"},
		},
		{
			name:  "ダブルクォート内のスペース",
			input: `find /tmp -name "*.md"`,
			want:  []string{"find", "/tmp", "-name", "*.md"},
		},
		{
			name:  "シングルクォート内のスペース",
			input: `find /tmp -name '*.md'`,
			want:  []string{"find", "/tmp", "-name", "*.md"},
		},
		{
			name:  "空文字列",
			input: "",
			want:  nil,
		},
		{
			name:  "リダイレクト付き",
			input: "find /tmp -name foo 2>/dev/null",
			want:  []string{"find", "/tmp", "-name", "foo", "2>/dev/null"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenize(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("tokenize(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitCommands(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "単一コマンド",
			input: "find /tmp -name foo",
			want:  []string{"find /tmp -name foo"},
		},
		{
			name:  "パイプ",
			input: "find /tmp -name foo | head -10",
			want:  []string{"find /tmp -name foo ", " head -10"},
		},
		{
			name:  "AND チェーン",
			input: "git status && find /tmp -name foo",
			want:  []string{"git status ", " find /tmp -name foo"},
		},
		{
			name:  "OR チェーン",
			input: "find /tmp || echo not found",
			want:  []string{"find /tmp ", " echo not found"},
		},
		{
			name:  "セミコロン",
			input: "cd /tmp; find . -name foo",
			want:  []string{"cd /tmp", " find . -name foo"},
		},
		{
			name:  "複合",
			input: "git status && find /Users/u -name '*.md' 2>/dev/null | head -10",
			want:  []string{"git status ", " find /Users/u -name '*.md' 2>/dev/null ", " head -10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCommands(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("splitCommands(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	home := "/Users/masaru_uchida"
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "チルダのみ",
			token: "~",
			want:  home,
		},
		{
			name:  "チルダ + パス",
			token: "~/Downloads",
			want:  home + "/Downloads",
		},
		{
			name:  "$HOME のみ",
			token: "$HOME",
			want:  home,
		},
		{
			name:  "$HOME + パス",
			token: "$HOME/Downloads",
			want:  home + "/Downloads",
		},
		{
			name:  "展開不要",
			token: "/tmp/foo",
			want:  "/tmp/foo",
		},
		{
			name:  "相対パス",
			token: "./foo",
			want:  "./foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHome(tt.token, home)
			if got != tt.want {
				t.Errorf("expandHome(%q, %q) = %q, want %q", tt.token, home, got, tt.want)
			}
		})
	}
}

func TestParseFindPaths(t *testing.T) {
	home := "/Users/masaru_uchida"
	tests := []struct {
		name   string
		tokens []string
		want   []string
	}{
		{
			name:   "単一パス",
			tokens: []string{"find", "/tmp", "-name", "foo"},
			want:   []string{"/tmp"},
		},
		{
			name:   "複数パス",
			tokens: []string{"find", "/path1", "/path2", "-type", "f"},
			want:   []string{"/path1", "/path2"},
		},
		{
			name:   "チルダ展開",
			tokens: []string{"find", "~", "-maxdepth", "5"},
			want:   []string{home},
		},
		{
			name:   "カレントディレクトリ",
			tokens: []string{"find", ".", "-name", "foo"},
			want:   []string{"."},
		},
		{
			name:   "パスなし (find のみ)",
			tokens: []string{"find"},
			want:   nil,
		},
		{
			name:   "パスなし (即座にオプション)",
			tokens: []string{"find", "-name", "foo"},
			want:   nil,
		},
		{
			name:   "リダイレクトをスキップ",
			tokens: []string{"find", "/tmp", "-name", "foo", "2>/dev/null"},
			want:   []string{"/tmp"},
		},
		{
			name:   "括弧で始まるトークン",
			tokens: []string{"find", "/tmp", "(", "-name", "foo", ")"},
			want:   []string{"/tmp"},
		},
		{
			name:   "否定で始まるトークン",
			tokens: []string{"find", "/tmp", "!", "-name", "foo"},
			want:   []string{"/tmp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFindPaths(tt.tokens, home)
			if !slices.Equal(got, tt.want) {
				t.Errorf("parseFindPaths(%v, %q) = %v, want %v", tt.tokens, home, got, tt.want)
			}
		})
	}
}

func TestExtractScanTargets(t *testing.T) {
	home := "/Users/masaru_uchida"
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "find + ホームディレクトリ",
			command: `find /Users/masaru_uchida -name "*.md"`,
			want:    []string{"/Users/masaru_uchida"},
		},
		{
			name:    "find + チルダ",
			command: "find ~ -maxdepth 5 -type d",
			want:    []string{home},
		},
		{
			name:    "find + $HOME",
			command: `find $HOME -name "*.json"`,
			want:    []string{home},
		},
		{
			name:    "find + 複数パス",
			command: "find /path1 /path2 -type f",
			want:    []string{"/path1", "/path2"},
		},
		{
			name:    "find + カレントディレクトリ",
			command: `find . -name "foo"`,
			want:    []string{"."},
		},
		{
			name:    "find + パイプ",
			command: `find /Users/u -name "*.md" 2>/dev/null | head -10`,
			want:    []string{"/Users/u"},
		},
		{
			name:    "AND チェーン内の find",
			command: `git status && find /Users/u -name "*.md"`,
			want:    []string{"/Users/u"},
		},
		{
			name:    "du + チルダ",
			command: "du -sh ~/Downloads",
			want:    []string{home + "/Downloads"},
		},
		{
			name:    "tree + チルダ",
			command: "tree ~/obsidian",
			want:    []string{home + "/obsidian"},
		},
		{
			name:    "ls -R + チルダ",
			command: "ls -R ~/Downloads",
			want:    []string{home + "/Downloads"},
		},
		{
			name:    "ls (no -R): スキャンコマンドではない",
			command: "ls /tmp",
			want:    nil,
		},
		{
			name:    "git status: スキャンコマンドではない",
			command: "git status",
			want:    nil,
		},
		{
			name:    "echo: スキャンコマンドではない",
			command: "echo hello",
			want:    nil,
		},
		{
			name:    "空コマンド",
			command: "",
			want:    nil,
		},
		{
			name:    "du (パスなし)",
			command: "du -sh",
			want:    nil,
		},
		{
			name:    "tree (パスなし)",
			command: "tree",
			want:    nil,
		},
		{
			name:    "ls -la (再帰なし、オプションのみ)",
			command: "ls -la",
			want:    nil,
		},
		{
			name:    "ls --recursive",
			command: "ls --recursive ~/Downloads",
			want:    []string{home + "/Downloads"},
		},
		{
			name:    "find + サブコマンド付きのパイプチェーン",
			command: "cd /workspace && find /Users/u -type f -name 'SKILL.md' 2>/dev/null | head -10",
			want:    []string{"/Users/u"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractScanTargets(tt.command, home)
			if !slices.Equal(got, tt.want) {
				t.Errorf("extractScanTargets(%q, %q) = %v, want %v", tt.command, home, got, tt.want)
			}
		})
	}
}
