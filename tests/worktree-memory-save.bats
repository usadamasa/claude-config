#!/usr/bin/env bats
# worktree-memory-save.sh のテスト

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/hooks/worktree-memory-save.sh"

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
  # マーカー付きの worktree MEMORY.md (load 済み状態を再現)
  printf "# 既存の親 MEMORY\n\n<!-- worktree-memory-loaded -->\n# ワークツリー追記" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"既存の親 MEMORY"* ]]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"ワークツリー追記"* ]]
}

@test "追記フォーマット: ## [Merged from worktree: {branch}] ヘッダーが付く" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  # マーカー付きの worktree MEMORY.md
  printf "# 既存の親 MEMORY\n\n<!-- worktree-memory-loaded -->\n# ワークツリー追記" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"## [Merged from worktree: feature]"* ]]
}

# =============================================================================
# マーカーベースの差分追記
# =============================================================================

@test "マーカー付き MEMORY.md: 差分のみ親に追記される" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  # マーカー前 = load 時にコピーされた内容、マーカー後 = worktree で追記した新規内容
  printf "# 既存の親 MEMORY\n\n<!-- worktree-memory-loaded -->\n# 新規追記\nworktree で学んだこと" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  # 親に新規内容が追記されていること
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"新規追記"* ]]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"worktree で学んだこと"* ]]
  # マーカー前の内容 (既存の親 MEMORY) が二重にならないこと
  local count
  count=$(grep -c "既存の親 MEMORY" "$PARENT_MEM/MEMORY.md")
  [ "$count" -eq 1 ]
}

@test "マーカー付き MEMORY.md: 新規追加なしの場合はスキップ" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  # マーカーで終わっている (追記なし)
  printf "# 既存の親 MEMORY\n\n<!-- worktree-memory-loaded -->" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  # 親 MEMORY.md が変更されていないこと
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == "# 既存の親 MEMORY" ]]
  # SKIP ログが出力されること
  [[ "$output" == *"SKIP"* ]]
}

@test "親 MEMORY.md が存在しない + マーカー付き: マーカーを除去してコピー" {
  # 親 MEMORY.md は存在しない
  # worktree にはマーカー付き MEMORY.md + 新規内容
  printf "# ロードされた内容\n\n<!-- worktree-memory-loaded -->\n# 新規追記" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ -f "$PARENT_MEM/MEMORY.md" ]
  # マーカー行が除去されていること
  ! grep -q "<!-- worktree-memory-loaded -->" "$PARENT_MEM/MEMORY.md"
  # 内容は保持されること
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"ロードされた内容"* ]]
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == *"新規追記"* ]]
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

# =============================================================================
# 通常モード: 操作ログ
# =============================================================================

@test "通常モード: MKDIR ログが stderr に出力される" {
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"
  # PARENT_MEM は未作成

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[worktree-memory-save]"* ]]
  [[ "$output" == *"MKDIR"* ]]
}

@test "通常モード: SESSION_HANDOFF の CP ログが出力される" {
  echo "# SESSION HANDOFF" > "$WORKTREE_MEM/SESSION_HANDOFF.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[worktree-memory-save]"* ]]
  [[ "$output" == *"CP"* ]]
  [[ "$output" == *"SESSION_HANDOFF"* ]]
}

@test "通常モード: MEMORY.md 追記時に APPEND ログが出力される" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  # マーカー付きの worktree MEMORY.md
  printf "# 既存の親 MEMORY\n\n<!-- worktree-memory-loaded -->\n# ワークツリー追記" > "$WORKTREE_MEM/MEMORY.md"

  run bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[worktree-memory-save]"* ]]
  [[ "$output" == *"APPEND"* ]]
}

# =============================================================================
# dry-run モード
# =============================================================================

@test "dry-run: ファイルがコピーされない" {
  echo "# SESSION HANDOFF" > "$WORKTREE_MEM/SESSION_HANDOFF.md"
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [ ! -f "$PARENT_MEM/SESSION_HANDOFF_feature.md" ]
  [ ! -f "$PARENT_MEM/MEMORY.md" ]
}

@test "dry-run: [DRY-RUN] プレフィックスでログが出力される" {
  echo "# ワークツリー MEMORY" > "$WORKTREE_MEM/MEMORY.md"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
}

@test "dry-run: MEMORY.md 追記時に APPEND メッセージが出力される" {
  mkdir -p "$PARENT_MEM"
  echo "# 既存の親 MEMORY" > "$PARENT_MEM/MEMORY.md"
  # マーカー付きの worktree MEMORY.md
  printf "# 既存の親 MEMORY\n\n<!-- worktree-memory-loaded -->\n# ワークツリー追記" > "$WORKTREE_MEM/MEMORY.md"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"APPEND"* ]]
  # 親 MEMORY.md が変更されていないこと
  [[ "$(cat "$PARENT_MEM/MEMORY.md")" == "# 既存の親 MEMORY" ]]
}

@test "dry-run: worktree memory が存在しない場合に SKIP ログが出力される" {
  rm -rf "$WORKTREE_MEM"

  run env DRY_RUN=1 bash "$SCRIPT_PATH" "$TEST_WORKTREE"

  [ "$status" -eq 0 ]
  [[ "$output" == *"[DRY-RUN]"* ]]
  [[ "$output" == *"SKIP"* ]]
}
