// Package config manages loading config variables from environment
package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var OverwriteDiscordToken string
var OverwriteInsomniacIds []string
var OverwriteInsomniacRoles []string
var OverwriteDbConnectionStr string

type Config struct {
	DiscordToken    string
	InsomniacIds    []string
	InsomniacRoles  []string
	DbConnectionStr string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load dotenv: %v. Continuing...\n", err)
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

	var insomniacRoles []string
	if OverwriteInsomniacRoles != nil {
		insomniacRoles = OverwriteInsomniacRoles
	} else {
		insomniacRoles = strings.Split(os.Getenv("INSOMNIAC_ROLES"), ",")
	}

	var dbString string
	if OverwriteDbConnectionStr != "" {
		dbString = OverwriteDbConnectionStr
	} else {
		dbString = os.Getenv("DB_URL")
	}

	cfg := Config{
		DiscordToken:    discordtoken,
		InsomniacIds:    insomniacIds,
		InsomniacRoles:  insomniacRoles,
		DbConnectionStr: dbString,
	}
	return &cfg, nil
}
