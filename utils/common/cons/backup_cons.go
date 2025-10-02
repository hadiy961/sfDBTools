package cons

// ConfigurationSource represents the source of database configuration
type ConfigurationSource int

const (
	SourceConfigFile ConfigurationSource = iota
	SourceFlags
	SourceDefaults
	SourceInteractive
)
