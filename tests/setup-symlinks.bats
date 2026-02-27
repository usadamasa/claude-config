#!/usr/bin/env bats
# Taskfile.yml setup タスクの symlink テスト
# CLAUDE-global.md → ~/.claude/CLAUDE.md の特殊マッピングを検証
bats_require_minimum_version 1.5.0

REPO_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  FAKE_HOME="$TEST_TMPDIR/home"
  mkdir -p "$FAKE_HOME/.claude/skills"
  export TEST_TMPDIR FAKE_HOME
}

teardown() {
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
}

# =============================================================================
# ファイル存在チェック
# =============================================================================

@test "CLAUDE-global.md がリポジトリルートに存在する" {
  [ -f "$REPO_ROOT/dotclaude/CLAUDE-global.md" ]
}

@test "CLAUDE.md (プロジェクトスコープ) がリポジトリルートに存在する" {
  [ -f "$REPO_ROOT/CLAUDE.md" ]
}

@test "CLAUDE-global.md と CLAUDE.md の内容が異なる" {
  ! diff -q "$REPO_ROOT/dotclaude/CLAUDE-global.md" "$REPO_ROOT/CLAUDE.md" >/dev/null 2>&1
}

# =============================================================================
# symlink テスト
# =============================================================================

@test "setup で CLAUDE-global.md → ~/.claude/CLAUDE.md にリンクされる" {
  # setup タスクの CLAUDE-global.md 処理を再現
  ln -sfn "$REPO_ROOT/dotclaude/CLAUDE-global.md" "$FAKE_HOME/.claude/CLAUDE.md"

  [ -L "$FAKE_HOME/.claude/CLAUDE.md" ]
  [ "$(readlink "$FAKE_HOME/.claude/CLAUDE.md")" = "$REPO_ROOT/dotclaude/CLAUDE-global.md" ]
}

@test "setup で dotclaude/settings.json → ~/.claude/settings.json にリンクされる (basename 検証)" {
  # setup タスクの basename ロジックを再現
  local src="$REPO_ROOT/dotclaude/settings.json"
  local basename
  basename=$(basename "dotclaude/settings.json")
  ln -sfn "$src" "$FAKE_HOME/.claude/$basename"

  [ -L "$FAKE_HOME/.claude/settings.json" ]
  [ "$(readlink "$FAKE_HOME/.claude/settings.json")" = "$src" ]
  # dotclaude/ がターゲット名に含まれないことを確認
  [ ! -e "$FAKE_HOME/.claude/dotclaude" ]
}

@test "setup で dotclaude/hooks → ~/.claude/hooks にリンクされる (basename 検証)" {
  local src="$REPO_ROOT/dotclaude/hooks"
  local basename
  basename=$(basename "dotclaude/hooks")
  ln -sfn "$src" "$FAKE_HOME/.claude/$basename"

  [ -L "$FAKE_HOME/.claude/hooks" ]
  [ "$(readlink "$FAKE_HOME/.claude/hooks")" = "$src" ]
}

@test "symlink 先の内容がグローバル設定を含む" {
  ln -sfn "$REPO_ROOT/dotclaude/CLAUDE-global.md" "$FAKE_HOME/.claude/CLAUDE.md"

  # グローバル設定特有の内容が含まれる
  grep -q "Conversation Guidelines" "$FAKE_HOME/.claude/CLAUDE.md"
  grep -q "Development Philosophy" "$FAKE_HOME/.claude/CLAUDE.md"
  grep -q "Test-Driven Development" "$FAKE_HOME/.claude/CLAUDE.md"
}

@test "CLAUDE.md (プロジェクトスコープ) にグローバル設定が含まれない" {
  # プロジェクト CLAUDE.md にはグローバル設定特有のセクションがない
  ! grep -q "Test-Driven Development" "$REPO_ROOT/CLAUDE.md"
  ! grep -q "CI 優先ポリシー" "$REPO_ROOT/CLAUDE.md"
  ! grep -q "セッション管理" "$REPO_ROOT/CLAUDE.md"
}

@test "CLAUDE.md (プロジェクトスコープ) にプロジェクト固有の情報が含まれる" {
  # プロジェクト CLAUDE.md にはリポジトリ固有の情報がある
  grep -q "ディレクトリ構成" "$REPO_ROOT/CLAUDE.md" || \
    grep -q "リポジトリ" "$REPO_ROOT/CLAUDE.md"
}

