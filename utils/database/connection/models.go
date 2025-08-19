package connection

// Config represents database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}
