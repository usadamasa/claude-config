---
name: obsidian-vault-management
description: >-
  Obsidian CLI を使った Vault 管理ワークフロー。タスク管理、健全性チェック、プラグイン管理、
  Sync 制御、バッチ処理など CLI ベースの Vault 運用を支援する。
  「Vaultを管理」「タスクを確認」「孤立ノートを調べて」「プラグイン管理」「Sync状態」
  「Vault健全性」「orphan確認」「deadend確認」のように Vault 管理を依頼されたときに使用する。
---

# Obsidian Vault 管理ワークフロー

Obsidian CLI を活用した Vault 管理・運用タスクのワークフローガイド。

## Context

- Obsidian 起動確認: !`pgrep -x Obsidian > /dev/null && echo "Running" || echo "NOT RUNNING - Obsidian を起動してください"`
- work Vault タスク数 (todo): !`obsidian tasks todo total vault=work 2>/dev/null || echo "取得失敗"`
- work Vault orphan 数: !`obsidian orphans total vault=work 2>/dev/null || echo "取得失敗"`

## 既存スキルとの棲み分け

| スキル | 方式 | 用途 |
|---|---|---|
| `setup-obsidian-mcp` | MCP | プラグイン導入・設定 |
| `restructure-obsidian-vault` | MCP+Script | 構造分析・CLAUDE.md 更新 |
| `obsidian-cli` | CLI | コマンドリファレンス・構文ガイド |
| **本スキル** | CLI | Vault 管理ワークフロー (タスク, 健全性, プラグイン, Sync) |

**棲み分けの原則**: MCP = セマンティック・リッチ操作 / CLI = バッチ・スクリプト・管理操作

## Vault 識別

| Vault | 用途 | CLI指定 |
|---|---|---|
| (業務用) | 業務ナレッジ・日誌 | `vault=work` |
| (個人用) | 個人ナレッジ | `vault=personal` |

CLI では `vault=work` / `vault=personal` で指定。省略時はデフォルト Vault。

---

## A. Daily Note ワークフロー

### 今日の Daily Note を作成/開く

```bash
# スクリプトで冪等に作成
~/.claude/skills/obsidian-vault-management/scripts/obs-daily.sh vault=work
```

### 手動で作成

```bash
DATE=$(date +%Y-%m-%d)
obsidian create path="01_Daily/${DATE}.md" template=Daily open vault=work
```

### Daily Note に追記

```bash
# 「やったこと」セクションに追記
obsidian append path="01_Daily/$(date +%Y-%m-%d).md" content="\n- 作業内容" vault=work
```

### Daily Note のタスクを確認

```bash
obsidian tasks daily todo vault=work
```

---

## B. タスク管理

### 全未完了タスクを表示

```bash
obsidian tasks todo vault=work
obsidian tasks todo verbose vault=work  # ファイル別グルーピング
```

### 特定ファイルのタスク

```bash
obsidian tasks todo file=日誌 vault=work
obsidian tasks file=概要 format=json vault=work
```

### タスクのトグル

```bash
# verbose で行番号を確認してからトグル
obsidian tasks todo verbose vault=work
obsidian task file=日誌 line=10 toggle vault=work
```

### タスクを完了/未完了にする

```bash
obsidian task file=日誌 line=10 done vault=work    # 完了
obsidian task file=日誌 line=10 todo vault=work    # 未完了に戻す
```

### タスク数のサマリ

```bash
echo "=== work Vault タスクサマリ ==="
echo "未完了: $(obsidian tasks todo total vault=work)"
echo "完了済: $(obsidian tasks done total vault=work)"
echo "合計:   $(obsidian tasks total vault=work)"
```

> **Kanban ボード操作** は MCP 経由のスキルを使用すること。CLI ではファイル内のチェックボックスタスクのみ操作可能。

---

## C. 健全性チェック

### 統合レポート (スクリプト)

```bash
~/.claude/skills/obsidian-vault-management/scripts/obs-vault-health.sh vault=work
```

### 個別チェック

```bash
# 孤立ノート (被リンクなし)
obsidian orphans vault=work
obsidian orphans total vault=work

# デッドエンド (発リンクなし)
obsidian deadends vault=work
obsidian deadends total vault=work

# 未解決リンク
obsidian unresolved verbose vault=work
obsidian unresolved total vault=work

# Vault 統計
obsidian vault vault=work
obsidian files total vault=work
obsidian folders total vault=work
```

### 健全性の判断基準

