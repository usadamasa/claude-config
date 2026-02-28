---
name: obsidian-cli
description: >-
  Obsidian CLI (v1.12+) のコマンドリファレンスと構文ガイド。
  「CLIで」「コマンドラインから」「obsidianコマンド」「CLIでノートを作成」「CLIで検索」
  「シェルからObsidian操作」のように CLI 経由の Obsidian 操作を依頼されたときに使用する。
---

# Obsidian CLI リファレンス

Obsidian v1.12+ で追加された公式 CLI のコマンドリファレンス。

## Context

- Obsidian 起動確認: !`pgrep -x Obsidian > /dev/null && echo "Running" || echo "NOT RUNNING - Obsidian を起動してください"`
- Vault 一覧: !`obsidian vaults verbose 2>/dev/null`

## 前提条件

- Obsidian v1.12 以上がインストール済み
- **Obsidian が起動中であること** (CLI は起動中の Obsidian プロセスと通信する)
- バイナリパス: `/Applications/Obsidian.app/Contents/MacOS/obsidian` (PATH 設定済み)

## 基本構文

```bash
obsidian <command> [options]
```

### パラメータ規則

| ルール | 例 |
|---|---|
| key=value 形式 | `vault=work` |
| スペースを含む値はクォート | `name="My Note"` |
| ファイル指定は `file=` (名前) または `path=` (パス) | `file=日誌` / `path=01_Daily/2026-02-28.md` |
| `file=` は wikilink 同様の名前解決 | `file=概要` → 最初にマッチするファイル |
| `path=` は Vault ルートからの相対パス | `path=Component/catalyst/概要.md` |
| 省略時はアクティブファイルが対象 | `obsidian read` → 現在開いているファイル |
| 改行は `\n`、タブは `\t` | `content="行1\n行2"` |
| フラグは値なし | `obsidian files total` |

### Vault 指定

```bash
obsidian <command> vault=work      # Work Vault
obsidian <command> vault=personal  # Personal Vault
# vault= 省略時はデフォルト Vault が対象
```

## CLI vs MCP 選択指針

| やりたいこと | CLI | MCP |
|---|---|---|
| バッチ処理 (一括プロパティ更新等) | **推奨** | - |
| シェルパイプライン連携 | **推奨** | - |
| プラグイン管理 | **推奨** | - |
| Sync 制御・履歴 | **推奨** | - |
| タスク一覧・トグル | **推奨** | - |
| orphan / deadend 検出 | **推奨** | - |
| スクリプトからの自動化 | **推奨** | - |
| セマンティック検索 (TF-IDF) | - | **推奨** |
| グラフ走査 (traverse/path) | - | **推奨** |
| Dataview クエリ (DQL) | - | **推奨** |
| Kanban ボード操作 | - | **推奨** |
| UI 操作 (タブ、ワークスペース) | どちらも可 | - |

## コマンドカテゴリ別早見表

> 各コマンドの詳細パラメータは `references/command-reference.md` を参照。

### ノート操作

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `create` | 新規ファイル作成 | `obsidian create name="メモ" content="# メモ"` |
| `read` | ファイル内容を読む | `obsidian read file=日誌` |
| `append` | 末尾に追記 | `obsidian append file=日誌 content="- 追記"` |
| `prepend` | 先頭に追記 | `obsidian prepend file=日誌 content="# 見出し"` |
| `open` | Obsidian で開く | `obsidian open file=概要` |
| `move` | ファイル移動 | `obsidian move file=メモ to=99_Archives/` |
| `rename` | ファイル名変更 | `obsidian rename file=旧名 name=新名` |
| `delete` | ファイル削除 | `obsidian delete file=不要` |
| `unique` | ユニークノート作成 | `obsidian unique name="Zettel"` |

### 検索

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `search` | テキスト検索 | `obsidian search query="gRPC"` |
| `search:context` | マッチ行のコンテキスト付き | `obsidian search:context query="API" limit=5` |
| `search:open` | Obsidian の検索ビューを開く | `obsidian search:open query="TODO"` |
| `files` | ファイル一覧 | `obsidian files folder=01_Daily ext=md` |
| `folders` | フォルダ一覧 | `obsidian folders` |

