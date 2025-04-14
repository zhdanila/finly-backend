package budget

type Budget struct {
	Name     string `json:"name"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type CreateBudgetRequest struct {
	UserID   string `header:"User-Id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type CreateBudgetResponse struct{}

type GetBudgetByIDRequest struct {
	UserID   string `header:"User-Id" validate:"required"`
	BudgetID string `param:"budget_id" validate:"required,uuid"`
}

type GetBudgetByIDResponse struct {
	Budget
}

type ListBudgetsByIDRequest struct {
	UserID string `header:"User-Id" validate:"required"`
}

type ListBudgetsByIDResponse struct {
	Budgets []Budget `json:"budgets"`
}

type DeleteBudgetRequest struct {
	UserID   string `header:"User-Id" validate:"required"`
	BudgetID string `param:"budget_id" validate:"required,uuid"`
}

type DeleteBudgetResponse struct{}
