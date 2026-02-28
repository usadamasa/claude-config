#!/bin/bash
set -euo pipefail
# 広域ワイルドカードパーミッション自動除去フック
# SessionStart 時に settings.local.json から広域パターンを除去する
readonly INPUT=$(cat)

# SCRIPT_DIR の解決 (テスト時にオーバーライド可能)
if [ -z "${SCRIPT_DIR:-}" ]; then
  SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
fi

# ロガー読み込み
HOOK_NAME="guard-broad-wildcard"
if [ ! -f "$SCRIPT_DIR/lib/hook-logger.sh" ]; then
  echo "ERROR: [$HOOK_NAME] hook-logger.sh not found at $SCRIPT_DIR/lib/hook-logger.sh" >&2
  exit 1
fi
source "$SCRIPT_DIR/lib/hook-logger.sh"

# jq の存在チェック
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

# cwd 取得 (入力JSON のパースエラーを検出)
CWD=""
if ! CWD=$(echo "$INPUT" | jq -r '.cwd // empty' 2>&1); then
  log_error "failed to parse hook input JSON: $CWD"
  exit 1
fi
readonly CWD

# cwd が指定されている場合は絶対パスであることを検証
if [ -n "$CWD" ] && [[ "$CWD" != /* ]]; then
  log_error "cwd from hook input is not an absolute path: $CWD"
  exit 1
fi

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

  # JSON として有効かチェック
  if ! jq -e '.' "$settings_file" >/dev/null 2>&1; then
    log_error "$settings_file is not valid JSON"
    return 1
  fi

  # permissions または permissions.allow が存在しなければスキップ
  if ! jq -e '.permissions.allow' "$settings_file" >/dev/null 2>&1; then
    log_skip "permissions.allow not found in: $settings_file"
    return 0
  fi

  # 除去対象エントリを検出 (文字列のみを対象にし、非文字列エントリではエラーにしない)
  local removed
  if ! removed=$(jq -r --arg pat "$BLOCK_PATTERN" \
    '[.permissions.allow[] | strings | select(test($pat))] | .[]' \
    "$settings_file" 2>&1); then
    log_error "jq failed to scan $settings_file: $removed"
    return 1
  fi

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
  if ! tmpfile=$(mktemp "${settings_file}.XXXXXX"); then
    log_error "failed to create temporary file alongside $settings_file"
    return 1
  fi

  if jq --arg pat "$BLOCK_PATTERN" \
    '.permissions.allow |= [.[] | if type == "string" then select(test($pat) | not) else . end]' \
    "$settings_file" > "$tmpfile"; then
    if ! mv "$tmpfile" "$settings_file"; then
      log_error "failed to replace $settings_file (tmpfile left at $tmpfile)"
      return 1
    fi
  else
    rm -f "$tmpfile"
    log_error "jq failed to filter $settings_file"
    return 1
  fi
}

overall_ok=0

# プロジェクト settings.local.json
if [ -n "$CWD" ]; then
  if ! clean_settings_local "$CWD/.claude/settings.local.json"; then
    log_error "failed to clean project settings.local.json"
    overall_ok=1
  fi
fi

# グローバル settings.local.json
if ! clean_settings_local "$HOME/.claude/settings.local.json"; then
  log_error "failed to clean global settings.local.json"
  overall_ok=1
fi

exit "$overall_ok"
