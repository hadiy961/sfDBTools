package installcmd

import (
	"fmt"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"
)

// PrintInstallSummary prints a concise installation summary to stdout
func PrintInstallSummary(result *mariadb_utils.InstallResult) {
	if result == nil {
		return
	}

	status := "FAILED"
	if result.Success {
		status = "SUCCESS"
	}

	terminal.PrintHeader("MariaDB Install Summary")
	terminal.PrintInfo("Status:       " + status)
	terminal.PrintInfo("Version:      " + result.Version)
	terminal.PrintInfo("OS:           " + result.OperatingSystem)
	if result.Distribution != "" {
		terminal.PrintInfo("Distro:       " + result.Distribution)
	}
	terminal.PrintInfo(fmt.Sprintf("Port:         %d", result.Port))
	terminal.PrintInfo("Data dir:     " + result.DataDir)
	terminal.PrintInfo("Log dir:      " + result.LogDir)
	terminal.PrintInfo("Binlog dir:   " + result.BinlogDir)
	terminal.PrintInfo("Service:      " + result.ServiceStatus)
	if result.Duration != 0 {
		terminal.PrintInfo("Duration:     " + result.Duration.String())
	}
}
