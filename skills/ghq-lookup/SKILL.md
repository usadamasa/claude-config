---
name: ghq-lookup
description: リポジトリ URL が提示されたときにローカルの ghq 管理下クローンを優先参照する。GitHub/GitLab 等のリポジトリ URL (HTTPS/SSH) や「このリポジトリのコードを見て」「ソースを調査して」のような依頼時に使用する。WebFetch でリモートを見に行く前にローカルを確認する。
---

# ghq-lookup

リポジトリ URL が提示されたコード調査時に、ghq 管理下のローカルクローンを優先的に参照する。

## Context

- GHQ_ROOT: !`ghq root`

## Step 1: URL 正規化

提示された URL を `host/owner/repo` 形式に正規化する。

| 入力形式 | 例 | 正規化結果 |
|----------|---|-----------|
| HTTPS | `https://github.com/x-motemen/ghq` | `github.com/x-motemen/ghq` |
| HTTPS (サブパス付き) | `https://github.com/x-motemen/ghq/tree/main/cmd` | `github.com/x-motemen/ghq` |
| HTTPS (.git 付き) | `https://github.com/x-motemen/ghq.git` | `github.com/x-motemen/ghq` |
| SSH | `git@github.com:x-motemen/ghq.git` | `github.com/x-motemen/ghq` |
| 短縮形 | `github.com/x-motemen/ghq` | `github.com/x-motemen/ghq` |

正規化ロジック:

```bash
# HTTPS URL からホスト/オーナー/リポジトリを抽出
REPO_PATH=$(echo "$URL" | sed -E 's|^https?://||; s|^git@([^:]+):|\\1/|; s|\.git$||; s|/tree/.*||; s|/blob/.*||; s|/pull/.*||; s|/issues/.*||; s|/actions/.*||; s|/releases/.*||' | cut -d'/' -f1-3)
```

## Step 2: ローカル存在確認

```bash
GHQ_ROOT=$(ghq root)
test -d "${GHQ_ROOT}/${REPO_PATH}"
```

## Step 3A: ローカルにある場合

ローカルクローンが存在する場合:

1. Read ツールで `${GHQ_ROOT}/${REPO_PATH}` 配下のファイルを直接読む
2. 継続的に多数のファイルを参照する場合は `/add-dir ${GHQ_ROOT}/${REPO_PATH}` をユーザーに提案する
3. WebFetch は使わない (ローカルの方が高速かつコンテキスト効率が良い)

## Step 3B: ローカルにない場合

ローカルクローンが存在しない場合:

1. `ghq get` でクローンする (確認不要):
   ```bash
   ghq get "${REPO_PATH}"
   ```
2. クローン完了後、Step 3A と同様にローカル参照に移行する

## 注意事項

### ファジー検索

正確なパスで見つからない場合、`ghq list` でファジー検索を試みる:

```bash
ghq list | grep "repo-name"
```

リポジトリ名の部分一致で候補を探し、複数ヒットした場合はユーザーに選択を促す。

### worktree 対応

ghq 管理下のリポジトリが git worktree を使用している場合がある。worktree ディレクトリは通常 `../worktrees/{gitroot}/` 配下にあるため、必要に応じてそちらも参照する。

### GitHub 以外のホスト対応

ghq は GitHub 以外のホスト (GitLab, Bitbucket, 自社 Git サーバー等) も管理できる。URL 正規化時にホスト部分を保持すること。

### WebFetch との使い分け

- コード調査、ファイル内容の確認 → ghq-lookup (ローカル優先)
- README やドキュメントの概要確認 → WebFetch でも可
- API ドキュメント、Issue、PR の内容確認 → `gh` CLI または WebFetch
