#!/bin/bash
set -euo pipefail
# PostToolUse hook: Write/Edit で .sh ファイルが変更されたら shellcheck を実行
# 警告のみ (exit 0) - ブロックはしない

INPUT=$(cat)

# lint ツールが利用可能か確認
if ! command -v shellcheck &>/dev/null; then
  exit 0
fi

# ツール入力からファイルパスを取得
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.notebook_path // empty')

# .sh ファイル以外はスキップ
case "$FILE_PATH" in
  *.sh) ;;
  *)    exit 0 ;;
esac

# ファイルが存在しない場合はスキップ
if [ ! -f "$FILE_PATH" ]; then
  exit 0
fi

# lint 実行 (結果は stderr に出力、exit code は無視)
if ! shellcheck "$FILE_PATH" 2>&1; then
  echo "[shellcheck] warnings found in $(basename "$FILE_PATH")" >&2
fi

# 常に exit 0 (ブロックしない)
exit 0
