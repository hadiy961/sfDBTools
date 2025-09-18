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
	// Determine possible config locations (order of preference):
	// 1) ./config/config.yaml (when running from project root via `go run`)
	// 2) <appDir>/config/config.yaml (preferred when running installed binary)
	// 3) /etc/sfDBTools/config/config.yaml (system-wide fallback)
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("gagal menentukan path executable: %w", err)
	}
	appDir := filepath.Dir(exePath)
	appConfigDir := filepath.Join(appDir, "config")
	systemConfigDir := "/etc/sfDBTools/config"

	// Also consider relative ./config in case we're running with `go run` from project root
	// or during tests. Use working directory's ./config as highest priority when present.
	cwd, _ := os.Getwd()
	cwdConfigDir := filepath.Join(cwd, "config")

	// Check for the presence of config.yaml in app path first, then system path
	var chosenConfigDir string
	if fileExists(filepath.Join(cwdConfigDir, "config.yaml")) {
		chosenConfigDir = cwdConfigDir
	} else if fileExists(filepath.Join(appConfigDir, "config.yaml")) {
		chosenConfigDir = appConfigDir
	} else if fileExists(filepath.Join(systemConfigDir, "config.yaml")) {
		chosenConfigDir = systemConfigDir
	} else {
		return nil, fmt.Errorf("file konfigurasi tidak ditemukan di %s, %s atau %s", cwdConfigDir, appConfigDir, systemConfigDir)
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
