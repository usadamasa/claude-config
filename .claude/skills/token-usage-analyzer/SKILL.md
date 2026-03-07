---
name: token-usage-analyzer
description: |
  Claude Codeセッションのtoken使用量を分析し、改善ポイントを提示するスキル。
  「token使用量を分析して」「token消費を調べて」「コスト分析」「使用量レポート」
  のようにtoken使用量の分析や改善提案を依頼されたときに使用する。
---

# Token使用量分析

## CLIツールの実行

`go.mod` は `cmd/` ディレクトリにあるため、claude-config リポジトリルートから `cd cmd &&` を付けて実行する。

```bash
# 直近30日間の分析(デフォルト)
cd cmd && go run ./analyze-tokens

# 期間指定
cd cmd && go run ./analyze-tokens --days 7

# Top N変更
cd cmd && go run ./analyze-tokens --top 20

# カスタムディレクトリ指定
cd cmd && go run ./analyze-tokens --dir ~/.claude/projects

# 警告のみ出力(定期チェック向け)
cd cmd && go run ./analyze-tokens --warnings-only

# worktree環境での実行(settings.jsonのパスを明示)
cd cmd && go run ./analyze-tokens --settings $(pwd)/dotclaude/settings.json
```

## 出力フォーマット

JSON形式で以下のセクションを出力する:

### summary (全体統計)
- `total_sessions`: セッション数
- `total_input_tokens`: 総input tokens
- `total_output_tokens`: 総output tokens
- `total_api_calls`: 総APIコール数
- `average_input_per_call`: 1 APIコールあたりの平均input tokens

### warnings (自動警告)
閾値ベースで自動検出された問題を表示する:
- `global_high_avg`: 全体の avg_input_per_call > 60K
- `high_avg_input`: プロジェクト別 avg_input_per_call > 80K
- `high_call_ratio`: セッション別 api_calls/user_messages > 50

- `too_many_plugins`: 有効プラグイン数 > 10
- `too_many_skills`: グローバルスキル数 > 15

各警告には `value` (実測値)、`threshold` (閾値)、`recommendation` (具体的な改善アクション) が含まれる｡
**warnings が空なら健全な状態**｡warnings セクションを最初に確認し、recommendation に従って対処する｡

### config_health (グローバル設定の健全性)
- `enabled_plugins`: 有効プラグイン数
- `global_skills`: グローバルスキル数 (`~/.claude/skills/` 配下のディレクトリ/symlink数)

### top_sessions (Top Nセッション)
- input tokens降順で上位セッションを表示
- 各セッションのproject, model, API call数, user message数を含む

### project_summary (プロジェクト別)
- プロジェクトごとの合計input/output tokens, セッション数, 平均input/call

### model_summary (モデル別)
- モデルごとのinput/output tokens, コール数

## レポートの解釈ガイド

### 注目指標

| 指標 | 健全な範囲 | 要注意 | 対策 |
|------|----------|--------|------|
| average_input_per_call | 30K-60K | >80K | システムプロンプト肥大化を疑う |
| api_calls / user_messages | 5-20x | >50x | subagent多段呼び出しを疑う |
| cache_read_tokens | >0 | =0 | プロンプトキャッシュ未活用 |

### 改善ワークフロー

warnings の `recommendation` フィールドに具体的なアクションが記載されている｡
以下の優先順位で対処する:

1. **config警告 (too_many_plugins/too_many_skills)** → 全プロジェクトに影響するため最優先
2. **global_high_avg** → 全体の baseline を下げる
3. **high_avg_input** → セッション数の多いプロジェクトから対処
4. **high_call_ratio** → 個別セッションの最適化

### 手動確認が必要な指標

| 指標 | 確認方法 |
|------|---------|
| cache_read_tokens = 0 | セッションが短すぎてキャッシュが効いていない可能性｡長めのセッションで確認 |
| MCP deferred tools数 | `settings.json` の MCP 設定と各 MCP サーバのツール数を確認 |

## プラグイン/スキル精査手順

`config_health` の `enabled_plugin_names` と `global_skill_names` を確認し、以下の基準で精査する:

### プラグイン精査基準
- **LSP系 (gopls, jdtls, typescript)**: 対象言語を使わないなら無効化
- **セキュリティ系**: 他プラグインと重複する機能があれば統合
- **開発支援系**: 使用頻度が低いものは無効化を検討

### スキル精査基準
- **プロジェクト固有スキル**: 特定リポジトリでしか使わないなら `.claude/skills/` に移動
- **同系統の重複**: 例えば Obsidian 関連が複数あれば統合を検討
- **使用頻度ゼロ**: 30日間のセッションで一度も invoke されていないスキルは移動候補
- **一回きりのセットアップ系**: 初期設定完了後は不要になるスキルは移動候補

### スキル統合の優先候補
- **Obsidian系 (4スキル)**: `obsidian-cli` を `obsidian-vault-management` に統合、`setup-obsidian-mcp` と `restructure-obsidian-vault` をプロジェクト固有に移動
- **オンデマンド有効化プラグイン**: `skill-creator`、`pr-review-toolkit` は日常不要｡必要時に `settings.json` で一時的に有効化する運用に切り替え

## 定期分析の推奨

月1回、以下の手順で分析を実施する:

1. `cd cmd && go run ./analyze-tokens --days 30` を実行
2. `warnings` セクションを確認し、`recommendation` に従って対処
3. `config_health` のプラグイン/スキル一覧を精査
4. 結果を前月と比較し、改善効果を確認
