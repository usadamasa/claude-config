#!/usr/bin/env bats
# guard-git-commit-main.sh のテスト
bats_require_minimum_version 1.5.0

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks/guard-git-commit-main.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  export TEST_TMPDIR

  # テスト用 git リポジトリを作成
  git init "$TEST_TMPDIR/repo" >/dev/null 2>&1
  cd "$TEST_TMPDIR/repo" || return
  # CI 環境では user.name/email が未設定のため、リポジトリローカルで設定
  git config user.name "test" >/dev/null 2>&1
  git config user.email "test@test.com" >/dev/null 2>&1
  git commit --allow-empty -m "initial" >/dev/null 2>&1
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
# main ブランチでの git commit → ブロック
# =============================================================================

@test "main ブランチで git commit: ブロック" {
  cd "$TEST_TMPDIR/repo" || return
  git checkout -b main >/dev/null 2>&1 || git checkout main >/dev/null 2>&1

  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit -m 'test'")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
  [[ "$output" == *'main ブランチへの直接コミット'* ]]
}

@test "main ブランチで git commit --amend: ブロック" {
  cd "$TEST_TMPDIR/repo" || return
  git checkout -b main >/dev/null 2>&1 || git checkout main >/dev/null 2>&1

  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit --amend -m 'fix'")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}

@test "master ブランチで git commit: ブロック" {
  cd "$TEST_TMPDIR/repo" || return
  # デフォルトブランチが master の場合 (既に master なら checkout のみ)
  git checkout -b master >/dev/null 2>&1 || git checkout master >/dev/null 2>&1

  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit -m 'test'")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
  [[ "$output" == *'master ブランチへの直接コミット'* ]]
}

# =============================================================================
# feature ブランチでの git commit → 通過
# =============================================================================

@test "feature ブランチで git commit: 通過" {
  cd "$TEST_TMPDIR/repo" || return
  git checkout -b feature/test >/dev/null 2>&1

  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit -m 'add feature'")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "develop ブランチで git commit: 通過" {
  cd "$TEST_TMPDIR/repo" || return
  git checkout -b develop >/dev/null 2>&1

  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit -m 'test'")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# git commit 以外のコマンド → 通過
# =============================================================================

@test "git status: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git status")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git push origin main: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git push origin main")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "git log: 通過" {
  run bash "$SCRIPT_PATH" <<< "$(make_input "git log --oneline")"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# Bash 以外のツール → 通過
# =============================================================================

@test "Bash 以外のツール: 通過" {
  run bash "$SCRIPT_PATH" <<< '{"tool_name":"Read","tool_input":{"file_path":"/tmp/test"}}'

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# パイプ・チェーンを含むコマンド
# =============================================================================

@test "main ブランチで git commit && echo done: ブロック" {
  cd "$TEST_TMPDIR/repo" || return
  git checkout -b main >/dev/null 2>&1 || git checkout main >/dev/null 2>&1

  run bash "$SCRIPT_PATH" <<< "$(make_input "git commit -m 'test' && echo done")"

  [ "$status" -eq 0 ]
  [[ "$output" == *'"permissionDecision":"deny"'* ]]
}
