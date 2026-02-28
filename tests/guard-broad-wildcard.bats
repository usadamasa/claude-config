#!/usr/bin/env bats
# guard-broad-wildcard.sh のテスト
bats_require_minimum_version 1.5.0

SCRIPT_PATH="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks/guard-broad-wildcard.sh"

setup() {
  TEST_TMPDIR=$(mktemp -d)
  MOCK_BIN="$TEST_TMPDIR/bin"
  mkdir -p "$MOCK_BIN"

  # hook-logger.sh をコピー
  REAL_HOOKS_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)/dotclaude/hooks"
  mkdir -p "$MOCK_BIN/lib"
  cp "$REAL_HOOKS_DIR/lib/hook-logger.sh" "$MOCK_BIN/lib/hook-logger.sh"

  # テスト用プロジェクトディレクトリ
  MOCK_PROJECT="$TEST_TMPDIR/project"
  mkdir -p "$MOCK_PROJECT/.claude"

  # テスト用 HOME
  ORIG_HOME="$HOME"
  MOCK_HOME="$TEST_TMPDIR/home"
  mkdir -p "$MOCK_HOME/.claude"

  export TEST_TMPDIR MOCK_BIN MOCK_PROJECT MOCK_HOME
  export SCRIPT_DIR="$MOCK_BIN"
  export HOME="$MOCK_HOME"
}

teardown() {
  export HOME="$ORIG_HOME"
  [ -n "$TEST_TMPDIR" ] && rm -rf "$TEST_TMPDIR"
  # SCRIPT_DIR をクリアしないと後続のテストがモックライブラリを参照し続ける
  unset SCRIPT_DIR
}

# ヘルパー: settings.local.json を作成
create_settings_local() {
  local dir="$1"
  local content="$2"
  echo "$content" > "$dir/.claude/settings.local.json"
}

# =============================================================================
# settings.local.json 不在
# =============================================================================

@test "settings.local.json が存在しない: 正常終了" {
  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"SKIP"* ]]
}

# =============================================================================
# permissions.allow が空
# =============================================================================

@test "permissions.allow が空配列: 正常終了 (変更なし)" {
  create_settings_local "$MOCK_PROJECT" '{"permissions":{"allow":[]}}'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"SKIP"* ]]
  # ファイル内容が変わっていないこと
  local result
  result=$(jq '.permissions.allow | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$result" -eq 0 ]
}

# =============================================================================
# Bash(*) 除去 (正当エントリは保持)
# =============================================================================

@test "Bash(*) が除去される (Bash(git status:*) と Bash(ls:*) は保持)" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Bash(*)", "Bash(git status:*)", "Bash(ls:*)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  # Bash(*) が除去されていること
  local count
  count=$(jq '[.permissions.allow[] | select(. == "Bash(*)")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 0 ]
  # 正当なエントリが保持されていること
  count=$(jq '[.permissions.allow[] | select(. == "Bash(git status:*)")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
  count=$(jq '[.permissions.allow[] | select(. == "Bash(ls:*)")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
}

# =============================================================================
# ベアエントリ除去
# =============================================================================

@test "Read, Write, Edit のベアエントリが除去される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Read", "Write", "Edit", "Read(~/src/**)", "WebSearch"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  local result
  result=$(jq -c '.permissions.allow' "$MOCK_PROJECT/.claude/settings.local.json")
  [[ "$result" == *'"Read(~/src/**)"'* ]]
  [[ "$result" == *'"WebSearch"'* ]]
  # ベアエントリが除去されていること
  local count
  count=$(jq '[.permissions.allow[] | select(. == "Read" or . == "Write" or . == "Edit")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 0 ]
}

# =============================================================================
# WebFetch(:*) 等のワイルドカード除去
# =============================================================================

@test "WebFetch(:*) と Fetch( *) が除去される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["WebFetch(:*)", "Fetch( *)", "WebFetch(domain:example.com)", "NotebookEdit(*)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  local result
  result=$(jq -c '.permissions.allow' "$MOCK_PROJECT/.claude/settings.local.json")
  # 正当なドメイン指定は保持
  [[ "$result" == *'"WebFetch(domain:example.com)"'* ]]
  # ワイルドカードは除去
  local count
  count=$(jq '[.permissions.allow[] | select(. == "WebFetch(:*)" or . == "Fetch( *)" or . == "NotebookEdit(*)")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 0 ]
}

# =============================================================================
# Glob, Grep のワイルドカード除去
# =============================================================================

@test "Glob(*) と Grep(*) が除去される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Glob(*)", "Grep(*)", "Glob", "Grep"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  local count
  count=$(jq '.permissions.allow | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 0 ]
}

# =============================================================================
# mcp__* 除去 (サーバー固有は保持)
# =============================================================================

@test "mcp__* が除去される (mcp__obsidian__* は保持)" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["mcp__*", "mcp__obsidian__*", "mcp__orm-discovery-mcp-go__*"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  local result
  result=$(jq -c '.permissions.allow' "$MOCK_PROJECT/.claude/settings.local.json")
  # mcp__* は除去
  local count
  count=$(jq '[.permissions.allow[] | select(. == "mcp__*")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 0 ]
  # サーバー固有は保持
  count=$(jq '[.permissions.allow[] | select(. == "mcp__obsidian__*")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
  count=$(jq '[.permissions.allow[] | select(. == "mcp__orm-discovery-mcp-go__*")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
}

# =============================================================================
# permissions キー不在
# =============================================================================

@test "permissions キー不在: 正常終了" {
  create_settings_local "$MOCK_PROJECT" '{"model":"sonnet"}'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"SKIP"* ]]
}

@test "permissions.allow キー不在: 正常終了" {
  create_settings_local "$MOCK_PROJECT" '{"permissions":{"deny":["Bash(rm:*)"]}}'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"SKIP"* ]]
}

# =============================================================================
# グローバル settings.local.json からも除去
# =============================================================================

@test "グローバル settings.local.json からも除去される" {
  create_settings_local "$MOCK_HOME" '{
    "permissions": {
      "allow": ["Bash(*)", "Write(*)", "Bash(git:*)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  local count
  count=$(jq '[.permissions.allow[] | select(. == "Bash(*)" or . == "Write(*)")] | length' "$MOCK_HOME/.claude/settings.local.json")
  [ "$count" -eq 0 ]
  # 正当なエントリは保持
  count=$(jq '[.permissions.allow[] | select(. == "Bash(git:*)")] | length' "$MOCK_HOME/.claude/settings.local.json")
  [ "$count" -eq 1 ]
}

# =============================================================================
# ログ出力
# =============================================================================

@test "除去時に REMOVE ログが出力される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Bash(*)", "Read"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"REMOVE"* ]]
  [[ "$output" == *"Bash(*)"* ]]
  [[ "$output" == *"Read"* ]]
}

