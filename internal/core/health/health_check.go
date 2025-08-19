package health

import (
	"fmt"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
	health_db "sfDBTools/utils/database/health"
)

// HealthCheckInfo represents complete health check information
type HealthCheckInfo struct {
	ServiceInfo         *health_db.ServiceInfo         `json:"service_info"`
	ConnectionInfo      *health_db.ConnectionInfo      `json:"connection_info"`
	CoreConfig          *health_db.CoreConfig          `json:"core_config"`
	LogsConfig          *health_db.LogsConfig          `json:"logs_config"`
	DatabasesInfo       *health_db.DatabasesInfo       `json:"databases_info"`
	ReplicationInfo     *health_db.ReplicationInfo     `json:"replication_info"`
	SystemResourcesInfo *health_db.SystemResourcesInfo `json:"system_resources_info"`
	Status              string                         `json:"status"`
	TotalWarnings       int                            `json:"total_warnings"`
	TotalErrors         int                            `json:"total_errors"`
}

// CollectHealthCheckInfo collects all health check information
func CollectHealthCheckInfo(config database.Config) (*HealthCheckInfo, error) {
	lg, _ := logger.Get()

	lg.Debug("Starting health check information collection")

	info := &HealthCheckInfo{}

	// Collect service information
	serviceInfo, err := health_db.GetServiceInfo()
	if err != nil {
		lg.Warn("Failed to collect service information", logger.Error(err))
	}
	info.ServiceInfo = serviceInfo

	// Collect connection information
	connectionInfo, err := health_db.GetConnectionInfo(config)
	if err != nil {
		lg.Warn("Failed to collect connection information", logger.Error(err))
	}
	info.ConnectionInfo = connectionInfo

	// Collect core configuration information
	coreConfig, err := health_db.GetCoreConfig(config)
	if err != nil {
		lg.Warn("Failed to collect core configuration", logger.Error(err))
	}
	info.CoreConfig = coreConfig

	// Collect logs configuration information
	logsConfig, err := health_db.GetLogsConfig(config)
	if err != nil {
		lg.Warn("Failed to collect logs configuration", logger.Error(err))
	}
	info.LogsConfig = logsConfig

	// Collect databases information
	databasesInfo, err := health_db.GetDatabasesInfo(config)
	if err != nil {
		lg.Warn("Failed to collect databases information", logger.Error(err))
	}
	info.DatabasesInfo = databasesInfo

	// Collect replication information
	replicationInfo, err := health_db.GetReplicationInfo(config)
	if err != nil {
		lg.Warn("Failed to collect replication information", logger.Error(err))
	}
	info.ReplicationInfo = replicationInfo

	// Collect system resources information
	systemResourcesInfo, err := health_db.GetSystemResourcesInfo(config)
	if err != nil {
		lg.Warn("Failed to collect system resources information", logger.Error(err))
	}
	info.SystemResourcesInfo = systemResourcesInfo

	// Determine overall status
	info.TotalErrors = 0
	info.TotalWarnings = 0

	if info.ServiceInfo != nil && !info.ServiceInfo.IsActive {
		info.TotalErrors++
	}

	if info.ConnectionInfo != nil && !info.ConnectionInfo.IsConnected {
		info.TotalErrors++
	}

	// Check for replication errors
	if info.ReplicationInfo != nil && info.ReplicationInfo.HasError {
		info.TotalErrors++
	}

	if info.TotalErrors > 0 {
		info.Status = "failed"
	} else if info.TotalWarnings > 0 {
		info.Status = "warning"
	} else {
		info.Status = "ok"
	}

	lg.Debug("Health check information collection completed")
	return info, nil
}

// DisplayHealthCheckInfo displays health check information in formatted text
func DisplayHealthCheckInfo(info *HealthCheckInfo) {
	lg, _ := logger.Get()

	lg.Info("--- MariaDB Health Check: Full Report ---")
	lg.Info("")
	lg.Info("[Section: Service & Connection Info]")

	// Display service information
	if info.ServiceInfo != nil {
		lg.Info(fmt.Sprintf("- Service Name: %s", info.ServiceInfo.ServiceName))
		lg.Info(fmt.Sprintf("- Status: %s", info.ServiceInfo.Status))
		lg.Info(fmt.Sprintf("- Uptime: %s", info.ServiceInfo.Uptime))
		lg.Info(fmt.Sprintf("- Process ID: %s", info.ServiceInfo.ProcessID))
	}

	// Display connection information
	if info.ConnectionInfo != nil {
		connectionDetails := health_db.FormatConnectionInfo(info.ConnectionInfo)
		for _, detail := range connectionDetails {
			lg.Info(detail)
		}
	}

	lg.Info("")
	lg.Info("[Section: Core Configuration]")

	// Display core configuration information
	if info.CoreConfig != nil {
		coreConfigDetails := health_db.FormatCoreConfig(info.CoreConfig)
		for _, detail := range coreConfigDetails {
			lg.Info(detail)
		}
	}

	lg.Info("")
	lg.Info("[Section: MariaDB Logs & Paths]")

	// Display logs configuration information
	if info.LogsConfig != nil {
		logsConfigDetails := health_db.FormatLogsConfig(info.LogsConfig)
		for _, detail := range logsConfigDetails {
			lg.Info(detail)
		}
	}

	lg.Info("")
	lg.Info("[Section: Databases]")

	// Display databases information
	if info.DatabasesInfo != nil {
		databasesDetails := health_db.FormatDatabasesInfo(info.DatabasesInfo)
		for _, detail := range databasesDetails {
			lg.Info(detail)
		}
	}

	lg.Info("")
	lg.Info("[Section: Replication]")

	// Display replication information
	if info.ReplicationInfo != nil {
		replicationDetails := health_db.FormatReplicationInfo(info.ReplicationInfo)
		for _, detail := range replicationDetails {
			lg.Info(detail)
		}
	}

	lg.Info("")
	lg.Info("[Section: System Resources]")

	// Display system resources information
	if info.SystemResourcesInfo != nil {
		systemResourcesDetails := health_db.FormatSystemResourcesInfo(info.SystemResourcesInfo)
		for _, detail := range systemResourcesDetails {
			lg.Info(detail)
		}
	}

	lg.Info("")
	lg.Info("---")
	lg.Info("Report Summary")
	lg.Info(fmt.Sprintf("Status: %s", info.Status))
	if info.TotalErrors > 0 || info.TotalWarnings > 0 {
		lg.Info(fmt.Sprintf("Issues found: %d errors, %d warnings", info.TotalErrors, info.TotalWarnings))
	} else {
		lg.Info("No issues found")
	}
}
