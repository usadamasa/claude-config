# Cloud Trace 操作リファレンス

GCP Cloud Trace の REST API を使った分散トレースの検索・取得方法。
`gcloud traces` コマンドは存在しないため、`curl` + REST API で操作する。

## 認証パターン

```bash
# アクセストークンの取得
TOKEN=$(gcloud auth print-access-token)

# プロジェクト ID の取得
PROJECT_ID=$(gcloud config get-value project)
```

> **Note**: トークンは 1 時間で失効する。長時間の調査では再取得が必要。

## REST API v1: トレース取得

### 単一トレースの取得

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://cloudtrace.googleapis.com/v1/projects/${PROJECT_ID}/traces/${TRACE_ID}" \
  | jq .
```

**レスポンス構造**:
```json
{
  "projectId": "my-project",
  "traceId": "abc123...",
  "spans": [
    {
      "spanId": "1234567890",
      "name": "GET /api/users",
      "startTime": "2024-01-01T00:00:00.000Z",
      "endTime": "2024-01-01T00:00:00.500Z",
      "parentSpanId": "0",
      "labels": {
        "http/method": "GET",
        "http/status_code": "200",
        "http/url": "/api/users"
      }
    }
  ]
}
```

### トレース一覧の取得

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://cloudtrace.googleapis.com/v1/projects/${PROJECT_ID}/traces?filter=${FILTER}&startTime=${START_TIME}&endTime=${END_TIME}&pageSize=20" \
  | jq .
```

**パラメータ**:

| パラメータ | 必須 | 説明 |
|-----------|------|------|
| `filter` | いいえ | フィルタ式 (後述) |
| `startTime` | いいえ | 開始時刻 (RFC3339) |
| `endTime` | いいえ | 終了時刻 (RFC3339) |
| `pageSize` | いいえ | 結果数 (デフォルト: 10, 最大: 1000) |
| `pageToken` | いいえ | ページネーショントークン |
| `orderBy` | いいえ | `startTime desc` など |
| `view` | いいえ | `MINIMAL`, `ROOTSPAN`, `COMPLETE` |

**view パラメータ**:
- `MINIMAL`: トレース ID と概要のみ (高速)
- `ROOTSPAN`: ルートスパンの情報を含む
- `COMPLETE`: 全スパンの詳細を含む (遅い)

### フィルタ構文

| フィルタ | 説明 | 例 |
|---------|------|-----|
| `root:` | ルートスパン名のプレフィックス | `root:/api/users` |
| `span:` | 任意のスパン名のプレフィックス | `span:datastore` |
| `latency:` | レイテンシフィルタ | `latency:>500ms`, `latency:>1s` |
| `method:` | HTTP メソッド | `method:GET` |
| `url:` | URL パスプレフィックス | `url:/api/` |
| `+<label>:` | ラベルフィルタ | `+/http/status_code:500` |

**複合フィルタ**:
```bash
# 500ms 以上のレイテンシで /api/ パスのトレース
FILTER="root:/api/ latency:>500ms"

# HTTP 500 エラーのトレース
FILTER="+/http/status_code:500"

# 特定のスパン名を含むトレース
FILTER="span:CloudSQL"
```

**URL エンコードが必要**:
```bash
# フィルタ文字列を URL エンコード
ENCODED_FILTER=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$FILTER'))")

curl -s -H "Authorization: Bearer $TOKEN" \
  "https://cloudtrace.googleapis.com/v1/projects/${PROJECT_ID}/traces?filter=${ENCODED_FILTER}&startTime=$(date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)&endTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)&pageSize=20&view=ROOTSPAN" \
  | jq .
```

## gcloud logging read によるトレース相関

Cloud Logging のログからトレース ID で相関する。

```bash
# 特定のトレース ID に紐づくログを取得
gcloud logging read \
  'trace="projects/'"${PROJECT_ID}"'/traces/'"${TRACE_ID}"'"' \
  --project="${PROJECT_ID}" \
  --format=json \
  --limit=50

# サービス名 + エラーレベルで絞り込み
gcloud logging read \
  'trace="projects/'"${PROJECT_ID}"'/traces/'"${TRACE_ID}"'" AND severity>=ERROR' \
  --project="${PROJECT_ID}" \
  --format=json

# 時間範囲指定
gcloud logging read \
  'trace="projects/'"${PROJECT_ID}"'/traces/'"${TRACE_ID}"'"' \
  --project="${PROJECT_ID}" \
  --format=json \
  --freshness=1h
```

### ログからトレース ID を抽出

```bash
# Cloud Logging でエラーログを検索し、トレース ID を抽出
gcloud logging read \
  'severity>=ERROR AND resource.type="k8s_container" AND resource.labels.namespace_name="production"' \
  --project="${PROJECT_ID}" \
  --format='value(trace)' \
  --limit=10 \
  | sed 's|projects/.*/traces/||'
```

## Cloud Trace v2 API

v2 API はスパンの書き込み用。読み取りには v1 API を使用する。

```bash
# v2 でスパンをバッチ書き込み (通常はアプリ側のライブラリが自動で行う)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "https://cloudtrace.googleapis.com/v2/projects/${PROJECT_ID}/traces:batchWrite" \
  -d '{"spans": [...]}'
```

## BigQuery 代替 (大量データ分析)

Cloud Trace データを BigQuery にエクスポートしている場合、SQL で分析可能。

```bash
# BigQuery でトレースデータを分析
bq query --use_legacy_sql=false '
SELECT
  trace_id,
  span_name,
  TIMESTAMP_DIFF(end_time, start_time, MILLISECOND) AS duration_ms,
  status.code AS status_code
FROM `'"${PROJECT_ID}"'.trace_export._AllSpans`
WHERE
  start_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 HOUR)
  AND span_name LIKE "/api/%"
ORDER BY duration_ms DESC
LIMIT 20
'

# エラーの多いスパンをランキング
bq query --use_legacy_sql=false '
SELECT
  span_name,
  COUNT(*) AS error_count,
  AVG(TIMESTAMP_DIFF(end_time, start_time, MILLISECOND)) AS avg_duration_ms
FROM `'"${PROJECT_ID}"'.trace_export._AllSpans`
WHERE
  start_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 24 HOUR)
  AND status.code != 0
GROUP BY span_name
ORDER BY error_count DESC
LIMIT 10
'
```

> **Note**: BigQuery エクスポートはプロジェクトで設定が必要。未設定の場合は REST API を使用する。

## トラブルシューティング

| 問題 | 原因 | 対処 |
|------|------|------|
| 401 Unauthorized | トークン失効 | `TOKEN=$(gcloud auth print-access-token)` で再取得 |
| 403 Forbidden | Cloud Trace API 未有効化 / 権限不足 | `gcloud services enable cloudtrace.googleapis.com` |
| トレースが見つからない | 保持期間超過 (デフォルト 30 日) | BigQuery エクスポートを検討 |
| スパンが欠落 | サンプリング | トレーサーのサンプリングレートを確認 |
