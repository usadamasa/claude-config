#!/bin/bash
set -euo pipefail
# main/master ブランチでの git commit をガードするフック
# feature ブランチでの作業を強制し、直接 main への commit を防止する

INPUT=$(cat)
readonly INPUT

TOOL_NAME=$(printf '%s' "$INPUT" | jq -r '.tool_name')
readonly TOOL_NAME

# Bash 以外のツールは対象外
if [ "$TOOL_NAME" != "Bash" ]; then
  exit 0
fi

COMMAND=$(printf '%s' "$INPUT" | jq -r '.tool_input.command')
readonly COMMAND

# コマンドの先頭トークンを抽出(パイプやチェーン前の最初のコマンド)
FIRST_CMD="${COMMAND%%|*}"
FIRST_CMD="${FIRST_CMD%%&&*}"
FIRST_CMD="${FIRST_CMD%%;*}"

# git commit 以外のコマンドは通過
if ! printf '%s' "$FIRST_CMD" | grep -qE '(^|[[:space:]])git[[:space:]]+commit([[:space:]]|$)'; then
  exit 0
fi

# 現在のブランチを取得
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || printf '')
readonly CURRENT_BRANCH

# main または master ブランチなら deny
if [ "$CURRENT_BRANCH" = "main" ] || [ "$CURRENT_BRANCH" = "master" ]; then
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"%s ブランチへの直接コミットはフックによりブロックされました。feature ブランチを作成してください。"}}\n' "$CURRENT_BRANCH"
  exit 0
fi

exit 0
