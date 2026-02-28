# Obsidian CLI コマンド詳細リファレンス

> 本ドキュメントは `obsidian-cli/SKILL.md` の早見表を補完する詳細リファレンスである。

## 凡例

- `(required)` - 必須パラメータ
- `(flag)` - 値なしフラグ
- `(default: ...)` - デフォルト値
- vault 共通オプション `vault=<name>` は全コマンドで使用可能 (省略時はデフォルト Vault)

---

## ノート操作

### create - ファイル作成

| パラメータ | 説明 |
|---|---|
| `name=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `content=<text>` | 初期内容 |
| `template=<name>` | 使用するテンプレート |
| `overwrite` (flag) | 既存ファイルを上書き |
| `open` (flag) | 作成後に開く |
| `newtab` (flag) | 新しいタブで開く |

```bash
obsidian create name="新規ノート" content="# 新規ノート\n\n本文"
obsidian create path="01_Daily/2026-02-28.md" template=Daily open
obsidian create name="既存" content="上書き内容" overwrite
```

### read - ファイル内容を読む

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian read file=概要
obsidian read path="Component/catalyst/概要.md"
obsidian read  # アクティブファイル
```

### append - 末尾に追記

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `content=<text>` (required) | 追記内容 |
| `inline` (flag) | 改行なしで追記 |

```bash
obsidian append file=日誌 content="- 15:00 ミーティング完了"
obsidian append path="01_Daily/2026-02-28.md" content="\n## 追加セクション\n- 項目1"
```

### prepend - 先頭に追記

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `content=<text>` (required) | 追記内容 |
| `inline` (flag) | 改行なしで追記 |

```bash
obsidian prepend file=メモ content="> 重要: この件は要確認"
```

### open - ファイルを開く

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `newtab` (flag) | 新しいタブで開く |

```bash
obsidian open file=概要
obsidian open path="01_Daily/2026-02-28.md" newtab
```

### move - ファイル移動

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `to=<path>` (required) | 移動先フォルダまたはパス |

```bash
obsidian move file=古いメモ to=99_Archives/
obsidian move path="00_Inbox/メモ.md" to="Tech/メモ.md"
```

### rename - ファイル名変更

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `name=<name>` (required) | 新しいファイル名 |

```bash
obsidian rename file=旧名 name=新名
```

### delete - ファイル削除

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `permanent` (flag) | ゴミ箱をスキップして完全削除 |

```bash
obsidian delete file=不要なメモ
obsidian delete path="temp/test.md" permanent
```

### unique - ユニークノート作成

| パラメータ | 説明 |
|---|---|
| `name=<text>` | ノート名 |
| `content=<text>` | 初期内容 |
| `paneType=tab\|split\|window` | 表示方法 |
| `open` (flag) | 作成後に開く |

```bash
obsidian unique name="Zettelkasten メモ" content="# 思考メモ" open
```

