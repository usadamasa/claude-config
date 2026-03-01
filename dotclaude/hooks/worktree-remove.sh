#!/bin/bash
# worktree-remove.sh
# Claude Code WorktreeRemove フック
# stdin JSON から worktree_path を読み取り、メモリをセーブしてから worktree を削除する
#
# stdin: {"worktree_path": "/path/to/worktree"}
# 注意: git wt -d は使わない (dotfile の wt.deletehook との依存を避ける)
# 環境変数: DRY_RUN=1 で副作用なしのシミュレーション

set -euo pipefail

# shellcheck disable=SC2034 # Used by sourced hook-logger.sh
HOOK_NAME="worktree-remove"
SCRIPT_DIR="${SCRIPT_DIR:-$(cd "$(dirname "$0")" && pwd)}"
# shellcheck disable=SC1091 # Dynamically resolved path
source "$SCRIPT_DIR/lib/hook-logger.sh"
# shellcheck disable=SC1091 # Dynamically resolved path
source "$SCRIPT_DIR/lib/sync-main-repo.sh"

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

# 削除前にメインリポジトリパスを解決 (.git ファイルが消える前に)
MAIN_REPO=$(resolve_main_repo "$WORKTREE_PATH" 2>/dev/null) || MAIN_REPO=""

# メモリセーブ (失敗しても続行、DRY_RUN は環境変数として自動伝播)
"$SCRIPT_DIR/worktree-memory-save.sh" "$WORKTREE_PATH" || true

# worktree 削除
if is_dry_run; then
  logged_cmd git worktree remove "$WORKTREE_PATH"
else
  log_info "CMD    git worktree remove $WORKTREE_PATH"
  if ! git worktree remove "$WORKTREE_PATH" 2>/dev/null; then
    # 失敗時は --force でリトライ
    logged_cmd git worktree remove --force "$WORKTREE_PATH"
    logged_cmd git worktree prune
  fi
fi

# 削除後にメインリポジトリを同期
if [ -n "$MAIN_REPO" ]; then
  sync_main_repo "$MAIN_REPO"
fi
