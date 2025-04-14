-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID,
    name             VARCHAR(255) NOT NULL,
    description      VARCHAR(255) NOT NULL,
    is_user_category BOOLEAN    DEFAULT FALSE,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE categories
-- +goose StatementEnd
