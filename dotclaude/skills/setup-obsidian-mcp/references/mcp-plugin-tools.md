# Semantic Notes Vault MCP (aaronsb版) ツールリファレンス

GitHub: https://github.com/aaronsb/obsidian-mcp-plugin
プラグインID: `semantic-vault-mcp`

## 概要

Semantic Notes Vault MCP (旧: Obsidian MCP Plugin) はObsidian VaultをAIアシスタントとModel Context Protocol (MCP)で接続する｡
グラフ走査やセマンティック検索をClaude Codeから直接利用できる｡

- **HTTPポート**: 3001 (デフォルト)
- **HTTPSポート**: 3443 (デフォルト)
- **MCP エンドポイント**: `http://localhost:3001/mcp`
- **トランスポート**: Streamable HTTP (`--transport http`)
- **Claude Code 登録コマンド**: `claude mcp add --transport http --scope user obsidian http://localhost:3001/mcp`

## ツールカテゴリ (8カテゴリ)

| カテゴリ | 説明 |
|---|---|
| vault | Vault全体の情報取得 (ファイル一覧､検索､統計) |
| edit | ノートの作成・編集・削除 |
| view | ノートの閲覧・メタデータ取得 |
| graph | グラフ走査・リンク分析 |
| workflow | ワークフロー自動化 |
| dataview | Dataviewクエリの実行 |
| bases | Obsidian Basesとの連携 |
| system | システム情報・設定 |

## 基本グラフ操作

### traverse - 接続ノードの探索

BFS (幅優先探索) で起点ファイルから接続されたノートを発見する｡

```json
{
  "operation": "graph",
  "action": "traverse",
  "sourcePath": "01_Daily/2024-01-15.md",
  "maxDepth": 3,
  "maxNodes": 50,
  "followBacklinks": true,
  "followForwardLinks": true,
  "followTags": false,
  "fileFilter": "",
  "folderFilter": ""
}
```

**パラメータ:**
- `sourcePath` (必須): 起点ファイルパス
- `maxDepth` (任意, default: 3): 探索ホップ数
- `maxNodes` (任意, default: 50): 返却ノード数上限
- `followBacklinks` (任意, default: true): 被リンクを辿る
- `followForwardLinks` (任意, default: true): 発リンクを辿る
- `followTags` (任意, default: false): タグベースの接続を辿る
- `fileFilter` (任意): ファイル名フィルタ (正規表現)
- `folderFilter` (任意): フォルダフィルタ

### neighbors - 直接接続の取得

特定ファイルに直接リンクしている/されているファイル一覧を返す｡

```json
{
  "operation": "graph",
  "action": "neighbors",
  "sourcePath": "Component/tableland/overview.md"
}
```

### path - ファイル間パスの探索

2つのファイル間の最短パスを検索する｡

```json
{
  "operation": "graph",
  "action": "path",
  "sourcePath": "Tech/gRPC.md",
  "targetPath": "Component/catalyst/design.md",
  "maxDepth": 5
}
```

### statistics - リンク統計

ファイルのリンク統計情報を取得する｡sourcePath省略でVault全体の統計を返す｡

```json
{
  "operation": "graph",
  "action": "statistics",
  "sourcePath": "index.md"
}
```

**レスポンス項目:**
- `inDegree`: 被リンク数
- `outDegree`: 発リンク数
- `totalDegree`: 総接続数
- `unresolvedCount`: リンク切れ数
- `tagCount`: タグ数

### backlinks - 被リンク一覧

指定ファイルへリンクしているファイル一覧を返す｡

```json
{
  "operation": "graph",
  "action": "backlinks",
  "sourcePath": "Component/tableland/overview.md"
}
```

### forwardlinks - 発リンク一覧

指定ファイルからリンクしているファイル一覧を返す｡

```json
{
  "operation": "graph",
  "action": "forwardlinks",
  "sourcePath": "Component/tableland/overview.md"
}
```

## 高度な検索・走査

### search-traverse - 検索付きグラフ走査

グラフ走査とテキスト検索を組み合わせて関連コンテンツを発見する｡

```json
{
  "action": "search-traverse",
  "startPath": "Component/tableland/overview.md",
  "searchQuery": "データパイプライン",
  "maxDepth": 2,
  "maxSnippetsPerNode": 2,
  "scoreThreshold": 0.5
}
```

### advanced-traverse - 高度なグラフ走査

複数の検索クエリとカスタマイズ可能な走査戦略を使用する｡

```json
{
  "action": "advanced-traverse",
  "sourcePath": "Component/catalyst/design.md",
  "searchQueries": ["認証", "gRPC", "JWT"],
  "strategy": "best-first",
  "maxDepth": 4
}
```

**戦略:**
- `breadth-first`: 幅優先探索
- `best-first`: 最良優先探索 (デフォルト)
- `beam-search`: ビームサーチ (`beamWidth` で幅を指定)

### tag-traverse - タグベース走査

指定タグに関連するノートを辿ってグラフを走査する｡

```json
{
  "action": "tag-traverse",
  "sourcePath": "01_Daily/2024-01-15.md",
  "tagFilter": ["#tableland", "#incident"],
  "maxDepth": 3
}
```

## タグ分析

### tag-analysis - タグ関係分析

タグの共起関係と頻度を分析する｡sourcePath省略でVault全体の分析｡

```json
{
  "action": "tag-analysis",
  "sourcePath": "Component/tableland/overview.md"
}
```

### shared-tags - 共有タグ検索

指定ノートとタグを共有するノートを検索する｡

```json
{
  "action": "shared-tags",
  "sourcePath": "Tech/gRPC.md",
  "minSharedTags": 2
}
```
