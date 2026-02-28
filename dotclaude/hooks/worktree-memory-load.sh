#!/bin/bash
# worktree-memory-load.sh
# worktree 作成時に親リポジトリの Claude Code auto-memory を worktree にコピーする
#
# Usage: worktree-memory-load.sh <worktree-path>
# 環境変数: DRY_RUN=1 で副作用なしのシミュレーション

set -euo pipefail

# shellcheck disable=SC2034 # Used by sourced hook-logger.sh
HOOK_NAME="worktree-memory-load"
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "$0")" && pwd)}"
# shellcheck disable=SC1091 # Dynamically resolved path
source "$SCRIPT_DIR/lib/hook-logger.sh"

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
if [ ! -d "$PARENT_MEM" ]; then
  log_skip "親 memory が存在しない"
  exit 0
fi

# 親 memory が空の場合はスキップ
# shellcheck disable=SC2012
if [ -z "$(ls -A "$PARENT_MEM" 2>/dev/null)" ]; then
  log_skip "親 memory が空"
  exit 0
fi

# worktree memory ディレクトリを自動作成
logged_mkdir "$WORKTREE_MEM"

# 親 memory の全ファイルを worktree にコピー (各ファイルごとにログ出力)
MEMORY_MARKER="<!-- worktree-memory-loaded -->"
for f in "$PARENT_MEM/"*; do
  [ -f "$f" ] || continue
  logged_cp "$f" "$WORKTREE_MEM/$(basename "$f")"
  # MEMORY.md にのみロードマーカーを追記
  if [ "$(basename "$f")" = "MEMORY.md" ]; then
    if is_dry_run; then
      log_info "MARKER $WORKTREE_MEM/MEMORY.md"
    else
      { echo ""; echo "$MEMORY_MARKER"; } >> "$WORKTREE_MEM/MEMORY.md"
    fi
  fi
done