### file - ファイル情報

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian file file=概要
# → name, path, size, created, modified 等の情報
```

---

## 検索

### search - テキスト検索

| パラメータ | 説明 |
|---|---|
| `query=<text>` (required) | 検索クエリ |
| `path=<folder>` | フォルダで絞り込み |
| `limit=<n>` | 最大ファイル数 |
| `total` (flag) | マッチ数のみ返す |
| `case` (flag) | 大文字小文字を区別 |
| `format=text\|json` | 出力形式 (default: text) |

```bash
obsidian search query="gRPC" vault=work
obsidian search query="TODO" path=01_Daily limit=10 format=json
obsidian search query="API" total  # マッチ数だけ
```

### search:context - コンテキスト付き検索

`search` と同じパラメータ。マッチ行の前後コンテキストを含む。

```bash
obsidian search:context query="error handling" vault=work limit=5
```

### search:open - 検索ビューを開く

| パラメータ | 説明 |
|---|---|
| `query=<text>` | 初期検索クエリ |

```bash
obsidian search:open query="TODO"
```

### files - ファイル一覧

| パラメータ | 説明 |
|---|---|
| `folder=<path>` | フォルダで絞り込み |
| `ext=<extension>` | 拡張子で絞り込み |
| `total` (flag) | ファイル数のみ返す |

```bash
obsidian files folder=01_Daily ext=md
obsidian files total vault=work  # ファイル総数
```

### folders - フォルダ一覧

| パラメータ | 説明 |
|---|---|
| `folder=<path>` | 親フォルダで絞り込み |
| `total` (flag) | フォルダ数のみ返す |

```bash
obsidian folders vault=work
obsidian folders folder=Component
```

### folder - フォルダ情報

| パラメータ | 説明 |
|---|---|
| `path=<path>` (required) | フォルダパス |
| `info=files\|folders\|size` | 特定情報のみ |

```bash
obsidian folder path=Component info=files
```

---

## メタデータ・プロパティ

### properties - プロパティ一覧

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル指定 |
| `path=<path>` | パス指定 |
| `name=<name>` | 特定プロパティの数 |
| `total` (flag) | プロパティ数のみ |
| `sort=count` | 使用頻度でソート |
| `counts` (flag) | 出現回数を含む |
| `format=yaml\|json\|tsv` | 出力形式 (default: yaml) |
| `active` (flag) | アクティブファイル |

```bash
obsidian properties counts sort=count vault=work  # 全 Vault のプロパティ使用頻度
obsidian properties file=概要 format=json           # 特定ファイルのプロパティ
```

### property:read - プロパティ値を読む

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | プロパティ名 |
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian property:read name=status file=概要
```

### property:set - プロパティを設定

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | プロパティ名 |
| `value=<value>` (required) | 値 |
| `type=text\|list\|number\|checkbox\|date\|datetime` | 型 |
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian property:set name=status value=completed file=概要
obsidian property:set name=priority value=1 type=number file=タスク
```

### property:remove - プロパティを削除

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | プロパティ名 |
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian property:remove name=draft file=概要
```

### tags - タグ一覧

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル指定 |
| `path=<path>` | パス指定 |
| `total` (flag) | タグ数のみ |
| `counts` (flag) | 出現回数を含む |
| `sort=count` | 使用頻度でソート |
| `format=json\|tsv\|csv` | 出力形式 (default: tsv) |
| `active` (flag) | アクティブファイル |

```bash
obsidian tags counts sort=count vault=work
obsidian tags file=概要
```

### tag - タグ情報

| パラメータ | 説明 |
|---|---|
| `name=<tag>` (required) | タグ名 |
| `total` (flag) | 出現回数のみ |
| `verbose` (flag) | ファイルリストと数を含む |

```bash
obsidian tag name=project verbose vault=work
```

### aliases - エイリアス一覧

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル指定 |
| `path=<path>` | パス指定 |
| `total` (flag) | エイリアス数のみ |
| `verbose` (flag) | ファイルパスを含む |
| `active` (flag) | アクティブファイル |

```bash
obsidian aliases verbose vault=work
obsidian aliases file=概要
```

---

## リンク・グラフ

### links - 発リンク一覧

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `total` (flag) | リンク数のみ |

```bash
obsidian links file=概要
obsidian links total  # アクティブファイルのリンク数
```

### backlinks - 被リンク一覧

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `counts` (flag) | リンク数を含む |
| `total` (flag) | 被リンク数のみ |
| `format=json\|tsv\|csv` | 出力形式 (default: tsv) |

```bash
obsidian backlinks file=概要 format=json
obsidian backlinks total file=概要  # 被リンク数
```

### orphans - 孤立ノート

| パラメータ | 説明 |
|---|---|
| `total` (flag) | 孤立ノート数のみ |
| `all` (flag) | 非 Markdown ファイルも含む |

```bash
obsidian orphans vault=work
obsidian orphans total vault=work  # 数だけ
```

### deadends - デッドエンド

| パラメータ | 説明 |
|---|---|
| `total` (flag) | デッドエンド数のみ |
| `all` (flag) | 非 Markdown ファイルも含む |

```bash
obsidian deadends vault=work
obsidian deadends total  # 数だけ
```

### unresolved - 未解決リンク

