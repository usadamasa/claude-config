package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		pinnedModel string
		stripFields []string
		want        string
		wantWarns   []string
	}{
		{
			name: "配列をアルファベット順にソート",
			input: `{
  "permissions": {
    "allow": [
      "Bash(git:*)",
      "Bash(docker:*)",
      "Bash(aws:*)"
    ],
    "deny": [
      "Bash(wget:*)",
      "Bash(curl:*)"
    ],
    "ask": [
      "Bash(rm -rf:*)",
      "Bash(brew install:*)"
    ]
  },
  "model": "claude-opus-4-6"
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": [
      "Bash(aws:*)",
      "Bash(docker:*)",
      "Bash(git:*)"
    ],
    "ask": [
      "Bash(brew install:*)",
      "Bash(rm -rf:*)"
    ],
    "deny": [
      "Bash(curl:*)",
      "Bash(wget:*)"
    ]
  }
}
`,
		},
		{
			name: "ランタイムフィールドを除去",
			input: `{
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": []
  },
  "effortLevel": "high",
  "teammateMode": "auto"
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: []string{"effortLevel", "teammateMode"},
			want: `{
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": []
  }
}
`,
		},
		{
			name: "sandbox.network.allowedDomains をソート",
			input: `{
  "model": "claude-opus-4-6",
  "sandbox": {
    "enabled": false,
    "network": {
      "allowedDomains": [
        "pkg.go.dev",
        "api.github.com",
        "*.anthropic.com"
      ]
    }
  }
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "model": "claude-opus-4-6",
  "sandbox": {
    "enabled": false,
    "network": {
      "allowedDomains": [
        "*.anthropic.com",
        "api.github.com",
        "pkg.go.dev"
      ]
    }
  }
}
`,
		},
		{
			name: "enabledPlugins をキー順ソート",
			input: `{
  "model": "claude-opus-4-6",
  "enabledPlugins": {
    "superpowers@claude-plugins-official": true,
    "atlassian@claude-plugins-official": true,
    "code-review@claude-plugins-official": true
  }
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "enabledPlugins": {
    "atlassian@claude-plugins-official": true,
    "code-review@claude-plugins-official": true,
    "superpowers@claude-plugins-official": true
  },
  "model": "claude-opus-4-6"
}
`,
		},
		{
			name: "model 不一致で警告",
			input: `{
  "model": "claude-sonnet-4-5-20250514",
  "permissions": {
    "allow": []
  }
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "model": "claude-sonnet-4-5-20250514",
  "permissions": {
    "allow": []
  }
}
`,
			wantWarns: []string{"model が期待値と不一致"},
		},
		{
			name: "未知のフィールドを保持",
			input: `{
  "model": "claude-opus-4-6",
  "customField": "preserved",
  "permissions": {
    "allow": []
  }
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "customField": "preserved",
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": []
  }
}
`,
		},
		{
			name: "空の stripFields は何も除去しない",
			input: `{
  "model": "claude-opus-4-6",
  "effortLevel": "high"
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "effortLevel": "high",
  "model": "claude-opus-4-6"
}
`,
		},
		{
			name: "sandbox の他のフィールドを保持",
			input: `{
  "model": "claude-opus-4-6",
  "sandbox": {
    "enabled": false,
    "autoAllowBashIfSandboxed": true,
    "network": {
      "allowedDomains": [
        "b.com",
        "a.com"
      ],
      "allowLocalBinding": true
    },
    "excludedCommands": [
      "docker",
      "git"
    ]
  }
}`,
			pinnedModel: "claude-opus-4-6",
			stripFields: nil,
			want: `{
  "model": "claude-opus-4-6",
  "sandbox": {
    "autoAllowBashIfSandboxed": true,
    "enabled": false,
    "excludedCommands": [
      "docker",
      "git"
    ],
    "network": {
      "allowLocalBinding": true,
      "allowedDomains": [
        "a.com",
        "b.com"
      ]
    }
  }
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warns, err := Normalize([]byte(tt.input), tt.pinnedModel, tt.stripFields)
			if err != nil {
				t.Fatalf("Normalize() error: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Normalize() output mismatch\ngot:\n%s\nwant:\n%s", string(got), tt.want)
			}

			if len(tt.wantWarns) != len(warns) {
				t.Errorf("warnings count: got %d, want %d", len(warns), len(tt.wantWarns))
			}
			for i, w := range tt.wantWarns {
				if i < len(warns) && warns[i] != w {
					t.Errorf("warning[%d]: got %q, want %q", i, warns[i], w)
				}
			}
		})
	}
}

func TestNormalizeInvalidJSON(t *testing.T) {
	_, _, err := Normalize([]byte(`{invalid}`), "claude-opus-4-6", nil)
	if err == nil {
		t.Error("Normalize() expected error for invalid JSON, got nil")
	}
}

func TestNormalizeFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	input := `{
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": [
      "Bash(git:*)",
      "Bash(docker:*)"
    ]
  },
  "effortLevel": "high"
}`
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	changed, warns, err := NormalizeFile(path, "claude-opus-4-6", []string{"effortLevel"})
	if err != nil {
		t.Fatalf("NormalizeFile() error: %v", err)
	}
	if !changed {
		t.Error("NormalizeFile() expected changed=true")
	}
	if len(warns) != 0 {
		t.Errorf("NormalizeFile() unexpected warnings: %v", warns)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": [
      "Bash(docker:*)",
      "Bash(git:*)"
    ]
  }
}
`
	if string(got) != want {
		t.Errorf("file content mismatch\ngot:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestNormalizeFileNoChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	input := `{
  "model": "claude-opus-4-6",
  "permissions": {
    "allow": [
      "Bash(docker:*)",
      "Bash(git:*)"
    ]
  }
}
`
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	changed, _, err := NormalizeFile(path, "claude-opus-4-6", nil)
	if err != nil {
		t.Fatalf("NormalizeFile() error: %v", err)
	}
	if changed {
		t.Error("NormalizeFile() expected changed=false for already normalized file")
	}
}

func TestNormalizePreservesJSONStructure(t *testing.T) {
	input := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "guard.sh"
          }
        ]
      }
    ]
  },
  "model": "claude-opus-4-6",
  "statusLine": {
    "type": "command",
    "command": "npx ccstatusline"
  }
}`

	got, _, err := Normalize([]byte(input), "claude-opus-4-6", nil)
	if err != nil {
		t.Fatalf("Normalize() error: %v", err)
	}

	// hooks と statusLine が構造的に保持されていることを検証
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := parsed["hooks"]; !ok {
		t.Error("hooks field missing from output")
	}
	if _, ok := parsed["statusLine"]; !ok {
		t.Error("statusLine field missing from output")
	}
}
