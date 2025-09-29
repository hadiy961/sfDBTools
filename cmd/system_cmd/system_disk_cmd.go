package system_cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common/format"
	"sfDBTools/utils/disk"
	"sfDBTools/utils/fs"

	"github.com/spf13/cobra"
)

// SystemDiskCmd checks disk space using system utilities
var SystemDiskCmd = &cobra.Command{
	Use:     "disk-check",
	Short:   "Periksa ruang disk pada path tertentu",
	Long:    "Periksa apakah tersedia ruang disk minimum (dalam MB) pada path yang diberikan.",
	Example: `sfDBTools system disk-check --path /var/backups --min-mb 1024`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, _ := logger.Get()

		path, _ := cmd.Flags().GetString("path")
		minMB, _ := cmd.Flags().GetInt64("min-mb")
		showDetails, _ := cmd.Flags().GetBool("details")

		if path == "" {
			path = string(os.PathSeparator)
		}

		if err := disk.CheckDiskSpace(path, minMB); err != nil {
			lg.Error("Disk check failed", logger.Error(err))
			fmt.Printf("Disk check failed: %v\n", err)
			os.Exit(1)
		}

		if showDetails {
			stats, _ := disk.GetUsageStatistics(path)
			fmt.Printf("Path: %s\nMount: %s (%s)\nFree: %s\nTotal: %s\nUsed: %s (%.1f%%)\n",
				stats.Path, stats.Mountpoint, stats.Fstype,
				format.FormatSizeWithPrecision(stats.Free, 2),
				format.FormatSizeWithPrecision(stats.Total, 2),
				format.FormatSizeWithPrecision(stats.Used, 2), stats.UsedPercent)
		} else {
			fmt.Printf("Disk check passed for %s (required %d MB)\n", path, minMB)
		}
	},
}

func init() {
	SystemDiskCmd.Flags().String("path", "", "Path to check (default root)")
	SystemDiskCmd.Flags().Int64("min-mb", 1024, "Minimum free space required in MB")
	SystemDiskCmd.Flags().Bool("details", false, "Show detailed usage information")
}

// SystemDiskMonitorCmd monitors disk usage and prints warning when threshold exceeded
var SystemDiskMonitorCmd = &cobra.Command{
	Use:   "disk-monitor",
	Short: "Monitor ruang disk dan beri peringatan jika melampaui threshold",
	Long:  "Monitor ruang disk secara periodik dan jalankan callback (stdout) jika persentase penggunaan melewati threshold.",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		intervalSec, _ := cmd.Flags().GetInt("interval")

		if path == "" {
			path = string(os.PathSeparator)
		}

		stop := disk.MonitorDisk(path, time.Duration(intervalSec)*time.Second, threshold, func(u *fs.DiskUsage) {
			fmt.Printf("[WARN] disk %s used %.1f%% (free %s)\n", u.Path, u.UsedPercent, format.FormatSizeWithPrecision(u.Free, 2))
		})

		fmt.Printf("Monitoring disk %s every %d seconds. Press CTRL+C to stop.\n", path, intervalSec)
		// Wait until interrupted
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		stop()
	},
}

func init() {
	SystemDiskMonitorCmd.Flags().String("path", "", "Path to monitor (default root)")
	SystemDiskMonitorCmd.Flags().Float64("threshold", 90.0, "Used percent threshold to trigger warning")
	SystemDiskMonitorCmd.Flags().Int("interval", 60, "Polling interval in seconds")
}
