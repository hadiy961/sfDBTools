package database

import (
	"database/sql"
	"fmt"
	"strings"

	"sfDBTools/internal/logger"
)

// GTIDInfo represents GTID information
type GTIDInfo struct {
	GTIDExecuted string `json:"gtid_executed"`
	GTIDPurged   string `json:"gtid_purged"`
	ServerUUID   string `json:"server_uuid"`
	HasGTID      bool   `json:"has_gtid"`
	BinlogFile   string `json:"binlog_file,omitempty"`   // From SHOW MASTER STATUS
	BinlogPos    int64  `json:"binlog_pos,omitempty"`    // From SHOW MASTER STATUS
	GTIDPosition string `json:"gtid_position,omitempty"` // From BINLOG_GTID_POS
}

// GetGTIDInfo retrieves GTID information from the database
func GetGTIDInfo(config Config) (*GTIDInfo, error) {
	lg, _ := logger.Get()

	// Connect without specifying a database for server-level info
	configWithoutDB := config
	configWithoutDB.DBName = ""

	db, err := GetWithoutDB(configWithoutDB)
	if err != nil {
		lg.Error("Failed to connect to database server", logger.Error(err))
		return nil, fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	gtidInfo := &GTIDInfo{
		HasGTID: false,
	}

	// Check if GTID is enabled
	hasGTID, err := checkGTIDEnabled(db)
	if err != nil {
		lg.Warn("Failed to check GTID status", logger.Error(err))
		return gtidInfo, nil
	}

	if !hasGTID {
		lg.Info("GTID is not enabled on this server")
		return gtidInfo, nil
	}

	gtidInfo.HasGTID = true

	// Get GTID_EXECUTED
	gtidExecuted, err := getGTIDExecuted(db)
	if err != nil {
		lg.Warn("Failed to get GTID_EXECUTED", logger.Error(err))
	} else {
		gtidInfo.GTIDExecuted = gtidExecuted
	}

	// Get GTID_PURGED
	gtidPurged, err := getGTIDPurged(db)
	if err != nil {
		lg.Warn("Failed to get GTID_PURGED", logger.Error(err))
	} else {
		gtidInfo.GTIDPurged = gtidPurged
	}

	// Get current binlog position for GTID calculation
	binlogFile, binlogPos, err := getMasterStatus(db)
	if err != nil {
		lg.Warn("Failed to get MASTER STATUS", logger.Error(err))
	} else {
		gtidInfo.BinlogFile = binlogFile
		gtidInfo.BinlogPos = binlogPos

		// Get GTID position using BINLOG_GTID_POS function
		gtidPosition, err := getBinlogGTIDPos(db, binlogFile, binlogPos)
		if err != nil {
			lg.Warn("Failed to get BINLOG_GTID_POS", logger.Error(err))
		} else {
			gtidInfo.GTIDPosition = gtidPosition
		}
	}

	lg.Info("GTID information collected",
		logger.Bool("has_gtid", gtidInfo.HasGTID),
		logger.String("server_uuid", gtidInfo.ServerUUID),
		logger.String("gtid_executed_length", fmt.Sprintf("%d chars", len(gtidInfo.GTIDExecuted))),
		logger.String("gtid_purged_length", fmt.Sprintf("%d chars", len(gtidInfo.GTIDPurged))),
		logger.String("binlog_file", gtidInfo.BinlogFile),
		logger.Int64("binlog_position", gtidInfo.BinlogPos),
		logger.String("gtid_position", gtidInfo.GTIDPosition))

	return gtidInfo, nil
}

// checkGTIDEnabled checks if GTID is enabled on the server
// Works for both MySQL and MariaDB by checking server capabilities
func checkGTIDEnabled(db *sql.DB) (bool, error) {
	lg, _ := logger.Get()

	// First check if this is MySQL or MariaDB
	version, err := getMySQLVersionString(db)
	if err != nil {
		lg.Warn("Could not determine database version", logger.Error(err))
	}

	isMariaDB := strings.Contains(strings.ToLower(version), "mariadb")

	if isMariaDB {
		// For MariaDB, check if gtid_domain_id exists and is configured
		var domainID sql.NullInt64
		err := db.QueryRow("SELECT @@GLOBAL.gtid_domain_id").Scan(&domainID)
		if err != nil {
			// gtid_domain_id doesn't exist, GTID not supported
			lg.Debug("MariaDB GTID not supported (no gtid_domain_id)", logger.Error(err))
			return false, nil
		}

		// Check if gtid_current_pos has any value (indicates GTID is working)
		var currentPos sql.NullString
		err = db.QueryRow("SELECT @@GLOBAL.gtid_current_pos").Scan(&currentPos)
		if err != nil {
			lg.Debug("MariaDB GTID variables not accessible", logger.Error(err))
			return false, nil
		}

		lg.Debug("MariaDB GTID status detected",
			logger.Bool("domain_id_valid", domainID.Valid),
			logger.String("current_pos", currentPos.String))

		return true, nil
	} else {
		// For MySQL, check gtid_mode variable
		var variableName, value string
		query := "SHOW VARIABLES LIKE 'gtid_mode'"

		err := db.QueryRow(query).Scan(&variableName, &value)
		if err != nil {
			if err == sql.ErrNoRows {
				// GTID not supported in this MySQL version
				lg.Debug("MySQL GTID not supported (no gtid_mode variable)")
				return false, nil
			}
			return false, err
		}

		enabled := strings.ToUpper(value) == "ON"
		lg.Debug("MySQL GTID status", logger.String("gtid_mode", value), logger.Bool("enabled", enabled))
		return enabled, nil
	}
}

// getGTIDExecuted retrieves the GTID_EXECUTED global variable (MySQL) or gtid_current_pos (MariaDB)
func getGTIDExecuted(db *sql.DB) (string, error) {
	version, err := getMySQLVersionString(db)
	if err != nil {
		// If we can't get version, try MySQL format first
		return getGTIDExecutedMySQL(db)
	}

	isMariaDB := strings.Contains(strings.ToLower(version), "mariadb")

	if isMariaDB {
		// MariaDB uses gtid_current_pos
		var gtidPos sql.NullString
		err := db.QueryRow("SELECT @@GLOBAL.gtid_current_pos").Scan(&gtidPos)
		if err != nil {
			return "", err
		}

		if gtidPos.Valid {
			return gtidPos.String, nil
		}
		return "", nil
	} else {
		// MySQL uses GTID_EXECUTED
		return getGTIDExecutedMySQL(db)
	}
}

// getGTIDExecutedMySQL retrieves GTID_EXECUTED for MySQL
func getGTIDExecutedMySQL(db *sql.DB) (string, error) {
	var gtidExecuted sql.NullString
	query := "SELECT @@GLOBAL.GTID_EXECUTED"

	err := db.QueryRow(query).Scan(&gtidExecuted)
	if err != nil {
		return "", err
	}

	if gtidExecuted.Valid {
		return gtidExecuted.String, nil
	}
	return "", nil
}

// getGTIDPurged retrieves the GTID_PURGED global variable (MySQL) or gtid_binlog_pos (MariaDB)
func getGTIDPurged(db *sql.DB) (string, error) {
	version, err := getMySQLVersionString(db)
	if err != nil {
		// If we can't get version, try MySQL format first
		return getGTIDPurgedMySQL(db)
	}

	isMariaDB := strings.Contains(strings.ToLower(version), "mariadb")

	if isMariaDB {
		// MariaDB uses gtid_binlog_pos (closest equivalent to GTID_PURGED)
		var gtidBinlogPos sql.NullString
		err := db.QueryRow("SELECT @@GLOBAL.gtid_binlog_pos").Scan(&gtidBinlogPos)
		if err != nil {
			return "", err
		}

		if gtidBinlogPos.Valid {
			return gtidBinlogPos.String, nil
		}
		return "", nil
	} else {
		// MySQL uses GTID_PURGED
		return getGTIDPurgedMySQL(db)
	}
}

// getGTIDPurgedMySQL retrieves GTID_PURGED for MySQL
func getGTIDPurgedMySQL(db *sql.DB) (string, error) {
	var gtidPurged sql.NullString
	query := "SELECT @@GLOBAL.GTID_PURGED"

	err := db.QueryRow(query).Scan(&gtidPurged)
	if err != nil {
		return "", err
	}

	if gtidPurged.Valid {
		return gtidPurged.String, nil
	}
	return "", nil
}

// getServerUUID retrieves the SERVER_UUID global variable
func getServerUUID(db *sql.DB) (string, error) {
	var serverUUID sql.NullString
	query := "SELECT @@GLOBAL.SERVER_UUID"

	err := db.QueryRow(query).Scan(&serverUUID)
	if err != nil {
		return "", err
	}

	if serverUUID.Valid {
		return serverUUID.String, nil
	}
	return "", nil
}

// getBinlogGTIDPos gets GTID position for a specific binlog file and position
// This function only works on MariaDB, returns empty for MySQL
func getBinlogGTIDPos(db *sql.DB, binlogFile string, binlogPos int64) (string, error) {
	version, err := getMySQLVersionString(db)
	if err != nil {
		// Can't determine version, skip BINLOG_GTID_POS
		return "", nil
	}

	isMariaDB := strings.Contains(strings.ToLower(version), "mariadb")

	if !isMariaDB {
		// BINLOG_GTID_POS function only exists in MariaDB
		return "", nil
	}

	// MariaDB: Use BINLOG_GTID_POS function
	query := fmt.Sprintf("SELECT BINLOG_GTID_POS('%s', %d)", binlogFile, binlogPos)

	var gtidPos sql.NullString
	err = db.QueryRow(query).Scan(&gtidPos)
	if err != nil {
		return "", fmt.Errorf("failed to get BINLOG_GTID_POS (MariaDB): %w", err)
	}

	if gtidPos.Valid {
		return gtidPos.String, nil
	}
	return "", nil
}

// GetBinaryLogInfo retrieves binary log information (for non-GTID setups)
type BinaryLogInfo struct {
	LogFile     string `json:"log_file"`
	LogPosition int64  `json:"log_position"`
	HasBinlog   bool   `json:"has_binlog"`
}

// GetBinaryLogInfo retrieves binary log position information
func GetBinaryLogInfo(config Config) (*BinaryLogInfo, error) {
	lg, _ := logger.Get()

	// Connect without specifying a database for server-level info
	configWithoutDB := config
	configWithoutDB.DBName = ""

	db, err := GetWithoutDB(configWithoutDB)
	if err != nil {
		lg.Error("Failed to connect to database server", logger.Error(err))
		return nil, fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()

	binlogInfo := &BinaryLogInfo{
		HasBinlog: false,
	}

	// Check if binary logging is enabled
	hasBinlog, err := checkBinaryLogEnabled(db)
	if err != nil {
		lg.Warn("Failed to check binary log status", logger.Error(err))
		return binlogInfo, nil
	}

	if !hasBinlog {
		lg.Info("Binary logging is not enabled on this server")
		return binlogInfo, nil
	}

	binlogInfo.HasBinlog = true

	// Get master status (current binlog position)
	logFile, logPosition, err := getMasterStatus(db)
	if err != nil {
		lg.Warn("Failed to get master status", logger.Error(err))
		return binlogInfo, nil
	}

	binlogInfo.LogFile = logFile
	binlogInfo.LogPosition = logPosition

	lg.Info("Binary log information collected",
		logger.Bool("has_binlog", binlogInfo.HasBinlog),
		logger.String("log_file", binlogInfo.LogFile),
		logger.Int64("log_position", binlogInfo.LogPosition))

	return binlogInfo, nil
}

// checkBinaryLogEnabled checks if binary logging is enabled
func checkBinaryLogEnabled(db *sql.DB) (bool, error) {
	var variableName, value string
	query := "SHOW VARIABLES LIKE 'log_bin'"

	err := db.QueryRow(query).Scan(&variableName, &value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return strings.ToUpper(value) == "ON", nil
}

// getMasterStatus retrieves the current binary log file and position
func getMasterStatus(db *sql.DB) (string, int64, error) {
	query := "SHOW MASTER STATUS"

	rows, err := db.Query(query)
	if err != nil {
		return "", 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", 0, fmt.Errorf("no master status available")
	}

	var file string
	var position int64
	var binlogDoDB sql.NullString
	var binlogIgnoreDB sql.NullString

	// MySQL 5.7+ has additional columns, but we only need the first two
	err = rows.Scan(&file, &position, &binlogDoDB, &binlogIgnoreDB)
	if err != nil {
		// Try with just file and position for older MySQL versions
		rows.Close()
		rows, err = db.Query(query)
		if err != nil {
			return "", 0, err
		}
		defer rows.Close()

		if !rows.Next() {
			return "", 0, fmt.Errorf("no master status available")
		}

		err = rows.Scan(&file, &position)
		if err != nil {
			return "", 0, err
		}
	}

	return file, position, nil
}

// ReplicationInfo represents complete replication information
type ReplicationInfo struct {
	GTIDInfo      *GTIDInfo      `json:"gtid_info,omitempty"`
	BinaryLogInfo *BinaryLogInfo `json:"binlog_info,omitempty"`
	MySQLVersion  string         `json:"mysql_version,omitempty"`
}

// GetReplicationInfo retrieves complete replication information (GTID + Binary Log)
func GetReplicationInfo(config Config) (*ReplicationInfo, error) {
	lg, _ := logger.Get()

	lg.Info("Collecting replication information",
		logger.String("host", config.Host),
		logger.Int("port", config.Port))

	replicationInfo := &ReplicationInfo{}

	// Get MySQL version
	version, err := GetMySQLVersion(config)
	if err != nil {
		lg.Warn("Failed to get MySQL version", logger.Error(err))
	} else {
		replicationInfo.MySQLVersion = version
	}

	// Get GTID information
	gtidInfo, err := GetGTIDInfo(config)
	if err != nil {
		lg.Warn("Failed to get GTID information", logger.Error(err))
	} else {
		replicationInfo.GTIDInfo = gtidInfo
	}

	// Get Binary Log information
	binlogInfo, err := GetBinaryLogInfo(config)
	if err != nil {
		lg.Warn("Failed to get binary log information", logger.Error(err))
	} else {
		replicationInfo.BinaryLogInfo = binlogInfo
	}

	lg.Info("Replication information collection completed",
		logger.String("mysql_version", replicationInfo.MySQLVersion),
		logger.Bool("has_gtid", replicationInfo.GTIDInfo != nil && replicationInfo.GTIDInfo.HasGTID),
		logger.Bool("has_binlog", replicationInfo.BinaryLogInfo != nil && replicationInfo.BinaryLogInfo.HasBinlog))

	return replicationInfo, nil
}

// getMySQLVersionString gets the version string from database
func getMySQLVersionString(db *sql.DB) (string, error) {
	var version string
	err := db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", err
	}
	return version, nil
}
