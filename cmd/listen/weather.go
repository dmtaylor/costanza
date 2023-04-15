package listen

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/dmtaylor/costanza/config"
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

var httpClientPool = sync.Pool{
	New: func() any {
		return new(http.Client)
	},
}

func (s *Server) weatherCommand(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name != weatherCommandName {
		return
	}
	var locations []string
	for _, option := range i.ApplicationCommandData().Options {
		if option.Name == "location" {
			locations = []string{option.StringValue()}
		}
	}
	if len(locations) < 1 {
		locations = config.GlobalConfig.Discord.DefaultWeatherLocations
	}
	msg, err := getWeatherString(locations)
	if err != nil {
		log.Printf("failed getting weather for %v: %s", locations, err)
		return
	}
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
	if err != nil {
		log.Printf("failed sending weather response: %s", err)
	}
}

func getWeatherString(locations []string) (string, error) {
	client := httpClientPool.Get().(*http.Client)
	defer httpClientPool.Put(client)
	lock := sync.Mutex{}
	b := strings.Builder{}
	for _, location := range locations {
		res, err := client.Get(weatherBase + "/" + url.PathEscape(location) + "?format=3")
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
		lock.Lock()
		_, err = b.Write(body)
		lock.Unlock()
		if err != nil {
			return "", fmt.Errorf("failed to grow buffer: %w", err)
		}

	}
	return b.String(), nil
}
