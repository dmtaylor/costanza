package stats

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmtaylor/costanza/internal/model"
)

// TODO add more tests here. Need to test some error cases

func TestNew(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock pool", err)
	want := Stats{
		pool: pool,
	}
	got := New(pool)
	assert.Equal(t, want, got, "unexpected new stats service")
}

func TestStats_GetLeadersSuccess(t *testing.T) {
	var guildId uint64 = 5555
	reportMonth := "2023-06"
	expectedResults := []*model.DiscordUsageStat{
		{
			9,
			guildId,
			9876,
			reportMonth,
			87,
		},
		{
			54,
			guildId,
			6783,
			reportMonth,
			44,
		},
		{
			89,
			guildId,
			34214,
			reportMonth,
			32,
		},
		{
			12,
			guildId,
			98987,
			reportMonth,
			20,
		},
		{
			43,
			guildId,
			4545,
			reportMonth,
			10,
		},
	}

	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock pool")
	rows := mockDb.NewRows([]string{"id", "guild_id", "user_id", "report_month", "message_count"}).
		AddRow(uint(9), guildId, uint64(9876), reportMonth, 87).
		AddRow(uint(54), guildId, uint64(6783), reportMonth, 44).
		AddRow(uint(89), guildId, uint64(34214), reportMonth, 32).
		AddRow(uint(12), guildId, uint64(98987), reportMonth, 20).
		AddRow(uint(43), guildId, uint64(4545), reportMonth, 10)

	mockDb.ExpectQuery(`SELECT \*\sFROM discord_usage_stats\sWHERE guild_id = \$1 AND report_month = \$2\sORDER BY message_count DESC\sLIMIT 5`).
		WithArgs(guildId, reportMonth).
		WillReturnRows(rows)

	s := New(mockDb)
	got, err := s.GetLeaders(context.Background(), guildId, reportMonth)
	require.Nil(t, err, "getting leaders failed with error")
	assert.Equal(t, expectedResults, got, "result mismatch")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet mock db expectations")

}

func TestStats_GetLeadersEmptyResults(t *testing.T) {
	guildId := uint64(9999)
	reportMonth := "2023-02"
	var expectedResults []*model.DiscordUsageStat

	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to create mock db")
	rows := mockDb.NewRows([]string{"id", "guild_id", "user_id", "report_month", "message_count"})
	mockDb.ExpectQuery(`SELECT \*\sFROM discord_usage_stats\sWHERE guild_id = \$1 AND report_month = \$2\sORDER BY message_count DESC\sLIMIT 5`).
		WithArgs(guildId, reportMonth).
		WillReturnRows(rows)
	s := New(mockDb)
	got, err := s.GetLeaders(context.Background(), guildId, reportMonth)
	require.Nil(t, err, "getting empty leader list failed with errors")
	assert.Equal(t, expectedResults, got, "did not get empty list")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet mock db expectations")

}

func TestStats_LogActivityCreateNew(t *testing.T) {
	var guildId uint64 = 1111
	var userId uint64 = 2222
	reportMonth := "2023-01"

	emptyMockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock")
	defer emptyMockDb.Close()
	emptyMockDb.ExpectQuery(`
SELECT id
FROM discord_usage_stats
`).
		WithArgs(guildId, userId, reportMonth).
		WillReturnError(pgx.ErrNoRows)
	emptyMockDb.ExpectExec("INSERT INTO discord_usage_stats").
		WithArgs(guildId, userId, reportMonth).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	stats := New(emptyMockDb)
	err = stats.LogActivity(context.Background(), guildId, userId, reportMonth)
	assert.Nil(t, err, "got error creating new stat row")
	assert.Nil(t, emptyMockDb.ExpectationsWereMet(), "unmet mock db expectations")
}

func TestStats_LogActivityUpdateCount(t *testing.T) {
	var guildId uint64 = 3333
	var userId uint64 = 4444
	reportMonth := "2023-02"
	var mockRowId uint = 5

	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock")
	defer mockDb.Close()

	rows := mockDb.NewRows([]string{"id"}).AddRow(mockRowId)
	mockDb.ExpectQuery(`
SELECT id
FROM discord_usage_stats`).
		WithArgs(guildId, userId, reportMonth).
		WillReturnRows(rows)
	mockDb.ExpectExec(`
UPDATE discord_usage_stats
SET message_count = message_count \+ 1`).
		WithArgs(mockRowId).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	stats := New(mockDb)
	err = stats.LogActivity(context.Background(), guildId, userId, reportMonth)
	assert.Nil(t, err, "got error updating stat row")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet mock db expectations")
}

