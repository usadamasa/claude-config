# CLI ワークフローパターン集

Obsidian CLI とシェルツールを組み合わせた実用的なパイプラインパターン。

## 検索・分析パターン

### 特定タグのノートを一覧

```bash
obsidian tag name=project verbose vault=work
```

### 最近更新されたファイル (ファイルシステム)

```bash
# CLI ではソート機能がないため、ファイルシステムで補完
find ~/obsidian/work -name "*.md" -mtime -7 -exec basename {} \; | sort
```

### プロパティ値でフィルタ (JSON パイプライン)

```bash
# status=draft のファイルを探す
obsidian files vault=work | while read -r f; do
  val=$(obsidian property:read name=status path="$f" vault=work 2>/dev/null)
  [ "$val" = "draft" ] && echo "$f"
done
```

### リンク密度の高いノート

```bash
# 被リンク数付きで全ファイルを表示 (重いので注意)
obsidian backlinks format=json vault=work | jq -r '.[] | "\(.count)\t\(.file)"' | sort -rn | head -20
```

## バッチ更新パターン

### フォルダ内の全ファイルにプロパティ追加

```bash
obsidian files folder=99_Archives vault=work | while read -r f; do
  obsidian property:set name=status value=archived path="$f" vault=work
done
```

### 空ファイルの検出

```bash
obsidian files vault=work | while read -r f; do
  content=$(obsidian read path="$f" vault=work 2>/dev/null)
  [ -z "$content" ] && echo "Empty: $f"
done
```

### 未解決リンクの一括表示 (ソースファイル付き)

```bash
obsidian unresolved verbose format=json vault=work | jq -r '.[] | "\(.link) ← \(.sources | join(", "))"'
```

## タスク管理パターン

### 全 Vault のタスクサマリ

```bash
for v in work personal; do
  todo=$(obsidian tasks todo total vault=$v 2>/dev/null || echo "0")
  done=$(obsidian tasks done total vault=$v 2>/dev/null || echo "0")
  echo "$v: todo=$todo done=$done"
done
```

### 特定ファイルのタスクを全て完了にする

```bash
# verbose で行番号を取得してループ
obsidian tasks todo verbose file=日誌 vault=work | awk -F: '{print $2}' | while read -r line; do
  obsidian task file=日誌 line=$line done vault=work
done
```

### 今日の Daily Note のタスクだけ確認

```bash
obsidian tasks daily todo vault=work
```

## 健全性チェックパターン

### orphan を一括アーカイブ

```bash
# まず確認
obsidian orphans vault=work

# 移動 (1つずつ確認しながら)
obsidian orphans vault=work | while read -r f; do
  echo "Move to archive? $f"
  # obsidian move path="$f" to=99_Archives/ vault=work
done
```

### deadend の発リンク追加候補

```bash
# deadend ノートの見出しを表示して、リンク先を考える材料にする
obsidian deadends vault=work | while read -r f; do
  echo "=== $f ==="
  obsidian outline path="$f" vault=work 2>/dev/null
done
```

## Sync パターン

### バッチ操作のセーフティパターン

```bash
# Sync を一時停止
obsidian sync off vault=work

# バッチ操作
obsidian files folder=old_folder vault=work | while IFS= read -r f; do
  obsidian move path="$f" to=99_Archives/ vault=work
done

# Sync を再開
obsidian sync on vault=work
```

### Sync コンフリクト検出

```bash
# "conflicted" を含むファイルを検索
obsidian search query="conflicted" vault=work
obsidian files vault=work | grep -i conflict
```

## プラグイン管理パターン

### 両 Vault のプラグイン差分

```bash
diff <(obsidian plugins:enabled filter=community vault=work | sort) \
     <(obsidian plugins:enabled filter=community vault=personal | sort)
```

### プラグインのバージョン確認

```bash
obsidian plugins filter=community versions format=json vault=work | jq '.[] | "\(.id): \(.version)"'
```

## 出力形式の使い分け

| 目的 | 形式 | 例 |
|---|---|---|
| jq でフィルタ | `format=json` | `obsidian tags format=json \| jq ...` |
| cut/awk で列抽出 | `format=tsv` | `obsidian tags format=tsv \| cut -f1` |
| スプレッドシートへ | `format=csv` | `obsidian backlinks format=csv` |
| 人間が読む | `format=text` | `obsidian tasks todo` |
| Markdown テーブル | `format=md` | `obsidian base:query format=md` |
