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

	// Get Server UUID
	serverUUID, err := getServerUUID(db)
	if err != nil {
		lg.Warn("Failed to get SERVER_UUID", logger.Error(err))
	} else {
		gtidInfo.ServerUUID = serverUUID
	}

	lg.Info("GTID information collected",
		logger.Bool("has_gtid", gtidInfo.HasGTID),
		logger.String("server_uuid", gtidInfo.ServerUUID),
		logger.String("gtid_executed_length", fmt.Sprintf("%d chars", len(gtidInfo.GTIDExecuted))),
		logger.String("gtid_purged_length", fmt.Sprintf("%d chars", len(gtidInfo.GTIDPurged))))

	return gtidInfo, nil
}

// checkGTIDEnabled checks if GTID is enabled on the server
func checkGTIDEnabled(db *sql.DB) (bool, error) {
	var variableName, value string
	query := "SHOW VARIABLES LIKE 'gtid_mode'"

	err := db.QueryRow(query).Scan(&variableName, &value)
	if err != nil {
		if err == sql.ErrNoRows {
			// GTID might not be supported in this MySQL version
			return false, nil
		}
		return false, err
	}

	// GTID is enabled if gtid_mode is ON
	return strings.ToUpper(value) == "ON", nil
}

// getGTIDExecuted retrieves the GTID_EXECUTED global variable
func getGTIDExecuted(db *sql.DB) (string, error) {
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

// getGTIDPurged retrieves the GTID_PURGED global variable
func getGTIDPurged(db *sql.DB) (string, error) {
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
