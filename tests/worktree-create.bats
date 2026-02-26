#!/usr/bin/env bats
# worktree-create.sh のテスト

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/hooks/worktree-create.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  MOCK_BIN="$TEST_TMPDIR/bin"
  mkdir -p "$MOCK_BIN"

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

  run bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

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

  run bash "$SCRIPT_PATH" <<< "{\"name\":\"my-branch\",\"cwd\":\"$TEST_TMPDIR\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == "$MOCK_WORKTREE_PATH" ]]
}
