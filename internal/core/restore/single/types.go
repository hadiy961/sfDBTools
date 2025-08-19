package single

// Options holds restore configuration
type Options struct {
	Host           string
	Port           int
	User           string
	Password       string
	DBName         string
	File           string
	VerifyChecksum bool
}
