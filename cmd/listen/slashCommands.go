package listen

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	helpSlashCommand,
	weatherSlashCommand,
	rollSlashCommand,
	shadowrunRollSlashCommand,
	worldOfDarknessCommand,
	darkHeresyTestSlashCommand,
}
