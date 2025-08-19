package info

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// DatabaseInfo represents basic information about a database
type DatabaseListItem struct {
	Name string `json:"name"`
}

// ListDatabases returns a list of all databases (excluding system databases)
func ListDatabases(config database.Config) ([]string, error) {
	lg, _ := logger.Get()

	// Connect without specifying a database
	configWithoutDB := config
	configWithoutDB.DBName = ""

	db, err := database.GetWithoutDB(configWithoutDB)
	if err != nil {
		lg.Error("Failed to connect to database server", logger.Error(err))
		return nil, fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	// Query to get all databases excluding system databases
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN (
			'information_schema', 'performance_schema', 'mysql', 'sys'
		) 
		ORDER BY schema_name
	`

	rows, err := db.Query(query)
	if err != nil {
		// Fallback to SHOW DATABASES
		return listDatabasesFallback(db)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		databases = append(databases, dbName)
	}

	lg.Debug("Retrieved database list", logger.Int("count", len(databases)))
	return databases, nil
}

// listDatabasesFallback uses SHOW DATABASES as fallback
func listDatabasesFallback(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	systemDatabases := map[string]bool{
		"information_schema": true,
		"performance_schema": true,
		"mysql":              true,
		"sys":                true,
	}

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		// Skip system databases
		if !systemDatabases[dbName] {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// SelectDatabaseInteractive displays available databases and lets user choose one
func SelectDatabaseInteractive(config database.Config) (string, error) {
	databases, err := ListDatabases(config)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve database list: %w", err)
	}

	if len(databases) == 0 {
		fmt.Println("❌ No databases found.")
		return "", fmt.Errorf("no databases found")
	}

	// Display available databases
	fmt.Println("📁 Available Databases:")
	fmt.Println("======================")
	for i, db := range databases {
		fmt.Printf("   %d. %s\n", i+1, db)
	}

	// Let user choose
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect database (1-%d): ", len(databases))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(databases) {
		return "", fmt.Errorf("invalid selection: %s", choice)
	}

	return databases[index-1], nil
}

// SelectMultipleDatabasesInteractive displays available databases and lets user choose multiple
func SelectMultipleDatabasesInteractive(config database.Config) ([]string, error) {
	databases, err := ListDatabases(config)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve database list: %w", err)
	}

	if len(databases) == 0 {
		fmt.Println("❌ No databases found.")
		return nil, fmt.Errorf("no databases found")
	}

	// Display available databases
	fmt.Println("📁 Available Databases:")
	fmt.Println("======================")
	for i, db := range databases {
		fmt.Printf("   %d. %s\n", i+1, db)
	}

	// Let user choose multiple databases
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect databases (comma-separated, e.g. 1,3,5 or ranges like 1-3,5): ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read selection: %w", err)
	}

	choice = strings.TrimSpace(choice)
	if choice == "" {
		return nil, fmt.Errorf("no databases selected")
	}

	// Parse selection
	selectedIndexes, err := parseSelection(choice, len(databases))
	if err != nil {
		return nil, fmt.Errorf("invalid selection: %w", err)
	}

	// Get selected databases
	var selectedDatabases []string
	for _, index := range selectedIndexes {
		selectedDatabases = append(selectedDatabases, databases[index-1])
	}

	return selectedDatabases, nil
}

// parseSelection parses user selection string (e.g., "1,3,5" or "1-3,5")
func parseSelection(selection string, maxCount int) ([]int, error) {
	var indexes []int
	seen := make(map[int]bool)

	parts := strings.Split(selection, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			// Handle range (e.g., "1-3")
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start number in range: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end number in range: %s", rangeParts[1])
			}

			if start < 1 || end > maxCount || start > end {
				return nil, fmt.Errorf("invalid range: %d-%d (valid: 1-%d)", start, end, maxCount)
			}

			for i := start; i <= end; i++ {
				if !seen[i] {
					indexes = append(indexes, i)
					seen[i] = true
				}
			}
		} else {
			// Handle single number
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}

			if index < 1 || index > maxCount {
				return nil, fmt.Errorf("invalid selection: %d (valid: 1-%d)", index, maxCount)
			}

			if !seen[index] {
				indexes = append(indexes, index)
				seen[index] = true
			}
		}
	}

	return indexes, nil
}
