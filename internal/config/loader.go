package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadViper() (*viper.Viper, error) {
	// Determine possible config locations:
	// 1) <appDir>/config/config.yaml (preferred if present)
	// 2) /etc/sfDBTools/config/config.yaml (fallback)
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("gagal menentukan path executable: %w", err)
	}
	appDir := filepath.Dir(exePath)
	appConfigDir := filepath.Join(appDir, "config")
	systemConfigDir := "/etc/sfDBTools/config"

	// Check for the presence of config.yaml in app path first, then system path
	var chosenConfigDir string
	if fileExists(filepath.Join(appConfigDir, "config.yaml")) {
		chosenConfigDir = appConfigDir
	} else if fileExists(filepath.Join(systemConfigDir, "config.yaml")) {
		chosenConfigDir = systemConfigDir
	} else {
		return nil, fmt.Errorf("file konfigurasi tidak ditemukan di %s atau %s", appConfigDir, systemConfigDir)
	}

	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add chosen config path (app-local preferred)
	v.AddConfigPath(chosenConfigDir)

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