| 指標 | 良好 | 要注意 | 要対策 |
|---|---|---|---|
| orphan 率 | < 10% | 10-20% | > 20% |
| deadend 率 | < 15% | 15-30% | > 30% |
| 未解決リンク数 | 0 | 1-10 | > 10 |

---

## D. バッチ操作

### 一括プロパティ更新

```bash
# 特定フォルダのファイルに status=archived を設定
for f in $(obsidian files folder=99_Archives vault=work); do
  obsidian property:set name=status value=archived path="$f" vault=work
done
```

### ファイル移動 (アーカイブ)

```bash
# 特定ファイルをアーカイブに移動
obsidian move file=旧プロジェクト to=99_Archives/ vault=work
```

### タグの使用状況分析

```bash
# タグの使用頻度を降順で表示
obsidian tags counts sort=count vault=work

# 特定タグの詳細
obsidian tag name=project verbose vault=work
```

### プロパティの使用状況分析

```bash
obsidian properties counts sort=count vault=work
```

---

## E. プラグイン管理

### インストール済みプラグイン確認

```bash
obsidian plugins filter=community versions vault=work
obsidian plugins:enabled filter=community vault=work
```

### プラグインの有効化/無効化

```bash
obsidian plugin:enable id=dataview vault=work
obsidian plugin:disable id=kanban vault=work
```

### プラグインのインストール

```bash
obsidian plugin:install id=templater-obsidian enable vault=work
```

### プラグインのアンインストール

```bash
obsidian plugin:uninstall id=unused-plugin vault=work
```

### 制限モード

```bash
obsidian plugins:restrict          # 現在の状態
obsidian plugins:restrict on       # 有効化 (コミュニティプラグイン無効)
obsidian plugins:restrict off      # 無効化
```

---

## F. クロス Vault 操作

### 両方の Vault の状態を比較

```bash
echo "=== work Vault ==="
obsidian vault vault=work
echo ""
echo "=== personal Vault ==="
obsidian vault vault=personal
```

### 両方の Vault のタスクを確認

```bash
echo "=== work 未完了タスク ==="
obsidian tasks todo vault=work
echo ""
echo "=== personal 未完了タスク ==="
obsidian tasks todo vault=personal
```

### 両方の Vault の健全性

```bash
echo "=== work ==="
~/.claude/skills/obsidian-vault-management/scripts/obs-vault-health.sh vault=work
echo ""
echo "=== personal ==="
~/.claude/skills/obsidian-vault-management/scripts/obs-vault-health.sh vault=personal
```

> CLI では Vault 間のファイルコピー/移動はできない。ファイルシステム上で直接操作すること。

---

## G. Sync 管理

### Sync 状態確認

```bash
obsidian sync:status vault=work
```

### Sync の一時停止/再開

```bash
obsidian sync off vault=work   # 大量変更前に一時停止
# ... 作業 ...
obsidian sync on vault=work    # 作業後に再開
```

### Sync 履歴の確認

```bash
obsidian sync:history file=概要 vault=work
```

### ファイルの復元

```bash
# Sync 版を確認
obsidian sync:history file=概要 vault=work
obsidian sync:read file=概要 version=1 vault=work

# 復元
obsidian sync:restore file=概要 version=2 vault=work
```

### ローカル履歴

```bash
obsidian history file=概要 vault=work
obsidian history:read file=概要 version=1 vault=work
obsidian history:restore file=概要 version=1 vault=work
```

### バージョン差分

```bash
obsidian diff file=概要 from=1 to=2 vault=work
```

---

## 注意事項

- **Obsidian が起動中であること**: CLI は起動中の Obsidian プロセスと通信する
- **バッチ操作は Sync を一時停止してから**: 大量変更時は `obsidian sync off` → 作業 → `obsidian sync on` の順で
- **Kanban は MCP で**: CLI ではチェックボックスタスクのみ操作可能。Kanban ボードの列移動は MCP 経由のスキルを使う
- **破壊的操作は慎重に**: `delete permanent`, `property:remove` は取り消し不可
- **`file=` の曖昧マッチに注意**: 正確な操作には `path=` を使う

## リファレンス

| ファイル | 内容 | 参照タイミング |
|---|---|---|
| `references/cli-workflow-patterns.md` | シェルパイプラインパターン | CLI を組み合わせた応用操作をしたいとき |
| `scripts/obs-vault-health.sh` | Vault 健全性チェック | 定期的な健全性確認時 |
| `scripts/obs-daily.sh` | Daily Note 作成 | 日次ルーティン時 |
