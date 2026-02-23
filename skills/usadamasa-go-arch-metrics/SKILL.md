---
name: go-arch-metrics
description: >-
  Use when asked to measure, introduce, or improve Go project architecture quality.
  Applies modularity (coupling/cohesion) and testability (cognitive complexity, cyclomatic
  complexity) metrics based on "Software Architecture Metrics". Triggers for requests like
  "measure code quality", "set up go-arch-lint", "improve testability", or
  "configure golangci-lint for metrics".
---

# Go アーキテクチャメトリクス導入ガイド

「ソフトウェアアーキテクチャメトリクス (ISBN: 9784814400607)」の観点に基づき、
Go プロジェクトのモジュール性とテスト可能性を測定・改善するワークフロー。

## 対象メトリクス早見表

| カテゴリ | メトリクス | ツール | しきい値 |
|---------|-----------|--------|---------|
| モジュール性 | パッケージ依存方向 | go-arch-lint | 違反 = 0 |
| モジュール性 | import 禁止リスト | golangci-lint/depguard | 指定パッケージ = 0 |
| テスト可能性 | 認知的複雑度 | gocognit | ≤ 20 |
| テスト可能性 | 循環複雑度 | gocyclo | ≤ 20 |
| テスト可能性 | 関数の長さ | funlen | ≤ 100行 / 60文 |
| テスト可能性 | ネストの深さ | nestif | ≤ 5 |
| 保守性 | 保守性指数 | maintidx | ≥ 20 |
| 保守性 | 到達不能コード | deadcode | 違反 = 0 |
| 静的解析 | 高度なバグ検出 | staticcheck | デフォルト有効 |

## golangci-lint のバージョンについて

> **重要**: golangci-lint v1 と v2 は設定ファイルが非互換。**必ず v2 を使うこと。**
>
> | 項目 | v1 (旧) | v2 (現行) |
> |------|---------|----------|
> | linter 設定 | `linters-settings:` (トップレベル) | `linters.settings:` (linters 内) |
> | テスト除外 | `issues.exclude-rules:` | `linters.exclusions.rules:` |
> | JSON 出力 | `--out-format json` | `--output.json.path stdout` |
> | バージョン宣言 | `version: "1"` or なし | `version: "2"` |
>
> v1 形式のキーは v2 では **警告なしに無視される** ため、設定が効いていないことに気づきにくい。

## 前提: ツールの準備

### golangci-lint のインストール

```bash
# ローカル: go install で最新 v2 をインストール
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# または aqua で管理 (バージョン固定したい場合)
# aqua.yaml に golangci/golangci-lint@v2.x.x を追加
```

**GitHub Actions では aqua 不要**: `golangci/golangci-lint-action@v6` が自動インストールする。

### go-arch-lint のインストール (モジュール性チェックが必要な場合のみ)

```bash
# aqua で管理する (aqua が未導入の場合は先にインストール)
brew install aquaproj/aqua/aqua
# aqua.yaml に fe3dback/go-arch-lint@v1.14.0 を追加して:
aqua install
```

詳細は `references/tools.md` の「ツール管理: aqua + direnv」セクションを参照。

## 5 ステップワークフロー

### Step 1: ベースライン測定

`scripts/baseline.sh` を使って現状のメトリクスを把握する。

```bash
# golangci-lint が入っていればそのままベースライン測定できる
# (go-arch-lint がなくても golangci-lint 部分だけ測定される)
bash ~/.claude/skills/usadamasa-go-arch-metrics/scripts/baseline.sh ./

# go-arch-lint も使いたい場合は aqua でインストールしてから:
# aqua install && bash ~/.claude/skills/usadamasa-go-arch-metrics/scripts/baseline.sh ./
```

出力されたサマリを確認し、違反件数と重大度を記録する。

### Step 2: 設定ファイルの配置

2種類の設定ファイルをプロジェクトルートに作成する:

- **`.golangci.yml`** → テスト可能性・保守性メトリクス
  → テンプレートは `references/golangci-config.md` を参照
- **`.go-arch-lint.yml`** → パッケージ依存方向ルール (モジュール性)
  → テンプレートは `references/arch-lint-config.md` を参照

### Step 3: 結果の分析

違反を以下の3カテゴリに分類する:

| カテゴリ | 影響 | 代表的な違反 |
|---------|------|------------|
| モジュール性 | 変更の伝播リスク | 依存方向の逆転, 循環参照 |
| テスト可能性 | テスト難易度の上昇 | 認知的複雑度 > 20, 関数長 > 100行 |
| 保守性 | 長期的な腐敗 | 保守性指数 < 20, 到達不能コード |

### Step 4: 是正の優先順位付け

`references/remediation.md` に基づき対応順序を決定する:

1. **High**: モジュール性違反 (依存方向の逆転) → アーキテクチャ崩壊の根本原因
2. **Medium**: テスト可能性違反 (複雑度 > 30, 関数長 > 200行) → テストが書けない
3. **Low**: 保守性違反 (maintidx < 20) → 段階的に改善

### Step 5: CI への統合

`.github/workflows/` に golangci-lint と go-arch-lint のジョブを追加する。
設定例は `references/ci-integration.md` を参照。

## リファレンス

| ファイル | 内容 | 参照タイミング |
|---------|------|--------------|
| `references/tools.md` | 各ツールの詳細・インストール・出力例 | ツールを初めて使うとき / しきい値の根拠を確認したいとき |
| `references/golangci-config.md` | `.golangci.yml` テンプレート | Step 2 でテスト可能性設定を作成するとき |
| `references/arch-lint-config.md` | `.go-arch-lint.yml` テンプレート | Step 2 でパッケージ依存ルールを設定するとき |
| `references/remediation.md` | 違反カテゴリ別の是正手順 | Step 4 で具体的なリファクタリング方法を調べるとき |
| `references/ci-integration.md` | GitHub Actions ジョブ定義 | Step 5 で CI 設定を追加するとき |
