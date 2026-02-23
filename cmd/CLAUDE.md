# Go CLI ツール開発ガイド

## ディレクトリ構成

```
${REPO_ROOT}/cmd/
└── <コマンド名>/
    ├── main.go           # エントリポイント (package main)
    ├── main_test.go      # main.go のテスト
    ├── <機能>.go          # 機能ごとにファイル分割
    └── <機能>_test.go     # 対応するテスト
```

- 1コマンド = 1ディレクトリ｡全ファイル `package main`
- `go.mod` はリポジトリルートに配置済み (module: `github.com/usadamasa/dotfile`)
- 外部依存は最小限に｡標準ライブラリで済むならそれで良い

## タスク実行

```sh
task go:test     # 全 cmd パッケージのテスト実行
task go:test-v   # 詳細モード (-v)
task go:vet      # 静的解析
task go:build    # ビルド
task test        # bats + Go テストを統合実行
```

## 新しいコマンドの追加手順

1. `cmd/<コマンド名>/` ディレクトリを作成
2. `main.go` に `package main` と `func main()` を定義
3. テストファイル `*_test.go` を同ディレクトリに配置
4. `task go:test` で自動的にテスト対象に含まれる (`./cmd/...`)

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
