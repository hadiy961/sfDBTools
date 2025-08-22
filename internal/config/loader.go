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
	// First, check if config file exists at the required path
	requiredPath := "/etc/sfDBTools/config/config.yaml"

	if !fileExists(requiredPath) {
		return nil, fmt.Errorf("file konfigurasi tidak ditemukan di %s", requiredPath)
	}

	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add only the system-wide config path
	v.AddConfigPath("/etc/sfDBTools/config") // system-wide config (for root)

	// Default values (opsional)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")

	// ENV override (opsional)
	bindEnvironment(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("gagal membaca config dari %s: %w", v.ConfigFileUsed(), err)
	}

	return v, nil
}
