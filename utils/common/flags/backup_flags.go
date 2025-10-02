package flags

import "github.com/spf13/cobra"

func AddBackupUserFlags(cmd *cobra.Command) {
	// Reuse common backup flags
	// AddCommonBackupFlags(cmd)

	// User grants specific options
	cmd.Flags().Bool("include-user", false, "include user grants in separate file")
}
