#!/bin/bash
set -euo pipefail

IMAGE_NAME="claude-config-verify"
VOLUME_NAME="claude-config-verify-local"
CONFIG_VOLUME_NAME="claude-config-verify-dotclaude"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STAGING_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/claude-config-verify"

usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS] [COMMAND...]

Dockerコンテナ内でclaude-config設定を検証する。

Options:
  --check     一括検証を実行し構造化結果を出力
  --rebuild   Dockerイメージを強制再ビルド
  --help      このヘルプを表示

Examples:
  $(basename "$0")                           # bashシェルを起動
  $(basename "$0") --check                   # 一括検証を実行
  $(basename "$0") claude --version          # claude --version を実行
  $(basename "$0") claude -p "Reply OK"      # claudeにプロンプトを送信
  $(basename "$0") --rebuild                 # イメージを再ビルドしてbash起動
EOF
}

# オプション解析
FORCE_REBUILD=false
RUN_CHECK=false
ARGS=()
for arg in "$@"; do
  case "$arg" in
    --rebuild) FORCE_REBUILD=true ;;
    --check) RUN_CHECK=true ;;
    --help) usage; exit 0 ;;
    *) ARGS+=("$arg") ;;
  esac
done
set -- "${ARGS[@]+"${ARGS[@]}"}"

# 1. Dockerイメージのビルド (Dockerfile動的生成)
build_image() {
  tmpdir=$(mktemp -d)

  # entrypoint.sh を生成
  cat > "$tmpdir/entrypoint.sh" << 'ENTRYPOINT_EOF'
#!/bin/bash
set -e
# Docker volume権限をclaudeユーザーに設定 (インストール前に必要)
chown -R claude:claude /home/claude/.local
chown -R claude:claude /home/claude/.claude

# read-onlyのstaging dirからDocker volumeに設定をコピー
# .claude.json は ~/.claude/ 内ではなく ~/ 直下に配置するため除外
if [ -d /staging ]; then
  find /staging -mindepth 1 -maxdepth 1 -not -name '.claude.json' \
    -exec cp -a {} /home/claude/.claude/ \;
  chown -R claude:claude /home/claude/.claude
  if [ -f /staging/.claude.json ]; then
    cp /staging/.claude.json /home/claude/.claude.json
    chown claude:claude /home/claude/.claude.json
  fi
fi

# セッションJSONL (過去の会話履歴) をsymlink
if [ -d /sessions ]; then
  ln -sf /sessions /home/claude/.claude/projects
fi

# ローカル環境と同じバージョンをネイティブインストーラーでインストール
INSTALLED_VERSION=$(/home/claude/.local/bin/claude --version 2>/dev/null | awk '{print $1}' || echo "none")
if [ "$INSTALLED_VERSION" != "${CLAUDE_VERSION:-}" ]; then
  su -s /bin/bash claude -c "curl -fsSL https://claude.ai/install.sh | bash -s ${CLAUDE_VERSION}"
fi

# claudeユーザーで実行
# claudeコマンドにはDocker内で安全な --dangerously-skip-permissions を自動付与
if [ "${1:-}" = "claude" ]; then
  shift
  exec setpriv --reuid="$(id -u claude)" --regid="$(id -g claude)" \
    --init-groups env HOME=/home/claude claude --dangerously-skip-permissions "$@"
else
  exec setpriv --reuid="$(id -u claude)" --regid="$(id -g claude)" \
    --init-groups env HOME=/home/claude "$@"
fi
ENTRYPOINT_EOF

  # Dockerfile を生成
  cat > "$tmpdir/Dockerfile" << 'DOCKERFILE_EOF'
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl ca-certificates jq bash file && rm -rf /var/lib/apt/lists/*
RUN useradd -m -s /bin/bash claude
ENV PATH="/home/claude/.local/bin:${PATH}"
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
WORKDIR /workspace
ENTRYPOINT ["/entrypoint.sh"]
CMD ["bash"]
DOCKERFILE_EOF

  docker build -t "$IMAGE_NAME" "$tmpdir"
  rm -rf "$tmpdir"
}

# 2. リポジトリのファイルをstaging dirに展開 (コンテナにはread-onlyでマウント)
sync_config() {
  # staging dirをクリアして最新状態にする
  rm -rf "$STAGING_DIR"
  mkdir -p "$STAGING_DIR"

  # .claude.json: ホストからコピー (オンボーディング済み状態を引き継ぐ)
  if [ -f "$HOME/.claude.json" ]; then
    cp "$HOME/.claude.json" "$STAGING_DIR/.claude.json"
  else
    echo '{"hasCompletedOnboarding":true,"theme":"dark"}' > "$STAGING_DIR/.claude.json"
  fi

  # CLAUDE-global.md → CLAUDE.md (名前マッピング)
  cp "$REPO_ROOT/dotclaude/CLAUDE-global.md" "$STAGING_DIR/CLAUDE.md"

  # settings.json: MCP plugins無効化 + statusLine削除
  jq '.enabledPlugins |= map_values(false) | del(.statusLine)' \
    "$REPO_ROOT/dotclaude/settings.json" > "$STAGING_DIR/settings.json"

  # env.sh: 実ファイルがあればそれを、なければ env.sh.example、どちらもなければ空ファイル
  if [ -f "$REPO_ROOT/dotclaude/env.sh" ]; then
    cp "$REPO_ROOT/dotclaude/env.sh" "$STAGING_DIR/env.sh"
  elif [ -f "$REPO_ROOT/dotclaude/env.sh.example" ]; then
    cp "$REPO_ROOT/dotclaude/env.sh.example" "$STAGING_DIR/env.sh"
  else
    : > "$STAGING_DIR/env.sh"
  fi

  # hooks/
  cp -a "$REPO_ROOT/dotclaude/hooks" "$STAGING_DIR/hooks"

  # bin/ - Linux 用クロスコンパイル済みバイナリを優先
  # docker/bin/ (Linux用) > dotclaude/bin/ (ホスト用、フォールバック)
  if [ -d "$REPO_ROOT/docker/bin" ]; then
    cp -a "$REPO_ROOT/docker/bin" "$STAGING_DIR/bin"
  elif [ -d "$REPO_ROOT/dotclaude/bin" ]; then
    echo "WARNING: docker/bin/ not found, falling back to dotclaude/bin/ (may be wrong architecture)" >&2
    cp -a "$REPO_ROOT/dotclaude/bin" "$STAGING_DIR/bin"
  fi

  # skills/
  mkdir -p "$STAGING_DIR/skills"
  for skill_dir in "$REPO_ROOT/dotclaude/skills"/*/; do
    [ -d "$skill_dir" ] || continue
    cp -a "$skill_dir" "$STAGING_DIR/skills/$(basename "$skill_dir")"
  done
}

