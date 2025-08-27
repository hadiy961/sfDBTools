package configure

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/terminal"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseManager handles database and user creation
type DatabaseManager struct {
	settings   *MariaDBSettings
	clientCode string
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(settings *MariaDBSettings, clientCode string) *DatabaseManager {
	return &DatabaseManager{
		settings:   settings,
		clientCode: clientCode,
	}
}

// SetupDatabasesAndUsers creates default databases and users
func (d *DatabaseManager) SetupDatabasesAndUsers() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Setting up default databases and users...")

	// Try programmatic connection first
	db, err := d.connectToMariaDB()
	if err != nil {
		lg.Warn("Programmatic connection failed, trying command line method", logger.Error(err))

		// Fallback to command line method
		return d.setupUsingCommandLine()
	}
	defer db.Close()

	// Create users
	if err := d.createUsers(db); err != nil {
		return fmt.Errorf("failed to create users: %w", err)
	}

	// Grant privileges
	if err := d.grantPrivileges(db); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	// Create databases
	if err := d.createDatabases(db); err != nil {
		return fmt.Errorf("failed to create databases: %w", err)
	}

	// Flush privileges
	if err := d.flushPrivileges(db); err != nil {
		return fmt.Errorf("failed to flush privileges: %w", err)
	}

	lg.Info("Database setup completed successfully")
	terminal.PrintSuccess("Default databases and users created successfully")
	return nil
}

// connectToMariaDB establishes connection to MariaDB
func (d *DatabaseManager) connectToMariaDB() (*sql.DB, error) {
	lg, _ := logger.Get()

	// Try different connection methods for root user
	socketPath := fmt.Sprintf("%s/mysql.sock", d.settings.DataDir)
	connectionMethods := []string{
		fmt.Sprintf("root@unix(%s)/", socketPath),               // Unix socket (preferred)
		"root:@tcp(localhost:%d)/",                              // TCP without password
		fmt.Sprintf("root@tcp(localhost:%d)/", d.settings.Port), // TCP without password (explicit port)
	}

	var db *sql.DB
	var err error

	for i, dsnTemplate := range connectionMethods {
		var dsn string
		if i == 0 {
			dsn = dsnTemplate // Unix socket - no port needed
		} else if i == 1 {
			dsn = fmt.Sprintf(dsnTemplate, d.settings.Port)
		} else {
			dsn = dsnTemplate
		}

		lg.Info("Attempting database connection", logger.String("method", dsn))

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			lg.Warn("Failed to open database connection", logger.String("dsn", dsn), logger.Error(err))
			continue
		}

		// Test connection
		if err = db.Ping(); err != nil {
			lg.Warn("Failed to ping database", logger.String("dsn", dsn), logger.Error(err))
			db.Close()
			continue
		}

		lg.Info("Connected to MariaDB successfully", logger.String("method", dsn))
		return db, nil
	}

	// If all connection methods failed
	lg.Error("All database connection methods failed", logger.Error(err))
	return nil, fmt.Errorf("unable to connect to MariaDB: all connection methods failed")
}

// createUsers creates all required users
func (d *DatabaseManager) createUsers(db *sql.DB) error {
	lg, _ := logger.Get()

	users := []struct {
		username string
		password string
	}{
		{"papp", "P@ssw0rdpapp!@#"},
		{"sysadmin", "P@ssw0rdsys!@#"},
		{"dbaDO", "DataOn24!!"},
		{"sst_user", "P@ssw0rdsst!@#"},
		{"backup_user", "P@ssw0rdBackup!@#"},
		{"restore_user", "P@ssw0rdRestore!@#"},
		{"maxscale", "P@ssw0rdMaxscale!@#"},
		{fmt.Sprintf("sfnbc_%s_admin", d.clientCode), "P@ssw0rdadm!@#"},
		{fmt.Sprintf("sfnbc_%s_user", d.clientCode), "P@ssw0rduser!@#"},
		{fmt.Sprintf("sfnbc_%s_fin", d.clientCode), "P@ssw0rdfin!@#"},
	}

	for _, user := range users {
		query := fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", user.username, user.password)
		if _, err := db.Exec(query); err != nil {
			lg.Error("Failed to create user",
				logger.String("username", user.username),
				logger.Error(err))
			return err
		}

		lg.Info("User created successfully", logger.String("username", user.username))
	}

	terminal.PrintSuccess("All users created successfully")
	return nil
}

