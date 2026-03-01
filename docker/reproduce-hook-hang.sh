#!/bin/bash
set -euo pipefail

# reproduce-hook-hang.sh
# worktree-create.sh のハング問題を Docker コンテナ内で再現・診断する
#
# テスト条件:
#   1. remote あり + ネットワーク接続あり (通常環境)
#   2. remote あり + ネットワーク切断 (--network none)
#   3. remote なし (resolve_main_repo の fallback)

IMAGE_NAME="claude-config-verify"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STAGING_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/claude-config-reproduce"

usage() {
  printf '%s\n' "Usage: $(basename "$0") [OPTIONS]"
  printf '%s\n' ""
  printf '%s\n' "worktree-create.sh のハング問題を Docker コンテナ内で再現する"
  printf '%s\n' ""
  printf '%s\n' "Options:"
  printf '%s\n' "  --rebuild   Docker イメージを強制再ビルド"
  printf '%s\n' "  --timeout N SYNC_FETCH_TIMEOUT の秒数を指定 (デフォルト: 5)"
  printf '%s\n' "  --help      このヘルプを表示"
}

# オプション解析
FORCE_REBUILD=false
FETCH_TIMEOUT=5
for arg in "$@"; do
  case "$arg" in
    --rebuild) FORCE_REBUILD=true ;;
    --timeout)
      shift
      FETCH_TIMEOUT="$1"
      ;;
    --help) usage; exit 0 ;;
  esac
  shift 2>/dev/null || true
done

# イメージビルド (verify.sh と同じイメージを使用)
if [ "$FORCE_REBUILD" = true ] || ! docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
  printf '%s\n' "=== Building Docker image ==="
  "$REPO_ROOT/docker/verify.sh" --rebuild --check || true
fi

# hooks をステージングにコピー
prepare_staging() {
  rm -rf "$STAGING_DIR"
  mkdir -p "$STAGING_DIR/hooks/lib"
  cp "$REPO_ROOT/dotclaude/hooks/worktree-create.sh" "$STAGING_DIR/hooks/"
  cp "$REPO_ROOT/dotclaude/hooks/lib/hook-logger.sh" "$STAGING_DIR/hooks/lib/"
  cp "$REPO_ROOT/dotclaude/hooks/lib/sync-main-repo.sh" "$STAGING_DIR/hooks/lib/"

  # worktree-memory-load.sh のスタブ (テストには不要)
  cat > "$STAGING_DIR/hooks/worktree-memory-load.sh" << 'EOF'
#!/bin/bash
exit 0
EOF
  chmod +x "$STAGING_DIR/hooks/worktree-memory-load.sh"
  chmod +x "$STAGING_DIR/hooks/worktree-create.sh"
}

# コンテナ内で実行するテストスクリプト (変数展開はコンテナ内で行う)
# shellcheck disable=SC2016
CONTAINER_SCRIPT='
set -euo pipefail

FETCH_TIMEOUT="${SYNC_FETCH_TIMEOUT:-5}"
HOOKS_DIR="/staging/hooks"

printf "%s\n" "=== Test setup ==="
cd /tmp
git init test-repo >/dev/null 2>&1
cd test-repo
git config user.email "test@test.com"
git config user.name "test"
git commit --allow-empty -m "init" >/dev/null 2>&1

setup_remote() {
  git remote add origin https://github.com/usadamasa/claude-config.git 2>/dev/null || true
}

remove_remote() {
  git remote remove origin 2>/dev/null || true
}

printf "%s\n" ""
printf "%s\n" "=== Running worktree-create.sh with bash -x ==="
printf "%s\n" "SYNC_FETCH_TIMEOUT=$FETCH_TIMEOUT"
printf "%s\n" ""

# git wt は不在なので、sync_main_repo までの動作を確認する
# worktree-create.sh は git wt がなくても sync 部分は実行される

if [ "${TEST_CONDITION:-}" = "no-remote" ]; then
  remove_remote
  printf "%s\n" "--- Condition: no remote ---"
else
  setup_remote
  printf "%s\n" "--- Condition: remote present ---"
fi

export SCRIPT_DIR="$HOOKS_DIR"
export SYNC_FETCH_TIMEOUT="$FETCH_TIMEOUT"
export DRY_RUN=0

printf "%s\n" ""
printf "%s\n" "--- bash -x trace start ---"

# worktree-create.sh 全体を実行すると git wt が必要になるため、
# sync 部分のみを直接テストする
bash -x -c "
source \"$HOOKS_DIR/lib/hook-logger.sh\"
source \"$HOOKS_DIR/lib/sync-main-repo.sh\"
export HOOK_NAME=reproduce-test
export SYNC_FETCH_TIMEOUT=$FETCH_TIMEOUT

# resolve_main_repo は .git ファイルが必要 (worktree 構造)
# 通常リポジトリなので resolve は失敗 → CWD をそのまま使用
MAIN_REPO=\$(resolve_main_repo /tmp/test-repo 2>/dev/null) || MAIN_REPO=\"\"
if [ -z \"\$MAIN_REPO\" ] && [ -d /tmp/test-repo/.git ]; then
  MAIN_REPO=/tmp/test-repo
fi
printf \"%s\n\" \"MAIN_REPO=\$MAIN_REPO\"

if [ -n \"\$MAIN_REPO\" ]; then
  sync_main_repo \"\$MAIN_REPO\" && printf \"%s\n\" \"sync: SUCCESS\" || printf \"%s\n\" \"sync: FAILED (exit: \$?)\"
else
  printf \"%s\n\" \"sync: SKIPPED (no main repo)\"
fi
" 2>&1

printf "%s\n" "--- bash -x trace end ---"
printf "%s\n" ""
'

run_test() {
  local condition="$1"
  local network_flag="$2"
  local description="$3"

  printf '%s\n' "=============================================="
  printf '%s\n' "Test: $description"
  printf '%s\n' "=============================================="

  local docker_args=(--rm -i
    --entrypoint bash
    -v "$STAGING_DIR:/staging:ro"
    -e "TEST_CONDITION=$condition"
    -e "SYNC_FETCH_TIMEOUT=$FETCH_TIMEOUT"
  )

  if [ -n "$network_flag" ]; then
    docker_args+=("$network_flag")
  fi

  local exit_code=0
  timeout 60 docker run "${docker_args[@]}" "$IMAGE_NAME" \
    -c "$CONTAINER_SCRIPT" 2>&1 || exit_code=$?

  if [ "$exit_code" -eq 124 ]; then
    printf '%s\n' ""
    printf '%s\n' "*** RESULT: HUNG (timed out after 60s) ***"
  elif [ "$exit_code" -eq 0 ]; then
    printf '%s\n' ""
    printf '%s\n' "*** RESULT: COMPLETED NORMALLY ***"
  else
    printf '%s\n' ""
    printf '%s\n' "*** RESULT: FAILED (exit: $exit_code) ***"
  fi

  printf '%s\n' ""
}

# ステージング準備
prepare_staging

printf '%s\n' "Docker reproduce-hook-hang.sh"
printf '%s\n' "SYNC_FETCH_TIMEOUT=${FETCH_TIMEOUT}s"
printf '%s\n' ""

# Test 1: remote あり + ネットワーク接続あり
run_test "with-remote" "" "Remote + Network (normal)"

# Test 2: remote あり + ネットワーク切断
run_test "with-remote" "--network=none" "Remote + No Network (reproduce hang)"

# Test 3: remote なし
run_test "no-remote" "" "No Remote (fallback)"

printf '%s\n' "=============================================="
printf '%s\n' "All tests completed"
printf '%s\n' "=============================================="
