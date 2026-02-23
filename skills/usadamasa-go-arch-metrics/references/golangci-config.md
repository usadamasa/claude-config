# .golangci.yml テンプレート (アーキテクチャメトリクス向け)

プロジェクトルートに `.golangci.yml` として配置する。

> **注意: golangci-lint v2 と v1 の設定ファイルは非互換**
>
> - v2 では `linters-settings:` が `linters.settings:` (linters ブロック内) に移動
> - v2 では `issues.exclude-rules` が `linters.exclusions.rules` に移動
> - v1 形式のキーは v2 では **警告なしに無視される** ため、動作しているように見えて効いていない点に注意
> - `version: "2"` を指定している場合は必ず v2 形式を使う

```yaml
# .golangci.yml
# アーキテクチャメトリクス向け golangci-lint v2 設定
# 参考: ソフトウェアアーキテクチャメトリクス (ISBN: 9784814400607)
version: "2"

linters:
  # デフォルト有効 linter を維持しつつ、メトリクス系を追加
  default: standard
  enable:
    # --- テスト可能性メトリクス ---
    - gocognit   # 認知的複雑度: コードの理解しやすさを数値化
    - gocyclo    # 循環複雑度: 分岐の多さを数値化 (McCabe)
    - cyclop     # 循環複雑度: パッケージ全体も対象
    - funlen     # 関数の長さ: 行数と文数の上限
    - nestif     # ネストの深さ: if/for の入れ子レベル
    # --- 保守性メトリクス ---
    - maintidx   # 保守性指数: 複雑度・行数・Halstead の複合指標
    - deadcode   # 到達不能コード: 実行されないコードの検出
    # --- 高度な静的解析 ---
    - staticcheck  # standard に含まれるが明示的に指定
    # --- 依存管理 ---
    - depguard   # 禁止パッケージの import を防ぐ

  # v2 では linters.settings (linters ブロック内) に記述する
  settings:
    # 認知的複雑度: 20 超えは要リファクタリング
    gocognit:
      min-complexity: 20

    # 循環複雑度: 20 超えは要リファクタリング
    gocyclo:
      min-complexity: 20

    # 循環複雑度 (cyclop)
    cyclop:
      max-complexity: 20

    # 関数の長さ
    funlen:
      lines: 100      # 行数上限
      statements: 60  # 文数上限 (コメント・空行除く)

    # ネストの深さ
    nestif:
      min-complexity: 5

    # 保守性指数: 20 未満は保守困難
    maintidx:
      under: 20

    # 禁止パッケージ設定 (プロジェクトに合わせてカスタマイズ)
    depguard:
      rules:
        # 例: sync/atomic より atomic パッケージを使う場合
        # no-sync-atomic:
        #   deny:
        #     - pkg: "sync/atomic"
        #       desc: "Use go.uber.org/atomic instead"

  # v2 では linters.exclusions (linters ブロック内) に記述する
  exclusions:
    # 生成コードは全チェックを除外
    paths:
      - ".*\\.pb\\.go$"
      - ".*_gen\\.go$"
    # テストファイルは関数長・複雑度の制限を緩める
    rules:
      - path: "_test\\.go"
        linters:
          - cyclop
          - funlen
          - gocognit
          - gocyclo
          - maintidx  # テーブル駆動テストは保守性指数が低くなりがち
```

## カスタマイズガイド

### しきい値の調整

初期導入時は既存コードの最大値より少し上にしきい値を設定し、段階的に絞る:

```yaml
# 例: 既存コードの最大認知的複雑度が 45 の場合、まず 50 から始める
linters:
  settings:
    gocognit:
      min-complexity: 50
```

### 段階的な導入 (new-from-rev オプション) - v2 では使用不可

> **注意**: v2 では `issues.new` / `issues.new-from-rev` は廃止された。
> 代わりに GitHub Actions の `golangci/golangci-lint-action` が差分チェックをサポートしている。

```yaml
# v2 での段階的導入は、しきい値を現在の最大値以上に設定して導入し、
# 段階的に引き下げる方式が推奨される。
```

### depguard でレイヤー間の直接依存を禁止する例

```yaml
linters:
  settings:
    depguard:
      rules:
        no-infra-in-domain:
          files:
            - "**/domain/**"
          deny:
            - pkg: "database/sql"
              desc: "Domain layer must not depend on infrastructure directly"
            - pkg: "github.com/your-org/app/infrastructure"
              desc: "Domain must not import infrastructure packages"
```
