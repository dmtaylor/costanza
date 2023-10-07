package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDailyGameWinStat_FormatWins(t *testing.T) {
	type fields struct {
		Id            uint
		GuildId       uint64
		UserId        uint64
		ReportMonth   string
		PlayCount     int
		GuessCount    int
		WinCount      int
		CurrentStreak int
		MaxStreak     int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"baseline",
			fields{
				Id:            42,
				GuildId:       43,
				ReportMonth:   "2023-10",
				PlayCount:     30,
				GuessCount:    105,
				WinCount:      20,
				CurrentStreak: 9,
				MaxStreak:     10,
			},
			"20 wins (win rate 66.67%, average guesses 3.50, longest streak 10)",
		},
		{
			"zero_plays_safety_catch",
			fields{
				Id:            44,
				GuildId:       45,
				ReportMonth:   "2023-10",
				PlayCount:     0,
				GuessCount:    0,
				WinCount:      0,
				CurrentStreak: 0,
				MaxStreak:     0,
			},
			"Zero plays",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DailyGameWinStat{
				Id:            tt.fields.Id,
				GuildId:       tt.fields.GuildId,
				UserId:        tt.fields.UserId,
				ReportMonth:   tt.fields.ReportMonth,
				PlayCount:     tt.fields.PlayCount,
				GuessCount:    tt.fields.GuessCount,
				WinCount:      tt.fields.WinCount,
				CurrentStreak: tt.fields.CurrentStreak,
				MaxStreak:     tt.fields.MaxStreak,
			}
			got := d.FormatWins()
			assert.Equal(t, tt.want, got)
		})
	}
}