func TestStats_RemoveMonthActivity(t *testing.T) {
	month := "2023-01"

	db, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock db")
	db.ExpectExec("DELETE FROM discord_usage_stats WHERE report_month").WithArgs(month).WillReturnResult(pgxmock.NewResult("DELETE", 5))

	s := Stats{pool: db}
	err = s.RemoveMonthActivity(context.Background(), month)
	assert.Nil(t, err, "got error when deleting stats")
	assert.Nil(t, db.ExpectationsWereMet(), "unmet mock db expectations")
}

func TestStats_RemoveReactionLogForMonth(t *testing.T) {
	month := "2024-01"
	db, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock db")
	db.ExpectExec("DELETE FROM discord_reaction_stats WHERE report_month").WithArgs(month).WillReturnResult(pgxmock.NewResult("DELETE", 10))
	s := Stats{pool: db}
	err = s.RemoveReactionLogForMonth(context.Background(), month)
	assert.Nil(t, err, "got error when deleting data")
	assert.Nil(t, db.ExpectationsWereMet(), "unmet mock db expectations")
}

func TestStats_LogDailyGameActivityUpdate(t *testing.T) {
	var guildId uint64 = 5555
	var userId uint64 = 6666
	reportMonth := "2023-10"

	gamePlay := model.DailyGamePlay{
		GuildId: guildId,
		UserId:  userId,
		Tries:   2,
		Win:     true,
	}
	dbModel := model.DailyGameWinStat{
		Id:            8,
		GuildId:       guildId,
		UserId:        userId,
		ReportMonth:   reportMonth,
		PlayCount:     5,
		GuessCount:    12,
		WinCount:      3,
		CurrentStreak: 2,
		MaxStreak:     2,
	}

	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock")
	defer mockDb.Close()
	rows := mockDb.NewRows([]string{"id", "guild_id", "user_id", "report_month", "play_count", "guess_count", "win_count", "current_streak", "max_streak"}).
		AddRow(dbModel.Id, dbModel.GuildId, dbModel.UserId, dbModel.ReportMonth, dbModel.PlayCount, dbModel.GuessCount, dbModel.WinCount, dbModel.CurrentStreak, dbModel.MaxStreak)
	mockDb.ExpectQuery(`
SELECT \*
FROM daily_game_win_stats`).
		WithArgs(guildId, userId, reportMonth).WillReturnRows(rows)
	mockDb.ExpectExec(`
UPDATE daily_game_win_stats
SET play_count = play_count \+ 1`).
		WithArgs(dbModel.GuessCount+int(gamePlay.Tries), dbModel.WinCount+1, dbModel.CurrentStreak+1, dbModel.MaxStreak+1, dbModel.Id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	stats := New(mockDb)
	err = stats.LogDailyGameActivity(context.Background(), gamePlay, reportMonth)
	assert.Nil(t, err, "got error when updating stats")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet db expectations")
}

func TestStats_LogDailyGameActivityNew(t *testing.T) {
	var guildId uint64 = 7777
	var userId uint64 = 7778
	reportMonth := "2023-10"
	gamePlay := model.DailyGamePlay{
		GuildId: guildId,
		UserId:  userId,
		Tries:   1,
		Win:     true,
	}
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock")
	defer mockDb.Close()
	mockDb.ExpectQuery(`
SELECT \*
FROM daily_game_win_stats`).
		WithArgs(guildId, userId, reportMonth).WillReturnError(pgx.ErrNoRows)
	mockDb.ExpectExec(`
INSERT INTO daily_game_win_stats\(guild_id, user_id, report_month, play_count, guess_count, win_count, current_streak, max_streak\)`).
		WithArgs(guildId, userId, reportMonth, uint(1), 1, 1, 1).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	stats := New(mockDb)
	err = stats.LogDailyGameActivity(context.Background(), gamePlay, reportMonth)
	assert.Nil(t, err, "got error when updating stats")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet db expectations")
}

func TestStats_LogReactionNew(t *testing.T) {
	var guildId uint64 = 1234
	var userId uint64 = 5678
	reportMonth := "2024-01"
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock")
	defer mockDb.Close()
	mockDb.ExpectQuery(`
SELECT id
FROM discord_reaction_stats
WHERE guild_id`).WithArgs(guildId, userId, reportMonth).WillReturnError(pgx.ErrNoRows)
	mockDb.ExpectExec(`INSERT INTO discord_reaction_stats\(guild_id, user_id, report_month\) VALUES `).
		WithArgs(guildId, userId, reportMonth).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	stats := New(mockDb)
	err = stats.LogReaction(context.Background(), guildId, userId, reportMonth)
	assert.Nil(t, err, "got err when adding stat")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet db expectations")
}

func TestStats_LogReactionUpdate(t *testing.T) {
	var guildId uint64 = 4321
	var userId uint64 = 8765
	reportMonth := "2023-01"
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock")
	mockRowId := uint(9921)
	rows := mockDb.NewRows([]string{"id"}).AddRow(mockRowId)
	mockDb.ExpectQuery(`
SELECT id
FROM discord_reaction_stats
WHERE guild_id`).WithArgs(guildId, userId, reportMonth).WillReturnRows(rows)
	mockDb.ExpectExec(`
UPDATE discord_reaction_stats
SET message_count = message_count \+ 1
WHERE id =`).WithArgs(mockRowId).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	stats := New(mockDb)
	err = stats.LogReaction(context.Background(), guildId, userId, reportMonth)
	assert.Nil(t, err, "got err when adding stat")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet db expectations")
}

func TestStats_GetDailyGameLeadersSuccess(t *testing.T) {
	var guildId uint64 = 8888
	reportMonth := "2023-10"
	expectedResults := []*model.DailyGameWinStat{
		{
			Id:            192,
			GuildId:       guildId,
			UserId:        9888,
			ReportMonth:   reportMonth,
			PlayCount:     31,
			GuessCount:    35,
			WinCount:      28,
			CurrentStreak: 5,
			MaxStreak:     11,
		},
		{
			Id:            870,
			GuildId:       guildId,
			UserId:        664,
			ReportMonth:   reportMonth,
			PlayCount:     28,
			GuessCount:    38,
			WinCount:      27,
			CurrentStreak: 10,
			MaxStreak:     10,
		},
		{
			Id:            58,
			GuildId:       guildId,
			UserId:        9034,
			ReportMonth:   reportMonth,
			PlayCount:     28,
			GuessCount:    38,
			WinCount:      26,
			CurrentStreak: 11,
			MaxStreak:     12,
		},
	}
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock pool")
	rows := mockDb.NewRows([]string{"id", "guild_id", "user_id", "report_month", "play_count", "guess_count", "win_count", "current_streak", "max_streak"}).
		AddRow(uint(192), guildId, uint64(9888), reportMonth, 31, 35, 28, 5, 11).
		AddRow(uint(870), guildId, uint64(664), reportMonth, 28, 38, 27, 10, 10).
		AddRow(uint(58), guildId, uint64(9034), reportMonth, 28, 38, 26, 11, 12)
	mockDb.ExpectQuery(`SELECT \*\sFROM daily_game_win_stats.*ORDER BY win_count DESC\sLIMIT 5`).
		WithArgs(guildId, reportMonth).
		WillReturnRows(rows)

	s := New(mockDb)
	got, err := s.GetDailyGameLeaders(context.Background(), guildId, reportMonth)
	require.Nil(t, err, "getting game leaders failed with error")
	assert.Equal(t, expectedResults, got, "result mismatch")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet mock db expectations")

}

func TestStats_RemoveDailyGameLeadersForMonth(t *testing.T) {
	month := "2023-10"

	db, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock db")
	db.ExpectExec(`DELETE FROM daily_game_win_stats WHERE report_month`).WithArgs(month).WillReturnResult(pgxmock.NewResult("DELETE", 5))
	s := Stats{db}
	err = s.RemoveDailyGameLeadersForMonth(context.Background(), month)
	assert.Nil(t, err, "got error when deleting stats")
	assert.Nil(t, db.ExpectationsWereMet(), "unmet mock db expectations")
}
