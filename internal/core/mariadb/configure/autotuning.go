package configure

import (
	"context"
	"fmt"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	"sfDBTools/utils/system"
)

// PerformAutoTuning melakukan auto-tuning berdasarkan hardware sistem
// Sesuai dengan Step 12-14 dalam flow implementasi
func PerformAutoTuning(ctx context.Context, config *mariadb_config.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	if config == nil {
		return fmt.Errorf("nil config passed to PerformAutoTuning")
	}

	lg.Info("Starting hardware-based auto-tuning")

	// Step 14: Detect hardware info
	hardwareInfo, err := system.GetHardwareInfo()
	if err != nil {
		return fmt.Errorf("failed to get hardware information: %w", err)
	}

	lg.Info("Hardware detected",
		logger.Int("cpu_cores", hardwareInfo.CPUCores),
		logger.Int("total_ram_gb", hardwareInfo.TotalRAMGB),
		logger.Int("total_ram_mb", hardwareInfo.TotalRAMMB))

	// Auto-tune buffer pool settings
	if err := autoTuneBufferPool(config, hardwareInfo); err != nil {
		return fmt.Errorf("failed to auto-tune buffer pool: %w", err)
	}

	// Auto-tune other performance settings
	if err := autoTunePerformanceSettings(config, hardwareInfo); err != nil {
		return fmt.Errorf("failed to auto-tune performance settings: %w", err)
	}

	lg.Info("Auto-tuning completed",
		logger.String("buffer_pool_size", config.InnodbBufferPoolSize),
		logger.Int("buffer_pool_instances", config.InnodbBufferPoolInstances))

	return nil
}

// AutoTuneConfig is the exported entrypoint to perform hardware-based auto-tuning
// It uses background context and delegates to PerformAutoTuning.
func AutoTuneConfig(config *mariadb_config.MariaDBConfigureConfig) error {
	return PerformAutoTuning(context.Background(), config)
}

// autoTuneBufferPool melakukan auto-tuning untuk InnoDB buffer pool
func autoTuneBufferPool(config *mariadb_config.MariaDBConfigureConfig, hw *system.HardwareInfo) error {
	lg, _ := logger.Get()

	// Jika user sudah set manual, skip auto-tuning
	if config.InnodbBufferPoolSize != "" && config.InnodbBufferPoolSize != "128M" {
		lg.Info("Buffer pool size already configured manually, skipping auto-tuning")
		return nil
	}

	// Use system package helper to compute optimal buffer pool size
	optimal := hw.CalculateOptimalInnoDBBufferPool()
	config.InnodbBufferPoolSize = optimal

	lg.Info("Auto-tuned buffer pool size",
		logger.Int("total_ram_mb", hw.TotalRAMMB),
		logger.String("buffer_pool_size", config.InnodbBufferPoolSize))

	return nil
}

// autoTuneBufferPoolInstances melakukan auto-tuning untuk buffer pool instances
func autoTuneBufferPoolInstances(config *mariadb_config.MariaDBConfigureConfig, hw *system.HardwareInfo) error {
	lg, _ := logger.Get()

	// Jika user sudah set manual, skip auto-tuning
	if config.InnodbBufferPoolInstances > 0 && config.InnodbBufferPoolInstances != 8 {
		lg.Info("Buffer pool instances already configured manually, skipping auto-tuning")
		return nil
	}

	// Use system package helper to compute optimal instances
	instances := hw.CalculateOptimalInnoDBBufferPoolInstances()
	config.InnodbBufferPoolInstances = instances

	lg.Info("Auto-tuned buffer pool instances",
		logger.Int("cpu_cores", hw.CPUCores),
		logger.Int("instances", instances))

	return nil
}

// autoTunePerformanceSettings melakukan auto-tuning untuk setting performance lainnya
func autoTunePerformanceSettings(config *mariadb_config.MariaDBConfigureConfig, hw *system.HardwareInfo) error {
	lg, _ := logger.Get()

	// Auto-tune buffer pool instances
	if err := autoTuneBufferPoolInstances(config, hw); err != nil {
		return fmt.Errorf("failed to auto-tune buffer pool instances: %w", err)
	}

	// TODO: Bisa ditambahkan auto-tuning untuk setting lain seperti:
	// - innodb_log_file_size
	// - innodb_io_capacity
	// - max_connections
	// - query_cache_size (jika digunakan)
	// - table_open_cache
	// - innodb_flush_method

	lg.Info("Performance settings auto-tuning completed")
	return nil
}

// Helper functions
