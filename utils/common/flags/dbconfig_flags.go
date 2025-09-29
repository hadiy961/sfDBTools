package flags

import "github.com/spf13/cobra"

// AddGenerateFlags adds flags specific to the generate command
func AddGenerateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("encryption-password", "e", "", "Encryption password (or set SFDB_ENCRYPTION_PASSWORD env variable)")
}

// AddCommonDbConfigFlags adds shared flags used across dbconfig commands
func AddCommonDbConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Specific encrypted config file")
}

// AddDeleteFlags adds flags specific to the delete command
func AddDeleteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Skip confirmation prompts")
	cmd.Flags().Bool("all", false, "Delete all encrypted config files")
}
