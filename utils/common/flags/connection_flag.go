package flags

import "github.com/spf13/cobra"

func ConnectionFlag(cmd *cobra.Command) {
	// Database connection options
	cmd.Flags().String("source_db", "", "database name")
	cmd.Flags().String("source_host", "", "source database host")
	cmd.Flags().Int("source_port", 0, "source database port")
	cmd.Flags().String("source_user", "", "source database user")
	cmd.Flags().String("source_password", "", "source database password")
	cmd.Flags().String("config", "", "encrypted configuration file (.cnf.enc)")
}
