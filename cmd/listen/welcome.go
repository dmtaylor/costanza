package listen

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

const welcomeMessageFmt = `Welcome to the party %s!`

func welcomeMessage(sess *discordgo.Session, j *discordgo.GuildMemberAdd) {
	if j.User.ID == sess.State.User.ID { // Don't welcome yourself
		return
	}
	if j.User.Bot { // Don't welcome robots
		return
	}
	ctx := context.WithValue(context.Background(), "memberId", j.User.ID)
	ctx = context.WithValue(ctx, "guildId", j.GuildID)

	channels, err := sess.GuildChannels(j.GuildID)
	if err != nil {
		log.Printf("error getting channel list: %s\n", err)
		return
	}
	if len(channels) < 1 {
		log.Printf("no channels in guild pulled\n")
		return
	}
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildText && channel.Position == 0 {
			_, err = sess.ChannelMessageSend(channel.ID, fmt.Sprintf(welcomeMessageFmt, j.User.Mention()))
			if err != nil {
				log.Printf("failed to send message to channel %s: %s\n", channel.ID, err)
			}
		}
	}
}