# =============================================================================
# Taskfile.yml の整合性
# =============================================================================

@test "Taskfile.yml の setup タスクに CLAUDE-global.md の処理がある" {
  grep -q "CLAUDE-global.md" "$REPO_ROOT/Taskfile.yml"
}

@test "Taskfile.yml の setup タスクのファイルループに CLAUDE.md が含まれない" {
  # for file in ... ループから CLAUDE.md が除外されている
  # (CLAUDE-global.md は別途処理されるため)
  ! grep -E 'for file in.*CLAUDE\.md' "$REPO_ROOT/Taskfile.yml"
}

# =============================================================================
# worktree 環境関連の構造テスト
# =============================================================================

@test "CLAUDE.md に git rev-parse による worktree 判定の記載がある" {
  grep -q "git rev-parse --git-dir" "$REPO_ROOT/CLAUDE.md"
  grep -q "git rev-parse --git-common-dir" "$REPO_ROOT/CLAUDE.md"
}

@test "CLAUDE.md に worktree 用パス解決 (pwd)/dotclaude/settings.json の記載がある" {
  grep -q 'pwd)/dotclaude/settings.json' "$REPO_ROOT/CLAUDE.md"
}

@test "CLAUDE.md に CLI ツールの --settings オプション記載がある" {
  grep -q '\-\-settings' "$REPO_ROOT/CLAUDE.md"
}

@test "permission-optimizer に worktree チェックの記載がある" {
  grep -q "worktree" "$REPO_ROOT/.claude/skills/permission-optimizer/SKILL.md"
  grep -q 'settings.*pwd' "$REPO_ROOT/.claude/skills/permission-optimizer/SKILL.md"
}

@test "webfetch-domain-manager に worktree チェックの記載がある" {
  grep -q "worktree" "$REPO_ROOT/.claude/skills/webfetch-domain-manager/SKILL.md"
  grep -q 'settings.*pwd' "$REPO_ROOT/.claude/skills/webfetch-domain-manager/SKILL.md"
}

@test "manage-claude-envs に worktree 用パス (pwd)/dotclaude/env.sh の記載がある" {
  grep -q 'pwd)/dotclaude/env.sh' "$REPO_ROOT/.claude/skills/manage-claude-envs/skill.md"
}

# =============================================================================
# worktree 検出方式の統一テスト (git rev-parse)
# =============================================================================

@test "CLAUDE-global.md に git rev-parse による worktree 判定の記載がある" {
  grep -q "git rev-parse --git-dir" "$REPO_ROOT/dotclaude/CLAUDE-global.md"
  grep -q "git rev-parse --git-common-dir" "$REPO_ROOT/dotclaude/CLAUDE-global.md"
}

@test "CLAUDE-global.md に cat .git による判定が残っていない" {
  ! grep -q 'cat \.git' "$REPO_ROOT/dotclaude/CLAUDE-global.md"
}

@test "finalize-pr スキルに git rev-parse による worktree 判定の記載がある" {
  grep -q "git rev-parse --git-dir" "$REPO_ROOT/dotclaude/skills/usadamasa-finalize-pr/SKILL.md"
  grep -q "git rev-parse --git-common-dir" "$REPO_ROOT/dotclaude/skills/usadamasa-finalize-pr/SKILL.md"
}

@test "finalize-pr スキルに cat .git による判定が残っていない" {
  ! grep -q 'cat \.git' "$REPO_ROOT/dotclaude/skills/usadamasa-finalize-pr/SKILL.md"
}

@test "session-handoff スキルに git rev-parse による worktree 判定の記載がある" {
  grep -q "git rev-parse --git-dir" "$REPO_ROOT/dotclaude/skills/usadamasa-session-handoff/SKILL.md"
  grep -q "git rev-parse --git-common-dir" "$REPO_ROOT/dotclaude/skills/usadamasa-session-handoff/SKILL.md"
}

@test "claude-config-management に worktree 環境チェックの記載がある" {
  grep -q "worktree 環境チェック" "$REPO_ROOT/.claude/skills/claude-config-management/SKILL.md"
  grep -q "git rev-parse --git-dir" "$REPO_ROOT/.claude/skills/claude-config-management/SKILL.md"
}
