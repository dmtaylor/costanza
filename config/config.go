package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type Config struct {
	DiscordToken string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load dotenv")
	}
	cfg := Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
	}
	return &cfg, nil
}
