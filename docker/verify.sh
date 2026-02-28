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
  --rebuild   Dockerイメージを強制再ビルド
  --help      このヘルプを表示

Examples:
  $(basename "$0")                           # bashシェルを起動
  $(basename "$0") claude --version          # claude --version を実行
  $(basename "$0") claude -p "Reply OK"      # claudeにプロンプトを送信
  $(basename "$0") --rebuild                 # イメージを再ビルドしてbash起動
EOF
}

# オプション解析
FORCE_REBUILD=false
ARGS=()
for arg in "$@"; do
  case "$arg" in
    --rebuild) FORCE_REBUILD=true ;;
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

# NOTE: curl | bash によるインストールはサプライチェーンリスクがある。
# POCとしてはネイティブインストーラーを使用するが、
# 本番運用ではバージョン固定+チェックサム検証を推奨する。
if [ ! -x /home/claude/.local/bin/claude ]; then
  su -s /bin/bash claude -c 'curl -fsSL https://claude.ai/install.sh | bash'
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
    git curl ca-certificates jq bash && rm -rf /var/lib/apt/lists/*
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

  # skills/
  mkdir -p "$STAGING_DIR/skills"
  for skill_dir in "$REPO_ROOT/dotclaude/skills"/*/; do
    [ -d "$skill_dir" ] || continue
    cp -a "$skill_dir" "$STAGING_DIR/skills/$(basename "$skill_dir")"
  done
}

# 3. イメージビルド (必要な場合のみ)
if [ "$FORCE_REBUILD" = true ] || ! docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
  build_image
fi

# 4. 設定同期 + config volumeリセット (変更を確実に反映)
sync_config
docker volume create "$VOLUME_NAME" > /dev/null 2>&1 || true
docker volume rm "$CONFIG_VOLUME_NAME" > /dev/null 2>&1 || true
docker volume create "$CONFIG_VOLUME_NAME" > /dev/null 2>&1 || true

# ホスト環境変数を透過 (DISABLE_AUTOUPDATERはデフォルトで有効)
ENV_FLAGS=(-e "DISABLE_AUTOUPDATER=${DISABLE_AUTOUPDATER:-1}")
for var in CLAUDE_CODE_USE_VERTEX ANTHROPIC_VERTEX_PROJECT_ID \
  CLOUD_ML_REGION ANTHROPIC_MODEL \
  DISABLE_PROMPT_CACHING ANTHROPIC_DEFAULT_HAIKU_MODEL \
  ANTHROPIC_DEFAULT_OPUS_MODEL ANTHROPIC_DEFAULT_SONNET_MODEL; do
  [ -n "${!var:-}" ] && ENV_FLAGS+=(-e "$var=${!var}")
done

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
