package fs

// This file contains pattern constant lists used by pattern matching operations.

var (
	// Log file patterns
	LogExtensions = []string{".log", ".err", ".pid", ".out"}
	LogPrefixes   = []string{"mysql-bin.", "mysql-relay-bin.", "slow", "error", "general", "access", "audit"}
	LogNames      = []string{
		"mysql.log", "mysqld.log", "error.log", "slow.log",
		"general.log", "relay.log", "mysqld.pid", "access.log",
		"audit.log", "binary.log", "update.log",
	}

	// Config file patterns
	ConfigExtensions = []string{".cnf", ".conf", ".cfg", ".ini", ".config", ".yaml", ".yml", ".json", ".toml"}
	ConfigPrefixes   = []string{"my.", "mysql.", "mariadb.", "config.", "settings."}
	ConfigNames      = []string{
		"my.cnf", "mysql.cnf", "mariadb.cnf", "server.cnf",
		"client.cnf", "mysqld.cnf", "mysql.conf", "config",
	}

	// Database file patterns
	DBExtensions = []string{
		".frm", ".ibd", ".MYD", ".MYI", ".opt", ".ARZ", ".ARM",
		".CSM", ".CSV", ".db", ".sqlite", ".sqlite3",
	}
	DBFiles = []string{"ibdata1", "ib_logfile0", "ib_logfile1", "auto.cnf"}

	// Backup file patterns
	BackupExtensions = []string{".bak", ".backup", ".dump", ".sql", ".gz", ".tar", ".zip", ".7z"}
	BackupPrefixes   = []string{"backup", "dump", "export", "mysqldump"}
	BackupSuffixes   = []string{"backup", "dump", "export", "old", "orig"}

	// Temporary file patterns
	TempExtensions = []string{".tmp", ".temp", ".swp", ".~", ".bak"}
	TempPrefixes   = []string{"tmp", "temp", ".", "#"}
	TempSuffixes   = []string{"~", ".tmp", ".temp"}

	// Directory lists
	SkipDirs = []string{"mysql", "performance_schema", "information_schema", "sys", "test"}

	SystemDirs = []string{
		"mysql", "performance_schema", "information_schema", "sys",
		"proc", "dev", "tmp", "var", "etc", "bin", "sbin", "usr",
		"lib", "lib64", "boot", "root", "home",
	}

	LogDirs = []string{"logs", "log", "var", "tmp", "spool"}
)
