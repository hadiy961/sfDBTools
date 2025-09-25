package connection

import (
	"fmt"

	"sfDBTools/internal/logger"
)

// EnsureDatabase checks if the database exists and creates it if missing
func EnsureDatabase(config Config) error {
	lg, err := getLogger()
	if err != nil {
		return err
	}

	dsn := buildDSN(config, false)
	db, err := createConnection(dsn)
	if err != nil {
		lg.Error("Failed to open database connection", logger.Error(err))
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	var exists bool
	query := "SELECT COUNT(*) > 0 FROM information_schema.schemata WHERE schema_name = ?"
	if err := db.QueryRow(query, config.DBName).Scan(&exists); err != nil {
		lg.Error("Failed to check if database exists", logger.Error(err), logger.String("database", config.DBName))
		return fmt.Errorf("failed to check if database '%s' exists: %w", config.DBName, err)
	}

	if !exists {
		lg.Info("Creating database", logger.String("database", config.DBName))
		if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", config.DBName)); err != nil {
			lg.Error("Failed to create database", logger.Error(err), logger.String("database", config.DBName))
			return fmt.Errorf("failed to create database '%s': %w", config.DBName, err)
		}
	}

	lg.Info("Database ready", logger.String("database", config.DBName))
	return nil
}
