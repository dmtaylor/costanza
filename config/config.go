// Package config manages loading config variables from environment
package config

import (
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configName = "parameters"
var etcPath = "/etc/costanza"

var TokenPath = "discord.token"
var InsomniacIdsPath = "discord.insomniac_ids"
var InsomniacRolesPath = "discord.insomniac_roles"
var DbConnectionPath = "db.connection"

type Config struct {
	DiscordToken    string
	InsomniacIds    []string
	InsomniacRoles  []string
	DbConnectionStr string
}

func SetConfigDefaults() {
	viper.SetConfigName(configName)
	viper.AddConfigPath(etcPath)
	viper.AddConfigPath(".")
}

func Load() (*Config, error) {
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("config file not found. Continuing...\n")
		} else {
			return nil, errors.Wrap(err, "failed to load config file")
		}
	}
	cfg := Config{
		DiscordToken:    viper.GetString(TokenPath),
		InsomniacIds:    viper.GetStringSlice(InsomniacIdsPath),
		InsomniacRoles:  viper.GetStringSlice(InsomniacRolesPath),
		DbConnectionStr: viper.GetString(DbConnectionPath),
	}

	return &cfg, nil
}
