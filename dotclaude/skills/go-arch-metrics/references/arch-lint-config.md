# .go-arch-lint.yml テンプレート (依存方向ルール)

プロジェクトルートに `.go-arch-lint.yml` として配置する。

## 設定時の重要な注意事項

### 1. `canUse: []` (空配列) は使用不可

go-arch-lint ツール v3 以降で設定ファイル形式 version: 2 を使う場合、`canUse: []` と書くと以下のエラーが発生する:

```
should have ref in 'mayDependOn'/'canUse' or at least one flag of ['anyProjectDeps', 'anyVendorDeps']
```

「何にも依存しない」を表現するには:
- deps エントリ自体を省略する → 他コンポーネントへの依存を禁止 (Go 標準ライブラリは常に利用可)
- `anyVendorDeps: true` → 任意のサードパーティ依存を許可するフラグ (stdlib 限定ではない)
- `allow.depGuards` で `deps: []` → 他コンポーネントへの依存を禁止 (version: 2 形式)

### 2. 同一モジュール内パッケージは `vendors` ではなく `components` に定義する

`vendors` セクションは外部（サードパーティ）パッケージ専用。
同一モジュール内の生成コード (例: `browser/generated/api`) を `vendors` に定義すると
go-arch-lint が認識せず、依存違反が大量に報告される。

```yaml
# NG: 同一モジュール内を vendors に定義
vendors:
  $generated:
    in: "github.com/myorg/myapp/browser/generated/**"

# OK: 同一モジュール内は components に定義
components:
  $generated:
    in: "browser/generated/**"
```

### 3. YAML インライン記法 ({...}) は使用しない

yamllint のデフォルト設定はブレース内スペースを禁止するため、
インライン記法は lint エラーになる。必ずブロック記法を使う:

```yaml
# NG: インライン記法
$pkg:
  in: ["pkg/**"]

# OK: ブロック記法
$pkg:
  in: "pkg/**"
```

### 4. テストファイルは `excludeFiles` で除外する

`_test.go` を分析対象に含めると `testify/assert`, `testify/require` などの
テスト専用ライブラリで全コンポーネントに大量の違反が出る。
必ず最初から除外を設定する:

```yaml
workdir:
  root: .
  excludeFiles:
    - ".*_test\\.go$"
```

---

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
  excludeFiles:
    - ".*_test\\.go$"  # テストファイルは必ず除外 (testify 依存で大量違反になる)

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
  excludeFiles:
    - ".*_test\\.go$"  # テストファイルは必ず除外

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
| `should have ref in 'mayDependOn'/'canUse' or at least one flag` | `canUse: []` (空配列) を使っている | deps エントリを省略するか、`anyVendorDeps: true` でサードパーティ依存を許可 |
| 同一モジュール内パッケージで大量の依存違反 | `vendors` セクションに同一モジュールのパッケージを定義している | `components` セクションに移動する |
| testify などで大量の依存違反 | テストファイルが分析対象に含まれている | `workdir.excludeFiles` に `".*_test\\.go$"` を追加 |
