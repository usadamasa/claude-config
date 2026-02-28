#!/bin/bash
set -euo pipefail
# 広域ワイルドカードパーミッション自動除去フック
# SessionStart 時に settings.local.json から広域パターンを除去する
readonly INPUT=$(cat)

# SCRIPT_DIR の解決 (テスト時にオーバーライド可能)
if [ -z "${SCRIPT_DIR:-}" ]; then
  HOOK_REAL_PATH=$(readlink "$0" 2>/dev/null || echo "$0")
  SCRIPT_DIR=$(cd "$(dirname "$HOOK_REAL_PATH")" && pwd)
fi

# ロガー読み込み
HOOK_NAME="guard-broad-wildcard"
source "$SCRIPT_DIR/lib/hook-logger.sh"

# jq の存在チェック
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

# cwd 取得
readonly CWD=$(echo "$INPUT" | jq -r '.cwd // empty')

# ガード対象ツール (追加・変更はここで管理)
GUARDED_TOOLS=(
  Bash Read Write Edit Glob Grep
  WebFetch Fetch NotebookEdit
)

# jq フィルタ用の正規表現を動的生成
TOOLS_RE=$(IFS='|'; echo "${GUARDED_TOOLS[*]}")
readonly BLOCK_PATTERN="^(${TOOLS_RE})(\\([: ]*\\*\\))?$|^mcp__\\*$"

# settings.local.json をクリーンする関数
clean_settings_local() {
  local settings_file="$1"

  # ファイルが存在しなければスキップ
  if [ ! -f "$settings_file" ]; then
    log_skip "settings.local.json not found: $settings_file"
    return 0
  fi

  # permissions.allow が存在しなければスキップ
  if ! jq -e '.permissions.allow' "$settings_file" &>/dev/null; then
    log_skip "permissions.allow not found in: $settings_file"
    return 0
  fi

  # 除去対象エントリを検出
  local removed
  removed=$(jq -r --arg pat "$BLOCK_PATTERN" \
    '[.permissions.allow[] | select(test($pat))] | .[]' \
    "$settings_file" 2>/dev/null || true)

  if [ -z "$removed" ]; then
    log_skip "no broad wildcards found in: $settings_file"
    return 0
  fi

  # 除去対象をログ出力
  while IFS= read -r entry; do
    log_info "REMOVE $entry from $settings_file"
  done <<< "$removed"

  if is_dry_run; then
    return 0
  fi

  # アトミック書き換え
  local tmpfile
  tmpfile=$(mktemp "${settings_file}.XXXXXX")

  if jq --arg pat "$BLOCK_PATTERN" \
    '.permissions.allow |= [.[] | select(test($pat) | not)]' \
    "$settings_file" > "$tmpfile"; then
    mv "$tmpfile" "$settings_file"
  else
    rm -f "$tmpfile"
    log_info "ERROR: failed to filter $settings_file"
    return 1
  fi
}

# プロジェクト settings.local.json
if [ -n "$CWD" ]; then
  clean_settings_local "$CWD/.claude/settings.local.json"
fi

# グローバル settings.local.json
clean_settings_local "$HOME/.claude/settings.local.json"

exit 0
