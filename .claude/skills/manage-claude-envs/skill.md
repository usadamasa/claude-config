---
name: manage-claude-envs
description: >-
  Claude Code環境変数(env.sh)の管理と最新モデルへの更新を行う。
  Anthropic公式ドキュメントから最新のVertexモデル名を取得し、
  env.shとenv.sh.exampleのモデル設定を更新する。
  「モデルを更新して」「最新モデルに追従」「env更新」「モデル確認」
  のように依頼されたときに使用する。
---

# Claude Code 環境変数管理

### 0. worktree 環境チェック

CLAUDE.md の「worktree 環境でのファイルパス解決」を参照し、worktree 判定を行う。
worktree 環境の場合、`env.sh` の編集パスを `$(pwd)/env.sh` に読み替えること。
`env.sh` には `--settings` 相当のフラグがないため、Edit ツールで直接 `$(pwd)/env.sh` を指定する。
`env.sh.example` は worktree でも通常でも `$(pwd)/env.sh.example` で変更不要。

### 対象ファイル

- `env.sh` - 実際の環境変数ファイル (.gitignore 対象)
- `env.sh.example` - テンプレート (リポジトリ管理下)

## モデル更新手順

### 1. 現在の設定を確認

`env.sh` を Read して現在のモデル設定を表示する。

### 2. 最新モデルを確認

以下の Anthropic 公式ドキュメントから最新の Claude モデル名を取得する:

- WebFetch: `https://docs.anthropic.com/en/docs/about-claude/models` で最新モデル一覧を取得
- Vertex AI で利用可能なモデル ID を確認する

### 3. 更新対象の環境変数

以下の環境変数のモデル名を最新に更新する:

| 環境変数 | 用途 |
|---------|------|
| `ANTHROPIC_MODEL` | デフォルトモデル (通常 sonnet) |
| `ANTHROPIC_DEFAULT_HAIKU_MODEL` | haiku 指定時のモデル |
| `ANTHROPIC_DEFAULT_OPUS_MODEL` | opus 指定時のモデル |
| `ANTHROPIC_DEFAULT_SONNET_MODEL` | sonnet 指定時のモデル |
| `CLAUDE_CODE_SUBAGENT_MODEL` | subagent 用モデル (通常 sonnet) |

### 4. ユーザー確認

差分がある場合、変更内容をユーザーに表示して確認を求める。

### 5. ファイル更新

承認後、`env.sh` と `env.sh.example` の両方を Edit ツールで更新する。

## 初期セットアップ

`env.sh` が存在しない場合:

1. `env.sh.example` を Read する
2. ユーザーに `ANTHROPIC_VERTEX_PROJECT_ID` を確認する
3. 値を埋めて `env.sh` を Write する
4. `task setup` で symlink を作成するよう案内する
