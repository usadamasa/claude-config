#!/usr/bin/env bats
# sync-main-repo.sh のテスト
bats_require_minimum_version 1.5.0

HOOKS_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  MOCK_BIN="$TEST_TMPDIR/bin"
  mkdir -p "$MOCK_BIN"

  # hook-logger.sh を MOCK_BIN/lib にコピー
  mkdir -p "$MOCK_BIN/lib"
  cp "$HOOKS_DIR/lib/hook-logger.sh" "$MOCK_BIN/lib/hook-logger.sh"
  cp "$HOOKS_DIR/lib/sync-main-repo.sh" "$MOCK_BIN/lib/sync-main-repo.sh"

  # テスト用 worktree 構造を作成
  TEST_PARENT_REPO="$TEST_TMPDIR/parent-repo"
  TEST_WORKTREE="$TEST_TMPDIR/my-worktree"
  mkdir -p "$TEST_PARENT_REPO/.git/worktrees/feature"
  mkdir -p "$TEST_WORKTREE"

  # commondir: worktree gitdir から親 .git への相対パス
  echo "../.." > "$TEST_PARENT_REPO/.git/worktrees/feature/commondir"
  # HEAD: ブランチ名の特定に使用
  echo "ref: refs/heads/feature" > "$TEST_PARENT_REPO/.git/worktrees/feature/HEAD"
  # ワークツリーの .git ファイル (worktree 判定のキー)
  echo "gitdir: $TEST_PARENT_REPO/.git/worktrees/feature" > "$TEST_WORKTREE/.git"

  export TEST_TMPDIR MOCK_BIN TEST_PARENT_REPO TEST_WORKTREE
  export SCRIPT_DIR="$MOCK_BIN"
  export HOOK_NAME="test"
  export DRY_RUN=0
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
  unset SCRIPT_DIR HOOK_NAME DRY_RUN
}

# ヘルパー: sync-main-repo.sh をソースして関数を呼ぶラッパー
_run_resolve() {
  # shellcheck disable=SC1091
  source "$MOCK_BIN/lib/hook-logger.sh"
  # shellcheck disable=SC1091
  source "$MOCK_BIN/lib/sync-main-repo.sh"
  resolve_main_repo "$@"
}

_run_sync() {
  export PATH="$MOCK_BIN:$PATH"
  # shellcheck disable=SC1091
  source "$MOCK_BIN/lib/hook-logger.sh"
  # shellcheck disable=SC1091
  source "$MOCK_BIN/lib/sync-main-repo.sh"
  sync_main_repo "$@"
}

# =============================================================================
# resolve_main_repo
# =============================================================================

@test "resolve_main_repo: 正常系 - 親リポジトリルートを返す" {
  run _run_resolve "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == "$TEST_PARENT_REPO" ]]
}

@test "resolve_main_repo: 通常リポジトリ (.git がディレクトリ) は return 1" {
  local normal_repo="$TEST_TMPDIR/normal-repo"
  mkdir -p "$normal_repo/.git"

  run _run_resolve "$normal_repo"

  [ "$status" -eq 1 ]
}

@test "resolve_main_repo: .git ファイル未存在は return 1" {
  local no_git="$TEST_TMPDIR/no-git"
  mkdir -p "$no_git"

  run _run_resolve "$no_git"

  [ "$status" -eq 1 ]
}

@test "resolve_main_repo: commondir 未存在は return 1" {
  rm "$TEST_PARENT_REPO/.git/worktrees/feature/commondir"

  run _run_resolve "$TEST_WORKTREE"

  [ "$status" -eq 1 ]
}

@test "resolve_main_repo: gitdir が存在しないディレクトリを指す場合は return 1" {
  echo "gitdir: /nonexistent/path" > "$TEST_WORKTREE/.git"

  run _run_resolve "$TEST_WORKTREE"

  [ "$status" -eq 1 ]
}

# =============================================================================
# sync_main_repo: 正常系
# =============================================================================

@test "sync_main_repo: fetch + merge 正常系" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  # fetch が呼ばれた
  grep -q "fetch origin" "$TEST_TMPDIR/git-calls.log"
  # merge が呼ばれた
  grep -q "merge --ff-only origin/main" "$TEST_TMPDIR/git-calls.log"
}

@test "sync_main_repo: fetch 失敗時は return 1" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 1 ]
  # merge は呼ばれない
  run ! grep -q "merge" "$TEST_TMPDIR/git-calls.log"
}

@test "sync_main_repo: merge 失敗 (diverged) は return 1" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
  merge) exit 1 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 1 ]
  [[ "$output" == *"merge --ff-only failed"* ]]
}

@test "sync_main_repo: デフォルトブランチ不在 (main/master どちらもない) は return 0" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
case "$1" in
  config)
    if [ "$2" = "--get" ]; then
      echo "+refs/heads/*:refs/remotes/origin/*"
      exit 0
    fi
    ;;
  fetch) exit 0 ;;
  show-ref) exit 1 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  [[ "$output" == *"no default branch found"* ]]
}

@test "sync_main_repo: feature ブランチ中は merge スキップ" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
  symbolic-ref) echo "feature-branch"; exit 0 ;;
  merge) echo "MERGE_SHOULD_NOT_BE_CALLED"; exit 1 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  [[ "$output" == *"not on default branch"* ]]
  run ! grep -q "merge" "$TEST_TMPDIR/git-calls.log"
}

