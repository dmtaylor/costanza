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
