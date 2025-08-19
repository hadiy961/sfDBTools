package restore_utils

// RestoreConfig represents the resolved restore configuration
type RestoreConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	DBName         string
	File           string
	VerifyChecksum bool
}

// RestoreOptions represents the configuration for restore operations (backward compatibility)
type RestoreOptions struct {
	Host           string
	Port           int
	User           string
	Password       string
	DBName         string
	File           string
	VerifyChecksum bool
}

// RestoreUserConfig represents the resolved restore user grants configuration
type RestoreUserConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	File           string
	VerifyChecksum bool
}

// RestoreUserOptions represents the configuration for restore user grants operations
type RestoreUserOptions struct {
	Host           string
	Port           int
	User           string
	Password       string
	File           string
	VerifyChecksum bool
}

// ToRestoreOptions converts RestoreConfig to RestoreOptions for backward compatibility
func (rc *RestoreConfig) ToRestoreOptions() RestoreOptions {
	return RestoreOptions{
		Host:           rc.Host,
		Port:           rc.Port,
		User:           rc.User,
		Password:       rc.Password,
		DBName:         rc.DBName,
		File:           rc.File,
		VerifyChecksum: rc.VerifyChecksum,
	}
}

// ToRestoreUserOptions converts RestoreUserConfig to RestoreUserOptions for backward compatibility
func (ruc *RestoreUserConfig) ToRestoreUserOptions() RestoreUserOptions {
	return RestoreUserOptions{
		Host:           ruc.Host,
		Port:           ruc.Port,
		User:           ruc.User,
		Password:       ruc.Password,
		File:           ruc.File,
		VerifyChecksum: ruc.VerifyChecksum,
	}
}

// ConfigurationSource represents the source of restore configuration
type ConfigurationSource int

const (
	SourceConfigFile ConfigurationSource = iota
	SourceFlags
	SourceDefaults
	SourceInteractive
)