@test "除去対象がない場合は SKIP ログが出力される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Bash(git status:*)", "Read(~/src/**)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"SKIP"* ]]
  [[ "$output" == *"no broad wildcards"* ]]
}

# =============================================================================
# dry-run モード
# =============================================================================

@test "dry-run: 除去対象があっても実際には変更されない" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Bash(*)", "Bash(git:*)"]
    }
  }'

  run env DRY_RUN=1 bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  [[ "$output" == *"REMOVE"* ]]
  # ファイルは変更されていないこと
  local count
  count=$(jq '[.permissions.allow[] | select(. == "Bash(*)")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
}

# =============================================================================
# 複合テスト: プロジェクトとグローバル両方
# =============================================================================

@test "プロジェクトとグローバル両方から除去される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Edit(*)", "Bash(ls:*)"]
    }
  }'
  create_settings_local "$MOCK_HOME" '{
    "permissions": {
      "allow": ["Write", "Read(~/src/**)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  # プロジェクト側: Edit(*) が除去、Bash(ls:*) は保持
  local count
  count=$(jq '.permissions.allow | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
  count=$(jq '[.permissions.allow[] | select(. == "Bash(ls:*)")] | length' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
  # グローバル側: Write が除去、Read(~/src/**) は保持
  count=$(jq '.permissions.allow | length' "$MOCK_HOME/.claude/settings.local.json")
  [ "$count" -eq 1 ]
  count=$(jq '[.permissions.allow[] | select(. == "Read(~/src/**)")] | length' "$MOCK_HOME/.claude/settings.local.json")
  [ "$count" -eq 1 ]
}

# =============================================================================
# エラーパス: 不正な JSON
# =============================================================================

@test "settings.local.json が不正な JSON: エラー終了" {
  echo '{"invalid json' > "$MOCK_PROJECT/.claude/settings.local.json"

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -ne 0 ]
  [[ "$output" == *"ERROR"* ]]
  [[ "$output" == *"not valid JSON"* ]]
}

@test "フック入力が不正な JSON: エラー終了" {
  run bash "$SCRIPT_PATH" <<< "not json at all"

  [ "$status" -ne 0 ]
  [[ "$output" == *"ERROR"* ]]
}

# =============================================================================
# エラーパス: 非文字列エントリ
# =============================================================================

@test "permissions.allow に非文字列エントリが混在しても Bash(*) が除去される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": [1, null, "Bash(*)", "Bash(git:*)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  # Bash(*) が除去されていること
  local count
  count=$(jq '[.permissions.allow[] | select(type == "string" and . == "Bash(*)")] | length' \
    "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 0 ]
  # 正当な文字列エントリは保持
  count=$(jq '[.permissions.allow[] | select(type == "string" and . == "Bash(git:*)")] | length' \
    "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 1 ]
  # 非文字列エントリも保持
  count=$(jq '[.permissions.allow[] | select(type != "string")] | length' \
    "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$count" -eq 2 ]
}

# =============================================================================
# cwd 未指定: グローバルのみ処理
# =============================================================================

@test "cwd が指定されていない場合: グローバル settings.local.json のみ除去される" {
  create_settings_local "$MOCK_HOME" '{
    "permissions": {
      "allow": ["Bash(*)"]
    }
  }'

  run bash "$SCRIPT_PATH" <<< '{}'

  [ "$status" -eq 0 ]
  local count
  count=$(jq '[.permissions.allow[] | select(. == "Bash(*)")] | length' \
    "$MOCK_HOME/.claude/settings.local.json")
  [ "$count" -eq 0 ]
}

# =============================================================================
# 除去後の他キー保持
# =============================================================================

@test "除去後に他の設定キー (model 等) が保持される" {
  create_settings_local "$MOCK_PROJECT" '{
    "permissions": {
      "allow": ["Bash(*)", "Bash(git:*)"]
    },
    "model": "sonnet",
    "theme": "dark"
  }'

  run bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}"

  [ "$status" -eq 0 ]
  local model theme
  model=$(jq -r '.model' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$model" = "sonnet" ]
  theme=$(jq -r '.theme' "$MOCK_PROJECT/.claude/settings.local.json")
  [ "$theme" = "dark" ]
}

# =============================================================================
# stdout が空であることの検証
# =============================================================================

@test "正常終了時: stdout は空である" {
  create_settings_local "$MOCK_PROJECT" '{"permissions":{"allow":["Bash(git:*)"]}}'

  stdout=$(bash "$SCRIPT_PATH" <<< "{\"cwd\":\"$MOCK_PROJECT\"}" 2>/dev/null)

  [ -z "$stdout" ]
}
