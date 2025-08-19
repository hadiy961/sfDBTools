package database

import (
	"database/sql"
	"fmt"
	"strconv"

	"sfDBTools/internal/logger"
)

// MaxStatementTimeManager manages the global max_statement_time setting
type MaxStatementTimeManager struct {
	db           *sql.DB
	originalTime string
	lg           *logger.Logger
}

// NewMaxStatementTimeManager creates a new max_statement_time manager
func NewMaxStatementTimeManager(config Config) (*MaxStatementTimeManager, error) {
	lg, err := logger.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	// Convert to connection.Config
	connConfig := Config{
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		DBName:   "", // Don't need specific database for global settings
	}

	// Create connection without selecting specific database for global settings
	db, err := GetWithoutDB(connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	manager := &MaxStatementTimeManager{
		db: db,
		lg: lg,
	}

	// Get original max_statement_time value
	if err := manager.getOriginalTime(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to get original max_statement_time: %w", err)
	}

	return manager, nil
}

// getOriginalTime retrieves the current max_statement_time value
func (m *MaxStatementTimeManager) getOriginalTime() error {
	query := "SELECT @@global.max_statement_time"
	var timeValue interface{}

	err := m.db.QueryRow(query).Scan(&timeValue)
	if err != nil {
		return fmt.Errorf("failed to query max_statement_time: %w", err)
	}

	// Convert to string for storage
	switch v := timeValue.(type) {
	case int64:
		m.originalTime = strconv.FormatInt(v, 10)
	case uint64:
		m.originalTime = strconv.FormatUint(v, 10)
	case string:
		m.originalTime = v
	case []byte:
		m.originalTime = string(v)
	default:
		m.originalTime = fmt.Sprintf("%v", v)
	}

	m.lg.Info("Retrieved original max_statement_time",
		logger.String("value", m.originalTime))

	return nil
}

// SetUnlimited sets max_statement_time to 0 (unlimited)
func (m *MaxStatementTimeManager) SetUnlimited() error {
	query := "SET GLOBAL max_statement_time = 0"
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to set unlimited max_statement_time: %w", err)
	}

	m.lg.Info("Set max_statement_time to unlimited (0)")
	return nil
}

// Restore restores the original max_statement_time value
func (m *MaxStatementTimeManager) Restore() error {
	if m.originalTime == "" {
		m.lg.Warn("No original max_statement_time to restore")
		return nil
	}

	query := fmt.Sprintf("SET GLOBAL max_statement_time = %s", m.originalTime)
	_, err := m.db.Exec(query)
	if err != nil {
		m.lg.Error("Failed to restore max_statement_time",
			logger.Error(err),
			logger.String("original_value", m.originalTime))
		return fmt.Errorf("failed to restore max_statement_time: %w", err)
	}

	m.lg.Info("Restored max_statement_time to original value",
		logger.String("value", m.originalTime))
	return nil
}

// Close closes the database connection
func (m *MaxStatementTimeManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// setupMaxStatementTimeManager creates and configures max_statement_time manager
func SetupMaxStatementTimeManager(config Config, lg *logger.Logger) (*MaxStatementTimeManager, error) {
	timeManager, err := NewMaxStatementTimeManager(config)
	if err != nil {
		lg.Warn("Failed to create max_statement_time manager", logger.Error(err))
		return nil, err
	}

	// Set unlimited max_statement_time for the backup process
	if err := timeManager.SetUnlimited(); err != nil {
		lg.Warn("Failed to set unlimited max_statement_time", logger.Error(err))
	}

	return timeManager, nil
}

// cleanupMaxStatementTimeManager properly restores and closes max_statement_time manager
func CleanupMaxStatementTimeManager(timeManager *MaxStatementTimeManager) {
	if timeManager != nil {
		timeManager.Restore()
		timeManager.Close()
	}
}
