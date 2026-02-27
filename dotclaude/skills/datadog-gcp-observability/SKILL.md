---
name: datadog-gcp-observability
description: >-
  Datadog observability workflows for GCP environments using pup CLI and gcloud.
  Use when investigating dashboards, analyzing logs, tracing requests, or navigating
  Datadog services. Triggers: "Datadogで確認", "ログ調査", "トレース",
  "ダッシュボード", "SLO確認", "monitor確認", "pup", "障害調査", "observability".
---

# Datadog GCP Observability

Datadog の CLI ツール `pup` と GCP の `gcloud` を組み合わせた障害調査・運用監視ワークフロー。

## Context

- GCP project: !`gcloud config get-value project 2>/dev/null || echo "未設定"`
- DD_SITE: !`echo "${DD_SITE:-datadoghq.com}"`
- pup version: !`pup version 2>/dev/null || echo "未インストール"`
- pup auth: !`pup auth status 2>&1 | head -3 || echo "認証状態不明"`
- gcloud auth: !`gcloud auth list --filter=status:ACTIVE --format='value(account)' 2>/dev/null | head -1 || echo "未認証"`

## クイックリファレンス

| やりたいこと | コマンド |
|-------------|---------|
| ダッシュボード一覧 | `pup dashboards list` |
| ダッシュボード詳細 | `pup dashboards get <id>` |
| ログ検索 | `pup logs search --query="status:error service:<name>" --from=1h` |
| ログ集計 | `pup logs aggregate --query="status:error" --from=1h --compute="count" --group-by="service"` |
| モニター一覧 | `pup monitors list --tags="env:production"` |
| アラート中モニター | `pup monitors search --query="status:Alert"` |
| メトリクスクエリ | `pup metrics query --query="avg:<metric>{<filter>}" --from=1h` |
| SLO ステータス | `pup slos status <id>` |
| APM サービス一覧 | `pup apm services list` |
| トレース検索 | `pup traces list --query="service:<name>" --from=1h` |
| Cloud Trace 取得 | `curl -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://cloudtrace.googleapis.com/v1/projects/$(gcloud config get-value project)/traces/<trace_id>"` |
| インシデント一覧 | `pup incidents list --query="status:active"` |

## ワークフロー 1: ダッシュボード調査

ダッシュボードからメトリクスの異常値を確認する。

### Step 1: ダッシュボード一覧の取得

```bash
pup dashboards list
```

対象のダッシュボードを特定する。ID は `id` フィールド。

### Step 2: ダッシュボード詳細の取得

```bash
pup dashboards get <dashboard_id>
```

ウィジェット定義からメトリクス名やクエリを把握する。

### Step 3: コンソール URL の構築

```bash
# DD_DOMAIN を解決
DD_DOMAIN=$(case "${DD_SITE:-datadoghq.com}" in
  datadoghq.com) echo "app.datadoghq.com" ;;
  datadoghq.eu)  echo "app.datadoghq.eu" ;;
  *)             echo "${DD_SITE}" ;;
esac)

echo "https://${DD_DOMAIN}/dashboard/<dashboard_id>"
```

### Step 4: 次のアクション

ダッシュボードのメトリクスに異常値が見つかった場合:
- **エラーレート上昇** → ワークフロー 2 (ログ調査) へ
- **レイテンシ上昇** → ワークフロー 3 (トレース調査) へ
- **リソース使用率** → `pup metrics query` で詳細確認

## ワークフロー 2: ログ調査

ログの検索・集計からエラーパターンを分析する。

### Step 1: 情報収集

調査対象を特定する:
- サービス名
- 時間範囲
- エラーの種類 (ステータスコード、例外等)

### Step 2: ログ検索

```bash
# エラーログを検索
pup logs search --query="status:error service:<service_name>" --from=1h --limit=20

# 特定の属性で絞り込み
pup logs search --query="service:<name> @http.status_code:500" --from=1h --limit=20
```

### Step 3: ログ集計

```bash
# サービス別エラー数
pup logs aggregate --query="status:error" --from=1h --compute="count" --group-by="service"

# ステータスコード別集計
pup logs aggregate --query="service:<name>" --from=1h --compute="count" --group-by="@http.status_code"
```

> **重要**: 件数を知りたい場合は `aggregate` を使う。`search` で全件取得してローカルでカウントしない。

### Step 4: コンソール URL の構築

```bash
QUERY="status:error service:<service_name>"
ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$QUERY', safe=''))")
echo "https://${DD_DOMAIN}/logs?query=${ENCODED}"
```

### Step 5: トレース ID の抽出

ログに `dd.trace_id` 属性がある場合、トレース調査へ連携する。

```bash
# ログからトレース ID を抽出
pup logs search --query="status:error service:<name>" --from=1h --limit=5 \
  | jq -r '.logs[].attributes.attributes["dd.trace_id"] // empty'
```

トレース ID が取得できたら → ワークフロー 3 (トレース調査) へ。

## ワークフロー 3: 分散トレース調査

リクエストの分散トレースを取得・分析する。

### Step 1: GCP 設定の確認

```bash
PROJECT_ID=$(gcloud config get-value project)
TOKEN=$(gcloud auth print-access-token)
```

