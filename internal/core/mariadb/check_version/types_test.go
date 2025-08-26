package check_version

import (
	"testing"
)

func TestDefaultCheckVersionConfig(t *testing.T) {
	config := DefaultCheckVersionConfig()

	if config == nil {
		t.Fatal("DefaultCheckVersionConfig should not return nil")
	}

	if config.MinimumVersion != "10.6" {
		t.Errorf("Expected minimum version to be '10.6', got '%s'", config.MinimumVersion)
	}

	if config.APIBaseURL == "" {
		t.Error("API base URL should not be empty")
	}

	if config.APITimeout <= 0 {
		t.Error("API timeout should be greater than 0")
	}
}

func TestNewVersionService(t *testing.T) {
	config := DefaultCheckVersionConfig()
	service := NewVersionService(config)

	if service == nil {
		t.Fatal("NewVersionService should not return nil")
	}

	if service.config != config {
		t.Error("Service should use the provided config")
	}

	if service.client == nil {
		t.Error("Service should have an HTTP client")
	}
}

func TestNewVersionServiceWithNilConfig(t *testing.T) {
	service := NewVersionService(nil)

	if service == nil {
		t.Fatal("NewVersionService should not return nil even with nil config")
	}

	if service.config == nil {
		t.Error("Service should use default config when nil is provided")
	}
}

func TestNewCheckVersionRunner(t *testing.T) {
	config := DefaultCheckVersionConfig()
	runner := NewCheckVersionRunner(config)

	if runner == nil {
		t.Fatal("NewCheckVersionRunner should not return nil")
	}

	if runner.versionService == nil {
		t.Error("Runner should have a version service")
	}
}
