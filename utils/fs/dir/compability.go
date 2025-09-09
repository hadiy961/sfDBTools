package dir

// CreateDir is a convenience wrapper to create a directory using the
// utils/dir package (backwards compatibility)
func CreateDir(path string) error {
	return Create(path)
}

// ValidateDir validates a directory path (backwards compatibility)
func ValidateDir(path string) error {
	return Validate(path)
}
