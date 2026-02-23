---
name: permission-optimizer
description: >-
  settings.jsonのBash/Read/Write/Editパーミッションを最適化するスキル。
  セッションJSONLから過去30日間のツール使用状況を分析し、
  安全性評価に基づいてパーミッションの追加/削除を提案する。
  (1) パーミッションの棚卸し (2) 新規パーミッションの追加提案
  (3) 未使用パーミッションの削除提案 (4) ベアエントリ警告 に使用する。
---

# Permission Optimizer

settings.jsonの`permissions.allow`/`permissions.deny`/`permissions.ask`に登録されているBash/Read/Write/Editパーミッションを管理するワークフロー。

## ワークフロー

### 1. 分析の実行

以下のコマンドを実行してツール使用状況を集計する:

```bash
go run ./cmd/analyze-permissions --days 30
```

オプション:
- `--days N`: 集計期間を指定(デフォルト: 30日)
- `--settings PATH`: settings.jsonのパスを指定(デフォルト: ~/.claude/settings.json)

### 2. 結果の確認

出力はJSON形式で以下のセクションを含む:

- **metadata**: 分析の概要(期間、ファイル数、ツール呼び出し回数)
- **current_allow**: 現在のallowリスト
- **current_deny**: 現在のdenyリスト
- **current_ask**: 現在のaskリスト
- **recommendations.add**: 追加推奨パーミッション(safeカテゴリで未登録のもの)
- **recommendations.review**: 要確認パーミッション(review/askカテゴリまたはdenyすべきもの)
- **recommendations.unused**: 未使用パーミッション(リストにあるが使用されていない)
- **recommendations.bare_entry_warnings**: ベアエントリ警告(修飾子なしのエントリ)
- **all_patterns**: 全パターンの使用統計

### 3. settings.jsonの更新

ユーザーの承認を得た上で、以下の手順でsettings.jsonを更新する:

1. 承認されたパーミッションのみ適切なリスト(allow/deny/ask)に追加
2. パーミッション形式: `Tool(pattern:*)` (Bash) または `Tool(pattern)` (Read/Write/Edit)
3. 許可エントリはアルファベット順にソート
4. 不要と判断されたエントリは削除

## パーミッション評価順序の注意

Claude Codeのパーミッション評価は **deny → ask → allow** の順序で行われる。

### ベア(修飾子なし)エントリの禁止

`ask`や`allow`配列にベアの`Bash`や`Read`(修飾子なし)を入れてはいけない。

**誤った設定:**
```json
"ask": ["Bash"],
"allow": ["Bash(git status:*)", "Bash(go test:*)"]
```

この場合、ベアの`Bash`がすべてのBash呼び出しにマッチし、`allow`のコマンド別許可が全て無視される。

**正しい設定:**
```json
"allow": ["Bash(git status:*)", "Bash(go test:*)"],
"ask": ["Bash(git commit:*)", "Bash(git push:*)"]
```

### このスキル実行時の検証

パーミッション管理の更新時に、以下を必ず検証すること:

1. `ask`/`allow`配列にベアの`Bash`、`Read`、`Write`、`Edit`が含まれていないか確認する
2. 含まれている場合はユーザーに警告し、削除を提案する
3. レポートの`bare_entry_warnings`フィールドで自動検出される

## 安全性カテゴリ

### Bash コマンド

| カテゴリ | 説明 | 例 |
|----------|------|-----|
| safe | 読取系・ビルドツール | git status, go test, task, make, brew list |
| ask | 変更・破壊操作 | git commit, git push, git rebase, rm -rf |
| deny | 外部通信・特権操作 | curl, wget, sudo, ssh, scp, eval |

### Read/Write/Edit パス

| カテゴリ | 説明 | 例 |
|----------|------|-----|
| safe | プロジェクトファイル・設定 | src/**, CLAUDE.md, .claude/** |
| deny | 機密ファイル | ~/.ssh/**, ~/.aws/**, .env, credentials |
| review | 手動確認が必要 | 上記に該当しないパス |

## パーミッション形式

- Bash: `Bash(コマンドプレフィックス:*)` 例: `Bash(git status:*)`
- Read: `Read(パスパターン)` 例: `Read(~/.claude/**)`
- Write: `Write(パスパターン)` 例: `Write(src/**)`
- Edit: `Edit(パスパターン)` 例: `Edit(~/.claude/**)`
