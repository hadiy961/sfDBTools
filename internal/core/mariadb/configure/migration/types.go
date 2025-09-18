package migration

type DataMigration struct {
	Type        string
	Source      string
	Destination string
	Critical    bool
}
