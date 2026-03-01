#!/usr/bin/env bats
# guard-git-config.sh のテスト
bats_require_minimum_version 1.5.0

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks/guard-git-config.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  export TEST_TMPDIR
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
}

# ヘルパー: PreToolUse の JSON 入力を生成 (jq で安全にエスケープ)
make_input() {
  local command="$1"
  jq -n --arg cmd "$command" '{"tool_name":"Bash","tool_input":{"command":$cmd}}'
}

# =============================================================================
# git config 以外のコマンド → 通過
# =============================================================================

@test "git status: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git status")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git commit -m 'test': 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit -m 'test'")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git push origin main: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git push origin main")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git diff HEAD: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git diff HEAD")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "Bash 以外のツール: 通過" {
  run bash "$SCRIPT_PATH" <<< '{"tool_name":"Read","tool_input":{"file_path":"/tmp/test"}}'

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git 以外の Bash コマンド: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "ls -la")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# 読み取り系 git config → 通過
# =============================================================================

@test "git config --get user.name: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --get user.name")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config --get-all remote.origin.url: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --get-all remote.origin.url")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config --get-regexp 'remote.*': 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --get-regexp 'remote.*'")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config --get-urlmatch http https://example.com: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --get-urlmatch http https://example.com")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config --list: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --list")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config -l: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config -l")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config --show-origin user.name: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --show-origin user.name")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config --show-scope user.name: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --show-scope user.name")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# worktree 例外: remote.origin.fetch → 通過
# =============================================================================

@test "git config remote.origin.fetch '+refs/heads/*:refs/remotes/origin/*': 通過 (worktree 例外)" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config remote.origin.fetch \"+refs/heads/*:refs/remotes/origin/*\"")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config remote.origin.fetch: 通過 (読み取り)" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config remote.origin.fetch")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# 書き込み系 git config → ブロック
# =============================================================================

@test "git config user.name 'foo': ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config user.name \"foo\"")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "git config --global core.editor vim: ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --global core.editor vim")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "git config --unset foo: ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --unset foo")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "git config --unset-all foo: ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --unset-all foo")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "git config --replace-all foo bar: ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --replace-all foo bar")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "git config --add foo bar: ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --add foo bar")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "git config --local user.email 'test@test.com': ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --local user.email \"test@test.com\"")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

# =============================================================================
# git -c (インラインconfig) → 通過
# =============================================================================

@test "git -c sequence.editor='...' rebase: 通過 (インライン config)" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git -c sequence.editor=\"sed -i '' '2,\\\$s/^pick/fixup/'\" rebase -i abc123")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# パイプ・チェーンを含むコマンド
# =============================================================================

@test "git config --get ... が含まれるパイプコマンド: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config --get remote.origin.fetch | grep refs")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git config user.name ... && echo done: ブロック" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git config user.name \"foo\" && echo done")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}
