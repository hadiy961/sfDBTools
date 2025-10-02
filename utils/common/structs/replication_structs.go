package structs

// ReplicationMeta represents replication information in metadata
type ReplicationMeta struct {
	HasGTID      bool   `json:"has_gtid"`
	GTIDExecuted string `json:"gtid_executed,omitempty"`
	GTIDPurged   string `json:"gtid_purged,omitempty"`
	ServerUUID   string `json:"server_uuid,omitempty"`
	HasBinlog    bool   `json:"has_binlog"`
	LogFile      string `json:"log_file,omitempty"`
	LogPosition  int64  `json:"log_position,omitempty"`
	GTIDPosition string `json:"gtid_position,omitempty"` // From BINLOG_GTID_POS function
}
