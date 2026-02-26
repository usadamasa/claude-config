# URL 構築パターン

Datadog と GCP コンソールの URL 構築パターン。
調査中に関連ページへのリンクを提示する際に使用する。

## DD_SITE リージョン対応表

| リージョン | DD_SITE | ドメイン |
|-----------|---------|---------|
| US1 (デフォルト) | `datadoghq.com` | `app.datadoghq.com` |
| US3 | `us3.datadoghq.com` | `us3.datadoghq.com` |
| US5 | `us5.datadoghq.com` | `us5.datadoghq.com` |
| EU1 | `datadoghq.eu` | `app.datadoghq.eu` |
| AP1 | `ap1.datadoghq.com` | `ap1.datadoghq.com` |
| US1-FED | `ddog-gov.com` | `app.ddog-gov.com` |

**ドメイン解決ロジック**:
```
DD_SITE が "datadoghq.com" → "app.datadoghq.com"
DD_SITE が "datadoghq.eu"  → "app.datadoghq.eu"
それ以外                    → DD_SITE をそのまま使用
```

以下のパターンでは `${DD_DOMAIN}` を上記で解決したドメインとする。

## Datadog URL パターン

### ダッシュボード

```
# ダッシュボード一覧
https://${DD_DOMAIN}/dashboard/lists

# 特定ダッシュボード
https://${DD_DOMAIN}/dashboard/${DASHBOARD_ID}

# ダッシュボード (時間範囲指定)
https://${DD_DOMAIN}/dashboard/${DASHBOARD_ID}?from_ts=${FROM_UNIX_MS}&to_ts=${TO_UNIX_MS}

# ダッシュボード (テンプレート変数指定)
https://${DD_DOMAIN}/dashboard/${DASHBOARD_ID}?tpl_var_env=production&tpl_var_service=web-app
```

### ログエクスプローラー

```
# ログエクスプローラー (クエリ指定)
https://${DD_DOMAIN}/logs?query=${URL_ENCODED_QUERY}

# ログエクスプローラー (クエリ + 時間範囲)
https://${DD_DOMAIN}/logs?query=${URL_ENCODED_QUERY}&from_ts=${FROM_UNIX_MS}&to_ts=${TO_UNIX_MS}

# 特定ログのパーマリンク (ログ ID)
https://${DD_DOMAIN}/logs?query=${URL_ENCODED_QUERY}&log_id=${LOG_ID}

# 例: エラーログを1時間分表示
https://${DD_DOMAIN}/logs?query=status%3Aerror%20service%3Aweb-app&from_ts=1704063600000&to_ts=1704067200000
```

### モニター

```
# モニター一覧
https://${DD_DOMAIN}/monitors/manage

# 特定モニター
https://${DD_DOMAIN}/monitors/${MONITOR_ID}

# モニター編集
https://${DD_DOMAIN}/monitors/${MONITOR_ID}/edit

# アラート中のモニター
https://${DD_DOMAIN}/monitors/manage?q=status%3AAlert

# タグでフィルタ
https://${DD_DOMAIN}/monitors/manage?q=tag%3A"env%3Aproduction"
```

### SLO

```
# SLO 一覧
https://${DD_DOMAIN}/slo/manage

# 特定 SLO
https://${DD_DOMAIN}/slo?slo_id=${SLO_ID}
```

### インシデント

```
# インシデント一覧
https://${DD_DOMAIN}/incidents

# 特定インシデント
https://${DD_DOMAIN}/incidents/${INCIDENT_ID}

# アクティブなインシデント
https://${DD_DOMAIN}/incidents?query=status%3Aactive
```

### APM

```
# サービスマップ
https://${DD_DOMAIN}/apm/map

# サービス一覧
https://${DD_DOMAIN}/apm/services

# 特定サービスの概要
https://${DD_DOMAIN}/apm/services/${SERVICE_NAME}

# トレースエクスプローラー
https://${DD_DOMAIN}/apm/traces?query=${URL_ENCODED_QUERY}

# 特定トレース
https://${DD_DOMAIN}/apm/trace/${TRACE_ID}

# サービスのリソース一覧
https://${DD_DOMAIN}/apm/resource/${SERVICE_NAME}/${RESOURCE_NAME}
```

### インフラストラクチャ

```
# ホストマップ
https://${DD_DOMAIN}/infrastructure/map

# ホスト一覧
https://${DD_DOMAIN}/infrastructure

# 特定ホスト
https://${DD_DOMAIN}/infrastructure?host=${HOST_NAME}

# コンテナ一覧
https://${DD_DOMAIN}/containers
```

### メトリクスエクスプローラー

```
# メトリクスエクスプローラー
https://${DD_DOMAIN}/metric/explorer?exp_metric=${METRIC_NAME}&exp_agg=avg&exp_row_type=metric

# メトリクスサマリー
https://${DD_DOMAIN}/metric/summary?filter=${METRIC_NAME}
```

## GCP コンソール URL パターン

### Cloud Trace エクスプローラー

```
# トレースエクスプローラー (プロジェクト指定)
https://console.cloud.google.com/traces/list?project=${PROJECT_ID}

# 特定トレース
https://console.cloud.google.com/traces/list?project=${PROJECT_ID}&tid=${TRACE_ID}

# フィルタ付き
https://console.cloud.google.com/traces/list?project=${PROJECT_ID}&filter=${URL_ENCODED_FILTER}

# トレース詳細 (ウォーターフォールビュー)
https://console.cloud.google.com/traces/overview?project=${PROJECT_ID}&tid=${TRACE_ID}
```

### Cloud Logging

```
# ログエクスプローラー
https://console.cloud.google.com/logs/query?project=${PROJECT_ID}

# クエリ付き
https://console.cloud.google.com/logs/query;query=${URL_ENCODED_QUERY}?project=${PROJECT_ID}

# トレース ID で絞り込み
https://console.cloud.google.com/logs/query;query=trace%3D%22projects%2F${PROJECT_ID}%2Ftraces%2F${TRACE_ID}%22?project=${PROJECT_ID}
```

### Cloud Monitoring

```
# ダッシュボード一覧
https://console.cloud.google.com/monitoring/dashboards?project=${PROJECT_ID}

# メトリクスエクスプローラー
https://console.cloud.google.com/monitoring/metrics-explorer?project=${PROJECT_ID}

# アラートポリシー
https://console.cloud.google.com/monitoring/alerting?project=${PROJECT_ID}
```

## URL 構築ヘルパー

### Bash でのURL エンコード

```bash
# Python を使った URL エンコード
url_encode() {
  python3 -c "import urllib.parse; print(urllib.parse.quote('$1', safe=''))"
}

# 使用例
QUERY="status:error service:web-app"
ENCODED=$(url_encode "$QUERY")
echo "https://app.datadoghq.com/logs?query=${ENCODED}"
```

### DD_DOMAIN の解決

```bash
resolve_dd_domain() {
  local site="${DD_SITE:-datadoghq.com}"
  case "$site" in
    datadoghq.com) echo "app.datadoghq.com" ;;
    datadoghq.eu)  echo "app.datadoghq.eu" ;;
    *)             echo "$site" ;;
  esac
}

DD_DOMAIN=$(resolve_dd_domain)
```