# 3. 一括検証 (--check モード)
run_checks() {
  local pass=0
  local fail=0
  local total=0

  check() {
    local _section="$1" # Used for categorization context
    local name="$2"
    local result="$3"
    total=$((total + 1))
    if [ "$result" = "OK" ]; then
      pass=$((pass + 1))
      echo "[OK]   $name"
    else
      fail=$((fail + 1))
      echo "[FAIL] $name: $result"
    fi
  }

  # コンテナ内で検証スクリプトを実行
  local check_output
  check_output=$(docker run --rm -i \
    -v "$VOLUME_NAME:/home/claude/.local" \
    -v "$CONFIG_VOLUME_NAME:/home/claude/.claude" \
    -v "$STAGING_DIR:/staging:ro" \
    "${ENV_FLAGS[@]}" \
    "$IMAGE_NAME" \
    sh -c '
      HOME=/home/claude

      echo "=== STRUCTURE ==="
      for f in CLAUDE.md settings.json env.sh; do
        if [ -f "$HOME/.claude/$f" ]; then
          echo "OK:$f"
        else
          echo "FAIL:$f:not found"
        fi
      done
      for d in hooks skills; do
        if [ -d "$HOME/.claude/$d" ]; then
          echo "OK:$d/"
        else
          echo "FAIL:$d/:not found"
        fi
      done

      echo "=== SETTINGS ==="
      if jq empty "$HOME/.claude/settings.json" 2>/dev/null; then
        echo "OK:valid JSON"
        # MCP plugins が無効化されているか
        disabled=$(jq -r ".enabledPlugins | to_entries | map(select(.value == true)) | length" "$HOME/.claude/settings.json" 2>/dev/null)
        if [ "${disabled:-1}" = "0" ]; then
          echo "OK:MCP plugins disabled"
        else
          echo "FAIL:MCP plugins:${disabled} plugins still enabled"
        fi
        # statusLine が削除されているか
        has_statusline=$(jq -r "has(\"statusLine\")" "$HOME/.claude/settings.json" 2>/dev/null)
        if [ "$has_statusline" = "false" ]; then
          echo "OK:statusLine removed"
        else
          echo "FAIL:statusLine:not removed"
        fi
        # permissions が存在するか
        has_perms=$(jq -r "has(\"permissions\")" "$HOME/.claude/settings.json" 2>/dev/null)
        if [ "$has_perms" = "true" ]; then
          echo "OK:permissions present"
        else
          echo "FAIL:permissions:not found"
        fi
      else
        echo "FAIL:settings.json:invalid JSON"
      fi

      echo "=== ENV ==="
      if [ -f "$HOME/.claude/env.sh" ]; then
        if bash -n "$HOME/.claude/env.sh" 2>/dev/null; then
          echo "OK:env.sh syntax valid"
        else
          echo "FAIL:env.sh:syntax error"
        fi
      else
        echo "FAIL:env.sh:not found"
      fi

      echo "=== HOOKS ==="
      if [ -d "$HOME/.claude/hooks" ]; then
        hook_count=$(find "$HOME/.claude/hooks" -name "*.sh" -type f | wc -l)
        echo "OK:${hook_count} hook scripts found"
        # hooks のシンタックスチェック
        find "$HOME/.claude/hooks" -name "*.sh" -type f | while read -r hook; do
          name=$(basename "$hook")
          if bash -n "$hook" 2>/dev/null; then
            echo "OK:hook:${name} syntax valid"
          else
            echo "FAIL:hook:${name} syntax error"
          fi
        done
        # hooks/lib/ の存在確認
        if [ -d "$HOME/.claude/hooks/lib" ]; then
          echo "OK:hooks/lib/ present"
        else
          echo "FAIL:hooks/lib/:not found"
        fi
      else
        echo "FAIL:hooks/:not found"
      fi

      echo "=== BIN ==="
      if [ -d "$HOME/.claude/bin" ]; then
        echo "OK:bin/ present"
        if [ -f "$HOME/.claude/bin/realpath" ]; then
          # バイナリのアーキテクチャ確認
          arch_info=$(file "$HOME/.claude/bin/realpath" 2>/dev/null)
          if echo "$arch_info" | grep -qE "ELF.*executable"; then
            echo "OK:realpath is Linux binary"
            # 実行可能か確認
            if "$HOME/.claude/bin/realpath" --help > /dev/null 2>&1; then
              echo "OK:realpath executable"
            elif "$HOME/.claude/bin/realpath" / > /dev/null 2>&1; then
              echo "OK:realpath executable"
            else
              echo "FAIL:realpath:not executable or runtime error"
            fi
          else
            echo "FAIL:realpath:wrong architecture ($arch_info)"
          fi
        else
          echo "FAIL:realpath:not found"
        fi
      else
        echo "FAIL:bin/:not found"
      fi

      echo "=== TOOLS ==="
      for tool in realpath analyze-permissions analyze-webfetch analyze-tokens; do
        if [ -f "$HOME/.claude/bin/$tool" ]; then
          arch_info=$(file "$HOME/.claude/bin/$tool" 2>/dev/null)
          if echo "$arch_info" | grep -qE "ELF.*executable"; then
            echo "OK:$tool is Linux binary"
            if "$HOME/.claude/bin/$tool" --help > /dev/null 2>&1 || "$HOME/.claude/bin/$tool" / > /dev/null 2>&1; then
              echo "OK:$tool executable"
            else
              echo "FAIL:$tool:not executable or runtime error"
            fi
          else
            echo "FAIL:$tool:wrong architecture ($arch_info)"
          fi
        else
          echo "FAIL:$tool:not found in bin/"
        fi
      done

      echo "=== SKILLS ==="
      if [ -d "$HOME/.claude/skills" ]; then
        skill_count=$(find "$HOME/.claude/skills" -maxdepth 1 -mindepth 1 -type d | wc -l)
        echo "OK:${skill_count} skills found"
        find "$HOME/.claude/skills" -maxdepth 1 -mindepth 1 -type d | while read -r skill; do
          name=$(basename "$skill")
          if [ -f "$skill/SKILL.md" ]; then
            echo "OK:skill:${name} has SKILL.md"
          else
            echo "FAIL:skill:${name} missing SKILL.md"
          fi
        done
      else
        echo "FAIL:skills/:not found"
      fi

      echo "=== HOOKS_DEPS ==="
      # settings.json の hooks コマンドパスが実在するか確認
      # $HOME を実際のパスに展開して検証
      jq -r ".hooks // {} | .. | .command? // empty" "$HOME/.claude/settings.json" 2>/dev/null | \
        grep "\.sh$" | sort -u | while read -r cmd; do
          resolved=$(echo "$cmd" | sed "s|\\\$HOME|$HOME|g; s|~|$HOME|g")
          name=$(basename "$resolved")
          if [ -f "$resolved" ]; then
            echo "OK:hook ref:${name}"
          else
            echo "FAIL:hook ref:${name} (${resolved} not found)"
          fi
        done
    ')

  # 出力を解析して構造化表示
  local current_section=""
  while IFS= read -r line; do
    case "$line" in
      "=== "*)
        current_section="${line//=== /}"
        current_section="${current_section// ===/}"
        echo ""
        echo "=== $current_section ==="
        ;;
      "OK:"*)
        local detail="${line#OK:}"
        check "$current_section" "$detail" "OK"
        ;;
      "FAIL:"*)
        local detail="${line#FAIL:}"
        local name="${detail%%:*}"
        local reason="${detail#*:}"
        check "$current_section" "$name" "$reason"
        ;;
    esac
  done <<< "$check_output"

  echo ""
  echo "=== SUMMARY ==="
  echo "Total: $total checks, $pass passed, $fail failed"

  if [ "$fail" -gt 0 ]; then
    return 1
  fi
  return 0
}

