package system

import (
	"fmt"
	"net"
	"strings"
	"time"

	"sfDBTools/internal/logger"
)

// PortInfo berisi informasi tentang port
type PortInfo struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Status      string `json:"status"`
	ProcessName string `json:"process_name,omitempty"`
	PID         string `json:"pid,omitempty"`
}

// IsPortAvailable mengecek apakah port tersedia (tidak digunakan)
func IsPortAvailable(port int) bool {
	lg, _ := logger.Get()

	// Test TCP port
	tcpAddr := fmt.Sprintf(":%d", port)
	tcpListener, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		lg.Debug("Port TCP tidak tersedia",
			logger.Int("port", port),
			logger.Error(err))
		return false
	}
	tcpListener.Close()

	// Test UDP port
	udpAddr, err := net.ResolveUDPAddr("udp", tcpAddr)
	if err != nil {
		lg.Debug("Gagal resolve UDP address",
			logger.Int("port", port),
			logger.Error(err))
		return false
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		lg.Debug("Port UDP tidak tersedia",
			logger.Int("port", port),
			logger.Error(err))
		return false
	}
	udpConn.Close()

	lg.Debug("Port tersedia", logger.Int("port", port))
	return true
}

// CheckPortConflict mengecek apakah port mengalami konflik dengan service lain
func CheckPortConflict(port int) (*PortInfo, error) {
	lg, _ := logger.Get()
	lg.Debug("Checking port conflict", logger.Int("port", port))

	// Test dengan mencoba bind ke port
	if IsPortAvailable(port) {
		return &PortInfo{
			Port:     port,
			Protocol: "tcp",
			Status:   "available",
		}, nil
	}

	// Jika port tidak tersedia, coba dapatkan informasi process yang menggunakan
	processInfo, err := getPortUsage(port)
	if err != nil {
		lg.Warn("Gagal mendapatkan informasi process untuk port",
			logger.Int("port", port),
			logger.Error(err))
		return &PortInfo{
			Port:     port,
			Protocol: "tcp",
			Status:   "in_use",
		}, nil
	}

	return processInfo, nil
}

// getPortUsage mencoba mendapatkan informasi process yang menggunakan port
func getPortUsage(port int) (*PortInfo, error) {
	// Implementasi sederhana: coba connect ke port untuk mendeteksi service
	address := fmt.Sprintf("localhost:%d", port)

	// Test dengan timeout singkat
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		// Port mungkin digunakan tapi tidak menerima koneksi
		return &PortInfo{
			Port:     port,
			Protocol: "tcp",
			Status:   "in_use",
		}, nil
	}
	defer conn.Close()

	// Port digunakan dan menerima koneksi
	return &PortInfo{
		Port:     port,
		Protocol: "tcp",
		Status:   "in_use_accepting_connections",
	}, nil
}

// ValidatePortRange memvalidasi apakah port dalam range yang valid
func ValidatePortRange(port int) error {
	if port < 1 {
		return fmt.Errorf("port harus lebih besar dari 0, diberikan: %d", port)
	}
	if port > 65535 {
		return fmt.Errorf("port harus kurang dari atau sama dengan 65535, diberikan: %d", port)
	}
	if port < 1024 {
		return fmt.Errorf("port harus lebih besar dari atau sama dengan 1024 untuk non-root services, diberikan: %d", port)
	}
	return nil
}

// FindAvailablePort mencari port yang tersedia dalam range yang diberikan
func FindAvailablePort(startPort, endPort int) (int, error) {
	lg, _ := logger.Get()

	if err := ValidatePortRange(startPort); err != nil {
		return 0, fmt.Errorf("start port tidak valid: %w", err)
	}
	if err := ValidatePortRange(endPort); err != nil {
		return 0, fmt.Errorf("end port tidak valid: %w", err)
	}
	if startPort > endPort {
		return 0, fmt.Errorf("start port (%d) harus kurang dari atau sama dengan end port (%d)", startPort, endPort)
	}

	lg.Debug("Mencari port tersedia",
		logger.Int("start_port", startPort),
		logger.Int("end_port", endPort))

	for port := startPort; port <= endPort; port++ {
		if IsPortAvailable(port) {
			lg.Info("Ditemukan port tersedia", logger.Int("port", port))
			return port, nil
		}
	}

	return 0, fmt.Errorf("tidak ada port tersedia dalam range %d-%d", startPort, endPort)
}

// GetMariaDBDefaultPorts mengembalikan list default ports yang biasa digunakan MariaDB
func GetMariaDBDefaultPorts() []int {
	return []int{3306, 3307, 3308, 3309, 3310}
}

// SuggestAlternativePort menyarankan port alternatif jika port yang diminta tidak tersedia
func SuggestAlternativePort(requestedPort int) (int, error) {
	lg, _ := logger.Get()

	// Cek port yang diminta
	if IsPortAvailable(requestedPort) {
		return requestedPort, nil
	}

	lg.Info("Port yang diminta tidak tersedia, mencari alternatif",
		logger.Int("requested_port", requestedPort))

	// Cari di sekitar port yang diminta (±10)
	startPort := requestedPort - 10
	endPort := requestedPort + 10

	if startPort < 1024 {
		startPort = 1024
	}
	if endPort > 65535 {
		endPort = 65535
	}

	// Skip port yang diminta karena sudah dicek
	for port := startPort; port <= endPort; port++ {
		if port == requestedPort {
			continue
		}
		if IsPortAvailable(port) {
			lg.Info("Ditemukan port alternatif",
				logger.Int("requested_port", requestedPort),
				logger.Int("alternative_port", port))
			return port, nil
		}
	}

	// Jika tidak ada di range sekitar, coba range MariaDB default
	mariadbPorts := GetMariaDBDefaultPorts()
	for _, port := range mariadbPorts {
		if port == requestedPort {
			continue
		}
		if IsPortAvailable(port) {
			lg.Info("Ditemukan port alternatif dari MariaDB defaults",
				logger.Int("requested_port", requestedPort),
				logger.Int("alternative_port", port))
			return port, nil
		}
	}

	return 0, fmt.Errorf("tidak dapat menemukan port alternatif untuk port %d", requestedPort)
}

// String mengembalikan representasi string dari PortInfo
func (pi *PortInfo) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Port: %d", pi.Port))
	parts = append(parts, fmt.Sprintf("Protocol: %s", pi.Protocol))
	parts = append(parts, fmt.Sprintf("Status: %s", pi.Status))

	if pi.ProcessName != "" {
		parts = append(parts, fmt.Sprintf("Process: %s", pi.ProcessName))
	}
	if pi.PID != "" {
		parts = append(parts, fmt.Sprintf("PID: %s", pi.PID))
	}

	return strings.Join(parts, ", ")
}
