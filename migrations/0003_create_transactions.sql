-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID           NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    budget_id        UUID           NOT NULL REFERENCES budgets (id) ON DELETE CASCADE,
    category_id      UUID           NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    amount           DECIMAL(15, 2) NOT NULL,
    transaction_type VARCHAR(20)    NOT NULL CHECK (transaction_type IN ('deposit', 'withdrawal')),
    note             VARCHAR(255),
    created_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions
-- +goose StatementEnd
