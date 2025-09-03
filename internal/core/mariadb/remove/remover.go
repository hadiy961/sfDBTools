package remove

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/repository"
	"sfDBTools/utils/system"
	"sfDBTools/utils/terminal"
)

// Config for remover
type Config struct {
	SkipConfirm bool
}

// RemoveResult contains outcome
type RemoveResult struct {
	Success   bool
	Message   string
	RemovedAt time.Time
}

// Remover performs MariaDB removal
type Remover struct {
	cfg        *Config
	pkgManager system.PackageManager
	svcManager system.ServiceManager
	repoMgr    *repository.Manager
}

// NewRemover constructs a Remover
func NewRemover(cfg *Config) (*Remover, error) {
	if cfg == nil {
		cfg = &Config{SkipConfirm: false}
	}

	// Use existing helpers to detect OS
	osDetector := common.NewOSDetector()
	osInfo, err := osDetector.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS: %w", err)
	}

	r := &Remover{
		cfg:        cfg,
		pkgManager: system.NewPackageManager(),
		svcManager: system.NewServiceManager(),
		repoMgr:    repository.NewManager(osInfo),
	}

	return r, nil
}

// Remove performs destructive removal following safe prompts
func (r *Remover) Remove() (*RemoveResult, error) {
	lg, _ := logger.Get()

	terminal.ClearAndShowHeader("MariaDB Remove")
	terminal.PrintWarning("⚠️  This will remove MariaDB packages, data directories and configuration. This action is irreversible.")

	if !r.cfg.SkipConfirm {
		var confirm string
		fmt.Print("Are you sure you want to continue? (y/n): ")
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			terminal.PrintInfo("Operation cancelled by user")
			return &RemoveResult{Success: false, Message: "cancelled by user"}, nil
		}
	}

	// Step 1: Stop and disable service
	terminal.PrintInfo("Stopping and disabling MariaDB service...")
	_ = r.svcManager.Stop("mariadb")
	_ = r.svcManager.Stop("mysql")
	_ = r.svcManager.Disable("mariadb")
	_ = r.svcManager.Disable("mysql")

	// Step 2: Remove packages
	terminal.PrintInfo("Removing MariaDB packages...")
	packages := r.getPackagesToRemove()
	if len(packages) > 0 {
		if err := r.pkgManager.Remove(packages); err != nil {
			lg.Warn("Failed to remove packages", logger.Error(err))
			// continue to cleanup files even if package removal failed
		} else {
			terminal.PrintSuccess("Package removal completed")
		}
	}

	// Step 3: Remove common directories
	terminal.PrintInfo("Cleaning up data and configuration directories...")
	defaultDirs := []string{
		"/var/lib/mysql",
		"/etc/mysql",
		"/etc/my.cnf",
		"/etc/mysql/mariadb.conf.d",
		"/usr/lib/systemd/system/mariadb.service",
		"/var/log/mysql",
		"/var/run/mysqld",
	}

	for _, p := range defaultDirs {
		if exists(p) {
			terminal.PrintInfo(fmt.Sprintf("Removing %s", p))
			_ = os.RemoveAll(p)
		}
	}

	// Step 4: Detect custom my.cnf files and offer removal
	terminal.PrintInfo("Searching for custom my.cnf files (this may take a while)...")
	custom := findCustomMyCnf()
	if len(custom) > 0 {
		terminal.PrintWarning("Custom my.cnf files found:")
		for _, f := range custom {
			terminal.PrintInfo(fmt.Sprintf(" - %s", f))
		}

		if r.cfg.SkipConfirm {
			// remove corresponding datadirs automatically in skip-confirm mode
			for _, f := range custom {
				datadir := parseDataDirFromMyCnf(f)
				if datadir != "" && exists(datadir) {
					terminal.PrintInfo(fmt.Sprintf("Removing custom data dir %s", datadir))
					_ = os.RemoveAll(datadir)
				}
			}
		} else {
			var ans string
			fmt.Print("Remove discovered custom data dirs? (y/n): ")
			fmt.Scanln(&ans)
			if ans == "y" || ans == "Y" {
				for _, f := range custom {
					datadir := parseDataDirFromMyCnf(f)
					if datadir != "" && exists(datadir) {
						terminal.PrintInfo(fmt.Sprintf("Removing custom data dir %s", datadir))
						_ = os.RemoveAll(datadir)
					}
				}
			} else {
				terminal.PrintInfo("Skipping removal of custom data dirs")
			}
		}
	}

	// Try to clean repository entries
	_ = r.repoMgr.Clean()

	terminal.PrintSuccess("MariaDB cleanup finished")

	lg.Info("MariaDB removal completed")
	return &RemoveResult{Success: true, Message: "completed", RemovedAt: time.Now()}, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (r *Remover) getPackagesToRemove() []string {
	// try to remove common package names; PackageManager will ignore missing ones
	// Use OS detector to determine package type
	osDetector := common.NewOSDetector()
	osInfo, err := osDetector.DetectOS()
	if err != nil {
		// fallback to generic names
		return []string{"mariadb-server", "mariadb-client", "mariadb"}
	}

	switch osInfo.PackageType {
	case "deb":
		return []string{"^mariadb.*", "^mysql.*"}
	case "rpm":
		return []string{"mariadb-server", "mariadb-client", "mariadb"}
	default:
		return []string{"mariadb-server", "mariadb-client", "mariadb"}
	}
}

// findCustomMyCnf finds my.cnf files outside /etc/my.cnf
func findCustomMyCnf() []string {
	var results []string
	// search common locations quickly
	candidates := []string{"/etc/mysql/my.cnf", "/usr/local/etc/my.cnf", "/opt/my.cnf"}
	for _, c := range candidates {
		if exists(c) {
			results = append(results, c)
		}
	}

	// scan filesystem root for my.cnf excluding /etc to avoid duplicates
	filepath.Walk("/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "my.cnf" && path != "/etc/my.cnf" {
			results = append(results, path)
		}
		return nil
	})

	// deduplicate
	seen := map[string]struct{}{}
	uniq := []string{}
	for _, r := range results {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			uniq = append(uniq, r)
		}
	}
	return uniq
}

// parseDataDirFromMyCnf extracts datadir from my.cnf if present
func parseDataDirFromMyCnf(path string) string {
	// simple parse using scanner
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		// lower-case for case-insensitive check
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "datadir") {
			// split on '='
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				// remove surrounding quotes if any
				value = strings.Trim(value, "\"'")
				return filepath.Clean(value)
			}
		}
	}
	return ""
}
