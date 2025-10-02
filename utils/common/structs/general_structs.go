// File : utils/common/structs/general_structs.go
// Description : Structs untuk opsi koneksi database umum.
// Author : Hadiyatna Muflihun

package structs

// ConnectionOptions - Database connection related flags
type ConnectionOptions struct {
	Host     string `flag:"host" env:"SFDB_DB_HOST" default:"localhost"`
	Port     int    `flag:"port" env:"SFDB_DB_PORT" default:"3306"`
	User     string `flag:"user" env:"SFDB_DB_USER" default:"root"`
	Password string `flag:"password" env:"SFDB_DB_PASSWORD" default:""`
}

// SourceConnection - Source database connection structure
type SourceConnection struct {
	Host     string `flag:"source-host" env:"SFDB_SOURCE_DB_HOST" default:"localhost"`
	Port     int    `flag:"source-port" env:"SFDB_SOURCE_DB_PORT" default:"3306"`
	User     string `flag:"source-user" env:"SFDB_SOURCE_DB_USER" default:"root"`
	Password string `flag:"source-password" env:"SFDB_SOURCE_DB_PASSWORD" default:""`
}

// TargetConnection - Target database connection structure
type TargetConnection struct {
	Host     string `flag:"target-host" env:"SFDB_TARGET_DB_HOST" default:"localhost"`
	Port     int    `flag:"target-port" env:"SFDB_TARGET_DB_PORT" default:"3306"`
	User     string `flag:"target-user" env:"SFDB_TARGET_DB_USER" default:"root"`
	Password string `flag:"target-password" env:"SFDB_TARGET_DB_PASSWORD" default:""`
}

// EncryptionOptions - Encryption related flags
type EncryptionOptions struct {
	Required bool   `flag:"encrypt" env:"SFDB_ENCRYPT" default:"false"`
	Password string `flag:"encryption-password" env:"SFDB_ENCRYPTION_PASSWORD" default:""`
}
