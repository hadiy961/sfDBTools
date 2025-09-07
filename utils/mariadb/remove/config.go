package remove

import (
	removeCore "sfDBTools/internal/core/mariadb/remove"
	"sfDBTools/utils/common"

	"github.com/spf13/cobra"
)

// RemoveConfig holds configuration for MariaDB removal operations
type RemoveConfig struct {
	SkipConfirm bool `json:"skip_confirm"`
}

// AddCommonRemoveFlags adds common flags for remove commands
func AddCommonRemoveFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("yes", false, "Skip confirmations and run non-interactively (dangerous)")
}

// ResolveRemoveConfig resolves remove configuration from command flags and environment
func ResolveRemoveConfig(cmd *cobra.Command) (*RemoveConfig, error) {
	config := &RemoveConfig{}

	// Get flags using shared helpers
	config.SkipConfirm = common.GetBoolFlagOrEnv(cmd, "yes", "SFDBTOOLS_SKIP_CONFIRM", false)

	return config, nil
}

// ToConfig converts RemoveConfig to the internal Config struct
func (rc *RemoveConfig) ToConfig() *removeCore.Config {
	return &removeCore.Config{
		SkipConfirm: rc.SkipConfirm,
	}
}
