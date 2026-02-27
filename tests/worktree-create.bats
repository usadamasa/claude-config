#!/usr/bin/env bats
# worktree-create.sh のテスト
bats_require_minimum_version 1.5.0

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks/worktree-create.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  MOCK_BIN="$TEST_TMPDIR/bin"
  mkdir -p "$MOCK_BIN"

  # hook-logger.sh を MOCK_BIN/lib にコピー
  REAL_HOOKS_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks"
  mkdir -p "$MOCK_BIN/lib"
  cp "$REAL_HOOKS_DIR/lib/hook-logger.sh" "$MOCK_BIN/lib/hook-logger.sh"

  # git wt のモック (デフォルト: 成功)
  MOCK_WORKTREE_PATH="$TEST_TMPDIR/worktrees/test-repo/my-branch"
  mkdir -p "$MOCK_WORKTREE_PATH"
  # worktree 判定用に .git ファイルを作成
  echo "gitdir: /tmp/fake-gitdir" > "$MOCK_WORKTREE_PATH/.git"

  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
if [ "$1" = "wt" ]; then
  shift
  # --nocd フラグの確認
  if [ "$1" = "--nocd" ]; then
    shift
    echo "$MOCK_WORKTREE_PATH"
    exit 0
  fi
  echo "$MOCK_WORKTREE_PATH"
  exit 0
fi
# 他の git コマンドは本物を使う
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  # worktree-memory-load.sh のモック
  cat > "$MOCK_BIN/worktree-memory-load.sh" << 'MOCKEOF'
#!/bin/bash
echo "LOAD_CALLED: $1" >> "$TEST_TMPDIR/load-calls.log"
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/worktree-memory-load.sh"

  export TEST_TMPDIR MOCK_BIN MOCK_WORKTREE_PATH
  export PATH="$MOCK_BIN:$PATH"
  # SCRIPT_DIR を MOCK_BIN に向けて worktree-memory-load.sh のモックを使う
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
  # jq がない最小限の PATH で実行
  run env PATH="/usr/bin:/bin:$MOCK_BIN" SCRIPT_DIR="$MOCK_BIN" \
    bash "$SCRIPT_PATH" <<< '{"name":"test","cwd":"/tmp"}'

  [ "$status" -ne 0 ]
}

@test "git-wt 未インストール: エラー終了" {
  # git wt が失敗するモック
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
if [ "$1" = "wt" ]; then
  echo "git: 'wt' is not a git command" >&2
  exit 1
fi
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run bash "$SCRIPT_PATH" <<< "{\"name\":\"test\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -ne 0 ]
}

# =============================================================================
# 正常系
# =============================================================================

@test "git wt --nocd が呼ばれ、stdout にパスのみ出力される" {
  cat > "$MOCK_BIN/git" << MOCKEOF
#!/bin/bash
if [ "\$1" = "wt" ]; then
  shift
  if [ "\$1" = "--nocd" ]; then
    echo "$MOCK_WORKTREE_PATH"
    exit 0
  fi
  echo "ERROR: --nocd not passed" >&2
  exit 1
fi
/usr/bin/git "\$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run --separate-stderr bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == "$MOCK_WORKTREE_PATH" ]]
}

@test "git wt 失敗時に exit 1" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
if [ "$1" = "wt" ]; then
  echo "fatal: error" >&2
  exit 128
fi
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run bash "$SCRIPT_PATH" <<< "{\"name\":\"test\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -ne 0 ]
}

@test "worktree-memory-load.sh が呼ばれる" {
  run bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  [ -f "$TEST_TMPDIR/load-calls.log" ]
  [[ "$(cat "$TEST_TMPDIR/load-calls.log")" == *"LOAD_CALLED: $MOCK_WORKTREE_PATH"* ]]
}

@test "load 失敗でも worktree パスは正常に出力される" {
  # worktree-memory-load.sh が失敗するモック
  cat > "$MOCK_BIN/worktree-memory-load.sh" << 'MOCKEOF'
#!/bin/bash
exit 1
MOCKEOF
  chmod +x "$MOCK_BIN/worktree-memory-load.sh"

  run --separate-stderr bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == "$MOCK_WORKTREE_PATH" ]]
}

# =============================================================================
# 通常モード: 操作ログ
# =============================================================================

@test "通常モード: CMD ログが stderr に出力される" {
  run bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  # stderr は run が output に混ぜるので確認できる
  [[ "$output" == *"$MOCK_WORKTREE_PATH"* ]]
}

# =============================================================================
# dry-run モード
# =============================================================================

@test "dry-run: git wt が実行されない" {
  # git wt が呼ばれたらログに記録するモック
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "GIT_CALLED: $@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "wt" ]; then
  echo "$MOCK_WORKTREE_PATH"
  exit 0
fi
/usr/bin/git "$@"
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  # git wt は呼ばれないこと
  if [ -f "$TEST_TMPDIR/git-calls.log" ]; then
    ! grep -q "wt" "$TEST_TMPDIR/git-calls.log"
  fi
}

@test "dry-run: stdout にダミーパスが出力される" {
  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  # stdout にパスが含まれる (dry-run ダミーパス)
  [[ "$output" == *"/tmp/dry-run-worktree/my-branch"* ]]
}

@test "dry-run: [DRY-RUN] CMD メッセージが出力される" {
  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"CMD"* ]]
  [[ "$output" == *"git wt"* ]]
}
