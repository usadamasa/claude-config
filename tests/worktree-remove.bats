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
  cp "$REAL_HOOKS_DIR/lib/sync-main-repo.sh" "$MOCK_BIN/lib/sync-main-repo.sh"

  # テスト用 worktree パス
  MOCK_WORKTREE_PATH="$TEST_TMPDIR/worktrees/test-repo/my-branch"
  mkdir -p "$MOCK_WORKTREE_PATH"

  # resolve_main_repo 用の .git ファイル構造
  TEST_PARENT_REPO="$TEST_TMPDIR/parent-repo"
  mkdir -p "$TEST_PARENT_REPO/.git/worktrees/my-branch"
  echo "../.." > "$TEST_PARENT_REPO/.git/worktrees/my-branch/commondir"
  echo "ref: refs/heads/my-branch" > "$TEST_PARENT_REPO/.git/worktrees/my-branch/HEAD"
  echo "gitdir: $TEST_PARENT_REPO/.git/worktrees/my-branch" > "$MOCK_WORKTREE_PATH/.git"

  export TEST_PARENT_REPO

  # git のモック (デフォルト: worktree remove 成功、sync 関連も処理)
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
# -C 付きの sync 系コマンド
if [ "$1" = "-C" ]; then
  shift 2
  case "$1" in
    config)
      if [ "$2" = "--get" ]; then
        echo "+refs/heads/*:refs/remotes/origin/*"
        exit 0
      fi
      ;;
    fetch) exit 0 ;;
    show-ref)
      if [[ "$*" == *"refs/heads/main"* ]]; then exit 0; fi
      exit 1
      ;;
    symbolic-ref) echo "main"; exit 0 ;;
    merge) exit 0 ;;
  esac
  exit 0
fi
exit 0
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
  # 最初の remove は失敗、--force 付きで成功 + sync 系コマンドも処理
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
if [ "$1" = "-C" ]; then
  shift 2
  case "$1" in
    config)
      if [ "$2" = "--get" ]; then
        echo "+refs/heads/*:refs/remotes/origin/*"
        exit 0
      fi
      ;;
    fetch) exit 0 ;;
    show-ref)
      if [[ "$*" == *"refs/heads/main"* ]]; then exit 0; fi
      exit 1
      ;;
    symbolic-ref) echo "main"; exit 0 ;;
    merge) exit 0 ;;
  esac
  exit 0
fi
exit 0
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

# =============================================================================
# sync 連携
# =============================================================================

@test "sync が worktree 削除後に実行される" {
  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  # fetch が呼ばれた (sync が実行された)
  grep -q "fetch origin" "$TEST_TMPDIR/git-calls.log"
  # worktree remove が fetch より先に呼ばれたことを確認
  local remove_line fetch_line
  remove_line=$(grep -n "worktree remove" "$TEST_TMPDIR/git-calls.log" | head -1 | cut -d: -f1)
  fetch_line=$(grep -n "fetch origin" "$TEST_TMPDIR/git-calls.log" | head -1 | cut -d: -f1)
  [ "$remove_line" -lt "$fetch_line" ]
}

@test "sync 失敗時にフック全体が失敗する" {
  # fetch が失敗するモック
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "worktree" ]; then
  if [ "$2" = "remove" ]; then exit 0; fi
  if [ "$2" = "prune" ]; then exit 0; fi
fi
if [ "$1" = "-C" ]; then
  shift 2
  case "$1" in
    config)
      if [ "$2" = "--get" ]; then
        echo "+refs/heads/*:refs/remotes/origin/*"
        exit 0
      fi
      ;;
    fetch) exit 1 ;;
  esac
  exit 0
fi
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -ne 0 ]
}

@test "worktree が通常リポジトリの場合 sync はスキップされる" {
  # .git ファイルを削除して .git ディレクトリに変更 (通常リポジトリ)
  rm "$MOCK_WORKTREE_PATH/.git"
  mkdir -p "$MOCK_WORKTREE_PATH/.git"

  run bash "$SCRIPT_PATH" <<< "{\"worktree_path\":\"$MOCK_WORKTREE_PATH\"}"

  [ "$status" -eq 0 ]
  # fetch は呼ばれない (sync スキップ)
  run ! grep -q "fetch origin" "$TEST_TMPDIR/git-calls.log"
}
