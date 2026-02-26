#!/bin/bash
# worktree-memory-load.sh
# worktree 作成時に親リポジトリの Claude Code auto-memory を worktree にコピーする
#
# Usage: worktree-memory-load.sh <worktree-path>

set -euo pipefail

WORKTREE_PATH="${1:-}"
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

# パスエンコード: Claude Code の auto-memory パス命名規則に合わせる
# / . _ を - に変換 (先頭 / も - になる)
encode_path() { echo "$1" | tr '/._' '-'; }

WORKTREE_ENC=$(encode_path "$WORKTREE_PATH")
PARENT_ENC=$(encode_path "$PARENT_ROOT")

WORKTREE_MEM="$HOME/.claude/projects/$WORKTREE_ENC/memory"
PARENT_MEM="$HOME/.claude/projects/$PARENT_ENC/memory"

# 親 memory ディレクトリが存在しない場合はスキップ
[ -d "$PARENT_MEM" ] || exit 0

# 親 memory が空の場合はスキップ
# shellcheck disable=SC2012
[ -n "$(ls -A "$PARENT_MEM" 2>/dev/null)" ] || exit 0

# worktree memory ディレクトリを自動作成
mkdir -p "$WORKTREE_MEM"

# 親 memory の全ファイルを worktree にコピー
cp "$PARENT_MEM/"* "$WORKTREE_MEM/"