| パラメータ | 説明 |
|---|---|
| `total` (flag) | 未解決リンク数のみ |
| `counts` (flag) | リンク数を含む |
| `verbose` (flag) | ソースファイルを含む |
| `format=json\|tsv\|csv` | 出力形式 (default: tsv) |

```bash
obsidian unresolved verbose vault=work
obsidian unresolved total  # 数だけ
```

### outline - 見出し構造

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `format=tree\|md\|json` | 出力形式 (default: tree) |
| `total` (flag) | 見出し数のみ |

```bash
obsidian outline file=概要 format=tree
obsidian outline format=md  # アクティブファイルを Markdown 形式で
```

---

## タスク

### tasks - タスク一覧

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイルで絞り込み |
| `path=<path>` | パスで絞り込み |
| `total` (flag) | タスク数のみ |
| `done` (flag) | 完了タスクのみ |
| `todo` (flag) | 未完了タスクのみ |
| `status="<char>"` | ステータス文字で絞り込み |
| `verbose` (flag) | ファイル別グルーピング (行番号付き) |
| `format=json\|tsv\|csv` | 出力形式 (default: text) |
| `active` (flag) | アクティブファイル |
| `daily` (flag) | Daily note のタスク |

```bash
obsidian tasks todo vault=work                          # 全未完了タスク
obsidian tasks todo verbose vault=work                  # ファイル別に表示
obsidian tasks done file=日誌                            # 特定ファイルの完了タスク
obsidian tasks total vault=work                         # タスク総数
obsidian tasks todo format=json vault=work | jq '.'     # JSON で処理
obsidian tasks daily todo                                # 今日の Daily note のタスク
```

### task - タスク操作

| パラメータ | 説明 |
|---|---|
| `ref=<path:line>` | タスク参照 (パス:行番号) |
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `line=<n>` | 行番号 |
| `toggle` (flag) | ステータスをトグル |
| `done` (flag) | 完了にする |
| `todo` (flag) | 未完了にする |
| `daily` (flag) | Daily note を対象 |
| `status="<char>"` | ステータス文字を設定 |

```bash
obsidian task file=日誌 line=10 toggle          # トグル
obsidian task file=日誌 line=10 done            # 完了にする
obsidian task ref="01_Daily/2026-02-28.md:15" todo  # 未完了に戻す
obsidian task daily line=5 done                 # Daily note の5行目を完了
```

---

## テンプレート

### templates - テンプレート一覧

| パラメータ | 説明 |
|---|---|
| `total` (flag) | テンプレート数のみ |

```bash
obsidian templates vault=work
```

### template:read - テンプレート内容を読む

| パラメータ | 説明 |
|---|---|
| `name=<template>` (required) | テンプレート名 |
| `resolve` (flag) | テンプレート変数を解決 |
| `title=<title>` | 変数解決時のタイトル |

```bash
obsidian template:read name=Daily
obsidian template:read name=Daily resolve title="2026-02-28"
```

### template:insert - テンプレート挿入

| パラメータ | 説明 |
|---|---|
| `name=<template>` (required) | テンプレート名 |

```bash
obsidian template:insert name=Daily
```

---

## プラグイン

### plugins - プラグイン一覧

| パラメータ | 説明 |
|---|---|
| `filter=core\|community` | プラグインタイプで絞り込み |
| `versions` (flag) | バージョン番号を含む |
| `format=json\|tsv\|csv` | 出力形式 (default: tsv) |

```bash
obsidian plugins filter=community versions vault=work
obsidian plugins format=json vault=work | jq '.[].id'
```

### plugins:enabled - 有効プラグイン一覧

`plugins` と同じパラメータ。有効なプラグインのみ表示。

```bash
obsidian plugins:enabled filter=community vault=work
```

### plugin - プラグイン情報

| パラメータ | 説明 |
|---|---|
| `id=<plugin-id>` (required) | プラグイン ID |

```bash
obsidian plugin id=dataview vault=work
```

### plugin:enable / plugin:disable

| パラメータ | 説明 |
|---|---|
| `id=<id>` (required) | プラグイン ID |
| `filter=core\|community` | プラグインタイプ |

