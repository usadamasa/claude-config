#!/bin/bash
# ホームディレクトリ走査防止フック
# 許可パス以外の ${HOME} 配下アクセスを deny する
readonly INPUT=$(cat)
readonly TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name')
readonly PROJECT_DIR=$(echo "$INPUT" | jq -r '.cwd')

# 許可するパス一覧(追加削除はここで管理)
readonly ALLOWED_PATHS=(
  "$PROJECT_DIR"
  "$HOME/src"
  "$HOME/.claude"
  "$HOME/obsidian"
)

# ファイルパスを取得(ツールごとに異なるキー)
case "$TOOL_NAME" in
  Read|Edit|Write)
    FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
    ;;
  Glob|Grep)
    FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.path // empty')
    ;;
  *)
    exit 0
    ;;
esac

# パスが空ならデフォルト(CWD)として通過
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# 相対パスを絶対パスに変換
if [[ "$FILE_PATH" != /* ]]; then
  FILE_PATH="$PROJECT_DIR/$FILE_PATH"
fi
readonly FILE_PATH

# ホームディレクトリ配下かチェック
if [[ "$FILE_PATH" == "$HOME"/* || "$FILE_PATH" == "$HOME" ]]; then
  # 許可パスに一致すれば通過
  for allowed in "${ALLOWED_PATHS[@]}"; do
    if [[ "$FILE_PATH" == "$allowed"/* || "$FILE_PATH" == "$allowed" ]]; then
      exit 0
    fi
  done

  # それ以外のホームディレクトリ配下は deny
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"ホームディレクトリ走査防止: %s はプロジェクトディレクトリおよび許可パスの外にあるためアクセスできません"}}\n' "$FILE_PATH"
  exit 0
fi

# ホームディレクトリ外はそのまま通過
exit 0
