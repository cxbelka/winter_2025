CREATE SCHEMA IF NOT EXISTS merch_shop;

CREATE TABLE IF NOT EXISTS merch_shop.auth (
    login text PRIMARY KEY,
    password bytea NOT NULL,
    balance integer CONSTRAINT positive_balance CHECK (balance > 0),
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp DEFAULT NULL
);

----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS merch_shop.items (
    name text PRIMARY KEY,
    price integer,
    deleted_at timestamp DEFAULT NULL
);

----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS merch_shop.transfers (
    dt timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    src text  REFERENCES merch_shop.auth (login),
    dst text  REFERENCES merch_shop.auth (login),
    sum integer CONSTRAINT positive_sum CHECK (sum > 0)
);
CREATE INDEX IF NOT EXISTS idx_merch_shop_transfers_from
    ON merch_shop.transfers USING hash (src);

CREATE INDEX IF NOT EXISTS idx_merch_shop_transfers_to
    ON merch_shop.transfers USING hash (dst);

CREATE INDEX IF NOT EXISTS idx_merch_shop_transfers_dt
    ON merch_shop.transfers USING btree (dt); -- for ordering

----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS merch_shop.purchases (
    dt timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name text  REFERENCES merch_shop.auth (login),
    item text  REFERENCES merch_shop.items (name),
    sum integer -- если цена на товары может поменяться
);
CREATE INDEX IF NOT EXISTS idx_merch_shop_purchases_user
    ON merch_shop.purchases USING hash (name);

CREATE INDEX IF NOT EXISTS iidx_merch_shop_purchases_dt
    ON merch_shop.purchases USING btree (dt); -- for ordering

----------------------------------------------------------------------------

INSERT INTO merch_shop.items (name, price) VALUES 
    ('t-shirt', 80),
    ('cup', 20),
    ('book',50),
    ('pen',	10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500)
    ON CONFLICT (name) DO NOTHING;
