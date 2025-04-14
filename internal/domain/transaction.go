package domain

type Transaction struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
	Type       string `json:"type"`
	Note       string `json:"note"`
	CreatedAt  string `json:"created_at"`
}
