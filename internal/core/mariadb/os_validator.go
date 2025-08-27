package mariadb

import (
	"sfDBTools/utils/common"
)

// ValidateOperatingSystem checks if the current OS is supported for MariaDB operations
// This is now a wrapper around the common OS validation functionality
func ValidateOperatingSystem() error {
	supportedOS := common.MariaDBSupportedOS()
	return common.ValidateOperatingSystem(supportedOS)
}
