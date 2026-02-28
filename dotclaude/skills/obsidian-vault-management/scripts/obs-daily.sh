#!/usr/bin/env bash
# obs-daily.sh - Daily Note を冪等に作成/オープンする
# Usage: obs-daily.sh [vault=<name>] [date=YYYY-MM-DD]
set -euo pipefail

# デフォルト値
TARGET_DATE=$(date +%Y-%m-%d)
VAULT_ARGS=()

# 引数解析
for arg in "$@"; do
  case "$arg" in
    vault=*) VAULT_ARGS+=("$arg") ;;
    date=*)  TARGET_DATE="${arg#date=}" ;;
  esac
done

DAILY_PATH="01_Daily/${TARGET_DATE}.md"

# ファイルの存在チェック: CLI は exit code 0 を常に返すため、出力に "Error:" が含まれるかで判定
read_output=$(obsidian read path="${DAILY_PATH}" "${VAULT_ARGS[@]}" 2>/dev/null)

if echo "$read_output" | grep -q "^Error:"; then
  echo "Creating daily note: ${DAILY_PATH}"
  # テンプレートがあれば使用、なければフォールバック
  template_output=$(obsidian template:read name=Daily "${VAULT_ARGS[@]}" 2>/dev/null)
  if echo "$template_output" | grep -q "^Error:"; then
    obsidian create path="${DAILY_PATH}" open "${VAULT_ARGS[@]}" \
      content="## DAILY PLAN\n\n## TODO\n\n## やったこと\n\n## 後で読む\n![[あとで読む]]" 2>/dev/null
  else
    obsidian create path="${DAILY_PATH}" template=Daily open "${VAULT_ARGS[@]}" 2>/dev/null
  fi
else
  echo "Daily note already exists: ${DAILY_PATH}"
  obsidian open path="${DAILY_PATH}" "${VAULT_ARGS[@]}" 2>/dev/null
fi

echo "Done: ${DAILY_PATH}"
