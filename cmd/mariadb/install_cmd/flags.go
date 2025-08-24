package installcmd

import (
	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

// InitFlags registers flags on the provided cobra command.
// It attempts to read the default MariaDB version from the system config
// file (/etc/sfDBTools/config/config.yaml). If the config is missing or
// invalid, a sensible hardcoded fallback is used.
func InitFlags(cmd *cobra.Command) {
	lg, _ := logger.Get()

	// Prefer default version from application config; fallback to static value
	defaultVersion := "10.6.23"

	// Validate and load config file if present
	if err := config.ValidateConfigFile(); err == nil {
		if cfg, err := config.LoadConfig(); err == nil && cfg != nil {
			if v := cfg.MariaDB.DefaultVersion; v != "" {
				defaultVersion = v
			}
		} else {
			lg.Debug("Failed to load config for default version, using fallback", logger.Error(err))
		}
	} else {
		lg.Debug("Config file not found or unreadable, using fallback version", logger.Error(err))
	}

	cmd.Flags().String("version", defaultVersion, "MariaDB version to install")
	cmd.Flags().Bool("force", false, "Skip confirmation prompts")
}
