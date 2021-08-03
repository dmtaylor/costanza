package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const DEFAULT_CONNECTION_STR = "file:db.sqlite3?cache=shared&mode=ro"

var OverwriteDiscordToken string
var OverwriteInsomniacIds []string
var OverwriteDbConnectionStr string

type Config struct {
	DiscordToken    string
	InsomniacIds    []string
	DbConnectionStr string
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

	var dbString string
	if OverwriteDbConnectionStr != "" {
		dbString = OverwriteDbConnectionStr
	} else {
		dbString = DEFAULT_CONNECTION_STR
	}

	cfg := Config{
		DiscordToken:    discordtoken,
		InsomniacIds:    insomniacIds,
		DbConnectionStr: dbString,
	}
	return &cfg, nil
}
