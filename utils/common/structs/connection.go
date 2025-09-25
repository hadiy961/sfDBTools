package structs

// ConnectionOptions - Database connection related flags
type ConnectionOptions struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}
