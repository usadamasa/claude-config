#!/bin/bash
# worktree-create.sh
# Claude Code WorktreeCreate フック
# stdin JSON から name, cwd を読み取り、git wt で worktree を作成し、メモリをロードする
#
# stdin: {"name": "branch-name", "cwd": "/path/to/repo"}
# stdout: worktree パス (Claude Code が cd する先)

set -euo pipefail

SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "$0")" && pwd)}"

# 前提コマンドチェック
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

# stdin から JSON を読み取り
INPUT=$(cat)
NAME=$(echo "$INPUT" | jq -r '.name // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')

# git wt で worktree 作成 (--nocd でディレクトリ移動なし、パスのみ出力)
WORKTREE_PATH=$(cd "$CWD" && git wt --nocd "$NAME")
if [ -z "$WORKTREE_PATH" ]; then
  echo "ERROR: git wt failed" >&2
  exit 1
fi

# メモリロード (失敗しても続行)
"$SCRIPT_DIR/worktree-memory-load.sh" "$WORKTREE_PATH" || true

# stdout に worktree パスを出力
echo "$WORKTREE_PATH"
