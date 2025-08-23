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
	ClientCode string `mapstructure:"client_code"`
	AppName    string `mapstructure:"app_name"`
	Version    string `mapstructure:"version"`
	Author     string `mapstructure:"author"`
}

type LogConfig struct {
	Level    string         `mapstructure:"level"`
	Format   string         `mapstructure:"format"`
	Timezone string         `mapstructure:"timezone"`
	Output   LogOutput      `mapstructure:"output"`
	File     LogFileSetting `mapstructure:"file"`
}

type LogOutput struct {
	Console bool `mapstructure:"console"`
	File    bool `mapstructure:"file"`
	Syslog  bool `mapstructure:"syslog"`
}

type LogFileSetting struct {
	Dir           string `mapstructure:"dir"`
	RotateDaily   bool   `mapstructure:"rotate_daily"`
	RetentionDays int    `mapstructure:"retention_days"`
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
	OutputDir         string `mapstructure:"output_dir"`
	Compress          bool   `mapstructure:"compress"`
	Compression       string `mapstructure:"compression"`
	CompressionLevel  string `mapstructure:"compression_level"`
	IncludeData       bool   `mapstructure:"include_data"`
	Encrypt           bool   `mapstructure:"encrypt"`
	VerifyDisk        bool   `mapstructure:"verify_disk"`
	RetentionDays     int    `mapstructure:"retention_days"`
	CalculateChecksum bool   `mapstructure:"calculate_checksum"`
	IncludeDmart      bool   `mapstructure:"include_dmart"`
	IncludeArchive    bool   `mapstructure:"include_archive"`
	IncludeTemp       bool   `mapstructure:"include_temp"`
	Progress          bool   `mapstructure:"progress"`
	SystemUser        bool   `mapstructure:"system_user"`
}

type SystemUsers struct {
	Users []string `mapstructure:"users"`
}

type ConfigDirConfig struct {
	DatabaseConfig string `mapstructure:"database_config"`
	MariaDBConfig  string `mapstructure:"mariadb_config"`
	MariaDBKey     string `mapstructure:"mariadb_key"`
	DatabaseList   string `mapstructure:"database_list"`
}

type MariaDBConfig struct {
	DefaultVersion string               `mapstructure:"default_version"`
	Installation   MariaDBInstallConfig `mapstructure:"installation"`
}

type MariaDBInstallConfig struct {
	BaseDir             string `mapstructure:"base_dir"`
	DataDir             string `mapstructure:"data_dir"`
	LogDir              string `mapstructure:"log_dir"`
	BinlogDir           string `mapstructure:"binlog_dir"`
	Port                int    `mapstructure:"port"`
	KeyFile             string `mapstructure:"key_file"`
	SeparateDirectories bool   `mapstructure:"separate_directories"`
}
