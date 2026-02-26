# pup CLI コマンドリファレンス

Datadog API CLI ツール `pup` の主要コマンド詳細リファレンス。

## グローバルフラグ

| フラグ | デフォルト | 説明 |
|--------|----------|------|
| `--agent` | `false` | エージェントモード (AI コーディングアシスタントでは自動有効化) |
| `--output` | `json` | 出力形式 (`json`, `table`, `yaml`) |
| `--yes` | `false` | 確認プロンプトをスキップ |

> **Note**: Claude Code から実行する場合、`--agent` は自動検出されるため明示指定は不要。

## 認証

```bash
# OAuth2 (推奨)
pup auth login
pup auth status

# API キー方式
export DD_API_KEY="..."
export DD_APP_KEY="..."
export DD_SITE="datadoghq.com"  # リージョンに応じて変更
```

## 時間指定

`--from` / `--to` フラグで時間範囲を指定する。

| 形式 | 例 |
|------|-----|
| 相対 | `5s`, `30m`, `1h`, `4h`, `1d`, `7d`, `30d` |
| RFC3339 | `2024-01-01T00:00:00Z` |
| Unix ミリ秒 | `1704067200000` |

```bash
--from=1h                    # 1時間前から現在まで
--from=7d --to=1d            # 7日前から1日前まで
--from=2024-01-01T00:00:00Z  # 絶対時刻指定
```

> **重要**: `--from` を省略すると意図しない時間範囲になることがある。常に明示指定すること。

## ドメイン別コマンド

### dashboards

ダッシュボードの一覧・詳細取得。

```bash
# 一覧取得
pup dashboards list

# 特定ダッシュボードの詳細 (ウィジェット定義含む)
pup dashboards get <dashboard_id>

# ダッシュボード作成・更新・削除
pup dashboards create --title="..." --widgets='[...]'
pup dashboards update <dashboard_id> --title="..."
pup dashboards delete <dashboard_id>
```

### logs

ログの検索・集計・分析。

```bash
# キーワード・フィルタ検索
pup logs search --query="status:error service:web-app" --from=1h --limit=20

# 集計 (カウント、分布など)
pup logs aggregate --query="status:error" --from=1h --compute="count" --group-by="service"

# ログ一覧 (フィルタなし)
pup logs list --from=1h --limit=50
```

**クエリ構文**:
- `status:error` - ステータスフィルタ
- `service:web-app` - サービスフィルタ
- `@attr:val` - カスタム属性
- `host:i-*` - ワイルドカード
- `"exact phrase"` - 完全一致
- `AND` / `OR` / `NOT` - 論理演算子
- `-status:info` - 否定

> **ベストプラクティス**: ログの件数を知りたい場合は `logs aggregate --compute=count` を使う。`logs search` で全件取得してローカルでカウントしない。

### monitors

モニターの一覧・検索・詳細取得。

```bash
# タグでフィルタして一覧
pup monitors list --tags="env:production" --limit=50

# 名前でサブ文字列検索
pup monitors list --name="API latency"

# フルテキスト検索
pup monitors search --query="status:Alert"

# 特定モニターの詳細
pup monitors get <monitor_id>

# モニター作成・更新・削除
pup monitors create --type="metric alert" --query="..." --name="..."
pup monitors update <monitor_id> --name="..."
pup monitors delete <monitor_id>
```

### metrics

メトリクスのクエリ・検索。

```bash
# メトリクスクエリ (集約関数を必ず指定)
pup metrics query --query="avg:system.cpu.user{env:prod} by {host}" --from=1h

# メトリクス検索 (名前パターン)
pup metrics search --query="system.cpu"

# メトリクス一覧
pup metrics list

# メトリクスメタデータ
pup metrics metadata <metric_name>

# タグ情報
pup metrics tags <metric_name>
```

**クエリ構文**: `<aggregation>:<metric_name>{<filter>} by {<group>}`
- aggregation: `avg`, `sum`, `min`, `max`, `count`
- filter: `env:prod`, `service:web-app` 等
- group: `host`, `service` 等

