package listen

import (
	"context"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/dmtaylor/costanza/config"
)

func (s *Server) logMessageActivity(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	// Only log stats if channel included in configs
	if _, found := config.GlobalConfig.Discord.ListenChannelSet[m.GuildID]; !found {
		return
	}

	if m.Type == discordgo.MessageTypeDefault || m.Type == discordgo.MessageTypeReply {
		guildId, err := strconv.ParseUint(m.GuildID, 10, 64)
		if err != nil {
			log.Printf("error logging activity: %s\n", err)
			return
		}
		userId, err := strconv.ParseUint(m.Author.ID, 10, 64)
		if err != nil {
			log.Printf("error logging activity: %s\n", err)
			return
		}
		err = s.app.Stats.LogActivity(context.Background(), guildId, userId, m.Timestamp.Format("2006-01"))
		if err != nil {
			log.Printf("error creating activity log: %s\n", err)
		}
	}
}
