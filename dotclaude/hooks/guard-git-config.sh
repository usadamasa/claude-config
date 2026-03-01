#!/bin/bash
set -euo pipefail
# git config 書き込みガードフック
# 読み取り系と worktree 用 remote.origin.fetch 以外の git config 書き込みをブロックする

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
# "git config ..." の部分だけを判定対象にする
FIRST_CMD="${COMMAND%%|*}"
FIRST_CMD="${FIRST_CMD%%&&*}"
FIRST_CMD="${FIRST_CMD%%;*}"

# git config 以外のコマンドは通過
# git -c ... はインライン設定であり書き込みではないため通過
if ! printf '%s' "$FIRST_CMD" | grep -qE '(^|[[:space:]])git[[:space:]]+config([[:space:]]|$)'; then
  exit 0
fi

# 読み取り系オプション: 通過
if printf '%s' "$FIRST_CMD" | grep -qE '(^|[[:space:]])git[[:space:]]+config[[:space:]]+.*(--get|--get-all|--get-regexp|--get-urlmatch|--list|--show-origin|--show-scope|-l)([[:space:]]|$)'; then
  exit 0
fi

# worktree 例外: remote.origin.fetch への設定は許可
if printf '%s' "$FIRST_CMD" | grep -qE '(^|[[:space:]])git[[:space:]]+config[[:space:]]+remote\.origin\.fetch([[:space:]]|$)'; then
  exit 0
fi

# 引数なしの git config <key> (値の読み取り) を判定
# git config <key> は読み取り、git config <key> <value> は書き込み
# --global, --local, --system, --unset, --unset-all, --replace-all, --add 等がある場合は書き込み
if printf '%s' "$FIRST_CMD" | grep -qE '(^|[[:space:]])git[[:space:]]+config[[:space:]]+--(global|local|system|worktree|file|blob|unset|unset-all|replace-all|add|rename-section|remove-section|edit)([[:space:]]|$)'; then
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"git config の書き込みはフックによりブロックされました。読み取り (--get, --list) と remote.origin.fetch の設定のみ許可されています。"}}\n'
  exit 0
fi

# フラグなしの git config <key> <value> を検出
# git config <key> のみ(引数1つ)なら読み取り、<key> <value>(引数2つ以上)なら書き込み
# "git config" の後のトークン数をカウント
CONFIG_ARGS="${FIRST_CMD#*git config}"
CONFIG_ARGS=$(printf '%s' "$CONFIG_ARGS" | sed 's/^[[:space:]]*//')

# 空なら git config 単体(ヘルプ表示) → 通過
if [ -z "$CONFIG_ARGS" ]; then
  exit 0
fi

# トークンをカウント(シンプルなスペース区切り)
# 1トークン = 読み取り(key のみ)、2トークン以上 = 書き込み(key + value)
WORD_COUNT=0
IN_QUOTE=false
QUOTE_CHAR=""
for (( i=0; i<${#CONFIG_ARGS}; i++ )); do
  CHAR="${CONFIG_ARGS:$i:1}"
  if $IN_QUOTE; then
    if [ "$CHAR" = "$QUOTE_CHAR" ]; then
      IN_QUOTE=false
    fi
  elif [ "$CHAR" = '"' ] || [ "$CHAR" = "'" ]; then
    if [ "$WORD_COUNT" -eq 0 ] || [ "${CONFIG_ARGS:$((i-1)):1}" = " " ]; then
      IN_QUOTE=true
      QUOTE_CHAR="$CHAR"
      WORD_COUNT=$((WORD_COUNT + 1))
    fi
  elif [ "$CHAR" = " " ] || [ "$CHAR" = "	" ]; then
    continue
  else
    if [ "$i" -eq 0 ] || [ "${CONFIG_ARGS:$((i-1)):1}" = " " ] || [ "${CONFIG_ARGS:$((i-1)):1}" = "	" ]; then
      WORD_COUNT=$((WORD_COUNT + 1))
    fi
  fi
done

if [ "$WORD_COUNT" -ge 2 ]; then
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"git config の書き込みはフックによりブロックされました。読み取り (--get, --list) と remote.origin.fetch の設定のみ許可されています。"}}\n'
  exit 0
fi

# 1トークン以下 = 読み取り(git config <key>) → 通過
exit 0