### メタデータ・プロパティ

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `properties` | プロパティ一覧 | `obsidian properties file=概要 format=yaml` |
| `property:read` | プロパティ値を読む | `obsidian property:read name=status file=概要` |
| `property:set` | プロパティを設定 | `obsidian property:set name=status value=done file=概要` |
| `property:remove` | プロパティを削除 | `obsidian property:remove name=draft file=概要` |
| `tags` | タグ一覧 | `obsidian tags counts sort=count` |
| `tag` | タグ情報 | `obsidian tag name=project verbose` |
| `aliases` | エイリアス一覧 | `obsidian aliases file=概要` |

### リンク・グラフ

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `links` | 発リンク一覧 | `obsidian links file=概要` |
| `backlinks` | 被リンク一覧 | `obsidian backlinks file=概要 format=json` |
| `orphans` | 孤立ノート (被リンクなし) | `obsidian orphans total` |
| `deadends` | デッドエンド (発リンクなし) | `obsidian deadends total` |
| `unresolved` | 未解決リンク | `obsidian unresolved verbose` |
| `outline` | 見出し構造 | `obsidian outline file=概要 format=tree` |

### タスク

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `tasks` | タスク一覧 | `obsidian tasks todo vault=work` |
| `task` | タスク操作 | `obsidian task file=日誌 line=5 toggle` |

### テンプレート

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `templates` | テンプレート一覧 | `obsidian templates` |
| `template:read` | テンプレート内容 | `obsidian template:read name=Daily resolve` |
| `template:insert` | テンプレート挿入 | `obsidian template:insert name=Daily` |

### プラグイン

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `plugins` | インストール済みプラグイン | `obsidian plugins filter=community versions` |
| `plugins:enabled` | 有効なプラグイン | `obsidian plugins:enabled` |
| `plugin` | プラグイン情報 | `obsidian plugin id=dataview` |
| `plugin:enable` | プラグイン有効化 | `obsidian plugin:enable id=kanban` |
| `plugin:disable` | プラグイン無効化 | `obsidian plugin:disable id=kanban` |
| `plugin:install` | プラグインインストール | `obsidian plugin:install id=templater enable` |
| `plugin:uninstall` | プラグインアンインストール | `obsidian plugin:uninstall id=unused-plugin` |
| `plugin:reload` | プラグイン再読み込み (dev) | `obsidian plugin:reload id=my-plugin` |
| `plugins:restrict` | 制限モード切替 | `obsidian plugins:restrict on` |

### Sync・履歴

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `sync:status` | Sync 状態 | `obsidian sync:status` |
| `sync` | Sync の一時停止/再開 | `obsidian sync off` / `obsidian sync on` |
| `sync:history` | Sync 版履歴 | `obsidian sync:history file=概要` |
| `sync:read` | Sync 版を読む | `obsidian sync:read file=概要 version=1` |
| `sync:restore` | Sync 版を復元 | `obsidian sync:restore file=概要 version=2` |
| `sync:deleted` | Sync 削除済みファイル | `obsidian sync:deleted` |
| `history` | ローカル履歴 | `obsidian history file=概要` |
| `history:read` | ローカル版を読む | `obsidian history:read file=概要 version=1` |
| `history:restore` | ローカル版を復元 | `obsidian history:restore file=概要 version=1` |
| `history:list` | 履歴のあるファイル一覧 | `obsidian history:list` |
| `diff` | バージョン差分 | `obsidian diff file=概要 from=1 to=2` |

### UI・システム

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `vault` | Vault 情報 | `obsidian vault info=name` |
| `vaults` | Vault 一覧 | `obsidian vaults verbose` |
| `version` | バージョン表示 | `obsidian version` |
| `reload` | Vault 再読み込み | `obsidian reload` |
| `restart` | アプリ再起動 | `obsidian restart` |
| `workspace` | ワークスペースツリー | `obsidian workspace` |
| `tabs` | 開いているタブ | `obsidian tabs` |
| `random` | ランダムノートを開く | `obsidian random` |
| `recents` | 最近開いたファイル | `obsidian recents` |
| `command` | コマンド実行 | `obsidian command id=editor:toggle-fold` |
| `commands` | コマンド一覧 | `obsidian commands filter=editor` |
| `bookmarks` | ブックマーク一覧 | `obsidian bookmarks` |
| `bookmark` | ブックマーク追加 | `obsidian bookmark file=概要` |

