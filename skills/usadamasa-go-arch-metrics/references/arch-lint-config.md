# .go-arch-lint.yml テンプレート (依存方向ルール)

プロジェクトルートに `.go-arch-lint.yml` として配置する。

## レイヤードアーキテクチャ向け (基本テンプレート)

```yaml
# .go-arch-lint.yml
# パッケージ依存方向ルール定義
# 依存方向の原則: 外側のレイヤーは内側を参照できる。逆は禁止。
#
# 参考構造:
#   cmd/         → エントリポイント (最外層)
#   internal/handler/   → HTTP/gRPC ハンドラ (presentation)
#   internal/usecase/   → ビジネスロジック (application)
#   internal/domain/    → ドメインモデル (domain)
#   internal/infra/     → DB/外部API実装 (infrastructure)
#   pkg/         → 共有ユーティリティ (横断的関心事)

version: 2

workdir:
  root: .  # go.mod が存在するディレクトリ

allow:
  depGuards:
    # cmd は全レイヤーに依存できる (DI コンテナとして機能)
    - pkg: "**"
      deps:
        - "**"
      files:
        - "$cmd/**"

    # handler は usecase にのみ依存できる (domain, pkg も可)
    - pkg: "$handler"
      deps:
        - "$usecase"
        - "$domain"
        - "$pkg"

    # usecase は domain にのみ依存できる (pkg も可)
    - pkg: "$usecase"
      deps:
        - "$domain"
        - "$pkg"

    # domain は他のどのレイヤーにも依存できない (pkg のみ可)
    - pkg: "$domain"
      deps:
        - "$pkg"

    # infra は domain と pkg に依存できる (usecase のインタフェースを実装)
    - pkg: "$infra"
      deps:
        - "$domain"
        - "$pkg"

    # pkg (共通ユーティリティ) は外部パッケージのみ依存可
    - pkg: "$pkg"
      deps: []

components:
  $cmd:
    in: "cmd/**"
  $handler:
    in: "internal/handler/**"
  $usecase:
    in: "internal/usecase/**"
  $domain:
    in: "internal/domain/**"
  $infra:
    in: "internal/infra/**"
  $pkg:
    in: "pkg/**"
```

## クリーンアーキテクチャ向けテンプレート

```yaml
# .go-arch-lint.yml (クリーンアーキテクチャ版)
version: 2

workdir:
  root: .

allow:
  depGuards:
    # Frameworks & Drivers: すべてに依存可
    - pkg: "$frameworks"
      deps:
        - "$interface_adapters"
        - "$application"
        - "$entities"
        - "$external"

    # Interface Adapters: application と entities に依存可
    - pkg: "$interface_adapters"
      deps:
        - "$application"
        - "$entities"
        - "$external"

    # Application Business Rules: entities のみ
    - pkg: "$application"
      deps:
        - "$entities"

    # Enterprise Business Rules: 依存なし (最内層)
    - pkg: "$entities"
      deps: []

    # 外部パッケージ: 制限なし
    - pkg: "$external"
      deps:
        - "**"

components:
  $frameworks:
    in: "cmd/**"
  $interface_adapters:
    in: "internal/adapter/**"
  $application:
    in: "internal/usecase/**"
  $entities:
    in: "internal/domain/**"
  $external:
    in: "pkg/**"
```

## go-arch-lint の実行

```bash
# 違反チェック
go-arch-lint check ./...

# JSON 形式で出力 (CI 向け)
go-arch-lint check --json-output ./...

# 依存グラフの可視化 (graphviz が必要)
go-arch-lint graph ./... | dot -Tsvg > arch-dependency.svg
```

## よくあるエラーと対処

| エラーメッセージ | 原因 | 対処 |
|----------------|------|------|
| `package not found in components` | コンポーネント定義のパターンが合っていない | `in:` のパターンを `go list ./...` で確認 |
| `circular dependency detected` | 相互参照が存在する | どちらかを抽象 (interface) に分離 |
| `allow rule missing for package` | ルールが未定義のパッケージが存在 | 新しい component を追加するか既存ルールに含める |
