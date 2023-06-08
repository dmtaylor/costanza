package listen

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var interactionTimeout = time.Second * 2

var Commands = []*discordgo.ApplicationCommand{
	helpSlashCommand,
	licenseSlashCommand,
	weatherSlashCommand,
	rollSlashCommand,
	shadowrunRollSlashCommand,
	worldOfDarknessCommand,
	darkHeresyTestSlashCommand,
}
