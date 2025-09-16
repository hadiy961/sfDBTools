package model

type Config struct {
	General     GeneralConfig   `mapstructure:"general"`
	Log         LogConfig       `mapstructure:"log"`
	Mysqldump   MysqldumpConfig `mapstructure:"mysqldump"`
	Database    DatabaseConfig  `mapstructure:"database"`
	Backup      BackupConfig    `mapstructure:"backup"`
	SystemUsers SystemUsers     `mapstructure:"system_users"`
	ConfigDir   ConfigDirConfig `mapstructure:"config_dir"`
	MariaDB     MariaDBConfig   `mapstructure:"mariadb"`
}

type GeneralConfig struct {
	ClientCode string       `mapstructure:"client_code"`
	AppName    string       `mapstructure:"app_name"`
	Version    string       `mapstructure:"version"`
	Author     string       `mapstructure:"author"`
	Locale     LocaleConfig `mapstructure:"locale"`
}

type LocaleConfig struct {
	Timezone   string `mapstructure:"timezone"`
	DateFormat string `mapstructure:"date_format"`
	TimeFormat string `mapstructure:"time_format"`
}

type LogConfig struct {
	Level    string    `mapstructure:"level"`
	Format   string    `mapstructure:"format"`
	Timezone string    `mapstructure:"timezone"`
	Output   LogOutput `mapstructure:"output"`
}

type LogOutput struct {
	Console LogConsoleOutput `mapstructure:"console"`
	File    LogFileOutput    `mapstructure:"file"`
	Syslog  LogSyslogOutput  `mapstructure:"syslog"`
}

type LogConsoleOutput struct {
	Enabled bool `mapstructure:"enabled"`
}

type LogFileOutput struct {
	Enabled         bool            `mapstructure:"enabled"`
	Dir             string          `mapstructure:"dir"`
	FilenamePattern string          `mapstructure:"filename_pattern"`
	Rotation        LogFileRotation `mapstructure:"rotation"`
}

type LogFileRotation struct {
	Daily         bool   `mapstructure:"daily"`
	MaxSize       string `mapstructure:"max_size"`
	RetentionDays int    `mapstructure:"retention_days"`
	CompressOld   bool   `mapstructure:"compress_old"`
}

type LogSyslogOutput struct {
	Enabled  bool   `mapstructure:"enabled"`
	Facility string `mapstructure:"facility"`
	Tag      string `mapstructure:"tag"`
}

type MysqldumpConfig struct {
	Args string `mapstructure:"args"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type BackupConfig struct {
	MysqldumpArgs string             `mapstructure:"mysqldump_args"`
	Retention     BackupRetention    `mapstructure:"retention"`
	Compression   BackupCompression  `mapstructure:"compression"`
	Security      BackupSecurity     `mapstructure:"security"`
	Storage       BackupStorage      `mapstructure:"storage"`
	Verification  BackupVerification `mapstructure:"verification"`
}

type BackupRetention struct {
	Days            int    `mapstructure:"days"`
	CleanupEnabled  bool   `mapstructure:"cleanup_enabled"`
	CleanupSchedule string `mapstructure:"cleanup_schedule"`
}

type BackupCompression struct {
	Required  bool   `mapstructure:"required"`
	Algorithm string `mapstructure:"algorithm"`
	Level     string `mapstructure:"level"`
}

type BackupSecurity struct {
	EncryptionRequired   bool `mapstructure:"encryption_required"`
	ChecksumVerification bool `mapstructure:"checksum_verification"`
	IntegrityCheck       bool `mapstructure:"integrity_check"`
}

type BackupStorage struct {
	BaseDirectory string                 `mapstructure:"base_directory"`
	Structure     BackupStorageStructure `mapstructure:"structure"`
	Naming        BackupStorageNaming    `mapstructure:"naming"`
	TempDirectory string                 `mapstructure:"temp_directory"`
	CleanupTemp   bool                   `mapstructure:"cleanup_temp"`
}

type BackupStorageStructure struct {
	Pattern       string `mapstructure:"pattern"`
	CreateSubdirs bool   `mapstructure:"create_subdirs"`
}

type BackupStorageNaming struct {
	Pattern           string `mapstructure:"pattern"`
	IncludeHostname   bool   `mapstructure:"include_hostname"`
	IncludeClientCode bool   `mapstructure:"include_client_code"`
}

type BackupVerification struct {
	DiskSpaceCheck   bool   `mapstructure:"disk_space_check"`
	MinimumFreeSpace string `mapstructure:"minimum_free_space"`
	VerifyAfterWrite bool   `mapstructure:"verify_after_write"`
	CompareChecksums bool   `mapstructure:"compare_checksums"`
}

type SystemUsers struct {
	Users []string `mapstructure:"users"`
}

type ConfigDirConfig struct {
	DatabaseConfig        string `mapstructure:"database_config"`
	MariaDBConfigTemplate string `mapstructure:"mariadb_config_templates"`
	MariaDBKey            string `mapstructure:"mariadb_key"`
	DatabaseList          string `mapstructure:"database_list"`
}

type MariaDBConfig struct {
	Version             string `mapstructure:"version"`
	DataDir             string `mapstructure:"data_dir"`
	LogDir              string `mapstructure:"log_dir"`
	BinlogDir           string `mapstructure:"binlog_dir"`
	Port                int    `mapstructure:"port"`
	SocketPath          string `mapstructure:"socket_path"`
	InnodbEncryptTables bool   `mapstructure:"innodb_encrypt_tables"`
	EncryptionKeyFile   string `mapstructure:"encryption_key_file"`
	ConfigDir           string `mapstructure:"config_dir"`
	ServerID            int    `mapstructure:"server_id"`
}
