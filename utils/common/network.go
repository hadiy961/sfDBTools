package common

import (
	"fmt"
	"time"

	"sfDBTools/internal/logger"

	"github.com/go-resty/resty/v2"
)

// CheckInternetConnectivity performs simple connectivity check
func CheckInternetConnectivity() error {
	lg, _ := logger.Get()
	lg.Debug("Checking internet connectivity")

	// Quick HTTP check to reliable endpoints
	client := resty.New().
		SetTimeout(5*time.Second).
		SetRetryCount(0). // No retries for speed
		SetHeader("User-Agent", "sfDBTools")

	endpoints := []string{
		"https://www.google.com",
		"https://cloudflare.com",
	}

	for _, endpoint := range endpoints {
		resp, err := client.R().Head(endpoint)
		if err == nil && resp.StatusCode() < 400 {
			lg.Info("Internet connectivity verified")
			return nil
		}
	}

	return fmt.Errorf("no internet connectivity")
}

// CheckInternetConnectivityWithRetry checks connectivity with retry
func CheckInternetConnectivityWithRetry(maxRetries int, delay time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		if err := CheckInternetConnectivity(); err == nil {
			return nil
		}
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("connectivity failed after %d attempts", maxRetries)
}

// IsConnectedToInternet quick boolean check
func IsConnectedToInternet() bool {
	return CheckInternetConnectivity() == nil
}

// RequireInternetForOperation ensures connectivity for operations
func RequireInternetForOperation(operationName string) error {
	lg, _ := logger.Get()
	lg.Info("Checking internet connectivity", logger.String("operation", operationName))

	if err := CheckInternetConnectivity(); err != nil {
		return fmt.Errorf("internet connectivity is required for %s: %w", operationName, err)
	}

	lg.Info("Internet connectivity verified", logger.String("operation", operationName))
	return nil
}
