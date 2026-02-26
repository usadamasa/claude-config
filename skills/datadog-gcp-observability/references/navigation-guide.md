# 関連サービス間ナビゲーションガイド

障害調査・運用監視における調査フローの決定木とサービス間の接続パターン。

## 症状別 調査開始点

```
症状を特定
├── アラート発火
│   └── → モニター詳細 → 関連メトリクス/ログ → トレース
├── エラーレート上昇
│   └── → ログ検索 (status:error) → 集計 (by service) → トレース
├── レイテンシ悪化
│   └── → トレース (latency:>閾値) → スパン分析 → ボトルネック特定
├── SLO バジェット消費
│   └── → SLO ステータス → 関連モニター → ログ/トレース
├── インシデント発生
│   └── → インシデント詳細 → タイムライン → 関連モニター/ログ
├── ユーザー報告
│   └── → ログ検索 (リクエスト情報) → トレース → 依存サービス
└── 定期チェック
    └── → ダッシュボード → 異常値確認 → 深掘り
```

## 調査フロー詳細

### フロー 1: アラート起点の調査

```bash
# 1. アラート中のモニターを確認
pup monitors search --query="status:Alert"

# 2. モニター詳細からクエリを確認
pup monitors get <monitor_id>
# → query フィールドからメトリクス名やフィルタ条件を取得

# 3. 関連メトリクスを直接クエリ
pup metrics query --query="<monitor_query>" --from=1h

# 4. 関連ログを検索
pup logs search --query="service:<affected_service> status:error" --from=1h

# 5. トレースで詳細を確認
pup traces list --query="service:<affected_service> status:error" --from=1h --limit=10
```

### フロー 2: エラー起点の調査

```bash
# 1. エラーの分布を確認
pup logs aggregate --query="status:error" --from=1h --compute="count" --group-by="service"

# 2. 最もエラーが多いサービスのログを詳細確認
pup logs search --query="status:error service:<top_service>" --from=1h --limit=20

# 3. ログからトレース ID を抽出 (dd.trace_id 属性)
# → jq で dd.trace_id を抽出

# 4. Datadog APM でトレースを確認
pup traces list --query="service:<top_service> status:error" --from=1h --limit=5

# 5. GCP Cloud Trace で詳細確認 (GCP トレーシングの場合)
# → gcloud-traces.md の手順に従う
```

### フロー 3: レイテンシ起点の調査

```bash
# 1. 遅いトレースを検索
pup traces list --query="service:<service> @duration:>5000000000" --from=1h --limit=10

# 2. サービス依存関係を確認
pup apm dependencies list

# 3. 依存サービスのメトリクスを確認
pup metrics query --query="avg:trace.servlet.request.duration{service:<dep_service>} by {resource_name}" --from=1h

# 4. GCP Cloud Trace でウォーターフォール確認
# → 各スパンの所要時間からボトルネックを特定
```

### フロー 4: SLO 起点の調査

```bash
# 1. SLO ステータス確認
pup slos list
pup slos status <slo_id>

# 2. SLO に紐づくモニターを確認
pup slos get <slo_id>
# → monitor_ids フィールドからモニター ID を取得

# 3. 関連モニターの詳細
pup monitors get <monitor_id>

# 4. 根本原因の調査 (ログ/トレース)
# → フロー 1 or 2 に合流
```

## Datadog → GCP Cloud Trace への接続

Datadog のログ/トレースから GCP Cloud Trace のトレース ID を抽出する方法。

### パターン A: Datadog ログから dd.trace_id を抽出

```bash
# Datadog ログに dd.trace_id が含まれている場合
pup logs search --query="service:<service> status:error" --from=1h --limit=5 \
  | jq -r '.logs[].attributes.attributes["dd.trace_id"] // empty'
```

### パターン B: W3C Trace Context (traceparent) から変換

W3C Trace Context を使用している場合、Datadog と GCP で同じトレース ID を共有できる。

```
traceparent: 00-<trace_id>-<span_id>-01
                  ^^^^^^^^^ これが Cloud Trace の trace_id
```

### パターン C: GCP ログからトレース ID を抽出

```bash
# Cloud Logging でエラーログを検索し、トレース ID を抽出
gcloud logging read \
  'severity>=ERROR AND resource.labels.namespace_name="production"' \
  --project="${PROJECT_ID}" \
  --format='value(trace)' \
  --limit=10 \
  | sed 's|projects/.*/traces/||'
```

### パターン D: リクエスト情報から双方で検索

```bash
# Datadog 側
pup logs search --query='@http.url:"/api/users/123"' --from=1h --limit=5

# GCP Cloud Trace 側
TOKEN=$(gcloud auth print-access-token)
PROJECT_ID=$(gcloud config get-value project)
FILTER="url:/api/users/123"
ENCODED_FILTER=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$FILTER'))")
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://cloudtrace.googleapis.com/v1/projects/${PROJECT_ID}/traces?filter=${ENCODED_FILTER}&pageSize=5&view=ROOTSPAN" \
  | jq .
```

## pup 未対応機能の代替手段

| 機能 | pup の状況 | 代替手段 |
|------|-----------|---------|
| トレース詳細取得 | `traces list` のみ | Datadog UI (`/apm/trace/<id>`) で確認 |
| ダッシュボード URL 生成 | `get` でID取得可 | `url-patterns.md` のパターンで構築 |
| ログのリアルタイムテール | 未対応 | Datadog UI のライブテール |
| カスタムメトリクス送信 | `metrics submit` で対応 | - |
| サービスマップ表示 | `apm flow-map` で対応 | - |
| RUM セッションリプレイ | 未対応 | Datadog UI の RUM セクション |
| Continuous Profiler | 未対応 | Datadog UI の Profiling セクション |
| Watchdog Insights | 未対応 | Datadog UI の Watchdog セクション |
| Notebooks 共有・コラボ | CRUD のみ | Datadog UI |
| ダッシュボード共有リンク | 未対応 | Datadog UI の Share メニュー |

## サービス間の接続マップ

```
Datadog                          GCP
┌─────────────┐                  ┌─────────────────┐
│ Monitors    │──(alert)──────→  │                 │
│ Dashboards  │──(metrics)────→  │ Cloud           │
│ Logs        │◄─(export)─────── │ Logging         │
│ APM/Traces  │◄─(W3C trace)──── │ Cloud Trace     │
│ SLOs        │──(budget)─────→  │                 │
│ Incidents   │                  │ Cloud           │
│ Infra       │◄─(GCP integ)──── │ Monitoring      │
└─────────────┘                  └─────────────────┘
      │                                │
      └─── pup CLI ───┐  ┌── gcloud + curl ──┘
                       ▼  ▼
                  Claude Code
```
