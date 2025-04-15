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

CREATE TRIGGER set_is_user_category_trigger
    BEFORE INSERT OR UPDATE ON categories
                         FOR EACH ROW
                         EXECUTE FUNCTION set_is_user_category();

INSERT INTO categories (id, name, description, created_at, updated_at)
VALUES
    ('d59e13c0-13c6-4e87-9c85-9f5e36a74562', 'Food & Drink', 'Expenses related to food, drinks, and dining out.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('a816f5b0-38ed-4fae-b9b4-ea9b17b8155c', 'Transport', 'Expenses for transportation, including fuel, public transit, etc.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('e69f5b13-18a6-40a2-9baf-bb5a846a93b2', 'Housing', 'Costs related to housing, such as rent or mortgage payments.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('f8dfe097-e7c7-4670-824b-0e3d2f43b3f4', 'Utilities', 'Expenses related to electricity, water, gas, and other utilities.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('b2b0e5fe-13a9-4937-b45f-6c16a9fa18cf', 'Entertainment', 'Spending on movies, events, hobbies, and recreational activities.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('a274e679-d830-4eec-9f6c-69ab364ad50d', 'Healthcare', 'Expenses related to medical care, insurance, medications, etc.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('c2c3f0fb-7a45-4097-990f-b5d5c7b45990', 'Savings', 'Funds set aside for future financial goals and emergencies.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('7a2b3560-24c1-4ff1-b1da-cdb88d3c77ba', 'Investment', 'Money invested in stocks, bonds, real estate, etc.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('f5d9bb33-e4fe-4693-870b-b9c56f7f7022', 'Insurance', 'Premiums paid for life, health, home, and car insurance.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('d0e241a6-df76-45f6-9f78-ec6d4de3c9b3', 'Education', 'Expenditures related to tuition, books, courses, and learning materials.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_is_user_category_trigger ON categories;
DROP FUNCTION IF EXISTS set_is_user_category;
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
