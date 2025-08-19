package health

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// SystemResourcesInfo represents system resources information
type SystemResourcesInfo struct {
	CPUCores             int             `json:"cpu_cores"`
	TotalSystemMemory    string          `json:"total_system_memory"`
	UsedSystemMemory     string          `json:"used_system_memory"`
	SystemMemoryPercent  string          `json:"system_memory_percent"`
	MariaDBMemoryUsage   string          `json:"mariadb_memory_usage"`
	MariaDBMemoryPercent string          `json:"mariadb_memory_percent"`
	DiskUsage            []DiskUsageInfo `json:"disk_usage"`
	MariaDBDataDirSize   string          `json:"mariadb_data_dir_size"`
	MariaDBDataDirPath   string          `json:"mariadb_data_dir_path"`
}

// DiskUsageInfo represents disk usage information
type DiskUsageInfo struct {
	Filesystem string `json:"filesystem"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	Available  string `json:"available"`
	UsePercent string `json:"use_percent"`
	MountedOn  string `json:"mounted_on"`
}

// GetSystemResourcesInfo retrieves system resources information
func GetSystemResourcesInfo(config database.Config) (*SystemResourcesInfo, error) {
	lg, _ := logger.Get()

	info := &SystemResourcesInfo{}

	// Get CPU cores
	cpuCores, err := getCPUCores()
	if err != nil {
		lg.Warn("Failed to get CPU cores", logger.Error(err))
	}
	info.CPUCores = cpuCores

	// Get memory information
	totalMem, usedMem, memPercent, err := getMemoryInfo()
	if err != nil {
		lg.Warn("Failed to get memory information", logger.Error(err))
	} else {
		info.TotalSystemMemory = totalMem
		info.UsedSystemMemory = usedMem
		info.SystemMemoryPercent = memPercent
	}

	// Get MariaDB memory usage
	mariadbMem, mariadbPercent, err := getMariaDBMemoryUsage(totalMem)
	if err != nil {
		lg.Warn("Failed to get MariaDB memory usage", logger.Error(err))
	} else {
		info.MariaDBMemoryUsage = mariadbMem
		info.MariaDBMemoryPercent = mariadbPercent
	}

	// Get disk usage
	diskUsage, err := getDiskUsage()
	if err != nil {
		lg.Warn("Failed to get disk usage", logger.Error(err))
	}
	info.DiskUsage = diskUsage

	// Get MariaDB data directory information
	dataDir, dataDirSize, err := getMariaDBDataDirInfo(config)
	if err != nil {
		lg.Warn("Failed to get MariaDB data directory info", logger.Error(err))
	} else {
		info.MariaDBDataDirPath = dataDir
		info.MariaDBDataDirSize = dataDirSize
	}

	return info, nil
}

// getCPUCores gets the number of CPU cores
func getCPUCores() (int, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}
	defer file.Close()

	cores := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processor") {
			cores++
		}
	}

	if cores == 0 {
		return 0, fmt.Errorf("no processors found in /proc/cpuinfo")
	}

	return cores, nil
}

// getMemoryInfo gets system memory information
func getMemoryInfo() (string, string, string, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}
	defer file.Close()

	var totalKB, availableKB int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if strings.HasPrefix(line, "MemTotal:") {
			totalKB, _ = strconv.ParseInt(fields[1], 10, 64)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			availableKB, _ = strconv.ParseInt(fields[1], 10, 64)
		}
	}

	if totalKB == 0 {
		return "", "", "", fmt.Errorf("failed to parse memory information")
	}

	usedKB := totalKB - availableKB
	usedPercent := float64(usedKB) / float64(totalKB) * 100

	totalGB := float64(totalKB) / 1024 / 1024
	usedGB := float64(usedKB) / 1024 / 1024

	return fmt.Sprintf("%.1f GB", totalGB),
		fmt.Sprintf("%.1f GB", usedGB),
		fmt.Sprintf("%.2f%%", usedPercent), nil
}

// getMariaDBMemoryUsage gets MariaDB memory usage
func getMariaDBMemoryUsage(totalSystemMemory string) (string, string, error) {
	// Get MariaDB/MySQL process memory usage
	cmd := exec.Command("pgrep", "-f", "mysqld|mariadb")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to find MariaDB/MySQL process: %w", err)
	}

	pids := strings.Fields(strings.TrimSpace(string(output)))
	if len(pids) == 0 {
		return "0 GB", "0%", nil
	}

	// Get memory usage for the first PID (main process)
	pid := pids[0]
	statusFile := fmt.Sprintf("/proc/%s/status", pid)
	file, err := os.Open(statusFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read process status: %w", err)
	}
	defer file.Close()

	var vmRSSKB int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				vmRSSKB, _ = strconv.ParseInt(fields[1], 10, 64)
				break
			}
		}
	}

	if vmRSSKB == 0 {
		return "0 GB", "0%", nil
	}

	mariadbGB := float64(vmRSSKB) / 1024 / 1024

	// Calculate percentage of system memory
	// Parse total system memory
	var totalGB float64
	if strings.Contains(totalSystemMemory, "GB") {
		totalStr := strings.Replace(totalSystemMemory, " GB", "", -1)
		totalGB, _ = strconv.ParseFloat(totalStr, 64)
	}

	var percent float64
	if totalGB > 0 {
		percent = mariadbGB / totalGB * 100
	}

	return fmt.Sprintf("%.1f GB", mariadbGB),
		fmt.Sprintf("%.2f%% of system memory", percent), nil
}

// getDiskUsage gets disk usage information
func getDiskUsage() ([]DiskUsageInfo, error) {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute df command: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid df output")
	}

	var diskUsage []DiskUsageInfo
	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 6 {
			continue
		}

		// Skip tmpfs, devtmpfs, and other virtual filesystems
		if strings.HasPrefix(fields[0], "tmpfs") ||
			strings.HasPrefix(fields[0], "devtmpfs") ||
			strings.HasPrefix(fields[0], "udev") ||
			strings.HasPrefix(fields[0], "overlay") {
			continue
		}

		diskInfo := DiskUsageInfo{
			Filesystem: fields[0],
			Size:       fields[1],
			Used:       fields[2],
			Available:  fields[3],
			UsePercent: fields[4],
			MountedOn:  fields[5],
		}
		diskUsage = append(diskUsage, diskInfo)
	}

	return diskUsage, nil
}

// getMariaDBDataDirInfo gets MariaDB data directory information
func getMariaDBDataDirInfo(config database.Config) (string, string, error) {
	lg, _ := logger.Get()

	// Get data directory from MariaDB configuration
	db, err := database.GetWithoutDB(config)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var dataDir string
	query := "SHOW VARIABLES LIKE 'datadir'"
	rows, err := db.Query(query)
	if err != nil {
		lg.Warn("Failed to get data directory from database", logger.Error(err))
		dataDir = "/var/lib/mysql" // Default fallback
	} else {
		defer rows.Close()
		if rows.Next() {
			var variable, value string
			if err := rows.Scan(&variable, &value); err == nil {
				dataDir = value
			}
		}
		if dataDir == "" {
			dataDir = "/var/lib/mysql" // Default fallback
		}
	}

	// Get directory size
	size, err := getDirectorySize(dataDir)
	if err != nil {
		return dataDir, "Unknown", fmt.Errorf("failed to get directory size: %w", err)
	}

	return dataDir, size, nil
}

// getDirectorySize gets the size of a directory
func getDirectorySize(dirPath string) (string, error) {
	cmd := exec.Command("du", "-sh", dirPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get directory size: %w", err)
	}

	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) < 1 {
		return "", fmt.Errorf("invalid du output")
	}

	return fields[0], nil
}

// FormatSystemResourcesInfo formats system resources information for display
func FormatSystemResourcesInfo(info *SystemResourcesInfo) []string {
	var details []string

	if info.CPUCores > 0 {
		details = append(details, fmt.Sprintf("- CPU Cores: %d", info.CPUCores))
	}

	if info.TotalSystemMemory != "" {
		details = append(details, fmt.Sprintf("- Total System Memory: %s", info.TotalSystemMemory))
	}

	if info.UsedSystemMemory != "" && info.SystemMemoryPercent != "" {
		details = append(details, fmt.Sprintf("- Used System Memory: %s (%s)", info.UsedSystemMemory, info.SystemMemoryPercent))
	}

	if info.MariaDBMemoryUsage != "" && info.MariaDBMemoryPercent != "" {
		details = append(details, fmt.Sprintf("- MariaDB Memory Usage: %s (%s)", info.MariaDBMemoryUsage, info.MariaDBMemoryPercent))
	}

	if len(info.DiskUsage) > 0 {
		details = append(details, "- Disk Usage:")
		details = append(details, "  | Filesystem | Size   | Used   | Avail  | Use% | Mounted On        |")
		details = append(details, "  |------------|--------|--------|--------|------|-------------------|")

		for _, disk := range info.DiskUsage {
			details = append(details, fmt.Sprintf("  | %-10s | %-6s | %-6s | %-6s | %-4s | %-17s |",
				disk.Filesystem, disk.Size, disk.Used, disk.Available, disk.UsePercent, disk.MountedOn))
		}
	}

	if info.MariaDBDataDirSize != "" && info.MariaDBDataDirPath != "" {
		details = append(details, fmt.Sprintf("- MariaDB Data Directory Size: %s (from `%s`)", info.MariaDBDataDirSize, info.MariaDBDataDirPath))
	}

	return details
}

// ValidateSystemResources validates system resources and returns warnings/errors
func ValidateSystemResources(info *SystemResourcesInfo) (warnings []string, errors []string) {
	// Check memory usage
	if info.SystemMemoryPercent != "" {
		percentStr := strings.Replace(info.SystemMemoryPercent, "%", "", -1)
		if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
			if percent > 90 {
				errors = append(errors, "System memory usage is critical (>90%)")
			} else if percent > 80 {
				warnings = append(warnings, "System memory usage is high (>80%)")
			}
		}
	}

	// Check disk usage
	for _, disk := range info.DiskUsage {
		percentStr := strings.Replace(disk.UsePercent, "%", "", -1)
		if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
			if percent > 95 {
				errors = append(errors, fmt.Sprintf("Disk %s is critical (>95%% full)", disk.MountedOn))
			} else if percent > 85 {
				warnings = append(warnings, fmt.Sprintf("Disk %s is high (>85%% full)", disk.MountedOn))
			}
		}
	}

	return warnings, errors
}
