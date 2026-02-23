#!/bin/bash
# migrate-worktree-memory.sh
# worktree 削除前に Claude Code auto-memory を親リポジトリの memory パスに移行する
#
# Usage: migrate-worktree-memory.sh <worktree-path>
#        または環境変数 GIT_WT_WORKTREE_PATH

set -euo pipefail

WORKTREE_PATH="${1:-${GIT_WT_WORKTREE_PATH:-}}"
if [ -z "$WORKTREE_PATH" ]; then exit 0; fi

# .git がファイルかどうか確認 (worktree 判定)
GIT_FILE="$WORKTREE_PATH/.git"
[ -f "$GIT_FILE" ] || exit 0

# gitdir パスを取得
GIT_DIR=$(sed 's/^gitdir: //' "$GIT_FILE" | tr -d '\n')
[ -d "$GIT_DIR" ] || exit 0

# commondir ファイルから親 .git ディレクトリを特定
COMMON_DIR_FILE="$GIT_DIR/commondir"
[ -f "$COMMON_DIR_FILE" ] || exit 0

COMMON_REL=$(tr -d '\n' < "$COMMON_DIR_FILE")

# 相対パスを絶対パスに変換
if [[ "$COMMON_REL" == /* ]]; then
  COMMON_ABS="$COMMON_REL"
else
  COMMON_ABS="$(cd "$GIT_DIR" && cd "$COMMON_REL" && pwd)"
fi

PARENT_ROOT="$(dirname "$COMMON_ABS")"

# ブランチ名を取得 (HEAD ファイルから)
HEAD_FILE="$GIT_DIR/HEAD"
if [ -f "$HEAD_FILE" ]; then
  BRANCH_NAME=$(sed 's|ref: refs/heads/||' "$HEAD_FILE" | tr -d '\n')
else
  BRANCH_NAME="unknown"
fi
[ -z "$BRANCH_NAME" ] && BRANCH_NAME="unknown"

# パスエンコード: Claude Code の auto-memory パス命名規則に合わせる
# / . _ を - に変換 (先頭 / も - になる)
encode_path() { echo "$1" | tr '/._' '-'; }

WORKTREE_ENC=$(encode_path "$WORKTREE_PATH")
PARENT_ENC=$(encode_path "$PARENT_ROOT")

WORKTREE_MEM="$HOME/.claude/projects/$WORKTREE_ENC/memory"
PARENT_MEM="$HOME/.claude/projects/$PARENT_ENC/memory"

# worktree memory ディレクトリが存在しない場合はスキップ
[ -d "$WORKTREE_MEM" ] || exit 0

# 移行するファイルがない場合はスキップ
HAS_FILES=false
for f in MEMORY.md SESSION_HANDOFF.md; do
  [ -f "$WORKTREE_MEM/$f" ] && HAS_FILES=true && break
done
[ "$HAS_FILES" = true ] || exit 0

mkdir -p "$PARENT_MEM"

# SESSION_HANDOFF.md: ブランチ名付きファイルにコピー (複数 worktree の情報が混在しないよう)
if [ -f "$WORKTREE_MEM/SESSION_HANDOFF.md" ]; then
  cp "$WORKTREE_MEM/SESSION_HANDOFF.md" "$PARENT_MEM/SESSION_HANDOFF_${BRANCH_NAME}.md"
fi

# MEMORY.md: 親が存在する場合は末尾に追記、存在しない場合はコピー
if [ -f "$WORKTREE_MEM/MEMORY.md" ]; then
  if [ -f "$PARENT_MEM/MEMORY.md" ]; then
    DATE=$(date +%Y-%m-%d)
    {
      echo ""
      echo "## [Merged from worktree: $BRANCH_NAME] $DATE"
      echo ""
      cat "$WORKTREE_MEM/MEMORY.md"
    } >> "$PARENT_MEM/MEMORY.md"

    # MEMORY.md サイズ警告 (Claude Code は 200 行以降を切り捨てる)
    LINE_COUNT=$(wc -l < "$PARENT_MEM/MEMORY.md")
    if [ "$LINE_COUNT" -gt 150 ]; then
      echo "WARNING: MEMORY.md is ${LINE_COUNT} lines. Claude Code truncates after 200 lines. Consider manual cleanup." >&2
    fi
  else
    cp "$WORKTREE_MEM/MEMORY.md" "$PARENT_MEM/MEMORY.md"
  fi
fi