# 4. イメージビルド (必要な場合のみ)
if [ "$FORCE_REBUILD" = true ] || ! docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
  build_image
fi

# 5. 設定同期 + config volumeリセット (変更を確実に反映)
sync_config
docker volume create "$VOLUME_NAME" > /dev/null 2>&1 || true
docker volume rm "$CONFIG_VOLUME_NAME" > /dev/null 2>&1 || true
docker volume create "$CONFIG_VOLUME_NAME" > /dev/null 2>&1 || true

# ローカルのClaude Codeバージョンを検出してコンテナに渡す
CLAUDE_VERSION=$(claude --version 2>/dev/null | awk '{print $1}' || echo "latest")

# ホスト環境変数を透過 (DISABLE_AUTOUPDATERはデフォルトで有効)
ENV_FLAGS=(-e "DISABLE_AUTOUPDATER=${DISABLE_AUTOUPDATER:-1}" -e "CLAUDE_VERSION=$CLAUDE_VERSION")
for var in CLAUDE_CODE_USE_VERTEX ANTHROPIC_VERTEX_PROJECT_ID \
  CLOUD_ML_REGION ANTHROPIC_MODEL \
  DISABLE_PROMPT_CACHING ANTHROPIC_DEFAULT_HAIKU_MODEL \
  ANTHROPIC_DEFAULT_OPUS_MODEL ANTHROPIC_DEFAULT_SONNET_MODEL; do
  [ -n "${!var:-}" ] && ENV_FLAGS+=(-e "$var=${!var}")
done

# --check モード: 一括検証を実行して終了
if [ "$RUN_CHECK" = true ]; then
  run_checks
  exit $?
fi

# docker run引数を構築
DOCKER_ARGS=(--rm -i
  -v "$VOLUME_NAME:/home/claude/.local"
  -v "$CONFIG_VOLUME_NAME:/home/claude/.claude"
  -v "$STAGING_DIR:/staging:ro"
)

# TTYが利用可能な場合のみ-tを付与
if [ -t 0 ]; then
  DOCKER_ARGS+=(-t)
fi

# gcloudの認証情報ディレクトリが存在する場合のみマウント
if [ -d "$HOME/.config/gcloud" ]; then
  DOCKER_ARGS+=(-v "$HOME/.config/gcloud:/home/claude/.config/gcloud:ro")
fi

# セッションJSONL (過去の会話履歴) をマウント
if [ -d "$HOME/.claude/projects" ]; then
  DOCKER_ARGS+=(-v "$HOME/.claude/projects:/sessions:ro")
fi

exec docker run "${DOCKER_ARGS[@]}" \
  "${ENV_FLAGS[@]}" \
  "$IMAGE_NAME" \
  "${@:-bash}"
