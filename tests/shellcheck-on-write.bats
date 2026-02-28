#!/usr/bin/env bats
# .claude/hooks/shellcheck-on-write.sh のテスト

HOOK_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/.claude/hooks/shellcheck-on-write.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  export TEST_TMPDIR
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
}

# =============================================================================
# .sh ファイルに対してのみ実行される
# =============================================================================

@test "shellcheck-on-write: .sh ファイルに対して shellcheck が実行される" {
  # lint ツールが利用可能かチェック
  if ! command -v shellcheck &>/dev/null; then
    skip "shellcheck not installed"
  fi

  # テスト用の正しい .sh ファイルを作成
  cat > "$TEST_TMPDIR/good.sh" << 'EOF'
#!/bin/bash
set -euo pipefail
echo "hello"
EOF

  local input
  input=$(cat <<JSONEOF
{"tool_name":"Write","tool_input":{"file_path":"$TEST_TMPDIR/good.sh"}}
JSONEOF
  )

  run bash -c "echo '$input' | bash '$HOOK_PATH'"

  [ "$status" -eq 0 ]
}

# =============================================================================
# .md ファイルに対しては実行されない
# =============================================================================

@test "shellcheck-on-write: .md ファイルに対しては実行されない" {
  local input='{"tool_name":"Write","tool_input":{"file_path":"/tmp/test.md"}}'

  run bash -c "echo '$input' | bash '$HOOK_PATH'"

  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

# =============================================================================
# lint ツール未インストール時はスキップ
# =============================================================================

@test "shellcheck-on-write: shellcheck 未インストール時は exit 0 で終了" {
  # lint ツールを見つけられない PATH を作成 (jq 等は保持)
  local restricted_path
  restricted_path=$(mktemp -d)
  # jq と bash 等の必要コマンドだけリンク
  for cmd in jq bash cat; do
    local cmd_path
    cmd_path=$(command -v "$cmd" 2>/dev/null || true)
    [ -n "$cmd_path" ] && ln -sf "$cmd_path" "$restricted_path/$cmd"
  done

  cat > "$TEST_TMPDIR/test.sh" << 'EOF'
#!/bin/bash
echo "test"
EOF

  local input
  input=$(printf '{"tool_name":"Write","tool_input":{"file_path":"%s/test.sh"}}' "$TEST_TMPDIR")

  run bash -c "printf '%s' '$input' | PATH='$restricted_path' bash '$HOOK_PATH'"

  rm -rf "$restricted_path"
  [ "$status" -eq 0 ]
}

# =============================================================================
# lint エラーでも exit 0
# =============================================================================

@test "shellcheck-on-write: shellcheck エラーがあっても exit 0" {
  if ! command -v shellcheck &>/dev/null; then
    skip "shellcheck not installed"
  fi

  # lint 警告が出るファイルを作成
  cat > "$TEST_TMPDIR/bad.sh" << 'EOF'
#!/bin/bash
readonly INPUT=$(cat)
echo $INPUT
EOF

  local input
  input=$(cat <<JSONEOF
{"tool_name":"Write","tool_input":{"file_path":"$TEST_TMPDIR/bad.sh"}}
JSONEOF
  )

  run bash -c "echo '$input' | bash '$HOOK_PATH' 2>&1"

  [ "$status" -eq 0 ]
}

# =============================================================================
# ファイルが存在しない場合はスキップ
# =============================================================================

@test "shellcheck-on-write: ファイルが存在しない場合は exit 0" {
  local input='{"tool_name":"Write","tool_input":{"file_path":"/nonexistent/path/test.sh"}}'

  run bash -c "echo '$input' | bash '$HOOK_PATH'"

  [ "$status" -eq 0 ]
}
