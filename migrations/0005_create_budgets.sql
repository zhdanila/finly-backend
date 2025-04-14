-- +goose Up
-- +goose StatementBegin
CREATE TABLE budgets
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount     int,
    currency   VARCHAR(3) NOT NULL,
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE budget
-- +goose StatementEnd
