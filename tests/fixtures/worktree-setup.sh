#!/usr/bin/env bash
# worktree メモリ移行スクリプトのテスト用フィクスチャ

# パスエンコード: Claude Code の auto-memory パス命名規則に合わせる
# / . _ を - に変換 (先頭 / も - になる)
_encode_claude_path() {
  echo "$1" | tr '/._' '-'
}

# テスト用 worktree 環境を作成
# 変数セット: TEST_HOME, TEST_PARENT_REPO, TEST_WORKTREE, WORKTREE_MEM, PARENT_MEM
create_worktree_memory_env() {
  TEST_HOME=$(mktemp -d)
  TEST_PARENT_REPO=$(mktemp -d)
  TEST_WORKTREE=$(mktemp -d)

  # 親リポジトリの .git 構造を作成
  mkdir -p "$TEST_PARENT_REPO/.git/worktrees/feature"

  # commondir: worktree gitdir から親 .git への相対パス
  echo "../.." > "$TEST_PARENT_REPO/.git/worktrees/feature/commondir"

  # HEAD: ブランチ名の特定に使用
  echo "ref: refs/heads/feature" > "$TEST_PARENT_REPO/.git/worktrees/feature/HEAD"

  # ワークツリーの .git ファイル (worktree 判定のキー)
  echo "gitdir: $TEST_PARENT_REPO/.git/worktrees/feature" > "$TEST_WORKTREE/.git"

  # HOME をモック (スクリプトが書き込む先を tempdir にリダイレクト)
  export HOME="$TEST_HOME"

  # ワークツリーの memory パス
  local worktree_enc
  worktree_enc=$(_encode_claude_path "$TEST_WORKTREE")
  WORKTREE_MEM="$TEST_HOME/.claude/projects/$worktree_enc/memory"
  mkdir -p "$WORKTREE_MEM"

  # 親リポジトリの memory パス (テストケースによって存在有無が異なる)
  local parent_enc
  parent_enc=$(_encode_claude_path "$TEST_PARENT_REPO")
  PARENT_MEM="$TEST_HOME/.claude/projects/$parent_enc/memory"

  export TEST_HOME TEST_PARENT_REPO TEST_WORKTREE WORKTREE_MEM PARENT_MEM
}

# テスト環境のクリーンアップ
cleanup_worktree_memory_env() {
  [ -n "$TEST_HOME" ] && rm -rf "$TEST_HOME"
  [ -n "$TEST_PARENT_REPO" ] && rm -rf "$TEST_PARENT_REPO"
  [ -n "$TEST_WORKTREE" ] && rm -rf "$TEST_WORKTREE"
  unset TEST_HOME TEST_PARENT_REPO TEST_WORKTREE WORKTREE_MEM PARENT_MEM
}
