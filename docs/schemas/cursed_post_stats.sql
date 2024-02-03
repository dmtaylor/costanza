CREATE TABLE IF NOT EXISTS discord_cursed_posts_stats (
                                                          id SERIAL PRIMARY KEY,
                                                          guild_id NUMERIC NOT NULL,
                                                          user_id NUMERIC NOT NULL,
                                                          report_month VARCHAR(7) NOT NULL,
                                                          message_count INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX cursed_word_post_users ON discord_cursed_posts_stats(guild_id, user_id);
CREATE INDEX cursed_word_guild_months ON discord_cursed_posts_stats(guild_id, report_month);
