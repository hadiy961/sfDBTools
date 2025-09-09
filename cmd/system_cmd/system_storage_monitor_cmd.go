package system_cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sort"

	"github.com/spf13/cobra"
)

// SystemStorageMonitorCmd monitors sizes of subdirectories under data_dir in config
var SystemStorageMonitorCmd = &cobra.Command{
	Use:   "storage-monitor",
	Short: "Monitor storage usage per-database under configured data_dir",
	Long:  "Load data_dir from configuration and monitor size of each immediate subdirectory (databases) every interval seconds.",
	Run: func(cmd *cobra.Command, args []string) {
		lg, _ := logger.Get()

		// load config
		cfg, err := config.Get()
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			os.Exit(1)
		}

		dataDir := cfg.MariaDB.DataDir
		if dataDir == "" {
			fmt.Println("data_dir not configured in config.yaml")
			os.Exit(1)
		}

		interval, _ := cmd.Flags().GetInt("interval")
		topN, _ := cmd.Flags().GetInt("top")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		prev := map[string]int64{}

		fmt.Printf("Monitoring storage in %s every %d second(s). Press Ctrl+C to stop.\n", dataDir, interval)

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-sig:
				fmt.Println("stopping storage monitor")
				return
			case <-ticker.C:
				sizes, err := computeImmediateSubdirSizes(dataDir)
				if err != nil {
					fmt.Printf("error computing sizes: %v\n", err)
					continue
				}

				// prepare sortable list
				type entry struct {
					name  string
					size  int64
					delta int64
				}
				var list []entry
				for name, size := range sizes {
					delta := int64(0)
					if p, ok := prev[name]; ok {
						delta = size - p
					}
					list = append(list, entry{name: name, size: size, delta: delta})
				}

				// sort by size desc
				sort.Slice(list, func(i, j int) bool { return list[i].size > list[j].size })

				displayList := list
				if topN > 0 && topN < len(list) {
					displayList = list[:topN]
				}

				ts := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("\n[%s] Top %d databases by size under %s:\n", ts, len(list), dataDir)
				fmt.Printf("%-30s %-12s %-12s\n", "Database", "Size", "Growth/s")
				for _, e := range displayList {
					growthPerSec := e.delta / int64(interval)
					fmt.Printf("%-30s %-12s %-12s\n", e.name,
						common.FormatSizeWithPrecision(e.size, 2),
						common.FormatSizeWithPrecision(growthPerSec, 2))
				}

				// update prev for all seen directories (not only displayed)
				prev = map[string]int64{}
				for _, e := range list {
					prev[e.name] = e.size
				}
				_ = lg
			}
		}
	},
}

func init() {
	SystemStorageMonitorCmd.Flags().Int("interval", 1, "Polling interval in seconds")
	SystemStorageMonitorCmd.Flags().Int("top", 10, "Show top N directories (0 = all)")
}

// computeImmediateSubdirSizes returns sizes (in bytes) for immediate subdirectories of path
func computeImmediateSubdirSizes(path string) (map[string]int64, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64)
	var wg sync.WaitGroup
	mu := sync.Mutex{}

	for _, de := range entries {
		if !de.IsDir() {
			continue
		}
		name := de.Name()
		sub := filepath.Join(path, name)
		wg.Add(1)
		go func(n, p string) {
			defer wg.Done()
			var size int64
			filepath.WalkDir(p, func(_ string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					return nil
				}
				fi, err := d.Info()
				if err != nil {
					return nil
				}
				mu.Lock()
				size += fi.Size()
				mu.Unlock()
				return nil
			})
			mu.Lock()
			out[n] = size
			mu.Unlock()
		}(name, sub)
	}
	wg.Wait()
	return out, nil
}
