package utils

// RestoreOptions represents the configuration for a single database Restore
type RestoreOptions struct {
	Host           string
	Port           int
	User           string
	Password       string
	DBName         string
	File           string
	VerifyChecksum bool
}
