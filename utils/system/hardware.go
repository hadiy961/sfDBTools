package system

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"sfDBTools/internal/logger"
)

// HardwareInfo berisi informasi hardware sistem
type HardwareInfo struct {
	CPUCores      int   // Jumlah CPU cores
	TotalRAMBytes int64 // Total RAM dalam bytes
	TotalRAMGB    int   // Total RAM dalam GB (rounded)
	TotalRAMMB    int   // Total RAM dalam MB
}

// GetHardwareInfo mendeteksi informasi hardware sistem
func GetHardwareInfo() (*HardwareInfo, error) {
	lg, _ := logger.Get()

	// Deteksi CPU cores
	cpuCores := runtime.NumCPU()
	lg.Info("Detected CPU cores", logger.Int("cores", cpuCores))

	// Deteksi RAM
	totalRAMBytes, err := getTotalRAM()
	if err != nil {
		return nil, fmt.Errorf("gagal mendeteksi RAM: %w", err)
	}

	totalRAMMB := int(totalRAMBytes / (1024 * 1024))
	totalRAMGB := int(totalRAMBytes / (1024 * 1024 * 1024))

	lg.Info("Detected system RAM",
		logger.Int64("total_bytes", totalRAMBytes),
		logger.Int("total_mb", totalRAMMB),
		logger.Int("total_gb", totalRAMGB))

	return &HardwareInfo{
		CPUCores:      cpuCores,
		TotalRAMBytes: totalRAMBytes,
		TotalRAMGB:    totalRAMGB,
		TotalRAMMB:    totalRAMMB,
	}, nil
}

// getTotalRAM membaca total RAM dari /proc/meminfo
func getTotalRAM() (int64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, fmt.Errorf("gagal membuka /proc/meminfo: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// MemTotal biasanya dalam kB
				ramKB, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("gagal parsing MemTotal: %w", err)
				}
				// Convert kB ke bytes
				return ramKB * 1024, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading /proc/meminfo: %w", err)
	}

	return 0, fmt.Errorf("MemTotal tidak ditemukan di /proc/meminfo")
}

// CalculateOptimalInnoDBBufferPool menghitung optimal InnoDB buffer pool size
// Berdasarkan best practice: 70-80% dari total RAM
func (hi *HardwareInfo) CalculateOptimalInnoDBBufferPool() string {
	// Gunakan 75% dari total RAM
	optimalBytes := hi.TotalRAMBytes * 75 / 100

	// Convert ke format yang readable
	if optimalBytes >= 1024*1024*1024 {
		// Jika >= 1GB, tampilkan dalam GB
		optimalGB := optimalBytes / (1024 * 1024 * 1024)
		return fmt.Sprintf("%dG", optimalGB)
	} else {
		// Jika < 1GB, tampilkan dalam MB
		optimalMB := optimalBytes / (1024 * 1024)
		return fmt.Sprintf("%dM", optimalMB)
	}
}

// CalculateOptimalInnoDBBufferPoolInstances menghitung optimal jumlah buffer pool instances
// Berdasarkan best practice: min(CPU cores, buffer_pool_size/1GB)
func (hi *HardwareInfo) CalculateOptimalInnoDBBufferPoolInstances() int {
	// Hitung buffer pool size dalam GB
	bufferPoolBytes := hi.TotalRAMBytes * 75 / 100
	bufferPoolGB := int(bufferPoolBytes / (1024 * 1024 * 1024))

	// Minimal 1 instance, maksimal berdasarkan CPU cores atau buffer pool size
	instances := hi.CPUCores
	if bufferPoolGB < instances {
		instances = bufferPoolGB
	}

	// Minimal 1 instance
	if instances < 1 {
		instances = 1
	}

	// Maksimal 64 instance (MySQL/MariaDB limit)
	if instances > 64 {
		instances = 64
	}

	return instances
}

// String mengembalikan representasi string dari HardwareInfo
func (hi *HardwareInfo) String() string {
	return fmt.Sprintf("CPU Cores: %d, RAM: %d GB (%d MB), Optimal Buffer Pool: %s, Optimal Instances: %d",
		hi.CPUCores,
		hi.TotalRAMGB,
		hi.TotalRAMMB,
		hi.CalculateOptimalInnoDBBufferPool(),
		hi.CalculateOptimalInnoDBBufferPoolInstances())
}
