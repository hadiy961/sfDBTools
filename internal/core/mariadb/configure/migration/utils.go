package migration

// IsLogFile determines if a file is a log-related file
func (m *MigrationManager) IsLogFile(path string) bool {
	return m.fsMgr.Pattern().IsLogFile(path)
}

// IsDataDirectory determines if a directory contains database data that should be skipped during log migration
func (m *MigrationManager) IsDataDirectory(path, sourceRoot string) bool {
	return m.fsMgr.Pattern().IsDataDirectory(path, sourceRoot)
}
