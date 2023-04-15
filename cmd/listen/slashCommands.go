package listen

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	weatherSlashCommand,
}
