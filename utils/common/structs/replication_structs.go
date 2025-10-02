package structs

// ReplicationMeta mendefinisikan metadata replikasi (terutama GTID) yang ditangkap
// selama operasi backup, yang diperlukan untuk restore dan penyiapan replika baru.
type ReplicationMeta struct {
	// Status Master/Replica saat dump dilakukan
	IsMasterDataEnabled bool `json:"is_master_data_enabled"` // Mereplikasi opsi MasterData dari BackupAllDBOptions

	// Posisi Log Biner Tradisional (dari SHOW MASTER STATUS)
	BinaryLogFile string `json:"binary_log_file"` // Nama file log biner (misalnya: mysql-bin.000005)
	LogPosition   uint32 `json:"log_position"`    // Posisi dalam file log biner (misalnya: 123456)

	// Posisi GTID (dari SELECT BINLOG_GTID_POS())
	GTIDSet string `json:"gtid_set"` // Nilai GTID Set (misalnya: 9b7f5e3b-748a-11eb-a2a3-0242ac110002:1-123)

	// Opsi tambahan: Posisi log biner yang dicatat oleh mysqldump
	// mysqldump --master-data=2 biasanya menulis ini sebagai komentar di file dump.
	DumpedGTIDSet string `json:"dumped_gtid_set,omitempty"`
}
