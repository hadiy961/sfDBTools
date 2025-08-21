package command_database

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	dbConfig "sfDBTools/utils/database"
	dbAction "sfDBTools/utils/database/action"
	"sfDBTools/utils/database/info"

	"github.com/spf13/cobra"
)

var DatabaseDropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drop databases (all, single, list, or interactive selection) - system DBs always protected",
	Long: `Drop databases from a MySQL/MariaDB server (system databases TIDAK AKAN PERNAH dihapus: mysql, information_schema, performance_schema, sys).

Modes (mutually exclusive):
  --all                 : Drop all user databases (system DB otomatis di-skip)
  --source_db <name>    : Drop a single database
  --db_list <file>      : Drop databases dari file (satu per baris)
  (no mode flags)       : Interactive multi-select

Fitur:
  --exclude <db>        : Tambah pengecualian tambahan
  --dry-run             : Simulasi
  --force               : Lanjut walau sebagian gagal
  --yes                 : Skip konfirmasi bertingkat (tidak disarankan)

Contoh:
  sfDBTools database drop --config ./conf.cnf.enc --all
  sfDBTools database drop --config ./conf.cnf.enc --source_db appdb
  sfDBTools database drop --config ./conf.cnf.enc --db_list dblist.txt
  sfDBTools database drop --config ./conf.cnf.enc --exclude audit --exclude staging
  sfDBTools database drop --config ./conf.cnf.enc (interactive select)`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, _ := logger.Get()

		if err := executeDatabaseDrop(cmd); err != nil {
			lg.Error("Database drop failed", logger.Error(err))
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	backup_utils.AddCommonBackupFlags(DatabaseDropCmd)

	DatabaseDropCmd.Flags().Bool("all", false, "Drop all user databases (system DB selalu di-skip)")
	// `source_db` flag is provided by AddCommonBackupFlags; avoid redefining it here.
	DatabaseDropCmd.Flags().String("db_list", "", "Path to file containing database names to drop")
	DatabaseDropCmd.Flags().StringSlice("exclude", []string{}, "Database names to exclude from drop")
	DatabaseDropCmd.Flags().Bool("dry-run", false, "Simulate only, no actual drop")
	DatabaseDropCmd.Flags().Bool("force", false, "Continue dropping remaining databases even if one fails")
	DatabaseDropCmd.Flags().Bool("yes", false, "Skip all confirmations (DANGEROUS)")

	hideIrrelevantFlags(DatabaseDropCmd)
}

func hideIrrelevantFlags(cmd *cobra.Command) {
	toHide := []string{
		"output-dir", "compress", "compression", "compression-level",
		"encrypt", "data", "system-user", "retention-days",
		"verify-disk", "calculate-checksum",
	}
	for _, f := range toHide {
		_ = cmd.Flags().MarkHidden(f)
	}
}

func executeDatabaseDrop(cmd *cobra.Command) error {

	backupConfig, err := backup_utils.ResolveBackupConfigWithoutDB(cmd)
	if err != nil {
		return fmt.Errorf("failed to resolve configuration: %w", err)
	}

	allFlag, _ := cmd.Flags().GetBool("all")
	sourceDB, _ := cmd.Flags().GetString("source_db")
	dbListPath, _ := cmd.Flags().GetString("db_list")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	excludes, _ := cmd.Flags().GetStringSlice("exclude")

	used := 0
	if allFlag {
		used++
	}
	if sourceDB != "" {
		used++
	}
	if dbListPath != "" {
		used++
	}
	if used > 1 {
		return errors.New("cannot use --all, --source_db, and --db_list together")
	}

	dbCfg := dbConfig.Config{
		Host:     backupConfig.Host,
		Port:     backupConfig.Port,
		User:     backupConfig.User,
		Password: backupConfig.Password,
	}
	if err := backup_utils.TestDatabaseConnection(dbCfg); err != nil {
		return err
	}

	var mode dbAction.DropMode
	var targets []string

	switch {
	case allFlag:
		mode = dbAction.DropModeAll
	case sourceDB != "":
		mode = dbAction.DropModeExact
		targets = []string{sourceDB}
	case dbListPath != "":
		mode = dbAction.DropModeList
		listTargets, err := readDBListFile(dbListPath)
		if err != nil {
			return fmt.Errorf("failed reading db_list file: %w", err)
		}
		if len(listTargets) == 0 {
			return fmt.Errorf("db_list file is empty")
		}
		targets = listTargets
	default:
		mode = dbAction.DropModeList
		allNames, err := info.ListDatabases(dbCfg)
		if err != nil {
			return fmt.Errorf("failed listing databases: %w", err)
		}
		targets, err = interactiveSelectDatabases(allNames)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return fmt.Errorf("no databases selected - aborted")
		}
	}

	opts := dbAction.DropDatabasesOptions{
		Host:        backupConfig.Host,
		Port:        backupConfig.Port,
		User:        backupConfig.User,
		Password:    backupConfig.Password,
		Mode:        mode,
		TargetList:  targets,
		Exclude:     excludes,
		DryRun:      dryRun,
		Force:       force,
		SkipConfirm: skipConfirm,
	}

	res, err := dbAction.DropDatabases(opts)
	printDropSummary(res)
	return err
}

func readDBListFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func interactiveSelectDatabases(all []string) ([]string, error) {
	fmt.Println("=== Interactive Database Selection (system DB automatically skipped) ===")
	filtered := make([]string, 0, len(all))
	systemSet := map[string]struct{}{
		"mysql":              {},
		"information_schema": {},
		"performance_schema": {},
		"sys":                {},
	}
	for _, n := range all {
		if _, isSys := systemSet[strings.ToLower(n)]; isSys {
			continue
		}
		filtered = append(filtered, n)
	}
	if len(filtered) == 0 {
		return nil, errors.New("no user databases available")
	}
	for i, name := range filtered {
		fmt.Printf("[%d] %s\n", i+1, name)
	}
	fmt.Println("Enter:")
	fmt.Println("  *                -> select all above")
	fmt.Println("  comma numbers    -> e.g. 1,3,5")
	fmt.Println("  ranges           -> e.g. 2-4")
	fmt.Println("  names            -> e.g. db1,db2")
	fmt.Print("Selection: ")
	reader := bufio.NewReader(os.Stdin)
	raw, _ := reader.ReadString('\n')
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if raw == "*" {
		return filtered, nil
	}
	parts := strings.Split(raw, ",")
	selectedMap := map[string]struct{}{}
	var result []string

	add := func(name string) {
		if _, ok := selectedMap[name]; !ok {
			selectedMap[name] = struct{}{}
			result = append(result, name)
		}
	}

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			rp := strings.SplitN(p, "-", 2)
			if len(rp) == 2 {
				var sIdx, eIdx int
				if _, e1 := fmt.Sscanf(strings.TrimSpace(rp[0]), "%d", &sIdx); e1 == nil {
					if _, e2 := fmt.Sscanf(strings.TrimSpace(rp[1]), "%d", &eIdx); e2 == nil && sIdx > 0 && eIdx >= sIdx && eIdx <= len(filtered) {
						for i := sIdx; i <= eIdx; i++ {
							add(filtered[i-1])
						}
						continue
					}
				}
			}
		}
		var idx int
		if _, scanErr := fmt.Sscanf(p, "%d", &idx); scanErr == nil && idx > 0 && idx <= len(filtered) {
			add(filtered[idx-1])
			continue
		}
		found := false
		for _, candidate := range filtered {
			if candidate == p {
				add(candidate)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Warning: token '%s' not matched (ignored)\n", p)
		}
	}
	return result, nil
}

func printDropSummary(res *dbAction.DropDatabasesResult) {
	lg, _ := logger.Get()
	if res == nil {
		return
	}
	fmt.Println("====== DROP DATABASE SUMMARY ======")
	fmt.Printf("Mode             : %s\n", res.Mode)
	fmt.Printf("Dry Run          : %t\n", res.DryRun)
	fmt.Printf("Total Found      : %d\n", res.TotalFound)
	fmt.Printf("Planned Targets  : %d\n", len(res.TargetsPlanned))
	fmt.Printf("Dropped          : %d\n", len(res.Dropped))
	fmt.Printf("Skipped          : %d\n", len(res.Skipped))
	if len(res.SafeSkipped) > 0 {
		fmt.Printf("System Protected : %s\n", strings.Join(res.SafeSkipped, ", "))
	}
	if len(res.Dropped) > 0 {
		fmt.Printf("Dropped List     : %s\n", strings.Join(res.Dropped, ", "))
	}
	if len(res.Skipped) > 0 {
		fmt.Printf("Skipped List     : %s\n", strings.Join(res.Skipped, ", "))
	}
	if len(res.Errors) > 0 {
		fmt.Println("Errors:")
		for dbName, err := range res.Errors {
			fmt.Printf("  - %s: %v\n", dbName, err)
		}
	}
	fmt.Println("===================================")

	lg.Info("Drop summary",
		logger.Int("planned", len(res.TargetsPlanned)),
		logger.Int("dropped", len(res.Dropped)),
		logger.Int("skipped", len(res.Skipped)),
		logger.Strings("system_protected", res.SafeSkipped))
}
