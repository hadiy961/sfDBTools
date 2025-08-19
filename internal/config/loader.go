package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func loadViper() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config") // default path: ./config/config.yaml

	// Default values (opsional)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")

	// ENV override (opsional)
	bindEnvironment(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("gagal membaca config: %w", err)
	}

	return v, nil
}
