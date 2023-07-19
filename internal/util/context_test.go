package util

import (
	"context"
	"errors"
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
	}
	if channelId, ok := testCtx.Value("channelId").(string); !ok || channelId != "8901234" {
		t.Errorf("expected channelId context value \"8901234\", got %s", channelId)
	}
	if messageId, ok := testCtx.Value("messageId").(string); !ok || messageId != "1234567" {
		t.Errorf("expected messageId context value \"1234567\", got %s", messageId)
	}
	if messageType, ok := testCtx.Value("messageType").(discordgo.MessageType); !ok || messageType != discordgo.MessageTypeDefault {
		t.Errorf("expected messageType context value %d, got %d", discordgo.MessageTypeDefault, messageType)
	}
	if userId, ok := testCtx.Value("user").(string); !ok || userId != "8675" {
		t.Errorf("expected user id \"8675\", got %s", userId)
	}
}

func TestContextFromDiscordInteractionCreate(t *testing.T) {
	// Abridged data for InteractionCreate event. In production this will be more complete
	i := &discordgo.InteractionCreate{&discordgo.Interaction{
		ID:      "42",
		GuildID: "8723",
		Member: &discordgo.Member{
			GuildID: "8723",
			User: &discordgo.User{
				ID:            "9812",
				Username:      "vimes",
				Email:         "vimes@example.com",
				Discriminator: "01",
			},
		},
		ChannelID: "4567",
		Type:      discordgo.InteractionApplicationCommand,
		Data:      discordgo.ApplicationCommandInteractionData{Name: "testCommand", ID: "12345"},
	}}
	ctx, cancel := ContextFromDiscordInteractionCreate(context.Background(), i, time.Second*2)
	_, deadlineSet := ctx.Deadline()
	if !deadlineSet {
		t.Errorf("no deadline set in context: %v", ctx)
	}
	if id, ok := ctx.Value("interactionId").(string); !ok || id != "42" {
		t.Errorf("expected interaction id \"42\", got %v", id)
	}
	if guildId, ok := ctx.Value("guildId").(string); !ok || guildId != "8723" {
		t.Errorf("expected guild id \"8723\", got %v", guildId)
	}
	if user, ok := ctx.Value("user").(string); !ok || user != "vimes#01" {
		t.Errorf("expected user id \"9812\", got %v", user)
	}
	if channelId, ok := ctx.Value("channelId").(string); !ok || channelId != "4567" {
		t.Errorf("expected channel id \"4567\", got %v", channelId)
	}
	if name, ok := ctx.Value("commandName").(string); !ok || name != "testCommand" {
		t.Errorf("expected command name \"testCommand\" got %v", name)
	}

	defer cancel()
}

func TestCheckCtxTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		sleepTime time.Duration
		wantErr   error
	}{
		{
			"no_timeout",
			time.Millisecond * 200,
			time.Millisecond * 50,
			nil,
		},
		{
			"timeout",
			time.Millisecond * 5,
			time.Millisecond * 10,
			context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()
			time.Sleep(tt.sleepTime)
			err := CheckCtxTimeout(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ContextTimeoutCheck: expected = %v; got = %v", tt.wantErr, err)
			}
		})
	}
}
