package cmd

import (
	"sfDBTools/cmd/backup_cmd"
	"sfDBTools/cmd/dbconfig_cmd"
	mariadb_cmd "sfDBTools/cmd/mariadb_cmd"
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/core/menu"
	"sfDBTools/internal/logger"

	"github.com/spf13/cobra"
)

var cfg *model.Config
var lg *logger.Logger

var rootCmd = &cobra.Command{
	Use:   "sfDBTools",
	Short: "sfDBTools CLI",
	Run: func(cmd *cobra.Command, args []string) {
		menu.MenuUtama(lg, cfg)
	},
}

func Execute(config *model.Config, logger *logger.Logger) error {
	// store provided config for use by commands
	cfg = config
	lg = logger

	// initialize sub-command packages that need cfg/lg
	// ensure dbconfig subpackage has access to cfg/lg
	dbconfig_cmd.Init(cfg, lg)
	// ensure mariadb subpackage has access to cfg/lg as well
	mariadb_cmd.Init(cfg, lg)
	// ensure backup subpackage has access to cfg/lg
	backup_cmd.Init(cfg, lg)

	return rootCmd.Execute()
}