```bash
obsidian plugin:enable id=kanban vault=work
obsidian plugin:disable id=kanban vault=work
```

### plugin:install - プラグインインストール

| パラメータ | 説明 |
|---|---|
| `id=<id>` (required) | プラグイン ID |
| `enable` (flag) | インストール後に有効化 |

```bash
obsidian plugin:install id=templater-obsidian enable vault=work
```

### plugin:uninstall - プラグインアンインストール

| パラメータ | 説明 |
|---|---|
| `id=<id>` (required) | プラグイン ID |

```bash
obsidian plugin:uninstall id=unused-plugin vault=work
```

### plugin:reload - プラグイン再読み込み

| パラメータ | 説明 |
|---|---|
| `id=<id>` (required) | プラグイン ID |

```bash
obsidian plugin:reload id=my-plugin
```

### plugins:restrict - 制限モード

| パラメータ | 説明 |
|---|---|
| `on` (flag) | 制限モード有効化 |
| `off` (flag) | 制限モード無効化 |

```bash
obsidian plugins:restrict on
obsidian plugins:restrict off
obsidian plugins:restrict  # 現在の状態を表示
```

---

## Sync・履歴

### sync:status - Sync 状態

```bash
obsidian sync:status vault=work
```

### sync - Sync 一時停止/再開

| パラメータ | 説明 |
|---|---|
| `on` (flag) | Sync 再開 |
| `off` (flag) | Sync 一時停止 |

```bash
obsidian sync off vault=work  # 一時停止
obsidian sync on vault=work   # 再開
```

### sync:history - Sync 版履歴

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `total` (flag) | バージョン数のみ |

```bash
obsidian sync:history file=概要 vault=work
```

### sync:read - Sync 版を読む

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `version=<n>` (required) | バージョン番号 |

```bash
obsidian sync:read file=概要 version=1 vault=work
```

### sync:restore - Sync 版を復元

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `version=<n>` (required) | バージョン番号 |

```bash
obsidian sync:restore file=概要 version=2 vault=work
```

### sync:deleted - Sync 削除済みファイル

| パラメータ | 説明 |
|---|---|
| `total` (flag) | 削除ファイル数のみ |

```bash
obsidian sync:deleted vault=work
```

### sync:open - Sync 履歴を開く

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian sync:open file=概要
```

### history - ローカル履歴

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian history file=概要
```

### history:list - 履歴のあるファイル一覧

```bash
obsidian history:list vault=work
```

### history:read - ローカル版を読む

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `version=<n>` | バージョン番号 (default: 1) |

```bash
obsidian history:read file=概要 version=1
```

### history:restore - ローカル版を復元

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `version=<n>` (required) | バージョン番号 |

```bash
obsidian history:restore file=概要 version=1
```

### history:open - ファイルリカバリを開く

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |

```bash
obsidian history:open file=概要
```

### diff - バージョン差分

| パラメータ | 説明 |
|---|---|
| `file=<name>` | ファイル名 |
| `path=<path>` | ファイルパス |
| `from=<n>` | 差分元バージョン |
| `to=<n>` | 差分先バージョン |
| `filter=local\|sync` | バージョンソースで絞り込み |

```bash
obsidian diff file=概要 from=1 to=2
obsidian diff file=概要 filter=sync
```

---

## UI・システム

### vault - Vault 情報

| パラメータ | 説明 |
|---|---|
| `info=name\|path\|files\|folders\|size` | 特定情報のみ |

```bash
obsidian vault vault=work
obsidian vault info=files vault=work  # ファイル数だけ
```

### vaults - Vault 一覧

| パラメータ | 説明 |
|---|---|
| `total` (flag) | Vault 数のみ |
| `verbose` (flag) | パスを含む |

```bash
obsidian vaults verbose
```

### version - バージョン表示

```bash
obsidian version
```

### reload - Vault 再読み込み

```bash
obsidian reload vault=work
```

### restart - アプリ再起動

```bash
obsidian restart
```

### workspace - ワークスペースツリー

| パラメータ | 説明 |
|---|---|
| `ids` (flag) | ワークスペースアイテム ID を含む |

