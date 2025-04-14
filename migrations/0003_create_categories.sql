-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID REFERENCES users (id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    description      VARCHAR(255) NOT NULL,
    is_user_category BOOLEAN          DEFAULT FALSE,
    created_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

-- Функція для оновлення is_user_category
CREATE OR REPLACE FUNCTION set_is_user_category()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.user_id IS NOT NULL THEN
        NEW.is_user_category := TRUE;
ELSE
        NEW.is_user_category := FALSE;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Тригер для автоматичного оновлення поля is_user_category
CREATE TRIGGER set_is_user_category_trigger
    BEFORE INSERT OR UPDATE ON categories
                         FOR EACH ROW
                         EXECUTE FUNCTION set_is_user_category();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_is_user_category_trigger ON categories;
DROP FUNCTION IF EXISTS set_is_user_category;
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
