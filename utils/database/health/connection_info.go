package health

import (
	"net"
	"strconv"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// ConnectionInfo represents database connection information
type ConnectionInfo struct {
	Host             string `json:"host"`
	IPAddress        string `json:"ip_address"`
	Port             int    `json:"port"`
	ConnectionStatus string `json:"connection_status"`
	IsConnected      bool   `json:"is_connected"`
}

// GetConnectionInfo retrieves database connection information
func GetConnectionInfo(config database.Config) (*ConnectionInfo, error) {
	lg, _ := logger.Get()

	info := &ConnectionInfo{
		Host: config.Host,
		Port: config.Port,
	}

	// Resolve IP address
	ipAddress, err := resolveIPAddress(config.Host)
	if err != nil {
		lg.Warn("Failed to resolve IP address", logger.Error(err))
		info.IPAddress = config.Host
	} else {
		info.IPAddress = ipAddress
	}

	// Test database connection
	err = database.ValidateConnection(config)
	if err != nil {
		lg.Warn("Database connection failed", logger.Error(err))
		info.ConnectionStatus = "❌ Failed"
		info.IsConnected = false
	} else {
		info.ConnectionStatus = "✅ Successful"
		info.IsConnected = true
	}

	return info, nil
}

// resolveIPAddress resolves hostname to IP address
func resolveIPAddress(hostname string) (string, error) {
	// Handle localhost cases
	if hostname == "localhost" || hostname == "" {
		return "127.0.0.1", nil
	}

	// If it's already an IP address, return as is
	if net.ParseIP(hostname) != nil {
		return hostname, nil
	}

	// Resolve hostname to IP
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}

	// Return first IPv4 address found
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	// If no IPv4 found, return first IP
	if len(ips) > 0 {
		return ips[0].String(), nil
	}

	return hostname, nil
}

// FormatConnectionInfo formats connection info for display
func FormatConnectionInfo(info *ConnectionInfo) []string {
	return []string{
		"- Host: " + info.Host,
		"- IP Address: " + info.IPAddress,
		"- Port: " + strconv.Itoa(info.Port),
		"- Connection Status: " + info.ConnectionStatus,
	}
}
