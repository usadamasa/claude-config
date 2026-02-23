# bats-core 詳細リファレンス

## 目次

- [モックの作成](#モックの作成)
- [フィクスチャ](#フィクスチャ)
- [テスタブルなコード設計](#テスタブルなコード設計)
- [bash 互換性](#bash-互換性)
- [トラブルシューティング](#トラブルシューティング)

## モックの作成

### 外部コマンドのモック

peco のような対話的コマンドをモック化する例:

```bash
#!/usr/bin/env bash
# helpers/peco-mock.sh

PECO_MOCK_SELECTION=""

peco() {
  if [ -n "$PECO_MOCK_SELECTION" ]; then
    echo "$PECO_MOCK_SELECTION"
  else
    return 0  # キャンセル
  fi
}

export -f peco
```

テストでの使用:

```bash
@test "peco で選択した値が使われる" {
  PECO_MOCK_SELECTION="feature-branch"
  run my_function
  [[ "$output" == *"feature-branch"* ]]
}

@test "peco キャンセル時にエラーにならない" {
  PECO_MOCK_SELECTION=""
  run my_function
  [ "$status" -eq 0 ]
}
```

### zsh 専用コマンドのモック

bats は bash で動作するため、zsh 専用コマンドをモック化する必要があります。

```bash
#!/usr/bin/env bash
# helpers/zle-mock.sh

# zle 関連変数
BUFFER=""
LBUFFER=""

# zle コマンドのモック
zle() {
  local cmd="$1"
  case "$cmd" in
    accept-line|reset-prompt|clear-screen|-N) ;;
    *) ;;
  esac
}

# bindkey コマンドのモック
bindkey() {
  :  # 何もしない
}

export -f zle
export -f bindkey
export BUFFER LBUFFER
```

### モックの読み込み順序

**重要**: モックはテスト対象より先に読み込む!

```bash
setup() {
  DOTFILE_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/.." && pwd)"

  # ✅ 正しい順序: モックが先
  load 'helpers/zle-mock.sh'
  load 'helpers/peco-mock.sh'
  load 'fixtures/git-setup.sh'

  # テスト対象は最後
  source "$DOTFILE_DIR/config/zsh/funcs/peco-src.sh"
}
```

## フィクスチャ

### テスト用 Git リポジトリ

dotfiles のブランチを汚さないために、一時ディレクトリにテスト用リポジトリを作成します。

```bash
#!/usr/bin/env bash
# fixtures/git-setup.sh

create_test_repo() {
  TEST_REMOTE=$(mktemp -d)
  TEST_REPO=$(mktemp -d)

  # bare リポジトリ (remote 用)
  git init --bare "$TEST_REMOTE" >/dev/null 2>&1

  # 作業リポジトリ
  git clone "$TEST_REMOTE" "$TEST_REPO" >/dev/null 2>&1
  cd "$TEST_REPO" || return 1

  # Git 設定
  git config user.email "test@example.com"
  git config user.name "Test User"

  # 初期コミット
  echo "initial" > README.md
  git add README.md
  git commit -m "Initial commit" >/dev/null 2>&1
  git push origin main >/dev/null 2>&1 || git push origin master >/dev/null 2>&1

  export TEST_REPO TEST_REMOTE
}

cleanup_test_repo() {
  [ -n "$WORKTREE_PATH" ] && rm -rf "$WORKTREE_PATH"
  [ -n "$TEST_REPO" ] && rm -rf "$TEST_REPO"
  [ -n "$TEST_REMOTE" ] && rm -rf "$TEST_REMOTE"
  unset TEST_REPO TEST_REMOTE WORKTREE_PATH
}
```

### ローカルブランチの作成

```bash
create_local_branch() {
  local branch_name="$1"
  git checkout -b "$branch_name" >/dev/null 2>&1
  echo "$branch_name" > "$branch_name.txt"
  git add "$branch_name.txt"
  git commit -m "Add $branch_name" >/dev/null 2>&1
  git checkout main >/dev/null 2>&1 || git checkout master >/dev/null 2>&1
}
```

### リモートのみのブランチ作成

ローカルには存在しない、リモートのみのブランチを作成:

```bash
create_remote_only_branch() {
  local branch_name="$1"
  local original_dir=$(pwd)
  local tmp_clone=$(mktemp -d)

  # 別クローンで作成してプッシュ
  git clone "$TEST_REMOTE" "$tmp_clone" >/dev/null 2>&1
  cd "$tmp_clone" || return 1
  git config user.email "test@example.com"
  git config user.name "Test User"
  git checkout -b "$branch_name" >/dev/null 2>&1
  echo "$branch_name" > "$branch_name.txt"
  git add "$branch_name.txt"
  git commit -m "Add $branch_name" >/dev/null 2>&1
  git push origin "$branch_name" >/dev/null 2>&1

  cd "$original_dir" || return 1
  rm -rf "$tmp_clone"

  # 元のリポジトリで fetch
  git fetch origin >/dev/null 2>&1
}
```

### worktree の作成

```bash
create_worktree() {
  local branch_name="$1"

  # まずブランチを作成
  create_local_branch "$branch_name"

  # 一意な worktree パスを生成
  local worktree_dir=$(mktemp -d)
  rm -rf "$worktree_dir"  # git worktree add が作成するため削除

  # worktree を追加
  git worktree add "$worktree_dir" "$branch_name" >/dev/null 2>&1

  export WORKTREE_PATH="$worktree_dir"
}
```

## テスタブルなコード設計

### zle 依存の分離

zsh の zle ウィジェットはテストしにくいため、コアロジックを分離します。

```bash
# ❌ テストしにくい: 全てが1つの関数
peco-gcop() {
  local branches=$(git branch -a ...)
  local selected=$(echo "$branches" | peco)
  BUFFER="git checkout $selected"
  zle accept-line
}

# ✅ テストしやすい: コアロジックを分離
_peco_gcop_list_branches() {
  # コアロジック: stdout に出力
  git branch -a ...
}

_peco_gcop_checkout() {
  local branch="$1"
  git checkout "$branch"
}

peco-gcop() {
  # UI層: zle 依存
  local selected=$(_peco_gcop_list_branches | peco)
  if [ -n "$selected" ]; then
    _peco_gcop_checkout "$selected"
    zle accept-line
  fi
}
```

**命名規則**:
- プライベート関数は `_` プレフィックス
- コア関数は stdout に結果を出力
- UI 関数は zle/BUFFER を操作

### run を使わない直接実行

`run` はサブシェルで実行するため、グローバル変数の変更が親に反映されません。

```bash
# ❌ BUFFER の変更が確認できない
@test "BUFFER が設定される" {
  run _peco_gcop_checkout "branch"
  [[ "$BUFFER" == *"cd"* ]]  # 失敗
}

# ✅ 直接実行して変数を確認
@test "BUFFER が設定される" {
  BUFFER=""
  _peco_gcop_checkout "branch"
  local status=$?
  [ "$status" -eq 0 ]
  [[ "$BUFFER" == *"cd"* ]]  # 成功
}
```

## bash 互換性

bats は bash で動作するため、zsh 専用機能は使用できません。

| zsh 専用 | bash 互換の代替 |
|---------|----------------|
| `trap ... RETURN` | `trap ... EXIT` |
| `setopt local_options` | (使用しない) |
| `setopt pipefail` | `set -o pipefail` |
| `bindkey` | モックで対応 |
| `zle` | モックで対応 |

### trap の互換性

```bash
# ❌ zsh 専用
trap "rm -f '$tmp_file'" RETURN

# ✅ bash 互換
trap "rm -f '$tmp_file'" EXIT
```

## トラブルシューティング

### テストがロードできない

**症状**: `load` が失敗する

**解決**:
- ファイルパスを確認
- `.sh` 拡張子を省略: `load 'helpers/mock'` (not `load 'helpers/mock.sh'`)

### command not found: bindkey

**症状**: テスト対象の読み込みで失敗

**解決**:
1. `helpers/zle-mock.sh` を作成
2. テスト対象より先に読み込む

```bash
setup() {
  load 'helpers/zle-mock.sh'  # 先に読み込む
  source "$DOTFILE_DIR/script.sh"
}
```

### グローバル変数が変更されない

**症状**: `run` 後に変数が期待値と異なる

**原因**: `run` はサブシェルで実行される

**解決**: 変数変更を確認する場合は直接実行

```bash
BUFFER=""
my_function
[[ "$BUFFER" == "expected" ]]
```

### trap: undefined signal: RETURN

**症状**: テスト実行時にエラー

**原因**: `RETURN` は zsh 専用

**解決**: `EXIT` に変更

```bash
trap "cleanup" EXIT
```

## 参考リンク

- [bats-core 公式ドキュメント](https://bats-core.readthedocs.io/)
- [bats-core GitHub](https://github.com/bats-core/bats-core)
- [TAP (Test Anything Protocol)](https://testanything.org/)
