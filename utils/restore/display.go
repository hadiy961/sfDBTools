package restore_utils

import (
	"bufio"
	"fmt"
	"os"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/terminal"
	"strings"
	"time"
)

// DisplayRestoreParameters shows restore parameters before execution
func DisplayRestoreParameters(options RestoreOptions) {
	terminal.Headers("Restore Tools - Restore Single Database")
	terminal.PrintSubHeader("Restore Parameters")
	fmt.Printf("Target Database:  %s\n", options.DBName)
	fmt.Printf("Target Host:      %s:%d\n", options.Host, options.Port)
	fmt.Printf("Target User:      %s\n", options.User)
	fmt.Printf("Backup File:      %s\n", options.File)
	fmt.Printf("Verify Checksum:  %t\n", options.VerifyChecksum)
	terminal.PrintSeparator()
}

// DisplayRestoreResults shows restore results after completion with customizable title
func DisplayRestoreResults(success bool, duration time.Duration, title string) {
	if title == "" {
		title = "Restore Result"
	}
	fmt.Printf("\n=== %s ===\n", title)
	if success {
		fmt.Println("✅ Restore completed successfully!")
	} else {
		fmt.Println("❌ Restore failed!")
	}
	fmt.Printf("Duration: %s\n", common.FormatDuration(duration, "words"))
	fmt.Printf("=== End %s ===\n", title)
}

// DisplayRestoreOverview shows restore overview before execution
func DisplayRestoreOverview(options RestoreOptions, filePath string) {
	fmt.Println("\n=== Restore Overview ===")
	fmt.Printf("Source File:      %s\n", filePath)
	fmt.Printf("Target Database:  %s\n", options.DBName)
	fmt.Printf("Target Host:      %s:%d\n", options.Host, options.Port)
	fmt.Printf("Target User:      %s\n", options.User)
	fmt.Println("========================")
}

// DisplayRestoreSummary shows restore summary after completion
func DisplayRestoreSummary(options RestoreOptions, duration time.Duration, lg *logger.Logger) {
	fmt.Println("\n=== Restore Summary ===")
	fmt.Printf("Database:         %s\n", options.DBName)
	fmt.Printf("Host:             %s:%d\n", options.Host, options.Port)
	fmt.Printf("User:             %s\n", options.User)
	fmt.Printf("Source File:      %s\n", options.File)
	fmt.Printf("Duration:         %s\n", common.FormatDuration(duration, "words"))
	fmt.Printf("Verify Checksum:  %t\n", options.VerifyChecksum)
	fmt.Println("=======================")

	if lg != nil {
		lg.Info("Restore summary",
			logger.String("database", options.DBName),
			logger.String("host", fmt.Sprintf("%s:%d", options.Host, options.Port)),
			logger.String("user", options.User),
			logger.String("file", options.File),
			logger.String("duration", duration.String()),
			logger.Bool("verify_checksum", options.VerifyChecksum))
	}
}

// LogRestoreCompletion logs the successful completion of restore
func LogRestoreCompletion(options RestoreOptions, duration time.Duration, lg *logger.Logger) {
	if lg != nil {
		lg.Info("Restore completed successfully",
			logger.String("database", options.DBName),
			logger.String("source_file", options.File),
			logger.String("host", fmt.Sprintf("%s:%d", options.Host, options.Port)),
			logger.String("duration", duration.String()))
	}
}

// PromptRestoreConfirmation prompts user for confirmation before performing restore
func PromptRestoreConfirmation(options RestoreOptions) error {
	reader := bufio.NewReader(os.Stdin)
	if options.DBName == "" {
		options.DBName = "All Database"
	}
	terminal.PrintSubHeader("RESTORE CONFIRMATION")
	fmt.Printf("You are about to restore backup to:\n")
	fmt.Printf("  Database: %s\n", options.DBName)
	fmt.Printf("  Host:     %s:%d\n", options.Host, options.Port)
	fmt.Printf("  User:     %s\n", options.User)
	fmt.Printf("  File:     %s\n", options.File)
	fmt.Println("\n🚨 WARNING: This will overwrite existing data in the target database!")

	fmt.Print("\nDo you want to continue with the restore? [y/N]: ")
	confirmInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirm := strings.ToLower(strings.TrimSpace(confirmInput))
	if confirm != "y" && confirm != "yes" {
		return fmt.Errorf("restore operation cancelled by user")
	}

	terminal.PrintSubHeader("Proceeding with restore...")
	return nil
}

// DisplayRestoreUserParameters shows user grants restore parameters before execution
func DisplayRestoreUserParameters(options RestoreUserOptions) {
	fmt.Println("\n=== User Grants Restore Parameters ===")
	fmt.Printf("Target Host:      %s:%d\n", options.Host, options.Port)
	fmt.Printf("Target User:      %s\n", options.User)
	fmt.Printf("Grants File:      %s\n", options.File)
	fmt.Printf("Verify Checksum:  %t\n", options.VerifyChecksum)
	fmt.Println("=======================================")
}

// PromptRestoreUserConfirmation prompts user for confirmation before performing user grants restore
func PromptRestoreUserConfirmation(options RestoreUserOptions) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n⚠️  USER GRANTS RESTORE CONFIRMATION")
	fmt.Println("====================================")
	fmt.Printf("You are about to restore user grants to:\n")
	fmt.Printf("  Host:     %s:%d\n", options.Host, options.Port)
	fmt.Printf("  User:     %s\n", options.User)
	fmt.Printf("  File:     %s\n", options.File)
	fmt.Println("\n🚨 WARNING: This will execute GRANT statements on the target database server!")
	fmt.Println("🚨 WARNING: This may create new users or modify existing user privileges!")

	fmt.Print("\nDo you want to continue with the user grants restore? [y/N]: ")
	confirmInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirm := strings.ToLower(strings.TrimSpace(confirmInput))
	if confirm != "y" && confirm != "yes" {
		return fmt.Errorf("user grants restore operation cancelled by user")
	}

	fmt.Println("✅ Proceeding with user grants restore...")
	return nil
}
