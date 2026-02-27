#!/usr/bin/env bats
# worktree-remove.sh のテスト
bats_require_minimum_version 1.5.0

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks/worktree-remove.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  MOCK_BIN="$TEST_TMPDIR/bin"
  mkdir -p "$MOCK_BIN"

  # hook-logger.sh を MOCK_BIN/lib にコピー
  REAL_HOOKS_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks"
  mkdir -p "$MOCK_BIN/lib"
  cp "$REAL_HOOKS_DIR/lib/hook-logger.sh" "$MOCK_BIN/lib/hook-logger.sh"

  # テスト用 worktree パス
  MOCK_WORKTREE_PATH="$TEST_TMPDIR/worktrees/test-repo/my-branch"
  mkdir -p "$MOCK_WORKTREE_PATH"

  # git のモック (デフォルト: worktree remove 成功)
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "worktree" ]; then
  if [ "$2" = "remove" ]; then
    exit 0
  fi
  if [ "$2" = "prune" ]; then
    exit 0
  fi
fi
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  # worktree-memory-save.sh のモック
  cat > "$MOCK_BIN/worktree-memory-save.sh" << 'MOCKEOF'
#!/bin/bash
echo "SAVE_CALLED: $1" >> "$TEST_TMPDIR/save-calls.log"
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/worktree-memory-save.sh"

  export TEST_TMPDIR MOCK_BIN MOCK_WORKTREE_PATH
  export PATH="$MOCK_BIN:$PATH"
  export SCRIPT_DIR="$MOCK_BIN"
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
  unset SCRIPT_DIR
}

# =============================================================================
# 前提条件チェック
# =============================================================================

@test "jq 未インストール: エラー終了" {
  run env PATH="/usr/bin:/bin:$MOCK_BIN" SCRIPT_DIR="$MOCK_BIN" \
    bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -ne 0 ]
}

# =============================================================================
# 正常系
# =============================================================================

@test "worktree-memory-save.sh が worktree_path で呼ばれる" {
  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/save-calls.log" ]
  [[ "$(cat "$TEST_TMPDIR/save-calls.log")" == *"SAVE_CALLED: $MOCK_WORKTREE_PATH"* ]]
}

@test "git worktree remove が呼ばれる" {
  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/git-calls.log" ]
  [[ "$(cat "$TEST_TMPDIR/git-calls.log")" == *"worktree remove $MOCK_WORKTREE_PATH"* ]]
}

@test "save 失敗でも git worktree remove は実行される" {
  # save が失敗するモック
  cat > "$MOCK_BIN/worktree-memory-save.sh" << 'MOCKEOF'
#!/bin/bash
exit 1
MOCKEOF
  chmod +x "$MOCK_BIN/worktree-memory-save.sh"

  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/git-calls.log" ]
  [[ "$(cat "$TEST_TMPDIR/git-calls.log")" == *"worktree remove $MOCK_WORKTREE_PATH"* ]]
}

@test "worktree remove 失敗時は --force でリトライされる" {
  # 最初の remove は失敗、--force 付きで成功
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "worktree" ]; then
  if [ "$2" = "remove" ]; then
    if [[ "$*" == *"--force"* ]]; then
      exit 0
    fi
    echo "fatal: cannot remove" >&2
    exit 1
  fi
  if [ "$2" = "prune" ]; then
    exit 0
  fi
fi
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  # --force でリトライされたことを確認
  [[ "$(cat "$TEST_TMPDIR/git-calls.log")" == *"--force"* ]]
  # prune も呼ばれる
  [[ "$(cat "$TEST_TMPDIR/git-calls.log")" == *"worktree prune"* ]]
}

# =============================================================================
# 通常モード: 操作ログ
# =============================================================================

@test "通常モード: git worktree remove の CMD ログが出力される" {
  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[worktree-remove]"* ]]
  [[ "$output" == *"CMD"* ]]
  [[ "$output" == *"worktree remove"* ]]
}

# =============================================================================
# dry-run モード
# =============================================================================

@test "dry-run: git worktree remove が実行されない" {
  # git の呼び出しをログに記録するモック
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "GIT_CALLED: $@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "worktree" ]; then
  exit 0
fi
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  # git worktree remove は呼ばれないこと
  if [ -f "$TEST_TMPDIR/git-calls.log" ]; then
    ! grep -q "worktree remove" "$TEST_TMPDIR/git-calls.log"
  fi
}

@test "dry-run: [DRY-RUN] CMD メッセージが出力される" {
  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"CMD"* ]]
  [[ "$output" == *"worktree remove"* ]]
}

@test "dry-run: DRY_RUN が worktree-memory-save.sh に伝播する" {
  # save モックで DRY_RUN を記録
  cat > "$MOCK_BIN/worktree-memory-save.sh" << 'MOCKEOF'
#!/bin/bash
echo "SAVE_DRY_RUN=$DRY_RUN" >> "$TEST_TMPDIR/save-env.log"
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/worktree-memory-save.sh"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/save-env.log" ]
  [[ "$(cat "$TEST_TMPDIR/save-env.log")" == *"SAVE_DRY_RUN=1"* ]]
}
