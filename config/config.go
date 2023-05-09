// Package config manages loading config variables from environment
package config

import (
	"fmt"
	"log"

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
	AppId                   string         `mapstructure:"app_id"`
	Token                   string         `mapstructure:"token"`
	InsomniacIds            []string       `mapstructure:"insomniac_ids"`
	InsomniacRoles          []string       `mapstructure:"insomniac_roles"`
	ListenConfigs           []ListenConfig `mapstructure:"listen_configs"`
	DefaultWeatherLocations []string       `mapstructure:"default_weather_locations"`
	ListenChannelSet        map[string]bool
}

type DbConfig struct {
	Connection string
}

type Config struct {
	Discord  DiscordConfig
	Db       DbConfig
	LogLevel string
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
			Token:          "fake-default-value",
			InsomniacIds:   nil,
			InsomniacRoles: nil,
			ListenConfigs:  nil,
		},
		Db:       DbConfig{Connection: "postgres://costanza:myvoiceismypassportverifyme@localhost:5432/costanza?sslmode=disable"},
		LogLevel: "info",
	}
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("config file not found. Continuing...\n")
		} else {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}
	err = viper.Unmarshal(&GlobalConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	initializeLogger()
	GlobalConfig.Discord.ListenChannelSet = make(map[string]bool, len(GlobalConfig.Discord.ListenConfigs))
	for _, listenConfig := range GlobalConfig.Discord.ListenConfigs {
		GlobalConfig.Discord.ListenChannelSet[listenConfig.GuildId] = true
	}

	return nil
}
