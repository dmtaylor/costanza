package model

import "fmt"

type DailyGamePlay struct {
	GuildId uint64
	UserId  uint64
	Tries   uint
	Win     bool
}

type DailyGameWinStat struct {
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

func (d DailyGameWinStat) FormatWins() string {
	if d.PlayCount == 0 {
		return "Zero plays"
	} else {
		return fmt.Sprintf("%d wins (win rate %4.2f%%, average guesses %.2f, longest streak %d)",
			d.WinCount,
			float32(d.WinCount)/float32(d.PlayCount)*100,
			float32(d.GuessCount)/float32(d.PlayCount),
			d.MaxStreak)
	}
}
