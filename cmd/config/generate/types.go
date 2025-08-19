package generate

// EncryptedDatabaseConfig represents the encrypted database configuration
type EncryptedDatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}
