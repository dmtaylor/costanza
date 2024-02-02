// Package config manages loading config variables from environment
package config

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/viper"
)

var configName = "config"
var etcPath = "/etc/costanza"

var TokenPath = "discord.token"

type ListenConfig struct {
	GuildId         string `mapstructure:"guild_id"`
	ReportChannelId string `mapstructure:"report_channel_id"`
	StartTime       string `mapstructure:"start_time"` // Time in 24hr format UTC to run
}

type DiscordConfig struct {
	AppId                   string         `mapstructure:"app_id"`
	Token                   string         `mapstructure:"token"`
	ShardId                 uint64         `mapstructure:"shard_id"`
	ShardCount              uint64         `mapstructure:"shard_count"`
	InsomniacIds            []string       `mapstructure:"insomniac_ids"`
	InsomniacRoles          []string       `mapstructure:"insomniac_roles"`
	ListenConfigs           []ListenConfig `mapstructure:"listen_configs"`
	DefaultWeatherLocations []string       `mapstructure:"default_weather_locations"`
	ListenChannelSet        map[string]*ListenConfig
}

type MetricsConfig struct {
	HealthcheckEnabled bool   `mapstructure:"healthcheck_enabled"`
	MetricsEnabled     bool   `mapstructure:"metrics_enabled"`
	Appname            string `mapstructure:"appname"`
	LogLevel           string `mapstructure:"log_level"`
	MetricsPort        uint64 `mapstructure:"port"`
}

type DbConfig struct {
	Connection string
}

type Config struct {
	Discord DiscordConfig
	Metrics MetricsConfig
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
			Token:          "fake-default-value",
			InsomniacIds:   nil,
			InsomniacRoles: nil,
			ListenConfigs:  nil,
		},
		Metrics: MetricsConfig{
			HealthcheckEnabled: false,
			MetricsEnabled:     false,
			Appname:            "costanza-local",
			LogLevel:           "info",
		},
		Db: DbConfig{Connection: "postgres://costanza:myvoiceismypassportverifyme@localhost:5432/costanza?sslmode=disable"},
	}
	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
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
	GlobalConfig.Discord.ListenChannelSet = make(map[string]*ListenConfig, len(GlobalConfig.Discord.ListenConfigs))
	for _, listenConfig := range GlobalConfig.Discord.ListenConfigs {
		gid, _ := strconv.ParseUint(listenConfig.GuildId, 10, 64)
		if gid%GlobalConfig.Discord.ShardCount == GlobalConfig.Discord.ShardId {
			GlobalConfig.Discord.ListenChannelSet[listenConfig.GuildId] = &listenConfig
		}
	}

	return nil
}
