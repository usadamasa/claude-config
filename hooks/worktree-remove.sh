#!/bin/bash
# worktree-remove.sh
# Claude Code WorktreeRemove フック
# stdin JSON から worktree_path を読み取り、メモリをセーブしてから worktree を削除する
#
# stdin: {"worktree_path": "/path/to/worktree"}
# 注意: git wt -d は使わない (dotfile の wt.deletehook との依存を避ける)

set -euo pipefail

SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "$0")" && pwd)}"

# 前提コマンドチェック
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

# stdin から JSON を読み取り
INPUT=$(cat)
WORKTREE_PATH=$(echo "$INPUT" | jq -r '.worktree_path // empty')

if [ -z "$WORKTREE_PATH" ]; then
  echo "ERROR: worktree_path is required" >&2
  exit 1
fi

# メモリセーブ (失敗しても続行)
"$SCRIPT_DIR/worktree-memory-save.sh" "$WORKTREE_PATH" || true

# worktree 削除
if ! git worktree remove "$WORKTREE_PATH" 2>/dev/null; then
  # 失敗時は --force でリトライ
  git worktree remove --force "$WORKTREE_PATH"
  git worktree prune
fi
