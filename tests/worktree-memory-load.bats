#!/usr/bin/env bats
# worktree-memory-load.sh のテスト

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/hooks/worktree-memory-load.sh"

setup() {
  load 'fixtures/worktree-setup.sh'
  create_worktree_memory_env
}

teardown() {
  cleanup_worktree_memory_env
}

# =============================================================================
# 正常系: メモリロード
# =============================================================================

@test "親 MEMORY.md が存在: worktree memory にコピーされる" {
  mkdir -p "$PARENT_MEM"
  echo "# 親の MEMORY" > "$PARENT_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -f "$WORKTREE_MEM/MEMORY.md" ]
  [[ "$(cat "$WORKTREE_MEM/MEMORY.md")" == *"親の MEMORY"* ]]
}

@test "親に複数ファイル存在: 全てコピーされる" {
  mkdir -p "$PARENT_MEM"
  echo "# MEMORY" > "$PARENT_MEM/MEMORY.md"
  echo "# DEBUG" > "$PARENT_MEM/debugging.md"
  echo "# PATTERNS" > "$PARENT_MEM/patterns.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -f "$WORKTREE_MEM/MEMORY.md" ]
  [ -f "$WORKTREE_MEM/debugging.md" ]
  [ -f "$WORKTREE_MEM/patterns.md" ]
}

@test "worktree memory ディレクトリ未作成: 自動作成される" {
  # cleanup the pre-created worktree memory dir
  rm -rf "$WORKTREE_MEM"

  mkdir -p "$PARENT_MEM"
  echo "# MEMORY" > "$PARENT_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -d "$WORKTREE_MEM" ]
  [ -f "$WORKTREE_MEM/MEMORY.md" ]
}

# =============================================================================
# スキップケース
# =============================================================================

@test "引数なし: 正常終了" {
  run bash "$SCRIPT_PATH"

  [ "$status" -eq 0 ]
}

@test ".git がディレクトリ (通常 repo): スキップ" {
  local NORMAL_REPO
  NORMAL_REPO=$(mktemp -d)
  mkdir -p "$NORMAL_REPO/.git"

  run bash "$SCRIPT_PATH" "$NORMAL_REPO"

  [ "$status" -eq 0 ]
  rm -rf "$NORMAL_REPO"
}

@test "親 memory が存在しない: 何もせず正常終了" {
  # PARENT_MEM は未作成

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  # worktree memory に何もコピーされていないことを確認
  [ -z "$(ls -A "$WORKTREE_MEM" 2>/dev/null)" ]
}

@test "親 memory が空: 何もせず正常終了" {
  mkdir -p "$PARENT_MEM"
  # PARENT_MEM ディレクトリは存在するが中身は空

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -z "$(ls -A "$WORKTREE_MEM" 2>/dev/null)" ]
}
