package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateBasicSafety(t *testing.T) {
	fs := &fileSystem{}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		// Critical system paths should be blocked
		{
			name:    "root directory blocked",
			path:    "/",
			wantErr: true,
			errMsg:  "critical system directory",
		},
		{
			name:    "bin directory blocked",
			path:    "/bin",
			wantErr: true,
			errMsg:  "critical system directory",
		},
		{
			name:    "usr directory blocked",
			path:    "/usr",
			wantErr: true,
			errMsg:  "critical system directory",
		},

		// Database paths should be allowed
		{
			name:    "mysql data directory allowed",
			path:    "/var/lib/mysql",
			wantErr: false,
		},
		{
			name:    "mariadb data directory allowed",
			path:    "/var/lib/mariadb",
			wantErr: false,
		},
		{
			name:    "mysql config directory allowed",
			path:    "/etc/mysql",
			wantErr: false,
		},
		{
			name:    "mariadb config directory allowed",
			path:    "/etc/mariadb",
			wantErr: false,
		},
		{
			name:    "custom mysql path allowed",
			path:    "/opt/mysql/data",
			wantErr: false,
		},

		// Non-database paths should now be allowed (user choice respected)
		{
			name:    "custom path now allowed",
			path:    "/custom/database/path",
			wantErr: false,
		},
		{
			name:    "user directory now allowed",
			path:    "/home/user/mysql-backup",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.validateBasicSafety(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateBasicSafety() expected error for path %s, got nil", tt.path)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateBasicSafety() error = %v, want error containing %s", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateBasicSafety() unexpected error for path %s: %v", tt.path, err)
				}
			}
		})
	}
}

func TestContainsDatabaseMarkers(t *testing.T) {
	fs := &fileSystem{}

	// Create temporary test directories
	tempDir := t.TempDir()

	// Test directory with database markers
	dbDir := filepath.Join(tempDir, "mysql-data")
	err := os.MkdirAll(filepath.Join(dbDir, "mysql"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test directory without database markers
	nonDbDir := filepath.Join(tempDir, "random-dir")
	err = os.MkdirAll(nonDbDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "directory with mysql marker",
			path:     dbDir,
			expected: true,
		},
		{
			name:     "directory without database markers",
			path:     nonDbDir,
			expected: false,
		},
		{
			name:     "directory with database keyword in name",
			path:     filepath.Join(tempDir, "mysql-backup"),
			expected: true,
		},
		{
			name:     "non-existent path",
			path:     "/non/existent/path",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create directory if it has database keyword but doesn't exist yet
			if tt.expected && !fs.Exists(tt.path) {
				err := os.MkdirAll(tt.path, 0755)
				if err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			}

			result := fs.containsDatabaseMarkers(tt.path)
			if result != tt.expected {
				t.Errorf("containsDatabaseMarkers(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
