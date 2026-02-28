---
name: setup-obsidian-mcp
description: Obsidian MCP Plugin (aaronsb/obsidian-mcp-plugin) の導入と設定を支援する。Obsidian VaultをAIアシスタントと連携させたいとき、MCPサーバを導入したいとき、グラフ走査やセマンティック検索をClaude Codeから使いたいときに使用する。
---

# Obsidian MCP Plugin 導入ガイド

## Overview

Obsidian MCP Plugin (aaronsb版) を導入し、Claude CodeからObsidian Vaultのグラフ走査やセマンティック検索を利用できるようにする。

## 導入ワークフロー

### Step 1: BRAT経由でプラグインをインストール

aaronsb/obsidian-mcp-plugin (現在の正式名: **Semantic Notes Vault MCP**) は Community Plugin ストアに未公開のため、BRAT (Beta Reviewers Auto-update Tester) 経由でインストールする。

#### 1-1. BRATのインストール

1. Obsidianを開く
2. 設定 → Community plugins → Browse
3. "BRAT" で検索し、**Obsidian42 - BRAT** をインストール
4. BRAT を有効化

#### 1-2. obsidian-mcp-plugin のインストール

1. 設定 → Community plugins → BRAT → Settings
2. "Add Beta plugin" をクリック
3. GitHub リポジトリ URL に `aaronsb/obsidian-mcp-plugin` を入力
4. "Add Plugin" をクリック
5. Community plugins 一覧で **Semantic Notes Vault MCP** を有効化

> プラグインは `.obsidian/plugins/semantic-vault-mcp/` に配置される。
> 手動インストールの場合: GitHubリリースから `main.js`, `manifest.json`, `styles.css` をダウンロードし、同ディレクトリに配置する。

### Step 2: プラグイン設定

1. Obsidian設定 → Community plugins → MCP Plugin → 設定
2. サーバーポートを確認 (デフォルト: HTTP 3001 / HTTPS 3443)
3. 必要に応じてCORS設定を確認

### Step 3: Claude Code MCP設定

`claude mcp add` コマンドでMCPサーバーを登録する。

#### コマンド

```bash
claude mcp add --transport http --scope user obsidian http://localhost:3001/mcp
```

- `--transport http`: Streamable HTTP トランスポートを使用 (SSE は deprecated)
- `--scope user`: 全プロジェクトで利用可能にする (`~/.claude.json` に保存)
- ポート番号はプラグイン設定で変更した場合は合わせること

#### 確認

```bash
claude mcp list
```

`obsidian: http://localhost:3001/mcp (HTTP) - ✓ Connected` と表示されれば成功。

#### permissions設定

`~/.claude/settings.json` の `permissions.allow` に以下を追加:

```
"mcp__obsidian__*"
```

### Step 4: 接続テスト

Claude Codeを再起動し、MCPツールが利用可能か確認する。

テスト方法:
1. Claude Codeを起動
2. Vault全体の統計情報を取得してみる: `statistics` アクションを実行
3. ノートの検索や走査が正常に動作することを確認

## 利用可能なツール

詳細は `references/mcp-plugin-tools.md` を参照。

### ツールカテゴリ

| カテゴリ | 説明 |
|---|---|
| vault | Vault全体の情報取得 (ファイル一覧、検索、統計) |
| edit | ノートの作成・編集・削除 |
| view | ノートの閲覧・メタデータ取得 |
| graph | グラフ走査・リンク分析 (traverse, neighbors, path, backlinks) |
| workflow | ワークフロー自動化 |
| dataview | Dataviewクエリの実行 |
| bases | Obsidian Basesとの連携 |
| system | システム情報・設定 |

### 基本グラフ操作

| ツール | 説明 |
|---|---|
| `traverse` | BFSでノートグラフを走査 |
| `neighbors` | 直接接続されたノート一覧 |
| `path` | 2ノート間の最短パス検索 |
| `statistics` | リンク統計情報 (Vault全体/個別) |
| `backlinks` | 被リンク一覧 |
| `forwardlinks` | 発リンク一覧 |

### 高度な検索・走査

| ツール | 説明 |
|---|---|
| `search-traverse` | テキスト検索 + グラフ走査 |
| `advanced-traverse` | 複数クエリ + 走査戦略選択 |
| `tag-traverse` | タグベースのグラフ走査 |

### タグ分析

| ツール | 説明 |
|---|---|
| `tag-analysis` | タグの共起・頻度分析 |
| `shared-tags` | 共有タグを持つノート検索 |

## トラブルシューティング

### 接続できない場合

1. Obsidianが起動しているか確認
2. MCP Pluginが有効化されているか確認
3. ポート番号が設定と一致しているか確認
4. ファイアウォールでlocalhost:3001がブロックされていないか確認

### ツールが認識されない場合

1. Claude Codeを再起動
2. `claude mcp list` でサーバーが登録されているか確認
3. `claude mcp get obsidian` で設定内容を確認
4. MCPサーバーのURL形式が正しいか確認 (`http://localhost:3001/mcp`)

## リソース

### references/

- `mcp-plugin-tools.md`: aaronsb版 obsidian-mcp-plugin の全ツールリファレンス。各ツールのパラメータ、リクエスト例、レスポンス形式を記載。
