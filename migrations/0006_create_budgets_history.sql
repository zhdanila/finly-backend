-- +goose Up
-- +goose StatementBegin
CREATE TABLE budgets_history
(
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id      UUID           NOT NULL REFERENCES budgets (id) ON DELETE CASCADE,
    transaction_id UUID           NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    balance        DECIMAL(15, 2) NOT NULL,
    created_at     TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE budgets_history
-- +goose StatementEnd
