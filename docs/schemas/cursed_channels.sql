CREATE TABLE IF NOT EXISTS cursed_channels (
    id SERIAL PRIMARY KEY,
    guild_id NUMERIC NOT NULL,
    channel_id NUMERIC NOT NULL
);
CREATE INDEX guilds ON cursed_channels(guild_id);