# ツール詳細リファレンス

## ツール管理: aqua + direnv

本スキルのツール群はすべて **aqua** で管理する。
aqua は `aqua.yaml` に依存関係を宣言するだけで、バージョン固定・自動インストールが可能。

### aqua 自体のインストール

```bash
# Homebrew で aqua をインストール (推奨)
brew install aquaproj/aqua/aqua

# または Go でインストール
go install github.com/aquaproj/aqua/v2/cmd/aqua@latest

# PATH に追加 (~/.zshrc や ~/.bashrc に追記)
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
```

### aqua.yaml へのツール追加

プロジェクトルートの `aqua.yaml` に以下を追加する:

```yaml
---
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json
aqua_version: ">=2.0.0"

registries:
  - type: standard
    ref: v4.227.0  # 最新バージョンに更新してください

packages:
  # golangci-lint: テスト可能性・保守性メトリクス
  - name: golangci/golangci-lint@v1.62.2
  # go-arch-lint: パッケージ依存方向チェック
  - name: fe3dback/go-arch-lint@v1.14.0
```

aqua を使ったツールのインストール:

```bash
# aqua.yaml に記載したツールをインストール
aqua install

# または特定のツールだけインストール
aqua install golangci/golangci-lint
```

### direnv による自動読み込み

**direnv** を使うと、プロジェクトディレクトリに入るだけで `aqua` のパスが自動的に有効になる。

```bash
# direnv のインストール
brew install direnv

# ~/.zshrc に追加 (zsh の場合)
eval "$(direnv hook zsh)"

# ~/.bashrc に追加 (bash の場合)
eval "$(direnv hook bash)"
```

プロジェクトルートに `.envrc` を作成:

```bash
# .envrc
# aqua のバイナリパスを PATH に追加
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
```

`.envrc` を許可:

```bash
direnv allow .
```

これでプロジェクトディレクトリに入ると自動的に aqua のツールが利用可能になる。

---

## golangci-lint

### 概要

複数の linter をまとめて実行する Go 静的解析ツール。
アーキテクチャメトリクス向けには以下の linter を有効化する:

| Linter | 計測するメトリクス | しきい値 | 根拠 |
|--------|------------------|---------|------|
| `gocognit` | 認知的複雑度 | ≤ 20 | 20 超えるとテストが指数関数的に困難になる |
| `gocyclo` | 循環複雑度 (McCabe) | ≤ 20 | 20 超えると分岐網羅テストが非現実的 |
| `funlen` | 関数の行数・文数 | ≤ 100行 / 60文 | 画面1枚で収まる範囲 |
| `nestif` | ネストの深さ | ≤ 5 | 5 超えるとコードの流れが追えなくなる |
| `maintidx` | 保守性指数 | ≥ 20 | 20 未満は実質的なブラックボックス |
| `deadcode` | 到達不能コード | 0件 | 死んだコードは誤解とバグの温床 |
| `staticcheck` | 高度なバグ検出 | デフォルト | nil参照, 不正なAPIUsage等 |
| `depguard` | import 禁止リスト | 指定パッケージ=0 | 意図しない依存を防ぐ |

### インストールと実行

```bash
# aqua 経由 (推奨)
aqua install

# 手動インストール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 実行
golangci-lint run ./...

# 特定 linter のみ
golangci-lint run --enable-only gocognit,gocyclo ./...

# JSON 出力
golangci-lint run --out-format json ./... 2>/dev/null
```

### 出力例

```
cmd/server/handler.go:42:1: Function 'HandleRequest' has too high cognitive complexity (25 > 20) (gocognit)
cmd/server/handler.go:42:1: Function 'HandleRequest' is too long (145 > 100) (funlen)
internal/usecase/order.go:89:5: nestif: nesting depth 6 > 5 (nestif)
```

---

## go-arch-lint

### 概要

パッケージ間の依存方向を `.go-arch-lint.yml` で宣言し、違反を検出するツール。
レイヤードアーキテクチャ・クリーンアーキテクチャの依存方向を強制できる。

### インストールと実行

```bash
# aqua 経由 (推奨)
aqua install

# 手動インストール
go install github.com/fe3dback/go-arch-lint@latest

# 実行 (プロジェクトルートから)
go-arch-lint check ./...

# JSON 出力
go-arch-lint check --json-output ./...

# 依存関係のグラフ生成 (graphviz 要)
go-arch-lint graph ./... | dot -Tsvg > arch.svg
```

### 出力例

```
[ARCHITECTURE ERROR] package "github.com/example/app/infrastructure/db"
  imports "github.com/example/app/domain/usecase"
  which violates rule: infrastructure -> domain is not allowed
```

---

## しきい値の根拠

| しきい値 | 根拠 |
|---------|------|
| 認知的複雑度 ≤ 20 | Sonar 社の研究: 20 超えで理解コストが急上昇 |
| 循環複雑度 ≤ 20 | McCabe の原論文: 10 が理想、20 が実用上限 |
| 関数長 ≤ 100行 | Clean Code: 関数は画面に収まる長さが理想 |
| 保守性指数 ≥ 20 | Microsoft の定義: 0-9 は保守困難、10-19 は低品質 |

初期導入時は既存コードのベースライン値を記録し、段階的に目標値に近づけること。
最初から全違反を 0 にしようとするのは逆効果。
