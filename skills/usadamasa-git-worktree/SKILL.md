---
name: git-worktree
description: git-wt を使用した Git worktree 管理ガイド。worktree の作成、切り替え、削除、設定について説明します。
---

# git-worktree

git-wt は Git worktree をより便利に扱うためのツールです。ブランチ間の移動や複数ブランチの同時作業を効率化します。

## git-wt とは

Git worktree は同一リポジトリの複数ブランチを別ディレクトリで同時に開く機能です。git-wt はこれをシンプルなコマンドで操作できるようにします。

### インストール

```bash
brew install k1LoW/tap/git-wt
```

## 基本コマンド

### worktree 一覧表示

```bash
git wt
```

現在の worktree 一覧を表示します。

### worktree 作成/切り替え

```bash
git wt <branch>
```

- 既存のブランチ: そのブランチの worktree に切り替え（なければ作成）
- 新規ブランチ: worktree を作成してそこに移動

### worktree 削除

```bash
# 通常の削除
git wt -d <branch>

# 強制削除
git wt -D <branch>
```

## シェル統合

zsh で `git wt` 後に自動的にディレクトリ移動するには、以下を `.zshrc` に追加:

```bash
eval "$(git wt --init zsh)"
```

bash の場合:

```bash
eval "$(git wt --init bash)"
```

## 設定オプション

`~/.config/git/config` または各リポジトリの `.git/config` で設定できます。

### 設定項目

| 設定 | 説明 | デフォルト |
|------|------|----------|
| `wt.basedir` | worktree の作成先ディレクトリ | `../worktrees/{gitroot}` |
| `wt.copyignored` | .gitignore ファイルをコピー | false |
| `wt.copyuntracked` | 未追跡ファイルをコピー | false |
| `wt.copymodified` | 変更済みファイルをコピー | false |
| `wt.hook` | worktree 作成後に実行するコマンド | なし |

### 推奨設定例

```ini
[wt]
    copyignored = true
    copyuntracked = true
    copymodified = true
    hook = test -f .envrc && direnv allow || true
    basedir = ../worktrees/{gitroot}
```

**設定の解説:**

- `copyignored/copyuntracked/copymodified = true`: 作業中のファイルを新しい worktree にコピー。環境設定ファイルなどを引き継げる
- `hook = test -f .envrc && direnv allow || true`: direnv 環境変数を自動で許可。`.envrc` がない場合はスキップ
- `basedir = ../worktrees/{gitroot}`: worktree を親ディレクトリの `worktrees/` 以下に配置

## 使用例

### 機能開発中に別ブランチの作業が必要になった場合

```bash
# 現在のブランチで作業中
git wt hotfix-123
# → ../worktrees/myrepo/hotfix-123 に移動して作業

# 作業完了後、元のブランチに戻る
git wt feature-abc
```

### 複数の PR を同時にレビュー

```bash
# PR #1 のブランチを確認
git wt pr-1-branch
# レビュー...

# PR #2 のブランチを確認（別ウィンドウで）
git wt pr-2-branch
# レビュー...
```

## トラブルシューティング

### direnv エラーが出る

worktree 作成時に `.envrc: No such file or directory` のようなエラーが出る場合、hook 設定を確認してください:

```ini
# 修正前（.envrc がないとエラー出力）
hook = direnv allow ; true

# 修正後（.envrc がない場合はスキップ）
hook = test -f .envrc && direnv allow || true
```

### worktree が見つからない

```bash
# worktree の一覧と場所を確認
git worktree list
```

## worktree 削除前の Claude Code Memory 移行

worktree を削除すると、その worktree に保存された Claude Code の auto-memory (`MEMORY.md`, `SESSION_HANDOFF.md`) が消える。
`migrate-worktree-memory.sh` を使うと、削除前に親リポジトリの memory path へ自動退避できる。

### 手動実行

```bash
~/.claude/hooks/migrate-worktree-memory.sh /path/to/worktree
```

### git-wt deletehook への登録

`wt.deletehook` に登録することで、`git wt -d <branch>` 実行時に自動で実行される:

```ini
# ~/.config/git/config または各リポジトリの .git/config
[wt]
    deletehook = ~/.claude/hooks/migrate-worktree-memory.sh
```

**設定例 (推奨設定に追加):**

```ini
[wt]
    copyignored = true
    copyuntracked = true
    copymodified = true
    hook = test -f .envrc && direnv allow || true
    basedir = ../worktrees/{gitroot}
    deletehook = ~/.claude/hooks/migrate-worktree-memory.sh
```

### 動作内容

1. worktree の `.git` ファイルから親リポジトリのパスを導出する
2. 親リポジトリの Claude Code memory path (`~/.claude/projects/.../memory/`) に移行する
3. `SESSION_HANDOFF.md`: 上書きコピー (最新の引き継ぎ情報を優先)
4. `MEMORY.md`: 親が存在する場合は末尾に追記、存在しない場合はコピー

追記フォーマット:

```markdown
## [Merged from worktree: {branch-name}] YYYY-MM-DD

{worktree の MEMORY.md 内容}
```

## 関連リンク

- [git-wt GitHub](https://github.com/k1LoW/git-wt)
- [Git worktree 公式ドキュメント](https://git-scm.com/docs/git-worktree)