> **重要**: `metrics query` では集約関数 (`avg`, `sum`, `max`, `min`, `count`) の指定が必須。

### slos

SLO の一覧・詳細・ステータス確認。

```bash
# SLO 一覧
pup slos list

# 特定 SLO の詳細
pup slos get <slo_id>

# SLO ステータス (バジェット残量など)
pup slos status <slo_id>

# SLO 作成・更新・削除
pup slos create --name="..." --type="..." --thresholds='[...]'
pup slos update <slo_id> --name="..."
pup slos delete <slo_id>
```

### incidents

インシデントの一覧・詳細取得。

```bash
# インシデント一覧
pup incidents list --query="status:active"

# 特定インシデントの詳細
pup incidents get <incident_id>

# インシデントの添付ファイル
pup incidents attachments <incident_id>

# ハンドル (対応者) 情報
pup incidents handles <incident_id>
```

### apm

APM サービス・エンティティ・依存関係の管理。

```bash
# サービス一覧
pup apm services list

# エンティティ一覧
pup apm entities list

# サービス依存関係
pup apm dependencies list

# フローマップ
pup apm flow-map
```

### traces

APM トレースの検索 (Datadog 側)。

```bash
# トレース一覧
pup traces list --query="service:web-app @duration:>5000000000" --from=1h --limit=20
```

**クエリ構文**:
- `service:<name>` - サービス名
- `resource_name:<path>` - リソース名 (URL パス等)
- `@duration:>5000000000` - 所要時間 (**ナノ秒**)
- `status:error` - エラーステータス
- `operation_name:<op>` - オペレーション名
- `env:production` - 環境

> **重要**: APM の duration は **ナノ秒** 単位。1秒 = 1,000,000,000、5ms = 5,000,000。

### infrastructure

インフラストラクチャホストの管理。

```bash
# ホスト一覧
pup infrastructure hosts list

# ホスト詳細 (フィルタ指定)
pup infrastructure hosts list --filter="env:production"
```

### events

Datadog イベントの一覧・検索。

```bash
# イベント一覧
pup events list --from=1d --limit=50

# イベント検索
pup events search --query="sources:pagerduty status:error" --from=1d

# 特定イベント取得
pup events get <event_id>
```

**クエリ構文**: `sources:nagios,pagerduty status:error priority:normal tags:env:prod`

### cloud

クラウドインテグレーションの管理。

```bash
# GCP インテグレーション一覧
pup cloud gcp list

# AWS インテグレーション一覧
pup cloud aws list

# Azure インテグレーション一覧
pup cloud azure list
```

## その他の主要コマンド

| ドメイン | コマンド例 | 説明 |
|---------|-----------|------|
| `audit-logs` | `search --query="*" --from=1d` | 監査ログ検索 |
| `synthetics` | `tests list` | Synthetic テスト一覧 |
| `rum` | `events list --from=1h` | RUM イベント |
| `security` | `signals list --query="status:critical"` | セキュリティシグナル |
| `downtime` | `list` | ダウンタイム一覧 |
| `notebooks` | `list` | ノートブック一覧 |
| `service-catalog` | `list` | サービスカタログ |
| `on-call` | `teams list` | オンコールチーム |
| `investigations` | `list` | Bits AI 調査 |

## アンチパターン

| やりがちなこと | なぜダメか | 代わりにやること |
|---------------|----------|----------------|
| `--from` を省略する | 意図しない時間範囲になる | 常に `--from` を明示指定 |
| `--limit=1000` を最初から使う | レスポンスが遅い | 小さい limit で絞ってからクエリを調整 |
| 全モニターを一覧して grep | 大規模環境で遅い (>10k) | `--tags` や `--name` でフィルタ |
| ログを全件取得してカウント | 非効率、タイムアウトしやすい | `logs aggregate --compute=count` |
| duration を秒と仮定 | APM は **ナノ秒** | 1s = 1,000,000,000 ns |
| `--from=30d` を安易に使う | 非常に遅い | まず `1h` から始めて必要に応じて拡大 |
| 401 エラーでリトライ | 認証切れ | `pup auth login` で再認証 |
