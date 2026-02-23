# 是正ガイド (優先順位付き)

## 優先順位の基準

| 優先度 | 条件 | 理由 |
|--------|------|------|
| **High** | 依存方向の逆転 / 循環参照 | アーキテクチャ崩壊の根本原因。放置すると全体に波及 |
| **Medium** | 認知的複雑度 > 30 / 関数長 > 200行 | テストが実質書けない状態 |
| **Low** | 保守性指数 15-19 / ネスト深さ 6-8 | 徐々に劣化するが即座の破綻はない |

---

## カテゴリ別 是正手順

### モジュール性違反の是正

#### 依存方向の逆転 (go-arch-lint 違反)

**問題**: `domain` パッケージが `infra` パッケージを直接 import している

```go
// ❌ Bad: domain が infra に依存
package domain

import "github.com/example/app/internal/infra/db"

type OrderRepository struct {
    db *db.Client  // domain が infra に依存 → 逆転!
}
```

**是正**: インタフェースを domain に定義し、infra で実装する (DIP)

```go
// ✅ Good: domain はインタフェースのみ定義
package domain

// Repository インタフェースを domain に定義
type OrderRepository interface {
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    Save(ctx context.Context, order *Order) error
}

// ✅ Good: infra でインタフェースを実装
package infra

import "github.com/example/app/internal/domain"

type OrderRepositoryImpl struct {
    db *sql.DB
}

func (r *OrderRepositoryImpl) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
    // DB アクセス実装
}
```

#### 循環参照の解消

**問題**: パッケージ A が B を import し、B が A を import している

**是正手順**:
1. `go list -f '{{.ImportPath}} -> {{.Imports}}' ./...` で循環を特定
2. 共有される型・ロジックを第3の共通パッケージ (`pkg/` や `internal/shared/`) に移動
3. または一方を interface に抽象化

---

### テスト可能性違反の是正

#### 認知的複雑度が高い関数の分割

**問題**: 認知的複雑度 > 20

```go
// ❌ Bad: 複雑度 28 (if/else/switch/for の組み合わせ)
func ProcessOrder(order Order) error {
    if order.Status == "pending" {
        if order.Amount > 1000 {
            for _, item := range order.Items {
                if item.Stock > 0 {
                    // ...
                } else {
                    // ...
                }
            }
        } else {
            // ...
        }
    } else if order.Status == "processing" {
        // ...
    }
    return nil
}
```

**是正**: 早期リターンと関数分割

```go
// ✅ Good: 各責務を独立した関数に分割
func ProcessOrder(order Order) error {
    if err := validateOrder(order); err != nil {
        return err
    }
    return processOrderByStatus(order)
}

func validateOrder(order Order) error {
    if order.Status == "" {
        return ErrInvalidStatus
    }
    return nil
}

func processOrderByStatus(order Order) error {
    switch order.Status {
    case "pending":
        return processPendingOrder(order)
    case "processing":
        return processRunningOrder(order)
    default:
        return ErrUnknownStatus
    }
}
```

#### 関数が長すぎる場合

**是正パターン**:

1. **ステップの抽出**: 処理の各フェーズを独立した関数に分割
2. **ヘルパーの抽出**: 繰り返しロジックをまとめる
3. **構造体メソッド化**: 状態を持つロジックはレシーバメソッドにする

```go
// ❌ Bad: 150行の巨大関数
func HandleRequest(req *http.Request, resp http.ResponseWriter) {
    // 認証チェック (30行)
    // バリデーション (40行)
    // ビジネスロジック (50行)
    // レスポンス組み立て (30行)
}

// ✅ Good: 責務を分割
func HandleRequest(req *http.Request, resp http.ResponseWriter) {
    user, err := h.authenticate(req)
    if err != nil {
        h.writeError(resp, err)
        return
    }
    input, err := h.validate(req)
    if err != nil {
        h.writeError(resp, err)
        return
    }
    result, err := h.usecase.Execute(req.Context(), user, input)
    if err != nil {
        h.writeError(resp, err)
        return
    }
    h.writeSuccess(resp, result)
}
```

#### ネストが深すぎる場合

**是正**: ガード節 (早期リターン) パターン

```go
// ❌ Bad: ネスト深さ 6
func processItem(item Item) error {
    if item != nil {
        if item.IsValid() {
            if item.Stock > 0 {
                if item.Price > 0 {
                    // 実際の処理
                }
            }
        }
    }
    return nil
}

// ✅ Good: ガード節で早期リターン
func processItem(item Item) error {
    if item == nil {
        return ErrNilItem
    }
    if !item.IsValid() {
        return ErrInvalidItem
    }
    if item.Stock <= 0 {
        return ErrOutOfStock
    }
    if item.Price <= 0 {
        return ErrInvalidPrice
    }
    // 実際の処理
    return nil
}
```

---

### 保守性違反の是正

#### 保守性指数が低い (maintidx < 20)

保守性指数は複合指標なので、以下を組み合わせて改善する:

1. **コメントを追加** (Halstead の可読性コンポーネントを改善)
2. **関数を分割** (行数コンポーネントを改善)
3. **複雑度を下げる** (循環複雑度コンポーネントを改善)

#### 到達不能コード (deadcode)

```bash
# deadcode で検出
deadcode ./...

# 出力例
cmd/server/main.go:45: unreachable function "oldHandler"
internal/usecase/order.go:78: unreachable function "legacyCalculate"
```

**是正**: 単純に削除する。「将来使うかも」は Git の履歴に残っているため不要。

---

## 段階的改善のロードマップ

```
Sprint 1: ベースライン測定と High 優先度の解消
  → 依存方向の逆転をすべて解消
  → 循環複雑度 > 50 の関数を優先的にリファクタリング

Sprint 2: テスト可能性の改善
  → 認知的複雑度 > 30 の関数を分割
  → テストカバレッジを測定し、複雑な関数のテストを追加

Sprint 3: しきい値の段階的引き下げ
  → golangci-lint の設定を 30 → 25 → 20 と段階的に厳しくする
  → CI でのチェックを有効化

Sprint 4: 保守性の安定化
  → maintidx 違反の解消
  → deadcode の削除
```