```bash
obsidian workspace
obsidian workspace ids
```

### tabs - 開いているタブ

| パラメータ | 説明 |
|---|---|
| `ids` (flag) | タブ ID を含む |

```bash
obsidian tabs
obsidian tabs ids
```

### tab:open - 新しいタブを開く

| パラメータ | 説明 |
|---|---|
| `group=<id>` | タブグループ ID |
| `file=<path>` | 開くファイル |
| `view=<type>` | ビュータイプ |

```bash
obsidian tab:open file=概要
```

### random / random:read - ランダムノート

| パラメータ | 説明 |
|---|---|
| `folder=<path>` | フォルダで絞り込み |
| `newtab` (flag) | 新しいタブで開く (random のみ) |

```bash
obsidian random folder=Tech
obsidian random:read folder=Component  # 内容を出力 (開かない)
```

### recents - 最近開いたファイル

| パラメータ | 説明 |
|---|---|
| `total` (flag) | ファイル数のみ |

```bash
obsidian recents
```

### command - コマンド実行

| パラメータ | 説明 |
|---|---|
| `id=<command-id>` (required) | コマンド ID |

```bash
obsidian command id=editor:toggle-fold
obsidian command id=app:toggle-left-sidebar
```

### commands - コマンド一覧

| パラメータ | 説明 |
|---|---|
| `filter=<prefix>` | ID プレフィックスで絞り込み |

```bash
obsidian commands
obsidian commands filter=editor
obsidian commands filter=app
```

### hotkey / hotkeys - ホットキー

hotkeys:

| パラメータ | 説明 |
|---|---|
| `total` (flag) | ホットキー数のみ |
| `verbose` (flag) | カスタム/デフォルト表示 |
| `format=json\|tsv\|csv` | 出力形式 (default: tsv) |
| `all` (flag) | ホットキーなしのコマンドも含む |

hotkey:

| パラメータ | 説明 |
|---|---|
| `id=<command-id>` (required) | コマンド ID |
| `verbose` (flag) | カスタム/デフォルト表示 |

```bash
obsidian hotkeys format=json
obsidian hotkey id=editor:toggle-fold verbose
```

### bookmarks / bookmark

bookmarks:

| パラメータ | 説明 |
|---|---|
| `total` (flag) | ブックマーク数のみ |
| `verbose` (flag) | ブックマークタイプを含む |
| `format=json\|tsv\|csv` | 出力形式 (default: tsv) |

bookmark:

| パラメータ | 説明 |
|---|---|
| `file=<path>` | ファイルをブックマーク |
| `subpath=<subpath>` | サブパス (見出しやブロック) |
| `folder=<path>` | フォルダをブックマーク |
| `search=<query>` | 検索クエリをブックマーク |
| `url=<url>` | URL をブックマーク |
| `title=<title>` | ブックマークタイトル |

```bash
obsidian bookmarks verbose format=json
obsidian bookmark file=概要 title="重要ノート"
```

---

## Base (データベース)

### bases - Base ファイル一覧

```bash
obsidian bases vault=work
```

### base:views - Base のビュー一覧

```bash
obsidian base:views file=MyBase
```

### base:query - Base クエリ

| パラメータ | 説明 |
|---|---|
| `file=<name>` | Base ファイル名 |
| `path=<path>` | Base ファイルパス |
| `view=<name>` | ビュー名 |
| `format=json\|csv\|tsv\|md\|paths` | 出力形式 (default: json) |

```bash
obsidian base:query file=MyBase format=json
obsidian base:query file=MyBase view="Active Items" format=md
```

### base:create - Base アイテム作成

| パラメータ | 説明 |
|---|---|
| `file=<name>` | Base ファイル名 |
| `path=<path>` | Base ファイルパス |
| `view=<name>` | ビュー名 |
| `name=<name>` | 新規ファイル名 |
| `content=<text>` | 初期内容 |
| `open` (flag) | 作成後に開く |
| `newtab` (flag) | 新しいタブで開く |

```bash
obsidian base:create file=MyBase name="新しいアイテム" content="# 詳細" open
```

