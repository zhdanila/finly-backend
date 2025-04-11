-- +goose Up
-- +goose StatementBegin
CREATE TABLE standard_categories
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE standard_categories
-- +goose StatementEnd
