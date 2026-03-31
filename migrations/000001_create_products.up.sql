CREATE TABLE IF NOT EXISTS products (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    price       BIGINT      NOT NULL CHECK (price >= 0),
    stock       INTEGER     NOT NULL CHECK (stock >= 0),
    status      TEXT        NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