// grantPrivileges grants appropriate privileges to users
func (d *DatabaseManager) grantPrivileges(db *sql.DB) error {
	lg, _ := logger.Get()

	privileges := []string{
		// Administrative users
		"GRANT ALL PRIVILEGES ON *.* TO 'papp'@'%'",
		"GRANT ALL PRIVILEGES ON *.* TO 'sysadmin'@'%'",
		"GRANT ALL PRIVILEGES ON *.* TO 'dbaDO'@'%' WITH GRANT OPTION",

		// SST user
		"GRANT ALL PRIVILEGES ON *.* TO 'sst_user'@'%'",

		// Backup user
		"GRANT SELECT, SHOW VIEW, TRIGGER, LOCK TABLES, EVENT, EXECUTE, RELOAD, PROCESS, REPLICATION CLIENT ON *.* TO 'backup_user'@'%'",

		// Restore user - limited to secondary training databases
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training`.* TO 'restore_user'@'%%'", d.clientCode),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training_dmart`.* TO 'restore_user'@'%%'", d.clientCode),

		// MaxScale user
		"GRANT ALL PRIVILEGES ON *.* TO 'maxscale'@'%'",
	}

	// Add application user privileges
	appUsers := []string{
		fmt.Sprintf("sfnbc_%s_admin", d.clientCode),
		fmt.Sprintf("sfnbc_%s_user", d.clientCode),
		fmt.Sprintf("sfnbc_%s_fin", d.clientCode),
	}

	databases := []string{
		fmt.Sprintf("dbsf_nbc_%s", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_dmart", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_temp", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_archive", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_secondary_training", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_secondary_training_dmart", d.clientCode),
	}

	for _, database := range databases {
		for _, user := range appUsers {
			privilege := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", database, user)
			privileges = append(privileges, privilege)
		}
	}

	// Execute all privilege grants
	for _, privilege := range privileges {
		if _, err := db.Exec(privilege); err != nil {
			lg.Error("Failed to grant privilege",
				logger.String("privilege", privilege),
				logger.Error(err))
			return err
		}
	}

	lg.Info("All privileges granted successfully")
	terminal.PrintSuccess("User privileges granted successfully")
	return nil
}

// createDatabases creates default databases
func (d *DatabaseManager) createDatabases(db *sql.DB) error {
	lg, _ := logger.Get()

	databases := []string{
		fmt.Sprintf("dbsf_nbc_%s", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_dmart", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_temp", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_archive", d.clientCode),
		fmt.Sprintf("sfDBTools_%s", d.clientCode),
	}

	for _, database := range databases {
		query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", database)
		if _, err := db.Exec(query); err != nil {
			lg.Error("Failed to create database",
				logger.String("database", database),
				logger.Error(err))
			return err
		}

		lg.Info("Database created successfully", logger.String("database", database))
	}

	terminal.PrintSuccess("Default databases created successfully")
	return nil
}

// flushPrivileges flushes the privilege tables
func (d *DatabaseManager) flushPrivileges(db *sql.DB) error {
	lg, _ := logger.Get()

	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		lg.Error("Failed to flush privileges", logger.Error(err))
		return err
	}

	lg.Info("Privileges flushed successfully")
	return nil
}

// setupUsingCommandLine creates databases and users using mysql command line
func (d *DatabaseManager) setupUsingCommandLine() error {
	lg, _ := logger.Get()

	terminal.PrintInfo("Using command line method for database setup...")

	// Create SQL script content
	sqlScript := d.generateSQLScript()

	// Write script to temporary file
	scriptFile := "/tmp/mariadb_setup.sql"
	if err := d.writeScriptToFile(scriptFile, sqlScript); err != nil {
		return fmt.Errorf("failed to write SQL script: %w", err)
	}
	defer func() {
		// Clean up temporary file
		if err := os.Remove(scriptFile); err != nil {
			lg.Warn("Failed to remove temporary SQL script", logger.Error(err))
		}
	}()

	// Execute script using mysql command
	socketPath := fmt.Sprintf("%s/mysql.sock", d.settings.DataDir)
	cmd := fmt.Sprintf("mysql -u root --socket=%s < %s", socketPath, scriptFile)
	if err := d.executeCommand(cmd); err != nil {
		// Try alternative socket path
		cmd = fmt.Sprintf("mysql -u root --socket=/var/run/mysqld/mysqld.sock < %s", scriptFile)
		if err := d.executeCommand(cmd); err != nil {
			// Try TCP connection without password
			cmd = fmt.Sprintf("mysql -u root -h localhost -P %d < %s", d.settings.Port, scriptFile)
			if err := d.executeCommand(cmd); err != nil {
				return fmt.Errorf("all mysql command line methods failed: %w", err)
			}
		}
	}

	lg.Info("Database setup completed successfully using command line")
	terminal.PrintSuccess("Default databases and users created successfully")
	return nil
}

// generateSQLScript creates the SQL script for database and user setup
func (d *DatabaseManager) generateSQLScript() string {
	var script strings.Builder

	// Create users
	users := []struct {
		username string
		password string
	}{
		{"papp", "P@ssw0rdpapp!@#"},
		{"sysadmin", "P@ssw0rdsys!@#"},
		{"dbaDO", "DataOn24!!"},
		{"sst_user", "P@ssw0rdsst!@#"},
		{"backup_user", "P@ssw0rdBackup!@#"},
		{"restore_user", "P@ssw0rdRestore!@#"},
		{"maxscale", "P@ssw0rdMaxscale!@#"},
		{fmt.Sprintf("sfnbc_%s_admin", d.clientCode), "P@ssw0rdadm!@#"},
		{fmt.Sprintf("sfnbc_%s_user", d.clientCode), "P@ssw0rduser!@#"},
		{fmt.Sprintf("sfnbc_%s_fin", d.clientCode), "P@ssw0rdfin!@#"},
	}

	script.WriteString("-- Create users\n")
	for _, user := range users {
		script.WriteString(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s';\n", user.username, user.password))
	}

	script.WriteString("\n-- Grant privileges\n")

	// Administrative users
	script.WriteString("GRANT ALL PRIVILEGES ON *.* TO 'papp'@'%';\n")
	script.WriteString("GRANT ALL PRIVILEGES ON *.* TO 'sysadmin'@'%';\n")
	script.WriteString("GRANT ALL PRIVILEGES ON *.* TO 'dbaDO'@'%' WITH GRANT OPTION;\n")
	script.WriteString("GRANT ALL PRIVILEGES ON *.* TO 'sst_user'@'%';\n")
	script.WriteString("GRANT ALL PRIVILEGES ON *.* TO 'maxscale'@'%';\n")

	// Backup user
	script.WriteString("GRANT SELECT, SHOW VIEW, TRIGGER, LOCK TABLES, EVENT, EXECUTE, RELOAD, PROCESS, REPLICATION CLIENT ON *.* TO 'backup_user'@'%';\n")

	// Restore user
	script.WriteString(fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training`.* TO 'restore_user'@'%%';\n", d.clientCode))
	script.WriteString(fmt.Sprintf("GRANT ALL PRIVILEGES ON `dbsf_nbc_%s_secondary_training_dmart`.* TO 'restore_user'@'%%';\n", d.clientCode))

	// Application users
	databases := []string{
		fmt.Sprintf("dbsf_nbc_%s", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_dmart", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_temp", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_archive", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_secondary_training", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_secondary_training_dmart", d.clientCode),
	}

	appUsers := []string{
		fmt.Sprintf("sfnbc_%s_admin", d.clientCode),
		fmt.Sprintf("sfnbc_%s_user", d.clientCode),
		fmt.Sprintf("sfnbc_%s_fin", d.clientCode),
	}

	for _, db := range databases {
		for _, user := range appUsers {
			script.WriteString(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%';\n", db, user))
		}
	}

	// Create databases
	script.WriteString("\n-- Create databases\n")
	mainDatabases := []string{
		fmt.Sprintf("dbsf_nbc_%s", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_dmart", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_temp", d.clientCode),
		fmt.Sprintf("dbsf_nbc_%s_archive", d.clientCode),
	}

	for _, db := range mainDatabases {
		script.WriteString(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;\n", db))
	}

	script.WriteString("\nFLUSH PRIVILEGES;\n")

	return script.String()
}

// writeScriptToFile writes SQL script to a file
func (d *DatabaseManager) writeScriptToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// executeCommand executes a shell command
func (d *DatabaseManager) executeCommand(command string) error {
	lg, _ := logger.Get()

	lg.Info("Executing command", logger.String("command", command))

	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		lg.Error("Command execution failed",
			logger.String("command", command),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("command failed: %s, output: %s", err, string(output))
	}

	lg.Info("Command executed successfully",
		logger.String("command", command),
		logger.String("output", string(output)))

	return nil
}
