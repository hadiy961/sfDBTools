package configure

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
	mariadb_utils "sfDBTools/utils/mariadb"
	"sfDBTools/utils/system"
)

// PerformAutoTuning melakukan auto-tuning berdasarkan hardware sistem
// Sesuai dengan Step 12-14 dalam flow implementasi
func PerformAutoTuning(ctx context.Context, config *mariadb_utils.MariaDBConfigureConfig) error {
	lg, err := logger.Get()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
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
func AutoTuneConfig(config *mariadb_utils.MariaDBConfigureConfig) error {
	return PerformAutoTuning(context.Background(), config)
}

// autoTuneBufferPool melakukan auto-tuning untuk InnoDB buffer pool
func autoTuneBufferPool(config *mariadb_utils.MariaDBConfigureConfig, hw *system.HardwareInfo) error {
	lg, _ := logger.Get()

	// Jika user sudah set manual, skip auto-tuning
	if config.InnodbBufferPoolSize != "" && config.InnodbBufferPoolSize != "128M" {
		lg.Info("Buffer pool size already configured manually, skipping auto-tuning")
		return nil
	}

	// Hitung buffer pool size: 70-80% dari total RAM
	// Untuk sistem dengan RAM < 4GB, gunakan 60%
	// Untuk sistem dengan RAM >= 4GB, gunakan 75%
	var percentage float64
	if hw.TotalRAMGB < 4 {
		percentage = 0.60 // 60% untuk sistem RAM kecil
	} else {
		percentage = 0.75 // 75% untuk sistem RAM normal
	}

	bufferPoolMB := int(float64(hw.TotalRAMMB) * percentage)

	// Minimum 128MB, maksimum leave 1GB untuk OS
	minBufferMB := 128
	maxBufferMB := hw.TotalRAMMB - 1024 // Leave 1GB for OS

	if bufferPoolMB < minBufferMB {
		bufferPoolMB = minBufferMB
	}
	if bufferPoolMB > maxBufferMB && maxBufferMB > minBufferMB {
		bufferPoolMB = maxBufferMB
	}

	// Set buffer pool size
	if bufferPoolMB >= 1024 {
		config.InnodbBufferPoolSize = fmt.Sprintf("%dG", bufferPoolMB/1024)
	} else {
		config.InnodbBufferPoolSize = fmt.Sprintf("%dM", bufferPoolMB)
	}

	lg.Info("Auto-tuned buffer pool size",
		logger.Int("total_ram_mb", hw.TotalRAMMB),
		logger.Float64("percentage_used", percentage),
		logger.String("buffer_pool_size", config.InnodbBufferPoolSize))

	return nil
}

// autoTuneBufferPoolInstances melakukan auto-tuning untuk buffer pool instances
func autoTuneBufferPoolInstances(config *mariadb_utils.MariaDBConfigureConfig, hw *system.HardwareInfo) error {
	lg, _ := logger.Get()

	// Jika user sudah set manual, skip auto-tuning
	if config.InnodbBufferPoolInstances > 0 && config.InnodbBufferPoolInstances != 8 {
		lg.Info("Buffer pool instances already configured manually, skipping auto-tuning")
		return nil
	}

	// Parse buffer pool size untuk mendapatkan ukuran dalam MB
	bufferPoolMB, err := parseMemorySizeToMB(config.InnodbBufferPoolSize)
	if err != nil {
		lg.Warn("Could not parse buffer pool size for instances calculation", logger.Error(err))
		return nil
	}

	// Rumus: min(CPU cores, buffer_pool_size_GB)
	// Minimum 1, maksimum 64 (limit MariaDB)
	bufferPoolGB := bufferPoolMB / 1024
	if bufferPoolGB < 1 {
		bufferPoolGB = 1
	}

	instances := hw.CPUCores
	if bufferPoolGB < instances {
		instances = bufferPoolGB
	}

	// Minimum 1, maksimum 64
	if instances < 1 {
		instances = 1
	}
	if instances > 64 {
		instances = 64
	}

	config.InnodbBufferPoolInstances = instances

	lg.Info("Auto-tuned buffer pool instances",
		logger.Int("cpu_cores", hw.CPUCores),
		logger.Int("buffer_pool_gb", bufferPoolGB),
		logger.Int("instances", instances))

	return nil
}

// autoTunePerformanceSettings melakukan auto-tuning untuk setting performance lainnya
func autoTunePerformanceSettings(config *mariadb_utils.MariaDBConfigureConfig, hw *system.HardwareInfo) error {
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

// parseMemorySizeToMB mengkonversi string memory size ke MB
func parseMemorySizeToMB(size string) (int, error) {
	if size == "" {
		return 0, fmt.Errorf("empty memory size")
	}

	size = strings.ToUpper(strings.TrimSpace(size))

	// Extract numeric part and suffix
	var numStr string
	var suffix string

	if len(size) < 2 {
		return 0, fmt.Errorf("invalid memory size format: %s", size)
	}

	suffix = size[len(size)-1:]
	numStr = size[:len(size)-1]

	// Parse numeric part
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric part in memory size: %s", numStr)
	}

	// Convert to MB based on suffix
	switch suffix {
	case "K":
		return num / 1024, nil
	case "M":
		return num, nil
	case "G":
		return num * 1024, nil
	default:
		return 0, fmt.Errorf("unsupported memory size suffix: %s", suffix)
	}
}

// calculateOptimalBufferPoolSize menghitung ukuran optimal buffer pool
func calculateOptimalBufferPoolSize(totalRAMMB int) (string, error) {
	// Implementasi yang lebih advanced bisa mempertimbangkan:
	// - Workload type (OLTP vs OLAP)
	// - Apakah ada aplikasi lain yang berjalan
	// - Storage type (SSD vs HDD)

	var bufferPoolMB int

	if totalRAMMB <= 1024 { // <= 1GB
		bufferPoolMB = int(float64(totalRAMMB) * 0.5) // 50%
	} else if totalRAMMB <= 4096 { // <= 4GB
		bufferPoolMB = int(float64(totalRAMMB) * 0.6) // 60%
	} else if totalRAMMB <= 8192 { // <= 8GB
		bufferPoolMB = int(float64(totalRAMMB) * 0.7) // 70%
	} else { // > 8GB
		bufferPoolMB = int(float64(totalRAMMB) * 0.75) // 75%
	}

	// Minimum 128MB
	if bufferPoolMB < 128 {
		bufferPoolMB = 128
	}

	// Convert to human readable format
	if bufferPoolMB >= 1024 {
		return fmt.Sprintf("%dG", bufferPoolMB/1024), nil
	} else {
		return fmt.Sprintf("%dM", bufferPoolMB), nil
	}
}

// calculateOptimalInstances menghitung jumlah optimal buffer pool instances
func calculateOptimalInstances(cpuCores int, bufferPoolSizeMB int) int {
	// Rule: 1 instance per GB buffer pool, tapi tidak lebih dari CPU cores
	bufferPoolGB := bufferPoolSizeMB / 1024
	if bufferPoolGB < 1 {
		bufferPoolGB = 1
	}

	instances := cpuCores
	if bufferPoolGB < instances {
		instances = bufferPoolGB
	}

	// Minimum 1, maksimum 64
	if instances < 1 {
		instances = 1
	}
	if instances > 64 {
		instances = 64
	}

	return instances
}