### Step 2: トレースの検索・取得

#### Case A: Datadog APM でトレース検索

```bash
# サービス名でトレースを検索
pup traces list --query="service:<name> status:error" --from=1h --limit=10

# レイテンシが高いトレースを検索 (duration はナノ秒)
pup traces list --query="service:<name> @duration:>5000000000" --from=1h --limit=10
```

> **注意**: APM の duration は**ナノ秒**。1秒 = 1,000,000,000。

#### Case B: GCP Cloud Trace で検索

```bash
# 時間範囲を設定
START_TIME=$(date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)
END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# フィルタ付きでトレース一覧取得
FILTER="root:/api/ latency:>500ms"
ENCODED_FILTER=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$FILTER'))")

curl -s -H "Authorization: Bearer $TOKEN" \
  "https://cloudtrace.googleapis.com/v1/projects/${PROJECT_ID}/traces?filter=${ENCODED_FILTER}&startTime=${START_TIME}&endTime=${END_TIME}&pageSize=20&view=ROOTSPAN" \
  | jq .
```

#### Case C: 特定トレース ID で取得

```bash
# Cloud Trace API で特定トレースを取得
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://cloudtrace.googleapis.com/v1/projects/${PROJECT_ID}/traces/${TRACE_ID}" \
  | jq .
```

### Step 3: コンソール URL の構築

```bash
# GCP Cloud Trace コンソール
echo "https://console.cloud.google.com/traces/list?project=${PROJECT_ID}&tid=${TRACE_ID}"

# Datadog APM トレース
echo "https://${DD_DOMAIN}/apm/trace/${TRACE_ID}"
```

### Step 4: Datadog APM との相関

```bash
# サービスの依存関係を確認
pup apm dependencies list

# フローマップで全体像を把握
pup apm flow-map

# 関連ログを Cloud Logging で確認
gcloud logging read \
  'trace="projects/'"${PROJECT_ID}"'/traces/'"${TRACE_ID}"'"' \
  --project="${PROJECT_ID}" \
  --format=json \
  --limit=50
```

## ワークフロー 4: 関連サービスナビゲーション

調査中に関連サービスへ遷移するパターン。

### Case A: モニター調査

```bash
# アラート中のモニター
pup monitors search --query="status:Alert"

# タグでフィルタ
pup monitors list --tags="env:production,team:backend" --limit=50

# 特定モニターの詳細
pup monitors get <monitor_id>

# コンソール URL
echo "https://${DD_DOMAIN}/monitors/<monitor_id>"
```

### Case B: SLO 確認

```bash
# SLO 一覧
pup slos list

# SLO ステータス (バジェット残量)
pup slos status <slo_id>

# SLO 詳細 (紐づくモニター ID を確認)
pup slos get <slo_id>

# コンソール URL
echo "https://${DD_DOMAIN}/slo?slo_id=<slo_id>"
```

### Case C: インシデント確認

```bash
# アクティブなインシデント
pup incidents list --query="status:active"

# インシデント詳細
pup incidents get <incident_id>

# コンソール URL
echo "https://${DD_DOMAIN}/incidents/<incident_id>"
```

### Case D: APM サービス確認

```bash
# サービス一覧
pup apm services list

# サービスの依存関係
pup apm dependencies list

# サービスマップ
pup apm flow-map

# コンソール URL
echo "https://${DD_DOMAIN}/apm/services/<service_name>"
```

### Case E: インフラ確認

```bash
# ホスト一覧
pup infrastructure hosts list

# フィルタ付き
pup infrastructure hosts list --filter="env:production"

# コンソール URL
echo "https://${DD_DOMAIN}/infrastructure"
```

## リファレンス

| ファイル | 内容 | 参照タイミング |
|---------|------|--------------|
| `references/pup-commands.md` | pup CLI コマンド詳細、クエリ構文、アンチパターン | コマンドの詳細オプションを確認するとき |
| `references/gcloud-traces.md` | Cloud Trace REST API、フィルタ構文、BigQuery 代替 | GCP トレース操作の詳細を確認するとき |
| `references/url-patterns.md` | Datadog/GCP コンソール URL 構築パターン | ユーザーにコンソール URL を提示するとき |
| `references/navigation-guide.md` | 症状別調査フロー、サービス間接続パターン | 調査の方向性を決めるとき |

## 注意事項

- **エージェントモード**: Claude Code から `pup` を実行する場合、`--agent` は自動検出されるため明示指定不要
- **DD_SITE**: 未設定の場合 `datadoghq.com` (US1) がデフォルト。環境に応じて `DD_SITE` を設定すること
- **トークン失効**: `gcloud auth print-access-token` のトークンは 1 時間で失効する。長時間の調査では再取得が必要
- **時間指定**: `pup` の `--from` は常に明示指定する。省略すると意図しない時間範囲になる
- **APM duration**: ナノ秒単位。1秒 = 1,000,000,000
- **ログ件数**: `logs aggregate --compute=count` を使う。`logs search` で全件取得してカウントしない
- **大規模環境**: モニターやログの一覧取得時は `--tags` や `--query` でフィルタしてから取得する
