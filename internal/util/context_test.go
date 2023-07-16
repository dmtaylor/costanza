package util

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

var eventTime = time.Date(2023, time.July, 15, 9, 0, 0, 0, time.UTC)

func TestContextFromDiscordMessageCreate(t *testing.T) {
	// unused fields omitted for brevity. Real production payloads will contain more information
	m := &discordgo.MessageCreate{&discordgo.Message{
		ID:        "1234567",
		ChannelID: "8901234",
		GuildID:   "2345678",
		Content:   "Men, if they cannot attain what is necessary, tire themselves with that which is useless",
		Timestamp: eventTime,
		Author: &discordgo.User{
			ID:       "8675",
			Email:    "goethe@example.com",
			Username: "goatman",
			Bot:      false,
		},
		Type: discordgo.MessageTypeDefault,
	}}
	testCtx := ContextFromDiscordMessageCreate(context.Background(), m)
	if guildId, ok := testCtx.Value("guildId").(string); !ok || guildId != "2345678" {
		t.Errorf("expected guildId context value \"2345678\", got %s", guildId)
		return
	}
	if channelId, ok := testCtx.Value("channelId").(string); !ok || channelId != "8901234" {
		t.Errorf("expected channelId context value \"8901234\", got %s", channelId)
		return
	}
	if messageId, ok := testCtx.Value("messageId").(string); !ok || messageId != "1234567" {
		t.Errorf("expected messageId context value \"1234567\", got %s", messageId)
		return
	}
	if messageType, ok := testCtx.Value("messageType").(discordgo.MessageType); !ok || messageType != discordgo.MessageTypeDefault {
		t.Errorf("expected messageType context value %d, got %d", discordgo.MessageTypeDefault, messageType)
		return
	}
	if userId, ok := testCtx.Value("user").(string); !ok || userId != "8675" {
		t.Errorf("expected user id \"8675\", got %s", userId)
		return
	}
}

func TestContextFromDiscordInteractionCreate(t *testing.T) {
	//i := discordgo.InteractionCreate
}
