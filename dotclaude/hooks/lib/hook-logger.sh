#!/bin/bash
# hook-logger.sh
# worktree hook スクリプト共通のログ・dry-run ヘルパー
#
# 使い方:
#   HOOK_NAME="worktree-create"
#   source "$SCRIPT_DIR/lib/hook-logger.sh"
#
# 環境変数:
#   DRY_RUN=1  全操作をシミュレーション表示し、実際の副作用を起こさない
#   HOOK_NAME  ログプレフィックスに使用するスクリプト名

DRY_RUN="${DRY_RUN:-0}"
HOOK_NAME="${HOOK_NAME:-hook}"

# DRY_RUN=1 かどうかを判定
is_dry_run() {
  [ "$DRY_RUN" = "1" ]
}

# ログプレフィックスを返す
# dry-run: "[DRY-RUN]", 通常: "[$HOOK_NAME]"
_log_prefix() {
  if is_dry_run; then
    echo "[DRY-RUN]"
  else
    echo "[$HOOK_NAME]"
  fi
}

# mkdir -p + ログ出力。dry-run 時は mkdir しない
logged_mkdir() {
  local target="$1"
  if is_dry_run; then
    echo "$(_log_prefix) MKDIR  $target" >&2
  else
    echo "$(_log_prefix) MKDIR  $target" >&2
    mkdir -p "$target"
  fi
}

# cp + ログ出力。dry-run 時は cp しない
logged_cp() {
  local src="$1"
  local dst="$2"
  if is_dry_run; then
    echo "$(_log_prefix) CP     $src -> $dst" >&2
  else
    echo "$(_log_prefix) CP     $src -> $dst" >&2
    cp "$src" "$dst"
  fi
}

# コマンド実行 + ログ出力。dry-run 時は実行しない
logged_cmd() {
  if is_dry_run; then
    echo "$(_log_prefix) CMD    $*" >&2
  else
    echo "$(_log_prefix) CMD    $*" >&2
    "$@"
  fi
}

# スキップ理由の通知
log_skip() {
  local reason="$1"
  echo "$(_log_prefix) SKIP   ($reason)" >&2
}

# 情報メッセージ (常に出力)
log_info() {
  local msg="$1"
  echo "$(_log_prefix) INFO   $msg" >&2
}

# エラーメッセージ (常に出力、重大度: ERROR)
log_error() {
  local msg="$1"
  echo "$(_log_prefix) ERROR  $msg" >&2
}