---

## テーマ・スニペット

### theme - 現在のテーマ

| パラメータ | 説明 |
|---|---|
| `name=<name>` | テーマ名 (詳細表示用) |

```bash
obsidian theme
obsidian theme name="Minimal"
```

### themes - インストール済みテーマ

| パラメータ | 説明 |
|---|---|
| `versions` (flag) | バージョン番号を含む |

```bash
obsidian themes versions
```

### theme:set - テーマ変更

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | テーマ名 (空文字でデフォルト) |

```bash
obsidian theme:set name="Minimal"
obsidian theme:set name=""  # デフォルトに戻す
```

### theme:install - テーマインストール

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | テーマ名 |
| `enable` (flag) | インストール後に有効化 |

```bash
obsidian theme:install name="Minimal" enable
```

### theme:uninstall - テーマアンインストール

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | テーマ名 |

```bash
obsidian theme:uninstall name="OldTheme"
```

### snippets / snippets:enabled - CSS スニペット

```bash
obsidian snippets
obsidian snippets:enabled
```

### snippet:enable / snippet:disable

| パラメータ | 説明 |
|---|---|
| `name=<name>` (required) | スニペット名 |

```bash
obsidian snippet:enable name=custom-styles
obsidian snippet:disable name=custom-styles
```

---

## 開発者向け

### eval - JavaScript 実行

| パラメータ | 説明 |
|---|---|
| `code=<javascript>` (required) | 実行する JavaScript コード |

```bash
obsidian eval code="app.vault.getFiles().length"
obsidian eval code="app.vault.getName()"
```

### devtools - DevTools 表示切替

```bash
obsidian devtools
```

### dev:dom - DOM クエリ

| パラメータ | 説明 |
|---|---|
| `selector=<css>` (required) | CSS セレクタ |
| `total` (flag) | 要素数のみ |
| `text` (flag) | テキスト内容を返す |
| `inner` (flag) | innerHTML を返す |
| `all` (flag) | 全マッチを返す |
| `attr=<name>` | 属性値を取得 |
| `css=<prop>` | CSS プロパティ値を取得 |

```bash
obsidian dev:dom selector=".workspace-leaf" total
obsidian dev:dom selector=".markdown-preview" text
```

### dev:css - CSS 検査

| パラメータ | 説明 |
|---|---|
| `selector=<css>` (required) | CSS セレクタ |
| `prop=<name>` | プロパティ名で絞り込み |

```bash
obsidian dev:css selector=".markdown-preview" prop=font-size
```

### dev:screenshot - スクリーンショット

| パラメータ | 説明 |
|---|---|
| `path=<filename>` | 出力ファイルパス |

```bash
obsidian dev:screenshot path=~/Desktop/obsidian-screenshot.png
```

### dev:console - コンソールログ

| パラメータ | 説明 |
|---|---|
| `clear` (flag) | バッファをクリア |
| `limit=<n>` | 最大メッセージ数 (default: 50) |
| `level=log\|warn\|error\|info\|debug` | ログレベルで絞り込み |

```bash
obsidian dev:console level=error
obsidian dev:console clear
```

### dev:errors - エラーログ

| パラメータ | 説明 |
|---|---|
| `clear` (flag) | バッファをクリア |

```bash
obsidian dev:errors
```

### dev:debug - デバッガ接続

| パラメータ | 説明 |
|---|---|
| `on` (flag) | デバッガ接続 |
| `off` (flag) | デバッガ切断 |

```bash
obsidian dev:debug on
obsidian dev:debug off
```

### dev:mobile - モバイルエミュレーション

| パラメータ | 説明 |
|---|---|
| `on` (flag) | 有効化 |
| `off` (flag) | 無効化 |

```bash
obsidian dev:mobile on
```

### dev:cdp - Chrome DevTools Protocol

| パラメータ | 説明 |
|---|---|
| `method=<CDP.method>` (required) | CDP メソッド |
| `params=<json>` | メソッドパラメータ (JSON) |

```bash
obsidian dev:cdp method="Runtime.evaluate" params='{"expression":"1+1"}'
```
