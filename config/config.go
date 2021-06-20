package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var OverwriteDiscordToken string
var OverwriteInsomniacIds []string

type Config struct {
	DiscordToken string
	InsomniacIds []string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load dotenv")
	}

	var discordtoken string
	if OverwriteDiscordToken != "" {
		discordtoken = OverwriteDiscordToken
	} else {
		discordtoken = os.Getenv("DISCORD_TOKEN")
	}

	var insomniacIds []string
	if OverwriteInsomniacIds != nil {
		insomniacIds = OverwriteInsomniacIds
	} else {
		insomniacIds = strings.Split(os.Getenv("INSOMNIAC_IDS"), ",")
	}

	cfg := Config{
		DiscordToken: discordtoken,
		InsomniacIds: insomniacIds,
	}
	return &cfg, nil
}
