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
		fmt.Printf("- %s%s: %s -> %s\n", migration.Type, criticalText, migration.Source, migration.Destination)
	}

	fmt.Println()
	terminal.PrintWarning("This process will stop MariaDB service temporarily")
}
