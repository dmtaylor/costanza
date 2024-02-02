CREATE TABLE IF NOT EXISTS discord_cursed_channel_stats (
    id SERIAL PRIMARY KEY,
    guild_id NUMERIC NOT NULL,
    user_id NUMERIC NOT NULL,
    report_month VARCHAR(7) NOT NULL,
    message_count INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX cursed_channel_posts_users ON discord_reaction_stats(guild_id, user_id);
CREATE INDEX cursed_channel_posts_guild_months ON discord_cursed_channel_stats(guild_id, report_month);