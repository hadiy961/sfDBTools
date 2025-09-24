package backup_cmd

import (
	"sfDBTools/internal/config/model"
	"sfDBTools/internal/logger"
)

// Cfg and Lg are package-level variables that child commands in this
// package can use. Call Init from the application entrypoint to set them.
var Cfg *model.Config
var Lg *logger.Logger

// Init sets the package-level config and logger for the dbconfig_cmd package.
func Init(cfg *model.Config, lg *logger.Logger) {
	Cfg = cfg
	Lg = lg
}
