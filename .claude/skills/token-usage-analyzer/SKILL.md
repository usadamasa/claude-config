---
name: token-usage-analyzer
description: |
  Claude Codeセッションのtoken使用量を分析し、改善ポイントを提示するスキル。
  「token使用量を分析して」「token消費を調べて」「コスト分析」「使用量レポート」
  のようにtoken使用量の分析や改善提案を依頼されたときに使用する。
---

# Token使用量分析

## CLIツールの実行

```bash
# 直近30日間の分析(デフォルト)
go run ~/src/github.com/usadamasa/dotfile/main/cmd/analyze-tokens

# 期間指定
go run ~/src/github.com/usadamasa/dotfile/main/cmd/analyze-tokens --days 7

# Top N変更
go run ~/src/github.com/usadamasa/dotfile/main/cmd/analyze-tokens --top 20

# カスタムディレクトリ指定
go run ~/src/github.com/usadamasa/dotfile/main/cmd/analyze-tokens --dir ~/.claude/projects
```

## 出力フォーマット

JSON形式で以下のセクションを出力する:

### summary (全体統計)
- `total_sessions`: セッション数
- `total_input_tokens`: 総input tokens
- `total_output_tokens`: 総output tokens
- `total_api_calls`: 総APIコール数
- `average_input_per_call`: 1 APIコールあたりの平均input tokens

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

### 改善アクションの判断基準

1. **average_input_per_call > 80K の場合:**
   - 有効プラグイン数を確認 (`settings.json` の `enabledPlugins`)
   - グローバルスキル数を確認 (`~/.claude/skills/`)
   - MCP deferred toolsの数を確認
   - 不要なプラグイン/スキルを無効化または移動

2. **特定プロジェクトのinput tokensが突出している場合:**
   - プロジェクト固有のCLAUDE.mdが肥大化していないか確認
   - プロジェクトローカルのスキルやMCP設定を確認

3. **api_calls/user_messages比率が高い場合:**
   - subagentの使用パターンを見直す
   - Explore agentに `model: "haiku"` を指定しているか確認

4. **cache_read_tokens = 0 の場合:**
   - セッションが短すぎてキャッシュが効かない可能性
   - セッション内で同じプロンプトパターンを繰り返す場合はキャッシュが有効

## 定期分析の推奨

月1回、以下の手順で分析を実施する:

1. `go run ~/src/github.com/usadamasa/dotfile/main/cmd/analyze-tokens --days 30` を実行
2. average_input_per_call の推移を確認
3. 未使用プラグインやスキルがあれば無効化/移動
4. 結果を前月と比較し、改善効果を確認
