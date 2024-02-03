package model

type CursedPostStat struct {
	Id           uint
	GuildId      uint64
	UserId       uint64
	ReportMonth  string
	MessageCount int
}
