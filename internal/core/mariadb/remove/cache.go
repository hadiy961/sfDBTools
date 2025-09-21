package remove

// getDetectedConfig returns cached MariaDBConfig if present, otherwise detects once and caches it
func getDetectedConfig(deps *Dependencies) *MariaDBConfig {
	if deps == nil {
		return nil
	}
	if deps.DetectedConfig != nil {
		return deps.DetectedConfig
	}

	cfg, err := detectCustomDirectories()
	if err != nil {
		// Detection failed; leave nil
		return nil
	}
	deps.DetectedConfig = cfg
	return cfg
}
