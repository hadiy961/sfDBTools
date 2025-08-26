package mariadb

import (
	"testing"
)

func TestGetSupportedOSList(t *testing.T) {
	supportedOS := getSupportedOSList()

	expectedOS := []string{"centos", "ubuntu", "rhel", "rocky", "almalinux"}

	if len(supportedOS) != len(expectedOS) {
		t.Errorf("Expected %d supported OS, got %d", len(expectedOS), len(supportedOS))
	}

	// Check that all expected OS are present
	osMap := make(map[string]bool)
	for _, os := range supportedOS {
		osMap[os] = true
	}

	for _, expected := range expectedOS {
		if !osMap[expected] {
			t.Errorf("Expected OS '%s' not found in supported list", expected)
		}
	}
}

func TestIsSupportedOS(t *testing.T) {
	testCases := []struct {
		osID     string
		expected bool
	}{
		{"centos", true},
		{"ubuntu", true},
		{"rhel", true},
		{"rocky", true},
		{"almalinux", true},
		{"CentOS", true}, // Case insensitive
		{"UBUNTU", true}, // Case insensitive
		{"windows", false},
		{"macos", false},
		{"debian", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isSupportedOS(tc.osID)
		if result != tc.expected {
			t.Errorf("isSupportedOS('%s') = %v, expected %v", tc.osID, result, tc.expected)
		}
	}
}

func TestExtractOSID(t *testing.T) {
	testCases := []struct {
		osInfo   string
		expected string
	}{
		{`NAME="Ubuntu"
VERSION="20.04"
ID=ubuntu
ID_LIKE=debian`, "ubuntu"},
		{`NAME="CentOS Stream"
VERSION="9"
ID="centos"
ID_LIKE="rhel fedora"`, "centos"},
		{`NAME="Rocky Linux"
VERSION="8.5"
ID="rocky"`, "rocky"},
		{"No ID line", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		result := extractOSID(tc.osInfo)
		if result != tc.expected {
			t.Errorf("extractOSID() = '%s', expected '%s'", result, tc.expected)
		}
	}
}
