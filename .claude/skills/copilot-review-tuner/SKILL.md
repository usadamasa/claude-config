---
name: copilot-review-tuner
description: >-
  GitHub Copilot の PR レビューコメントを分析し、
  .github/copilot-instructions.md を改善するワークフロー。
  偽陽性パターンの特定と抑制ルールの追加を自動化する。
  「Copilotレビューを分析して」「copilot-instructionsを更新して」
  「Copilotの指摘パターンを調べて」「Copilot棚卸し」
  のように依頼されたときに使用する。
---

# Copilot Review Tuner

GitHub Copilot の PR レビューコメントを分析し、`.github/copilot-instructions.md` の "Do Not Flag" セクションを改善するワークフロー。

## ワークフロー

### Step 1: データ収集

直近の merged PR から Copilot のレビューコメントを取得する。

```bash
# merged PR のリストを取得 (デフォルト: 直近 20 件)
gh pr list --state merged --limit 20 --json number,title,mergedAt

# 各 PR の Copilot レビューコメントを取得
gh api "repos/{owner}/{repo}/pulls/{pr_number}/comments" \
  --jq '[.[] | select(.user.login == "copilot-pull-request-reviewer[bot]") | {id, path, body, line, created_at}]'
```

**注意**: read-only の gh コマンドのみ使用する。`--method POST/PUT/DELETE` や書き込み操作は一切使用しない。

### Step 2: 分類・集計

各コメントを以下のカテゴリに分類する:

| カテゴリ | 判定基準 | 意味 |
|----------|----------|------|
| **True Positive** | 指摘に基づくコード修正が実際に行われた | 有益な指摘 |
| **False Positive** | 無視された、またはリポジトリの文脈で不適切 | 抑制すべき |
| **Low Value** | 正しいが優先度が低い (スタイル、ドキュメント等) | 抑制を検討 |

判定のヒント:
- PR のコメントスレッドに "Fixed." 等の応答があれば True Positive の可能性が高い
- PR diff を確認し、指摘箇所が修正されていれば True Positive
- shellcheck でカバーされるシェル品質の指摘は False Positive
- ドキュメント微小修正の指摘は Low Value

### Step 3: ユーザーへ報告

分類結果をサマリとして表示する:

```
## Copilot レビュー分析レポート

### 集計 (直近 N PR)
- True Positive: X 件
- False Positive: Y 件
- Low Value: Z 件
- Signal-to-Noise Ratio: X / (X + Y + Z) = NN%

### False Positive パターン (頻度順)
1. パターン名 (N 件) - 説明
2. ...

### 新規抑制ルール提案
- "Do Not Flag" に追加すべきパターン:
  1. ...
  2. ...

### 既存ルールの評価
- 現在の "Do Not Flag" ルールで不要になったもの (該当指摘なし):
  1. ...
```

### Step 4: copilot-instructions.md 更新

ユーザーの承認を得てから `.github/copilot-instructions.md` を編集する:

1. "Do Not Flag" セクションに新パターンを追加
2. 不要になったルールがあれば削除
3. "Review Focus" セクションに True Positive パターンを反映 (必要な場合)

## 使用する gh コマンド (read-only のみ)

| コマンド | 用途 |
|----------|------|
| `gh pr list --state merged` | マージ済み PR の一覧取得 |
| `gh api repos/.../pulls/.../comments` (GET) | PR コメントの取得 |
| `gh pr view` | PR の詳細情報取得 |
| `gh pr diff` | PR の差分取得 |

## 注意事項

- このスキルは **分析と提案** のみを行う。copilot-instructions.md の変更はユーザー承認後に実施する
- gh api の書き込み操作 (`--method POST/PUT/DELETE`) は一切使用しない
- Copilot のコメントは `copilot-pull-request-reviewer[bot]` ユーザーから投稿される
