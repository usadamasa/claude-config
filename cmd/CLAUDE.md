# Go CLI ツール開発ガイド

## ディレクトリ構成

```
${REPO_ROOT}/cmd/
├── go.mod                # モジュール定義 (module: github.com/usadamasa/claude-config)
├── Taskfile.yml          # Go タスク定義
├── .go-arch-lint.yml     # パッケージ依存方向ルール
├── internal/             # 共通パッケージ (cmd/ 配下のコマンドからのみ参照可能)
│   ├── pathutil/         # パス解決ユーティリティ
│   ├── jsonlscan/        # JSONL ファイル走査
│   ├── settings/         # settings.json 読込
│   └── category/         # ドメイン/パーミッション分類の共通型
└── <コマンド名>/
    ├── main.go           # エントリポイント (package main)
    ├── main_test.go      # main.go のテスト
    ├── <機能>.go          # 機能ごとにファイル分割
    └── <機能>_test.go     # 対応するテスト
```

- 1コマンド = 1ディレクトリ｡全ファイル `package main`
- `go.mod` は `cmd/` 配下に配置 (module: `github.com/usadamasa/claude-config`)
- 外部依存は最小限に｡標準ライブラリで済むならそれで良い

## internal/ パッケージ

共通ロジックは `cmd/internal/` に配置する｡Go の `internal` 規約により `cmd/` 配下のパッケージからのみ参照可能｡

| パッケージ | 用途 | 使用元 |
|-----------|------|--------|
| `internal/pathutil` | `ResolveRealpath`, `ResolveProjectsDir`, `ResolveSettingsPath` | 全コマンド |
| `internal/jsonlscan` | `WalkJSONLFiles`, `NewScanner`, `CountUniqueFiles` | analyze-* |
| `internal/settings` | `Settings` 構造体, `Load()` | analyze-webfetch, analyze-permissions |
| `internal/category` | `Category` 型, `Result` 型, 共通定数 | analyze-webfetch, analyze-permissions |

### パッケージ依存ルール

`.go-arch-lint.yml` で依存方向を強制する:

- cmd パッケージ → internal パッケージ: 許可 (上記テーブルに従う)
- internal パッケージ → internal パッケージ: 禁止 (相互依存なし)
- internal パッケージ → cmd パッケージ: 禁止

## タスク実行

```sh
task go:test       # 全パッケージのテスト実行
task go:test-v     # 詳細モード (-v)
task go:test-cover # カバレッジ付きテスト
task go:vet        # 静的解析
task go:build      # ビルド
task go:lint       # golangci-lint
task go:lint:arch  # go-arch-lint (パッケージ依存方向チェック)
task go:lint:vuln  # govulncheck (脆弱性スキャン)
task go:lint:sec   # gosec (セキュリティ解析)
task go:lint:all   # 全静的解析を実行
task test          # bats + Go テストを統合実行
```

## 新しいコマンドの追加手順

1. `cmd/<コマンド名>/` ディレクトリを作成
2. `main.go` に `package main` と `func main()` を定義
3. テストファイル `*_test.go` を同ディレクトリに配置
4. `task go:test` で自動的にテスト対象に含まれる (`./...`)
5. `.go-arch-lint.yml` に新コンポーネントと依存ルールを追加

## テストの作法

- テーブル駆動テスト + `t.Run()` でサブテスト化
- 一時ファイルは `t.TempDir()` を使い自動クリーンアップ
- テスト用ヘルパー関数はテストファイル内に定義 (例: `writeTestFile`)
- 正常系と異常系の両方をカバー

## コーディング規約

- コメント､エラーメッセージ､CLI の説明文は日本語
- CLIフラグは `flag` パッケージを使用
- エラーは `fmt.Fprintf(os.Stderr, ...)` で出力し `os.Exit(1)`
- 構造化出力は JSON で stdout に出力
