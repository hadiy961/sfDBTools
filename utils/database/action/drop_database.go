package database

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
	"sfDBTools/utils/database/info"
)

// DropMode menentukan mode eksekusi
type DropMode string

const (
	DropModeAll   DropMode = "all"
	DropModeList  DropMode = "list"
	DropModeExact DropMode = "exact"
)

type DropDatabasesOptions struct {
	Host        string
	Port        int
	User        string
	Password    string
	Mode        DropMode
	TargetList  []string
	Exclude     []string
	DryRun      bool
	Force       bool
	SkipConfirm bool
}

type DropDatabasesResult struct {
	TotalFound        int
	TargetsPlanned    []string
	Dropped           []string
	Skipped           []string
	Errors            map[string]error
	SystemDatabases   []string
	ExcludedDatabases []string
	DryRun            bool
	Mode              DropMode
	SafeSkipped       []string // sistem DB yang otomatis di-skip
}

var systemDatabases = []string{"mysql", "information_schema", "performance_schema", "sys"}

func isSystemDB(name string) bool {
	ln := strings.ToLower(name)
	for _, s := range systemDatabases {
		if ln == s {
			return true
		}
	}
	return false
}

func sanitizeDBName(name string) string {
	return strings.ReplaceAll(name, "`", "``")
}

func DropDatabases(opts DropDatabasesOptions) (*DropDatabasesResult, error) {
	lg, _ := logger.Get()

	res := &DropDatabasesResult{
		Errors:          map[string]error{},
		DryRun:          opts.DryRun,
		Mode:            opts.Mode,
		SystemDatabases: systemDatabases,
	}

	cfg := database.Config{
		Host:     opts.Host,
		Port:     opts.Port,
		User:     opts.User,
		Password: opts.Password,
	}

	db, err := database.GetWithoutDB(cfg)
	if err != nil {
		return res, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer db.Close()

	all, err := info.ListDatabases(cfg)
	if err != nil {
		return res, fmt.Errorf("failed to list databases: %w", err)
	}
	res.TotalFound = len(all)

	excludeMap := map[string]struct{}{}
	for _, e := range opts.Exclude {
		excludeMap[strings.ToLower(e)] = struct{}{}
	}

	var planned []string
	switch opts.Mode {
	case DropModeAll:
		planned = all
	case DropModeList, DropModeExact:
		planned = opts.TargetList
	default:
		return res, fmt.Errorf("unknown drop mode: %s", opts.Mode)
	}

	for _, name := range planned {
		ln := strings.ToLower(name)

		// ALWAYS skip system DB (safety)
		if isSystemDB(name) {
			res.SafeSkipped = append(res.SafeSkipped, name)
			if !containsCaseInsensitive(res.Skipped, name) {
				res.Skipped = append(res.Skipped, name)
			}
			continue
		}

		if _, ok := excludeMap[ln]; ok {
			res.Skipped = append(res.Skipped, name)
			res.ExcludedDatabases = append(res.ExcludedDatabases, name)
			continue
		}
		res.TargetsPlanned = append(res.TargetsPlanned, name)
	}

	if len(res.TargetsPlanned) == 0 {
		lg.Warn("No (non-system) databases to drop after filtering")
		return res, nil
	}

	lg.Info("Planned databases for drop",
		logger.Int("count", len(res.TargetsPlanned)),
		logger.Strings("targets", res.TargetsPlanned),
		logger.Bool("dry_run", opts.DryRun))

	if len(res.SafeSkipped) > 0 {
		lg.Info("System databases automatically protected",
			logger.Strings("system_skipped", res.SafeSkipped))
	}

	if opts.DryRun {
		lg.Warn("Dry run enabled - no databases will be dropped")
		return res, nil
	}

	if !opts.SkipConfirm {
		if err := multiLevelConfirmation(os.Stdin, os.Stdout, res.TargetsPlanned); err != nil {
			return res, err
		}
	}

	for _, dbName := range res.TargetsPlanned {
		if err := dropOne(db, dbName); err != nil {
			res.Errors[dbName] = err
			res.Skipped = append(res.Skipped, dbName)
			lg.Error("Failed to drop database", logger.String("database", dbName), logger.Error(err))
			if !opts.Force {
				return res, fmt.Errorf("failed dropping %s: %w", dbName, err)
			}
			continue
		}
		res.Dropped = append(res.Dropped, dbName)
		lg.Info("Dropped database", logger.String("database", dbName))
	}

	return res, nil
}

func dropOne(db *sql.DB, dbName string) error {
	q := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", sanitizeDBName(dbName))
	if _, err := db.Exec(q); err != nil {
		return fmt.Errorf("DROP DATABASE failed: %w", err)
	}
	return nil
}

func multiLevelConfirmation(in *os.File, out *os.File, targets []string) error {
	reader := bufio.NewReader(in)
	fmt.Fprintf(out, "\n=== DROP DATABASE CONFIRMATION ===\n")
	fmt.Fprintf(out, "Targets (%d): %s\n", len(targets), strings.Join(targets, ", "))
	fmt.Fprintf(out, "Step 1: Type 'yes' to continue: ")
	first, _ := readLine(reader)
	if !strings.EqualFold(first, "yes") {
		return errors.New("aborted (step 1)")
	}

	fmt.Fprintf(out, "Step 2: Type the number of databases (%d): ", len(targets))
	second, _ := readLine(reader)
	if second != fmt.Sprintf("%d", len(targets)) {
		return errors.New("aborted (step 2)")
	}

	phrase := fmt.Sprintf("DROP %d DB", len(targets))
	fmt.Fprintf(out, "Step 3: Type phrase exactly: %q: ", phrase)
	third, _ := readLine(reader)
	if third != phrase {
		return errors.New("aborted (step 3)")
	}

	fmt.Fprintln(out, "Confirmation complete. Proceeding...")
	return nil
}

func readLine(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, os.ErrClosed) && !strings.Contains(err.Error(), "EOF") {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func containsCaseInsensitive(list []string, v string) bool {
	lv := strings.ToLower(v)
	for _, item := range list {
		if strings.ToLower(item) == lv {
			return true
		}
	}
	return false
}
