package dbconfig

import (
	"github.com/spf13/cobra"
)

// AddCommonDbConfigFlags adds shared flags used across dbconfig commands
func AddCommonDbConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Specific encrypted config file")
}

// AddGenerateFlags adds flags specific to the generate command
func AddGenerateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "", "Configuration name (without extension)")
	cmd.Flags().StringP("host", "H", "", "Database host")
	cmd.Flags().IntP("port", "p", 0, "Database port")
	cmd.Flags().StringP("user", "u", "", "Database username")
	cmd.Flags().BoolP("auto", "a", false, "Auto mode - skip confirmations")
}

// AddDeleteFlags adds flags specific to the delete command
func AddDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Skip confirmation prompts")
	cmd.Flags().Bool("all", false, "Delete all encrypted config files")
}
