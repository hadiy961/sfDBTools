package common

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"sfDBTools/internal/logger"
)

// NetworkConnectivityConfig holds configuration for network connectivity checks
type NetworkConnectivityConfig struct {
	Timeout        time.Duration
	CheckEndpoints []string
	DNSServers     []string
}

// DefaultNetworkConfig returns default network connectivity configuration
func DefaultNetworkConfig() *NetworkConnectivityConfig {
	return &NetworkConnectivityConfig{
		Timeout: 10 * time.Second,
		CheckEndpoints: []string{
			"https://www.google.com",
			"https://www.cloudflare.com",
			"https://downloads.mariadb.org",
			"https://archive.mariadb.org",
		},
		DNSServers: []string{
			"8.8.8.8:53",     // Google DNS
			"1.1.1.1:53",     // Cloudflare DNS
			"208.67.222.222", // OpenDNS
		},
	}
}

// CheckInternetConnectivity performs comprehensive internet connectivity check
func CheckInternetConnectivity() error {
	lg, _ := logger.Get()
	config := DefaultNetworkConfig()

	lg.Debug("Starting internet connectivity check")

	// 1. Check DNS resolution
	if err := checkDNSResolution(config); err != nil {
		lg.Error("DNS resolution failed", logger.Error(err))
		return fmt.Errorf("DNS resolution failed: %w", err)
	}
	lg.Debug("DNS resolution successful")

	// 2. Check HTTP connectivity to multiple endpoints
	if err := checkHTTPConnectivity(config); err != nil {
		lg.Error("HTTP connectivity failed", logger.Error(err))
		return fmt.Errorf("HTTP connectivity failed: %w", err)
	}
	lg.Debug("HTTP connectivity successful")

	lg.Info("Internet connectivity verified")
	return nil
}

// checkDNSResolution checks if DNS resolution is working
func checkDNSResolution(config *NetworkConnectivityConfig) error {
	lg, _ := logger.Get()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: config.Timeout,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Try to resolve a known domain
	testDomains := []string{"google.com", "cloudflare.com", "mariadb.org"}

	for _, domain := range testDomains {
		lg.Debug("Testing DNS resolution", logger.String("domain", domain))

		_, err := resolver.LookupHost(ctx, domain)
		if err == nil {
			lg.Debug("DNS resolution successful", logger.String("domain", domain))
			return nil
		}

		lg.Debug("DNS resolution failed for domain",
			logger.String("domain", domain),
			logger.Error(err))
	}

	return fmt.Errorf("failed to resolve any test domains")
}

// checkHTTPConnectivity checks HTTP connectivity to multiple endpoints
func checkHTTPConnectivity(config *NetworkConnectivityConfig) error {
	lg, _ := logger.Get()

	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	for _, endpoint := range config.CheckEndpoints {
		lg.Debug("Testing HTTP connectivity", logger.String("endpoint", endpoint))

		ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

		req, err := http.NewRequestWithContext(ctx, "HEAD", endpoint, nil)
		if err != nil {
			cancel()
			lg.Debug("Failed to create request",
				logger.String("endpoint", endpoint),
				logger.Error(err))
			continue
		}

		resp, err := client.Do(req)
		cancel()

		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 { // Accept any non-server error status
				lg.Debug("HTTP connectivity successful",
					logger.String("endpoint", endpoint),
					logger.Int("status_code", resp.StatusCode))
				return nil
			}
		}

		lg.Debug("HTTP connectivity failed",
			logger.String("endpoint", endpoint),
			logger.Error(err))
	}

	return fmt.Errorf("failed to connect to any HTTP endpoints")
}

// CheckInternetConnectivityWithRetry performs internet connectivity check with retry mechanism
func CheckInternetConnectivityWithRetry(maxRetries int, retryDelay time.Duration) error {
	lg, _ := logger.Get()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		lg.Debug("Internet connectivity check attempt",
			logger.Int("attempt", attempt),
			logger.Int("max_retries", maxRetries))

		err := CheckInternetConnectivity()
		if err == nil {
			if attempt > 1 {
				lg.Info("Internet connectivity restored", logger.Int("attempts", attempt))
			}
			return nil
		}

		if attempt < maxRetries {
			lg.Warn("Internet connectivity check failed, retrying",
				logger.Error(err),
				logger.Int("attempt", attempt),
				logger.String("retry_in", retryDelay.String()))
			time.Sleep(retryDelay)
		} else {
			lg.Error("Internet connectivity check failed after all retries",
				logger.Error(err),
				logger.Int("total_attempts", attempt))
			return fmt.Errorf("internet connectivity check failed after %d attempts: %w", maxRetries, err)
		}
	}

	return fmt.Errorf("internet connectivity check failed after %d attempts", maxRetries)
}

// IsConnectedToInternet performs a quick internet connectivity check
func IsConnectedToInternet() bool {
	err := CheckInternetConnectivity()
	return err == nil
}

// CheckMariaDBConnectivity checks connectivity specifically for MariaDB-related endpoints
func CheckMariaDBConnectivity() error {
	lg, _ := logger.Get()

	mariadbConfig := &NetworkConnectivityConfig{
		Timeout: 15 * time.Second, // Slightly longer timeout for MariaDB endpoints
		CheckEndpoints: []string{
			"https://downloads.mariadb.org",
			"https://archive.mariadb.org",
			"https://mirror.mariadb.org",
			"https://www.google.com", // Fallback general connectivity test
		},
		DNSServers: []string{
			"8.8.8.8:53",     // Google DNS
			"1.1.1.1:53",     // Cloudflare DNS
			"208.67.222.222", // OpenDNS
		},
	}

	lg.Debug("Starting MariaDB-specific connectivity check")

	// 1. Check DNS resolution
	if err := checkDNSResolution(mariadbConfig); err != nil {
		lg.Error("DNS resolution failed for MariaDB operations", logger.Error(err))
		return fmt.Errorf("DNS resolution failed for MariaDB operations: %w", err)
	}
	lg.Debug("DNS resolution successful for MariaDB operations")

	// 2. Check HTTP connectivity to MariaDB endpoints
	if err := checkHTTPConnectivity(mariadbConfig); err != nil {
		lg.Error("HTTP connectivity failed for MariaDB operations", logger.Error(err))
		return fmt.Errorf("HTTP connectivity failed for MariaDB operations: %w", err)
	}
	lg.Debug("HTTP connectivity successful for MariaDB operations")

	lg.Info("MariaDB connectivity verified")
	return nil
}

// RequireInternetForOperation checks internet connectivity and returns descriptive error for specific operations
func RequireInternetForOperation(operationName string) error {
	lg, _ := logger.Get()

	lg.Info("Checking internet connectivity", logger.String("operation", operationName))

	if err := CheckMariaDBConnectivity(); err != nil {
		lg.Error("Internet connectivity required for operation",
			logger.String("operation", operationName),
			logger.Error(err))
		return fmt.Errorf("internet connectivity is required for %s: %w", operationName, err)
	}

	lg.Info("Internet connectivity verified", logger.String("operation", operationName))
	return nil
}