### Base (データベース)

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `bases` | Base ファイル一覧 | `obsidian bases` |
| `base:views` | Base のビュー一覧 | `obsidian base:views file=MyBase` |
| `base:query` | Base クエリ | `obsidian base:query file=MyBase format=json` |
| `base:create` | Base アイテム作成 | `obsidian base:create file=MyBase name="新規"` |

### テーマ・スニペット

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `theme` | 現在のテーマ | `obsidian theme` |
| `themes` | インストール済みテーマ | `obsidian themes versions` |
| `theme:set` | テーマ変更 | `obsidian theme:set name="Minimal"` |
| `theme:install` | テーマインストール | `obsidian theme:install name="Minimal" enable` |
| `snippets` | CSS スニペット一覧 | `obsidian snippets` |
| `snippet:enable` | スニペット有効化 | `obsidian snippet:enable name=custom` |

### 開発者向け

| コマンド | 説明 | 代表的な使い方 |
|---|---|---|
| `eval` | JavaScript 実行 | `obsidian eval code="app.vault.getFiles().length"` |
| `devtools` | DevTools 表示切替 | `obsidian devtools` |
| `dev:dom` | DOM クエリ | `obsidian dev:dom selector=".workspace"` |
| `dev:css` | CSS 検査 | `obsidian dev:css selector=".markdown-preview"` |
| `dev:screenshot` | スクリーンショット | `obsidian dev:screenshot path=~/shot.png` |
| `dev:console` | コンソールログ | `obsidian dev:console level=error` |
| `dev:cdp` | CDP コマンド | `obsidian dev:cdp method="Runtime.evaluate"` |

## 出力形式ガイド

多くのコマンドが `format=` で出力形式を指定できる。

| 形式 | 用途 | 例 |
|---|---|---|
| `json` | プログラム処理 | `obsidian tags format=json \| jq '.[] \| select(.count > 5)'` |
| `tsv` | シェルパイプライン | `obsidian tags format=tsv \| cut -f1` |
| `csv` | スプレッドシート連携 | `obsidian backlinks format=csv file=概要` |
| `text` | 人間が読む (デフォルト) | `obsidian tasks todo` |
| `yaml` | プロパティ表示 | `obsidian properties format=yaml` |
| `tree` | 階層表示 | `obsidian outline format=tree` |
| `md` | Markdown テーブル | `obsidian base:query format=md` |
| `paths` | パスのみ (Base用) | `obsidian base:query format=paths` |

### パイプライン例

```bash
# 全 Vault のオープンタスクを数える
obsidian tasks todo total vault=work

# orphan ノートのリストを JSON で取得
obsidian orphans format=json vault=work | jq '.'

# 特定タグのノート一覧
obsidian tag name=project verbose vault=work

# 全プロパティの使用頻度
obsidian properties counts sort=count vault=work
```

## 注意事項

- Obsidian が起動していない場合、CLI コマンドは動作しない
- **CLI はエラー時も exit code 0 を返す**。エラー判定は出力に `Error:` が含まれるかで行うこと
- `file=` は曖昧マッチ (wikilink 同様)。意図しないファイルにマッチする可能性があるため、正確な操作には `path=` を使う
- 破壊的操作 (`delete permanent`, `property:remove`, `plugin:uninstall`) は確認なしに実行されるため注意
- `eval` コマンドは任意の JavaScript を実行できるため、Vault データに影響を与える操作は慎重に

## リファレンス

| ファイル | 内容 | 参照タイミング |
|---|---|---|
| `references/command-reference.md` | 全コマンドの詳細パラメータ・出力例 | 特定コマンドのパラメータを正確に知りたいとき |
