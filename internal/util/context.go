package util

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

const MessageEvtType = "message"
const InteractionEvtType = "interaction"
const ReactionEvtType = "reaction"

func CheckCtxTimeout(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func ContextFromDiscordMessageCreate(parent context.Context, m *discordgo.MessageCreate) context.Context {
	ctx := context.WithValue(parent, "guildId", m.GuildID)
	ctx = context.WithValue(ctx, "messageId", m.ID)
	ctx = context.WithValue(ctx, "messageType", m.Type)
	ctx = context.WithValue(ctx, "channelId", m.ChannelID)
	ctx = context.WithValue(ctx, "user", m.Author.ID)
	ctx = context.WithValue(ctx, "type", MessageEvtType)
	return ctx
}

func ContextFromDiscordInteractionCreate(parent context.Context, i *discordgo.InteractionCreate, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	ctx = context.WithValue(ctx, "interactionId", i.ID)
	ctx = context.WithValue(ctx, "guildId", i.GuildID)
	if i.User != nil {
		ctx = context.WithValue(ctx, "user", i.User.String())
	}
	if i.Member != nil {
		ctx = context.WithValue(ctx, "user", i.Member.User.String())
	}
	ctx = context.WithValue(ctx, "channelId", i.ChannelID)
	if i.Type == discordgo.InteractionApplicationCommand {
		ctx = context.WithValue(ctx, "commandName", i.ApplicationCommandData().Name)
	}
	ctx = context.WithValue(ctx, "type", InteractionEvtType)

	return ctx, cancel
}

func ContextFromListenConfig(parent context.Context, guildId, channelId string) context.Context {
	ctx := context.WithValue(parent, "guildId", guildId)
	ctx = context.WithValue(ctx, "reportChannelId", channelId)
	return ctx
}

func ContextFromDiscordReactionAdd(parent context.Context, r *discordgo.MessageReactionAdd) context.Context {
	ctx := context.WithValue(parent, "guildId", r.GuildID)
	ctx = context.WithValue(ctx, "messageId", r.MessageID)
	ctx = context.WithValue(ctx, "user", r.UserID)
	ctx = context.WithValue(ctx, "type", ReactionEvtType)

	return ctx
}
