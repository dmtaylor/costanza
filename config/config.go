// Package config manages loading config variables from environment
package config

import (
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configName = "config"
var etcPath = "/etc/costanza"

var TokenPath = "discord.token"

type ListenConfig struct {
	GuildId         string `mapstructure:"guild_id"`
	ReportChannelId string `mapstructure:"report_channel_id"`
}

type DiscordConfig struct {
	Token            string
	InsomniacIds     []string       `mapstructure:"insomniac_ids"`
	InsomniacRoles   []string       `mapstructure:"insomniac_roles"`
	ListenConfigs    []ListenConfig `mapstructure:"listen_configs"`
	ListenChannelSet map[string]bool
}

type DbConfig struct {
	Connection string
}

type Config struct {
	Discord DiscordConfig
	Db      DbConfig
}

var GlobalConfig Config

func SetConfigDefaults() {
	viper.SetConfigName(configName)
	viper.AddConfigPath(etcPath)
	viper.AddConfigPath(".")
}

func LoadConfig() error {
	GlobalConfig = Config{
		Discord: DiscordConfig{
			Token:            "fake-default-value",
			InsomniacIds:     nil,
			InsomniacRoles:   nil,
			ListenConfigs:    nil,
			ListenChannelSet: make(map[string]bool, 0),
		},
		Db: DbConfig{Connection: "postgres://costanza:myvoiceismypassportverifyme@localhost:5432/costanza?sslmode=disable"},
	}
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("config file not found. Continuing...\n")
		} else {
			return errors.Wrap(err, "failed to load config file")
		}
	}
	err = viper.Unmarshal(&GlobalConfig)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}
	for _, listenConfig := range GlobalConfig.Discord.ListenConfigs {
		GlobalConfig.Discord.ListenChannelSet[listenConfig.GuildId] = true
	}

	return nil
}
