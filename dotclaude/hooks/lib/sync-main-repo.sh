#!/bin/bash
# sync-main-repo.sh
# メインリポジトリを最新に同期するための共有ライブラリ
#
# 提供関数:
#   resolve_main_repo <worktree-path>  worktree パスから親リポジトリルートを stdout に出力
#   sync_main_repo <main-repo-path>    fetch + ff-only merge でメインリポジトリを同期
#
# 依存: hook-logger.sh (is_dry_run, log_info, log_error, log_skip, logged_cmd)

# worktree パスから親リポジトリのルートディレクトリを解決する
# .git ファイル → gitdir → commondir → 親リポジトリルート
# 通常リポジトリ (.git がディレクトリ) の場合は return 1
# bare リポジトリにも対応: commondir のベース名が .git でなければ bare とみなす
resolve_main_repo() {
  local worktree_path="$1"
  local git_file="$worktree_path/.git"

  # .git がファイルでなければ worktree ではない
  if [ ! -f "$git_file" ]; then
    return 1
  fi

  # gitdir パスを取得
  local git_dir
  git_dir=$(sed 's/^gitdir: //' "$git_file" | tr -d '\n')
  if [ ! -d "$git_dir" ]; then
    return 1
  fi

  # commondir ファイルから親 .git ディレクトリを特定
  local common_dir_file="$git_dir/commondir"
  if [ ! -f "$common_dir_file" ]; then
    return 1
  fi

  local common_rel
  common_rel=$(tr -d '\n' < "$common_dir_file")

  # 相対パスを絶対パスに変換
  local common_abs
  if [[ "$common_rel" == /* ]]; then
    common_abs="$common_rel"
  else
    common_abs="$(cd "$git_dir" && cd "$common_rel" && pwd)"
  fi

  # bare リポジトリ判定: commondir のベース名が .git なら通常リポジトリ
  local base
  base=$(basename "$common_abs")
  if [ "$base" = ".git" ]; then
    # 通常リポジトリ: .git ディレクトリの親がリポジトリルート
    dirname "$common_abs"
  else
    # bare リポジトリ: commondir 自体がリポジトリルート
    echo "$common_abs"
  fi
  return 0
}

# メインリポジトリを fetch + ff-only merge で同期する
# 実際の失敗 (fetch, merge, config) は return 1 で呼び出し元に伝播する
# skip 条件 (detached HEAD, feature ブランチ中, デフォルトブランチ不在) は return 0
sync_main_repo() {
  local main_repo="$1"

  if [ ! -d "$main_repo" ]; then
    log_error "sync: invalid path: $main_repo"
    return 1
  fi

  # remote.origin.fetch が空なら設定 (worktree 環境の既知問題)
  local fetch_config
  fetch_config=$(git -C "$main_repo" config --get remote.origin.fetch 2>/dev/null || true)
  if [ -z "$fetch_config" ]; then
    log_info "sync: setting remote.origin.fetch"
    if is_dry_run; then
      log_info "sync: CMD    git -C $main_repo config remote.origin.fetch +refs/heads/*:refs/remotes/origin/*"
    else
      if ! git -C "$main_repo" config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*" 2>/dev/null; then
        log_error "sync: failed to set remote.origin.fetch"
        return 1
      fi
    fi
  fi

  # git fetch origin
  log_info "sync: fetching origin"
  if is_dry_run; then
    log_info "sync: CMD    git -C $main_repo fetch origin"
  else
    if ! git -C "$main_repo" fetch origin 2>/dev/null; then
      log_error "sync: fetch failed"
      return 1
    fi
  fi

  # デフォルトブランチを検出 (main → master の順)
  local default_branch=""
  if is_dry_run; then
    default_branch="main"
    log_info "sync: CMD    git -C $main_repo show-ref --verify refs/heads/main"
  else
    if git -C "$main_repo" show-ref --verify --quiet "refs/heads/main" 2>/dev/null; then
      default_branch="main"
    elif git -C "$main_repo" show-ref --verify --quiet "refs/heads/master" 2>/dev/null; then
      default_branch="master"
    else
      log_skip "sync: no default branch found (main/master)"
      return 0
    fi
  fi

  # bare リポジトリ判定
  local is_bare="false"
  if ! is_dry_run; then
    is_bare=$(git -C "$main_repo" rev-parse --is-bare-repository 2>/dev/null) || {
      log_info "sync: could not determine bare status, assuming non-bare"
      is_bare="false"
    }
  fi

  if [ "$is_bare" = "true" ]; then
    # bare リポジトリ: merge できないので update-ref でブランチを進める
    # ローカルブランチが origin の先祖であることを確認 (ff-only 相当)
    log_info "sync: updating ref refs/heads/$default_branch (bare repo)"
    if is_dry_run; then
      log_info "sync: CMD    git -C $main_repo update-ref refs/heads/$default_branch refs/remotes/origin/$default_branch"
    else
      local local_ref remote_ref
      local_ref=$(git -C "$main_repo" rev-parse "refs/heads/$default_branch" 2>/dev/null) || {
        log_info "sync: local ref $default_branch not found (new branch?)"
        local_ref=""
      }
      remote_ref=$(git -C "$main_repo" rev-parse "refs/remotes/origin/$default_branch" 2>/dev/null) || {
        log_info "sync: remote ref origin/$default_branch not found"
        remote_ref=""
      }
      if [ -z "$remote_ref" ]; then
        log_skip "sync: remote ref origin/$default_branch not found"
        return 0
      fi
      if [ "$local_ref" = "$remote_ref" ]; then
        log_info "sync: already up to date"
        return 0
      fi
      # ff-only チェック: local が remote の先祖かどうか
      if [ -n "$local_ref" ] && ! git -C "$main_repo" merge-base --is-ancestor "$local_ref" "$remote_ref" 2>/dev/null; then
        log_error "sync: update-ref failed (diverged?)"
        return 1
      fi
      if ! git -C "$main_repo" update-ref "refs/heads/$default_branch" "$remote_ref" 2>/dev/null; then
        log_error "sync: update-ref failed"
        return 1
      fi
    fi
  else
    # 通常リポジトリ: merge --ff-only で進める
    # 現在のブランチを確認
    local current_branch
    if is_dry_run; then
      current_branch="$default_branch"
      log_info "sync: CMD    git -C $main_repo symbolic-ref --short HEAD"
    else
      current_branch=$(git -C "$main_repo" symbolic-ref --short HEAD 2>/dev/null || true)
      if [ -z "$current_branch" ]; then
        log_skip "sync: detached HEAD"
        return 0
      fi
    fi

    # デフォルトブランチをチェックアウト中でなければスキップ
    if [ "$current_branch" != "$default_branch" ]; then
      log_skip "sync: not on default branch (on $current_branch)"
      return 0
    fi

    # ff-only merge
    log_info "sync: merging origin/$default_branch"
    if is_dry_run; then
      log_info "sync: CMD    git -C $main_repo merge --ff-only origin/$default_branch"
    else
      if ! git -C "$main_repo" merge --ff-only "origin/$default_branch" 2>/dev/null; then
        log_error "sync: merge --ff-only failed (diverged?)"
        return 1
      fi
    fi
  fi

  log_info "sync: main repo synced successfully"
  return 0
}
