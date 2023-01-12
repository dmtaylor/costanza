CREATE TABLE IF NOT EXISTS discord_usage_stats (
    id SERIAL PRIMARY KEY,
    guild_id NUMERIC NOT NULL,
    uid NUMERIC NOT NULL,
    report_month VARCHAR(7) NOT NULL,
    message_count INTEGER DEFAULT 1
);

CREATE INDEX report_guild_users ON discord_usage_stats(guild_id, uid);