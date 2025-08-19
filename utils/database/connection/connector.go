package connection

import (
	"database/sql"
	"fmt"
	"time"

	"sfDBTools/internal/logger"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

// createConnection creates a new database connection with the given DSN
// It is a helper function used by other functions in this package
func createConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Set reasonable connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	return db, nil
}

// buildDSN creates a DSN string for MySQL connections
// If dbName is empty, it will connect to the MySQL server without selecting a database
func buildDSN(config Config, includeDBName bool) string {
	dbPart := ""
	if includeDBName && config.DBName != "" {
		dbPart = config.DBName
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Password, config.Host, config.Port, dbPart)
}

// getLogger gets the logger or returns an error
func getLogger() (*logger.Logger, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}
	return lg, nil
}
