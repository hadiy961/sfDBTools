package mariadb

import (
	"github.com/spf13/cobra"
)

// VersionConfig holds configuration for version checking operations
type VersionConfig struct {
	OutputFormat string `json:"output_format"` // json or default
}

// AddCommonVersionFlags adds common flags for version checking commands
func AddCommonVersionFlags(cmd *cobra.Command) {
	// No flags for now - keep it simple
}

// ResolveVersionConfig resolves version checking configuration from command flags
func ResolveVersionConfig(cmd *cobra.Command) (*VersionConfig, error) {
	config := &VersionConfig{}

	// Check if output flag exists and use it
	if cmd.Flags().Lookup("output") != nil {
		if output, err := cmd.Flags().GetString("output"); err == nil {
			config.OutputFormat = output
		}
	}

	return config, nil
}
