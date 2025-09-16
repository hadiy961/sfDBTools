package migration

import (
	"fmt"
	"sfDBTools/utils/terminal"
)

func ShowMigrationPlan(migrations []DataMigration) {
	terminal.PrintInfo("Data Migration Plan:")
	terminal.PrintInfo("====================")

	for _, migration := range migrations {
		criticalText := ""
		if migration.Critical {
			criticalText = " (CRITICAL)"
		}
		var label string
		switch migration.Type {
		case "data":
			label = "Data Dir"
		case "logs":
			label = "Logs Dir"
		case "binlogs":
			label = "Binlog Dir"
		default:
			label = migration.Type
		}
		fmt.Printf("- %s%s: %s -> %s\n", label, criticalText, migration.Source, migration.Destination)
	}

	fmt.Println()
	terminal.PrintWarning("This process will stop MariaDB service temporarily")
}
