package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type Config struct {
	DiscordToken string
}

var Values = Config{
	DiscordToken: "",
}

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return errors.Wrap(err, "failed to load dotenv")
	}
	Values.DiscordToken = os.Getenv("DISCORD_TOKEN")
	return nil
}
