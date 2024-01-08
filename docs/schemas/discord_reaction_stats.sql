CREATE TABLE IF NOT EXISTS discord_reaction_stats (
                                                      id SERIAL PRIMARY KEY,
                                                      guild_id NUMERIC NOT NULL,
                                                      user_id NUMERIC NOT NULL,
                                                      report_month VARCHAR(7) NOT NULL,
                                                      message_count INTEGER DEFAULT 1
);

CREATE INDEX reaction_guild_users ON discord_reaction_stats(guild_id, user_id);
CREATE INDEX reaction_guild_months ON discord_reaction_stats(guild_id, report_month);
