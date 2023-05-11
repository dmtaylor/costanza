package listen

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"

	"github.com/dmtaylor/costanza/config"
	"github.com/dmtaylor/costanza/internal/util"
)

const weatherBase = "https://wttr.in"
const weatherCommandName = "weather"

var weatherSlashCommand = &discordgo.ApplicationCommand{
	Name:        weatherCommandName,
	Type:        discordgo.ChatApplicationCommand,
	Description: "Gets weather for listed location, or costanza default",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "location",
			Description: "location for weather",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	},
}

func (s *Server) weatherCommand(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	// Ensure we only get options from slash commands
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.ApplicationCommandData().Name != weatherCommandName {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), interactionTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, "guildId", i.GuildID)
	ctx = context.WithValue(ctx, "interactionId", i.ID)
	ctx = context.WithValue(ctx, "commandName", weatherCommandName)
	var locations []string
	for _, option := range i.ApplicationCommandData().Options {
		if option.Name == "location" {
			locations = []string{option.StringValue()}
		}
	}
	if len(locations) < 1 {
		locations = config.GlobalConfig.Discord.DefaultWeatherLocations
	}
	slog.DebugCtx(ctx, "running weather command", "locations", locations)
	msg, err := getWeatherString(ctx, locations)
	if err != nil {
		slog.ErrorCtx(ctx, "failed getting weather data: "+err.Error(), "locations", locations)
		return
	}
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
	if err != nil {
		slog.ErrorCtx(ctx, "failed sending weather response: "+err.Error())
	}
	slog.DebugCtx(ctx, "finished weather command")
}

func getWeatherString(ctx context.Context, locations []string) (string, error) {
	b := strings.Builder{}
	for _, location := range locations {
		if err := util.CheckCtxTimeout(ctx); err != nil {
			return "", fmt.Errorf("context error: %w", err)
		}
		path := weatherBase + "/" + url.PathEscape(location) + "?format=3"
		slog.DebugCtx(ctx, "getting weather data", "location", location, "weatherCall", path)
		res, err := http.Get(path)
		if err != nil {
			return "", fmt.Errorf("failed getting weather data: %w", err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			return "", fmt.Errorf("failure from wttr code %d body %s: %w", res.StatusCode, body, err)
		}
		if err != nil {
			return "", fmt.Errorf("failure to get body: %w", err)
		}
		_, err = b.Write(body)
		if err != nil {
			return "", fmt.Errorf("failed to grow buffer: %w", err)
		}
		slog.DebugCtx(ctx, "got weather data", "location", location)

	}
	return b.String(), nil
}
