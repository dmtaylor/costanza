package model

import "fmt"

type DiscordReactionStat struct {
	Id           uint
	GuildId      uint64
	UserId       uint64
	ReportMonth  string
	MessageCount int
}

type DiscordReactionScore struct {
	GuildId     uint64
	UserId      uint64
	ReportMonth string
	Score       int
}

func (d DiscordReactionScore) FormatResult(userString string) string {
	return fmt.Sprintf("%s with score %d", userString, d.Score)
}
