package command_mariadb

import (
	"encoding/json"
	"fmt"
	"os"

	"sfDBTools/internal/core/health"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/common"
	"sfDBTools/utils/database"

	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Comprehensive health check for MariaDB/MySQL server",
	Long: `This command performs a comprehensive health check of your MariaDB/MySQL server, including:
- Service status and connection validation
- Database version and configuration
- GTID and replication status
- System resource usage (CPU, memory, disk)
- Log files and paths validation
- Database and user statistics

The command supports different output formats:
- Standard: Summary table format (default)
- Full report: Detailed breakdown with sections
- JSON: Machine-readable output for automation`,
	Example: `# Standard health check report (uses config.yaml)
sfDBTools mariadb check

# Full detailed report
sfDBTools mariadb check --full-report

# JSON output for automation
sfDBTools mariadb check --json`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.Get()
		if err != nil {
			lg.Error("Failed to initialize logger", logger.Error(err))
			os.Exit(1)
		}

		lg.Info("Starting MariaDB health check")

		// Parse and validate flags
		config, err := parseHealthCheckFlags(cmd)
		if err != nil {
			lg.Error("Command failed", logger.Error(err))
			os.Exit(1)
		}

		// Execute health check
		if err := executeHealthCheck(config, cmd); err != nil {
			lg.Error("Command failed", logger.Error(err))
			os.Exit(1)
		}

		lg.Info("MariaDB health check completed")
	},
}

func init() {
	// Output format flags
	CheckCmd.Flags().Bool("full-report", false, "Generate detailed full report")
	CheckCmd.Flags().Bool("json", false, "Output in JSON format")
}

// HealthCheckConfig holds the configuration for health check
type HealthCheckConfig struct {
	DBConfig   database.Config `json:"db_config"`
	FullReport bool            `json:"full_report"`
	JSONOutput bool            `json:"json_output"`
}

// parseHealthCheckFlags parses and validates command flags
func parseHealthCheckFlags(cmd *cobra.Command) (*HealthCheckConfig, error) {
	config := &HealthCheckConfig{}

	// Parse output format flags
	var err error
	config.FullReport, err = cmd.Flags().GetBool("full-report")
	if err != nil {
		return nil, fmt.Errorf("failed to get full-report flag: %w", err)
	}

	config.JSONOutput, err = cmd.Flags().GetBool("json")
	if err != nil {
		return nil, fmt.Errorf("failed to get json flag: %w", err)
	}

	// Load database configuration from config.yaml
	dbConfig, err := common.GetDatabaseConfigFromDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config from config.yaml: %w", err)
	}
	config.DBConfig = *dbConfig

	return config, nil
}

// executeHealthCheck performs the health check and displays results
func executeHealthCheck(config *HealthCheckConfig, cmd *cobra.Command) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	lg.Info("Starting MariaDB health check")

	// Collect health check information
	healthInfo, err := health.CollectHealthCheckInfo(config.DBConfig)
	if err != nil {
		lg.Error("Failed to collect health check information", logger.Error(err))
		return fmt.Errorf("health check failed: %w", err)
	}

	// Display results based on output format
	if config.JSONOutput {
		// Display JSON format
		healthJSON, err := json.Marshal(healthInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal health info to JSON: %w", err)
		}
		lg.Info(string(healthJSON))
	} else {
		// Display formatted text
		health.DisplayHealthCheckInfo(healthInfo)
	}

	lg.Info("MariaDB health check completed")

	return nil
}
