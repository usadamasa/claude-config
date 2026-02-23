#!/usr/bin/env bats
# migrate-worktree-memory.sh のテスト

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/hooks/migrate-worktree-memory.sh"

setup() {
  load 'fixtures/worktree-setup.sh'
  create_worktree_memory_env
}

teardown() {
  cleanup_worktree_memory_env
}

# =============================================================================
# 通常の worktree 動作
# =============================================================================

@test "通常の worktree: SESSION_HANDOFF.md を親 memory に SESSION_HANDOFF_feature.md としてコピー" {
  echo "# SESSION HANDOFF" > "$WORKTREE_MEM/SESSION_HANDOFF.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -f "$PARENT_MEM/SESSION_HANDOFF_feature.md" ]
  [[ "$(cat "$PARENT_MEM/SESSION_HANDOFF_feature.md")" == *"SESSION HANDOFF"* ]]
}

@test "worktree 削除時: 親の SESSION_HANDOFF.md は上書きされない" {
  mkdir -p "$PARENT_MEM"
  echo "# 親の元の HANDOFF" > "$PARENT_MEM/SESSION_HANDOFF.md"
  echo "# worktree の HANDOFF" > "$WORKTREE_MEM/SESSION_HANDOFF.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$(cat "$PARENT_MEM/SESSION_HANDOFF.md")" == *"親の元の HANDOFF"* ]]
  [ -f "$PARENT_MEM/SESSION_HANDOFF_feature.md" ]
}

@test "通常の worktree: MEMORY.md を親 memory にコピー" {
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -f "$PARENT_MEM/MEMORY.md" ]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"ワークツリー MEMORY"* ]]
}

# =============================================================================
# MEMORY.md マージ戦略
# =============================================================================

@test "worktree: 親 MEMORY.md が存在しない場合はコピー" {
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"
  # PARENT_MEM ディレクトリは未作成

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -f "$PARENT_MEM/MEMORY.md" ]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"ワークツリー MEMORY"* ]]
}

@test "worktree: 親 MEMORY.md が存在する場合は末尾に追記" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"既存の親 MEMORY"* ]]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"ワークツリー MEMORY"* ]]
}

@test "追記フォーマット: ## [Merged from worktree: {branch}] ヘッダーが付く" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"## [Merged from worktree: feature]"* ]]
}

# =============================================================================
# エラーケース / スキップ
# =============================================================================

@test "通常 repo (.git がディレクトリ): スクリプトが何もせず正常終了" {
  local NORMAL_REPO
  NORMAL_REPO=$(mktemp -d)
  mkdir -p "$NORMAL_REPO/.git"  # .git はディレクトリ

  run bash "$SCRIPT_PATH" "$NORMAL_REPO"

  [ "$status" -eq 0 ]
  rm -rf "$NORMAL_REPO"
}

@test "worktree: memory ファイルが存在しない場合はスキップ" {
  # WORKTREE_MEM ディレクトリは存在するが中身は空

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ ! -f "$PARENT_MEM/MEMORY.md" ]
  [ ! -f "$PARENT_MEM/SESSION_HANDOFF.md" ]
}

@test "worktree: 親 memory ディレクトリが存在しない場合は自動作成" {
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"
  # PARENT_MEM は未作成

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -d "$PARENT_MEM" ]
}

@test "引数なし: 正常終了" {
  run bash "$SCRIPT_PATH"

  [ "$status" -eq 0 ]
}