@test "sync_main_repo: detached HEAD は merge スキップ" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
  symbolic-ref) exit 1 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  [[ "$output" == *"detached HEAD"* ]]
}

@test "sync_main_repo: master ブランチ検出" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
case "$1" in
  config)
    if [ "$2" = "--get" ]; then
      echo "+refs/heads/*:refs/remotes/origin/*"
      exit 0
    fi
    ;;
  fetch) exit 0 ;;
  show-ref)
    if [[ "$*" == *"refs/heads/main"* ]]; then exit 1; fi
    if [[ "$*" == *"refs/heads/master"* ]]; then exit 0; fi
    exit 1
    ;;
  symbolic-ref) echo "master"; exit 0 ;;
  merge) exit 0 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  grep -q "merge --ff-only origin/master" "$TEST_TMPDIR/git-calls.log"
}

@test "sync_main_repo: DRY_RUN モード" {
  export DRY_RUN=1

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"sync: CMD"* ]]
  # 実際の git コマンドは呼ばれない (MOCK_BIN/git がないので)
}

@test "sync_main_repo: 無効パスは return 1" {
  run _run_sync "/nonexistent/path"

  [ "$status" -eq 1 ]
  [[ "$output" == *"invalid path"* ]]
}

@test "sync_main_repo: remote.origin.fetch 未設定時に自動設定" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
case "$1" in
  config)
    if [ "$2" = "--get" ]; then
      # 空を返す (未設定)
      exit 1
    fi
    # config set は成功
    exit 0
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
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  # config remote.origin.fetch が設定された
  grep -q 'config remote.origin.fetch' "$TEST_TMPDIR/git-calls.log"
}

# =============================================================================
# resolve_main_repo: bare リポジトリ
# =============================================================================

@test "resolve_main_repo: bare リポジトリの worktree からルートを解決" {
  # bare リポジトリ構造 (ルートが .git で終わらない)
  BARE_REPO="$TEST_TMPDIR/bare-repo.git"
  BARE_WORKTREE="$TEST_TMPDIR/bare-wt"
  mkdir -p "$BARE_REPO/worktrees/feature"
  mkdir -p "$BARE_WORKTREE"

  echo "../.." > "$BARE_REPO/worktrees/feature/commondir"
  echo "ref: refs/heads/feature" > "$BARE_REPO/worktrees/feature/HEAD"
  echo "gitdir: $BARE_REPO/worktrees/feature" > "$BARE_WORKTREE/.git"

  run _run_resolve "$BARE_WORKTREE"

  [ "$status" -eq 0 ]
  # bare リポジトリでは commondir 自体がリポジトリルート
  [[ "$output" == "$BARE_REPO" ]]
}

# =============================================================================
# sync_main_repo: bare リポジトリ
# =============================================================================

@test "sync_main_repo: bare リポジトリで update-ref が使われる" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
  rev-parse)
    if [ "$2" = "--is-bare-repository" ]; then
      echo "true"
      exit 0
    fi
    # rev-parse refs/heads/main
    if [[ "$*" == *"refs/heads/main"* ]]; then
      echo "abc123"
      exit 0
    fi
    # rev-parse refs/remotes/origin/main
    if [[ "$*" == *"refs/remotes/origin/main"* ]]; then
      echo "def456"
      exit 0
    fi
    ;;
  merge-base) exit 0 ;;
  update-ref) exit 0 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  # update-ref が呼ばれた (merge --ff-only ではなく)
  grep -q "update-ref refs/heads/main def456" "$TEST_TMPDIR/git-calls.log"
  run ! grep -q "merge --ff-only" "$TEST_TMPDIR/git-calls.log"
}

@test "sync_main_repo: bare リポジトリで diverged は return 1" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
  rev-parse)
    if [ "$2" = "--is-bare-repository" ]; then
      echo "true"
      exit 0
    fi
    if [[ "$*" == *"refs/heads/main"* ]]; then echo "abc123"; exit 0; fi
    if [[ "$*" == *"refs/remotes/origin/main"* ]]; then echo "def456"; exit 0; fi
    ;;
  merge-base) exit 1 ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 1 ]
  [[ "$output" == *"update-ref failed (diverged?)"* ]]
}

@test "sync_main_repo: bare リポジトリで already up to date" {
  cat > "$MOCK_BIN/git" << 'MOCKEOF'
#!/bin/bash
echo "$@" >> "$TEST_TMPDIR/git-calls.log"
if [ "$1" = "-C" ]; then shift 2; fi
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
  rev-parse)
    if [ "$2" = "--is-bare-repository" ]; then
      echo "true"
      exit 0
    fi
    # local と remote が同じ SHA
    echo "abc123"
    exit 0
    ;;
esac
exit 0
MOCKEOF
  chmod +x "$MOCK_BIN/git"

  run _run_sync "$TEST_PARENT_REPO"

  [ "$status" -eq 0 ]
  [[ "$output" == *"already up to date"* ]]
  run ! grep -q "update-ref" "$TEST_TMPDIR/git-calls.log"
}
