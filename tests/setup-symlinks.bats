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
  [ -f "$REPO_ROOT/CLAUDE-global.md" ]
}

@test "CLAUDE.md (プロジェクトスコープ) がリポジトリルートに存在する" {
  [ -f "$REPO_ROOT/CLAUDE.md" ]
}

@test "CLAUDE-global.md と CLAUDE.md の内容が異なる" {
  ! diff -q "$REPO_ROOT/CLAUDE-global.md" "$REPO_ROOT/CLAUDE.md" >/dev/null 2>&1
}

# =============================================================================
# symlink テスト
# =============================================================================

@test "setup で CLAUDE-global.md → ~/.claude/CLAUDE.md にリンクされる" {
  # setup タスクの CLAUDE-global.md 処理を再現
  ln -sfn "$REPO_ROOT/CLAUDE-global.md" "$FAKE_HOME/.claude/CLAUDE.md"

  [ -L "$FAKE_HOME/.claude/CLAUDE.md" ]
  [ "$(readlink "$FAKE_HOME/.claude/CLAUDE.md")" = "$REPO_ROOT/CLAUDE-global.md" ]
}

@test "symlink 先の内容がグローバル設定を含む" {
  ln -sfn "$REPO_ROOT/CLAUDE-global.md" "$FAKE_HOME/.claude/CLAUDE.md"

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
