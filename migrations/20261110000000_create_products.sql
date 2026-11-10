-- +goose Up
CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    price BIGINT NOT NULL CHECK (price > 0),
    stock BIGINT NOT NULL CHECK (stock >= 0),
    deleted_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS products;
