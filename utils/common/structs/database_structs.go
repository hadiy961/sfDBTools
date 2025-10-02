package structs

// DatabaseInfoMeta represents database information in metadata
type DatabaseInfoMeta struct {
	SizeBytes    int64   `json:"size_bytes"`
	SizeMB       float64 `json:"size_mb"`
	TableCount   int     `json:"table_count"`
	ViewCount    int     `json:"view_count"`
	RoutineCount int     `json:"routine_count"`
	TriggerCount int     `json:"trigger_count"`
	UserCount    int     `json:"user_count"`
}
