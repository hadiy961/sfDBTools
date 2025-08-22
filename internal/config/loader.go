package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

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

	// Check if config file exists first
	configFile := v.ConfigFileUsed()
	if configFile == "" {
		// Try to find config file manually
		possiblePaths := []string{
			"./config/config.yaml",
			"./config/config.yml",
			"config.yaml",
			"config.yml",
		}

		var foundPath string
		for _, path := range possiblePaths {
			if fileExists(path) {
				foundPath = path
				break
			}
		}

		if foundPath == "" {
			return nil, fmt.Errorf("file konfigurasi tidak ditemukan. Pastikan file config.yaml ada di direktori ./config/")
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("gagal membaca config dari %s: %w", v.ConfigFileUsed(), err)
	}

	return v, nil
}
