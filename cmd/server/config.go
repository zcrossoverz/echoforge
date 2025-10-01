// cmd/server/config.go
package main // ← Fix: Align với main.go

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`

	Database struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"database"`

	JWT struct {
		Secret    string `mapstructure:"secret"`
		ExpiresIn int    `mapstructure:"expires_in"`
	} `mapstructure:"jwt"`

	App struct {
		SiteID   string `mapstructure:"site_id"`
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"app"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)
	viper.AutomaticEnv() // Bind env vars (e.g., APP_SITE_ID=blog1)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
