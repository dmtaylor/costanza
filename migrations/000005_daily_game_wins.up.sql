CREATE TABLE IF NOT EXiSTS daily_game_win_stats (
    id SERIAL PRIMARY KEY,
    guild_id NUMERIC NOT NULL,
    user_id NUMERIC NOT NULL,
    report_month VARCHAR(7) NOT NULL,
    play_count INTEGER DEFAULT 0,
    guess_count INTEGER DEFAULT 0,
    win_count INTEGER DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    max_streak INTEGER DEFAULT 0
);

CREATE INDEX win_stats_guild_users ON daily_game_win_stats(guild_id, user_id);
CREATE INDEX win_stats_guild_month ON daily_game_win_stats(guild_id, report_month);