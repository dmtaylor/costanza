package model

type CursedChannelPost struct {
	Id           uint
	GuildId      uint64
	UserId       uint64
	ReportMonth  string
	MessageCount int
}
