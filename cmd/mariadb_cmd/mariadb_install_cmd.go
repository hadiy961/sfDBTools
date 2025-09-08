package mariadb_cmd

import (
	"github.com/spf13/cobra"
)

// InstallCmd installs MariaDB interactively
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB server with version selection",
	Long: `Install MariaDB server with simple version selection.

This command will:
1. Fetch available MariaDB versions from official API
2. Let you select a stable version to install  
3. Use official MariaDB repository setup script
4. Install MariaDB server and client packages
5. Start and enable MariaDB service

The installation is interactive and requires root privileges.

Examples:
  # Interactive MariaDB installation
  sudo sfdbtools mariadb install`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
