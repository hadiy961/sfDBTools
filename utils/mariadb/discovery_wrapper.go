package mariadb

import (
	"sfDBTools/utils/database"
	discovery "sfDBTools/utils/mariadb/discovery"
)

// Alias type so existing callers that expect mariadb.MariaDBInstallation continue to work
type MariaDBInstallation = discovery.MariaDBInstallation

// DiscoverMariaDBInstallation delegates to the discovery subpackage
func DiscoverMariaDBInstallation() (*MariaDBInstallation, error) {
	return discovery.DiscoverMariaDBInstallation()
}

// CreateDatabaseConfigFromInstallation creates a basic database.Config from installation info
func CreateDatabaseConfigFromInstallation(installation *MariaDBInstallation) *database.Config {
	if installation == nil {
		return nil
	}
	return &database.Config{
		Host:     "localhost",
		Port:     installation.Port,
		User:     "root",
		Password: "",
		DBName:   "",
	}
}
