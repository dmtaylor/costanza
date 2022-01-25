CREATE TABLE IF NOT EXISTS initiative_orders (
    id BIGSERIAL PRIMARY KEY,
    owner_snowflake NUMERIC NOT NULL, -- Ids for discord items are 64 bits unsigned, use numeric to fit
    size INTEGER NOT NULL DEFAULT 0,
    current_in_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX ON initiative_orders (owner_snowflake);

CREATE TABLE IF NOT EXISTS initiative_users (
    id BIGSERIAL PRIMARY KEY,
    initiative_id BIGINT,
    user_snowflake NUMERIC NOT NULL,
    user_order INTEGER,
    CONSTRAINT fk_owner
        FOREIGN KEY (initiative_id)
        REFERENCES initiative_orders(id)
        ON DELETE CASCADE
);

CREATE INDEX ON initiative_users (initiative_id);