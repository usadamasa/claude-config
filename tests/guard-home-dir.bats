#!/usr/bin/env bats
# guard-home-dir (Go バイナリ) のテスト
bats_require_minimum_version 1.5.0

BIN_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/bin/guard-home-dir"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  export TEST_TMPDIR
  # テスト用 HOME を設定
  export HOME="$TEST_TMPDIR/home"
  mkdir -p "$HOME"
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
}

# ヘルパー: ファイルツール用の JSON 入力を生成
make_file_input() {
  local tool_name="$1"
  local file_path="$2"
  local cwd="$3"
  local key
  case "$tool_name" in
    Read|Edit|Write) key="file_path" ;;
    NotebookEdit) key="notebook_path" ;;
    Glob|Grep) key="path" ;;
  esac
  jq -n --arg t "$tool_name" --arg f "$file_path" --arg c "$cwd" --arg k "$key" \
    '{tool_name:$t, tool_input:{($k):$f}, cwd:$c}'
}

# ヘルパー: Bash ツール用の JSON 入力を生成
make_bash_input() {
  local command="$1"
  local cwd="$2"
  jq -n --arg cmd "$command" --arg c "$cwd" \
    '{tool_name:"Bash", tool_input:{command:$cmd}, cwd:$c}'
}

# =============================================================================
# ファイルツール: 許可パス → 通過
# =============================================================================

@test "Read: 許可パス (.claude) → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read "$HOME/.claude/settings.json" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Read: 許可パス (src) → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read "$HOME/src/project/main.go" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Read: 許可パス (obsidian) → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read "$HOME/obsidian/note.md" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Read: 許可パス (tmp) → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read "$HOME/tmp/scratch.txt" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Read: 許可パス (workspace) → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read "$HOME/workspace/project" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Read: cwd 配下 → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read /workspace/main.go /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Read: ホーム外 → 通過" {
  run "$BIN_PATH" <<< "$(make_file_input Read /tmp/foo.txt /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# ファイルツール: 禁止パス → deny
# =============================================================================

@test "Read: 禁止パス (Downloads) → deny" {
  run "$BIN_PATH" <<< "$(make_file_input Read "$HOME/Downloads/secret.pdf" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Edit: 禁止パス (.ssh) → deny" {
  run "$BIN_PATH" <<< "$(make_file_input Edit "$HOME/.ssh/id_rsa" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Write: 禁止パス (Documents) → deny" {
  run "$BIN_PATH" <<< "$(make_file_input Write "$HOME/Documents/private.docx" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "NotebookEdit: 禁止パス → deny" {
  run "$BIN_PATH" <<< "$(make_file_input NotebookEdit "$HOME/Desktop/notebook.ipynb" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Glob: 禁止パス → deny" {
  run "$BIN_PATH" <<< "$(make_file_input Glob "$HOME/Downloads" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Grep: 禁止パス → deny" {
  run "$BIN_PATH" <<< "$(make_file_input Grep "$HOME/.aws" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

# =============================================================================
# Glob/Grep: path が空 → 通過 (cwd 使用)
# =============================================================================

@test "Glob: path が空 → 通過" {
  run "$BIN_PATH" <<< '{"tool_name":"Glob","tool_input":{"pattern":"**/*.go"},"cwd":"/workspace"}'

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Grep: path が空 → 通過" {
  run "$BIN_PATH" <<< '{"tool_name":"Grep","tool_input":{"pattern":"TODO"},"cwd":"/workspace"}'

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# Bash: スキャンコマンド → deny
# =============================================================================

@test "Bash: find ~/Downloads → deny" {
  run "$BIN_PATH" <<< "$(make_bash_input "find $HOME/Downloads -name '*.pdf'" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Bash: find ~ → deny (ホームディレクトリ自体)" {
  run "$BIN_PATH" <<< "$(make_bash_input "find $HOME -type f -name '*.md'" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Bash: du ~/Downloads → deny" {
  run "$BIN_PATH" <<< "$(make_bash_input "du -sh $HOME/Downloads" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Bash: tree ~/Desktop → deny" {
  run "$BIN_PATH" <<< "$(make_bash_input "tree $HOME/Desktop" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Bash: ls -R ~/Downloads → deny" {
  run "$BIN_PATH" <<< "$(make_bash_input "ls -R $HOME/Downloads" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "Bash: find + パイプ → deny" {
  run "$BIN_PATH" <<< "$(make_bash_input "find $HOME -name '*.md' 2>/dev/null | head -10" /workspace)"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

# =============================================================================
# Bash: 許可パスのスキャン → 通過
# =============================================================================

@test "Bash: find ~/src → 通過 (許可サブディレクトリ)" {
  run "$BIN_PATH" <<< "$(make_bash_input "find $HOME/src -name '*.go'" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Bash: find /tmp → 通過 (ホーム外)" {
  run "$BIN_PATH" <<< "$(make_bash_input "find /tmp -name '*.txt'" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# Bash: スキャンコマンドでない → 通過
# =============================================================================

@test "Bash: git status → 通過" {
  run "$BIN_PATH" <<< "$(make_bash_input "git status" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Bash: echo hello → 通過" {
  run "$BIN_PATH" <<< "$(make_bash_input "echo hello" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Bash: ls /tmp (再帰なし) → 通過" {
  run "$BIN_PATH" <<< "$(make_bash_input "ls /tmp" /workspace)"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# 未知のツール → 通過
# =============================================================================

@test "Agent: 通過" {
  run "$BIN_PATH" <<< '{"tool_name":"Agent","tool_input":{},"cwd":"/workspace"}'

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "WebFetch: 通過" {
  run "$BIN_PATH" <<< '{"tool_name":"WebFetch","tool_input":{},"cwd":"/workspace"}'

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}
