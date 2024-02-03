CREATE TABLE IF NOT EXISTS cursed_word_list (
    id SERIAL PRIMARY KEY,
    guild_id NUMERIC NOT NULL,
    word TEXT
);
CREATE INDEX cursed_word_guilds ON cursed_word_list(guild_id);
