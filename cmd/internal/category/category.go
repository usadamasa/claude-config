package category

// Category はパターンの安全性分類を表す｡
type Category string

const (
	CategorySafe   Category = "safe"
	CategoryReview Category = "review"
)

// Result はパターンの分類結果を保持する｡
type Result struct {
	Category       Category `json:"category"`
	Reason         string   `json:"reason"`
	DenyBypassRisk bool     `json:"deny_bypass_risk,omitempty"`
}
