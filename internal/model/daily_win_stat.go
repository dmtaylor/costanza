package model

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
