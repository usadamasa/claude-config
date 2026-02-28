#!/bin/bash
# worktree-create.sh
# Claude Code WorktreeCreate フック
# stdin JSON から name, cwd を読み取り、git wt で worktree を作成し、メモリをロードする
#
# stdin: {"name": "branch-name", "cwd": "/path/to/repo"}
# stdout: worktree パス (Claude Code が cd する先)
# 環境変数: DRY_RUN=1 で副作用なしのシミュレーション

set -euo pipefail

# shellcheck disable=SC2034 # Used by sourced hook-logger.sh
HOOK_NAME="worktree-create"
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "$0")" && pwd)}"
# shellcheck disable=SC1091 # Dynamically resolved path
source "$SCRIPT_DIR/lib/hook-logger.sh"

# 前提コマンドチェック
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

# stdin から JSON を読み取り
INPUT=$(cat)
NAME=$(echo "$INPUT" | jq -r '.name // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')

if is_dry_run; then
  # dry-run: git wt を実行せず、ダミーパスを返す
  logged_cmd git wt --nocd "$NAME"
  WORKTREE_PATH="/tmp/dry-run-worktree/$NAME"
  echo "$WORKTREE_PATH"
  exit 0
fi

# git wt で worktree 作成 (--nocd でディレクトリ移動なし、パスのみ出力)
log_info "git wt --nocd $NAME"
WORKTREE_PATH=$(cd "$CWD" && git wt --nocd "$NAME")
if [ -z "$WORKTREE_PATH" ]; then
  echo "ERROR: git wt failed" >&2
  exit 1
fi

# メモリロード (失敗しても続行)
"$SCRIPT_DIR/worktree-memory-load.sh" "$WORKTREE_PATH" || true

# stdout に worktree パスを出力
echo "$WORKTREE_PATH"
