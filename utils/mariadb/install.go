package mariadb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
)

// MariaDB Download API structures
type MariaDBDownloadFile struct {
	FileID          string            `json:"file_id"`
	FileName        string            `json:"file_name"`
	PackageType     string            `json:"package_type"`
	OS              string            `json:"os"`
	CPU             string            `json:"cpu"`
	Checksum        map[string]string `json:"checksum"`
	FileDownloadURL string            `json:"file_download_url"`
}

type MariaDBDownloadRelease struct {
	ReleaseID   string                `json:"release_id"`
	ReleaseName string                `json:"release_name"`
	Files       []MariaDBDownloadFile `json:"files"`
}

type MariaDBDownloadResponse struct {
	Releases map[string]MariaDBDownloadRelease `json:"releases"`
}

// InstallMariaDB performs the actual MariaDB installation
func InstallMariaDB(options InstallOptions, osInfo *OSInfo) (*InstallResult, error) {
	lg, _ := logger.Get()

	lg.Info("Starting MariaDB installation",
		logger.String("version", options.Version),
		logger.String("os", osInfo.ID))

	// Validate version for this OS
	if err := ValidateVersionForOS(options.Version, osInfo); err != nil {
		return nil, fmt.Errorf("version validation failed: %w", err)
	}

	if IsRHELBased(osInfo.ID) {
		return installRHELBased(options, osInfo)
	} else if IsDebianBased(osInfo.ID) {
		return installDebianBased(options, osInfo)
	}

	return nil, fmt.Errorf("unsupported operating system: %s", osInfo.ID)
}

// installRHELBased installs MariaDB on RHEL-based systems
func installRHELBased(options InstallOptions, osInfo *OSInfo) (*InstallResult, error) {
	lg, _ := logger.Get()

	result := &InstallResult{
		Version:         options.Version,
		Port:            options.Port,
		DataDir:         options.DataDir,
		LogDir:          options.LogDir,
		BinlogDir:       options.BinlogDir,
		OperatingSystem: osInfo.ID,
		Distribution:    fmt.Sprintf("%s %s", osInfo.Name, osInfo.Version),
	}

	// Try to install using REST API download first
	lg.Info("Attempting installation via MariaDB REST API download")
	if err := installFromRestAPI(options.Version, osInfo); err != nil {
		lg.Warn("REST API installation failed, trying native repository method", logger.Error(err))

		// For CentOS Stream 10, try native repository instead of MariaDB official repo
		if osInfo.ID == "centos" && osInfo.Version == "10" {
			lg.Info("Using native CentOS repository for better compatibility")
			if err := installFromNativeRepo(osInfo); err != nil {
				return result, fmt.Errorf("failed to install from native repository: %w", err)
			}
		} else {
			// Fallback to MariaDB official repository for other systems
			if err := setupMariaDBRepo(options.Version, osInfo); err != nil {
				return result, fmt.Errorf("failed to setup MariaDB repository: %w", err)
			}

			// Install packages
			if err := installRHELPackages(options.Version); err != nil {
				return result, fmt.Errorf("failed to install packages: %w", err)
			}
		}
	} // Configure MariaDB
	if err := configureMariaDB(options); err != nil {
		return result, fmt.Errorf("failed to configure MariaDB: %w", err)
	}

	// Start and enable service
	if err := startAndEnableService(); err != nil {
		return result, fmt.Errorf("failed to start service: %w", err)
	}

	result.Success = true
	result.ServiceStatus = "active"

	lg.Info("MariaDB installation completed successfully",
		logger.String("version", options.Version))

	return result, nil
}

// installFromRestAPI installs MariaDB using REST API download
func installFromRestAPI(version string, osInfo *OSInfo) error {
	lg, _ := logger.Get()

	lg.Info("Downloading MariaDB packages via REST API",
		logger.String("version", version),
		logger.String("os", osInfo.ID))

	// Get download URLs from REST API
	files, err := getMariaDBDownloadFiles(version, osInfo)
	if err != nil {
		return fmt.Errorf("failed to get download files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no suitable packages found for %s on %s", version, osInfo.ID)
	}

	// Download packages
	downloadDir := "/tmp/mariadb-packages"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}
	defer os.RemoveAll(downloadDir) // Cleanup

	var packagePaths []string
	for _, file := range files {
		lg.Info("Downloading package",
			logger.String("file", file.FileName),
			logger.String("url", file.FileDownloadURL))

		packagePath := filepath.Join(downloadDir, file.FileName)
		if err := downloadFile(file.FileDownloadURL, packagePath); err != nil {
			return fmt.Errorf("failed to download %s: %w", file.FileName, err)
		}

		// Verify checksum if available
		if file.Checksum["sha256sum"] != "" {
			if err := verifyChecksum(packagePath, file.Checksum["sha256sum"]); err != nil {
				lg.Warn("Checksum verification failed",
					logger.String("file", file.FileName),
					logger.Error(err))
			}
		}

		packagePaths = append(packagePaths, packagePath)
	}

	// Install packages using rpm/dnf
	return installRPMPackages(packagePaths)
}

// getMariaDBDownloadFiles gets download files from MariaDB REST API
func getMariaDBDownloadFiles(version string, osInfo *OSInfo) ([]MariaDBDownloadFile, error) {
	// Map OS to MariaDB API OS names
	osMap := map[string]string{
		"centos": "rhel",
		"rhel":   "rhel",
		"rocky":  "rhel",
		"alma":   "rhel",
	}

	apiOS := osMap[strings.ToLower(osInfo.ID)]
	if apiOS == "" {
		apiOS = strings.ToLower(osInfo.ID)
	}

	// Build API URL - for specific version and OS
	apiURL := fmt.Sprintf("https://downloads.mariadb.org/rest-api/mariadb/%s/?os=%s&cpu=x86_64",
		version, apiOS)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var response MariaDBDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Find the release and filter for RPM packages
	var allFiles []MariaDBDownloadFile
	for _, release := range response.Releases {
		for _, file := range release.Files {
			// Filter for RPM packages that we need
			if strings.Contains(strings.ToLower(file.PackageType), "rpm") &&
				(strings.Contains(strings.ToLower(file.FileName), "server") ||
					strings.Contains(strings.ToLower(file.FileName), "client") ||
					strings.Contains(strings.ToLower(file.FileName), "common")) {
				allFiles = append(allFiles, file)
			}
		}
	}

	return allFiles, nil
}

// downloadFile downloads a file from URL to local path
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// verifyChecksum verifies SHA256 checksum of a file
func verifyChecksum(filepath, expectedChecksum string) error {
	cmd := exec.Command("sha256sum", filepath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return fmt.Errorf("invalid checksum output")
	}

	actualChecksum := parts[0]
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// installRPMPackages installs RPM packages
func installRPMPackages(packagePaths []string) error {
	lg, _ := logger.Get()

	var cmd *exec.Cmd
	if CommandExists("dnf") {
		args := append([]string{"install", "-y"}, packagePaths...)
		cmd = exec.Command("dnf", args...)
	} else if CommandExists("yum") {
		args := append([]string{"localinstall", "-y"}, packagePaths...)
		cmd = exec.Command("yum", args...)
	} else if CommandExists("rpm") {
		args := append([]string{"-ivh"}, packagePaths...)
		cmd = exec.Command("rpm", args...)
	} else {
		return fmt.Errorf("no package manager found (dnf/yum/rpm)")
	}

	lg.Info("Installing RPM packages", logger.Strings("packages", packagePaths))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Package installation failed",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("package installation failed: %w", err)
	}

	lg.Info("RPM packages installed successfully")
	return nil
}

// installFromNativeRepo installs MariaDB from native OS repository
func installFromNativeRepo(osInfo *OSInfo) error {
	lg, _ := logger.Get()

	lg.Info("Installing MariaDB from native repository")

	// Use native package names for CentOS/RHEL
	packages := []string{
		"mariadb-server",
		"mariadb",
	}

	var cmd *exec.Cmd
	if CommandExists("dnf") {
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	} else if CommandExists("yum") {
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	} else {
		return fmt.Errorf("neither dnf nor yum found")
	}

	lg.Info("Installing native MariaDB packages", logger.Strings("packages", packages))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Native package installation failed",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("native package installation failed: %w", err)
	}

	lg.Info("Native MariaDB packages installed successfully")
	return nil
}

// installFromNativeRepo installs MariaDB from native repository (for CentOS Stream 10)
// installDebianBased installs MariaDB on Debian-based systems
func installDebianBased(options InstallOptions, osInfo *OSInfo) (*InstallResult, error) {
	lg, _ := logger.Get()

	result := &InstallResult{
		Version:         options.Version,
		Port:            options.Port,
		DataDir:         options.DataDir,
		LogDir:          options.LogDir,
		BinlogDir:       options.BinlogDir,
		OperatingSystem: osInfo.ID,
		Distribution:    fmt.Sprintf("%s %s", osInfo.Name, osInfo.Version),
	}

	// Update package list
	lg.Info("Updating package list")
	cmd := exec.Command("apt", "update")
	if err := cmd.Run(); err != nil {
		lg.Warn("Failed to update package list", logger.Error(err))
	}

	// Setup MariaDB repository
	if err := setupMariaDBRepoDebian(options.Version, osInfo); err != nil {
		return result, fmt.Errorf("failed to setup MariaDB repository: %w", err)
	}

	// Install packages
	if err := installDebianPackages(options.Version); err != nil {
		return result, fmt.Errorf("failed to install packages: %w", err)
	}

	// Configure MariaDB
	if err := configureMariaDB(options); err != nil {
		return result, fmt.Errorf("failed to configure MariaDB: %w", err)
	}

	// Start and enable service
	if err := startAndEnableService(); err != nil {
		return result, fmt.Errorf("failed to start service: %w", err)
	}

	result.Success = true
	result.ServiceStatus = "active"

	return result, nil
}

// setupMariaDBRepo sets up MariaDB repository for RHEL-based systems
func setupMariaDBRepo(version string, osInfo *OSInfo) error {
	lg, _ := logger.Get()

	series := GetVersionSeries(version)

	// For CentOS 10/Stream, use CentOS 9 repo as fallback since CentOS 10 might not be directly supported
	centosMajor := "9"
	if osInfo.Version == "8" {
		centosMajor = "8"
	} else if osInfo.Version == "7" {
		centosMajor = "7"
	}

	repoContent := fmt.Sprintf(`[mariadb]
name = MariaDB
baseurl = https://archive.mariadb.org/mariadb-%s/yum/centos/%s/x86_64
gpgkey = https://archive.mariadb.org/PublicKey
gpgcheck = 1
`, version, centosMajor)

	repoFile := "/etc/yum.repos.d/MariaDB.repo"

	lg.Info("Setting up MariaDB repository",
		logger.String("series", series),
		logger.String("repo_file", repoFile))

	if err := os.WriteFile(repoFile, []byte(repoContent), 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	// Import GPG key
	cmd := exec.Command("rpm", "--import", "https://archive.mariadb.org/PublicKey")
	if err := cmd.Run(); err != nil {
		lg.Warn("Failed to import GPG key", logger.Error(err))
	}

	return nil
}

// setupMariaDBRepoDebian sets up MariaDB repository for Debian-based systems
func setupMariaDBRepoDebian(version string, osInfo *OSInfo) error {
	lg, _ := logger.Get()

	series := GetVersionSeries(version)

	// Install necessary packages
	cmd := exec.Command("apt", "install", "-y", "software-properties-common", "dirmngr", "apt-transport-https")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install required packages: %w", err)
	}

	// Add MariaDB APT key
	cmd = exec.Command("apt-key", "adv", "--fetch-keys", "https://mariadb.org/mariadb_release_signing_key.asc")
	if err := cmd.Run(); err != nil {
		lg.Warn("Failed to add MariaDB APT key", logger.Error(err))
	}

	// Add repository
	repoURL := fmt.Sprintf("deb [arch=amd64] https://archive.mariadb.org/mariadb-%s/repo/%s/ %s main", version, osInfo.ID, osInfo.Codename)

	lg.Info("Adding MariaDB repository",
		logger.String("series", series),
		logger.String("repo_url", repoURL))

	cmd = exec.Command("add-apt-repository", "-y", repoURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add MariaDB repository: %w", err)
	}

	// Update package list
	cmd = exec.Command("apt", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update package list after adding repo: %w", err)
	}

	return nil
}

// installRHELPackages installs MariaDB packages on RHEL systems
func installRHELPackages(version string) error {
	lg, _ := logger.Get()

	packages := []string{
		"MariaDB-server",
		"MariaDB-client",
		"MariaDB-common",
	}

	var cmd *exec.Cmd
	if CommandExists("dnf") {
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	} else if CommandExists("yum") {
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	} else {
		return fmt.Errorf("neither dnf nor yum found")
	}

	lg.Info("Installing MariaDB packages", logger.Strings("packages", packages))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Package installation failed",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("package installation failed: %w", err)
	}

	lg.Info("MariaDB packages installed successfully")
	return nil
}

// installDebianPackages installs MariaDB packages on Debian systems
func installDebianPackages(version string) error {
	lg, _ := logger.Get()

	packages := []string{
		"mariadb-server",
		"mariadb-client",
		"mariadb-common",
	}

	cmd := exec.Command("apt", append([]string{"install", "-y"}, packages...)...)

	lg.Info("Installing MariaDB packages", logger.Strings("packages", packages))

	output, err := cmd.CombinedOutput()
	if err != nil {
		lg.Error("Package installation failed",
			logger.Error(err),
			logger.String("output", string(output)))
		return fmt.Errorf("package installation failed: %w", err)
	}

	lg.Info("MariaDB packages installed successfully")
	return nil
}

// configureMariaDB configures MariaDB with custom settings
func configureMariaDB(options InstallOptions) error {
	lg, _ := logger.Get()

	lg.Info("Configuring MariaDB",
		logger.Int("port", options.Port),
		logger.String("data_dir", options.DataDir))

	// Create directories if they don't exist
	dirs := []string{options.DataDir, options.LogDir, options.BinlogDir}
	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			// Set ownership to mysql user
			cmd := exec.Command("chown", "-R", "mysql:mysql", dir)
			if err := cmd.Run(); err != nil {
				lg.Warn("Failed to set ownership for directory",
					logger.String("directory", dir),
					logger.Error(err))
			}
		}
	}

	return nil
}

// startAndEnableService starts and enables MariaDB service
func startAndEnableService() error {
	lg, _ := logger.Get()

	// Create systemd environment file to fix environment variables issue
	lg.Info("Creating systemd environment file")
	envContent := `MYSQLD_OPTS=""
_WSREP_NEW_CLUSTER=""
`
	envFile := "/etc/default/mariadb"
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		lg.Warn("Failed to create environment file", logger.Error(err))
	}

	// Create systemd override directory
	overrideDir := "/etc/systemd/system/mariadb.service.d"
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		lg.Warn("Failed to create systemd override directory", logger.Error(err))
	} else {
		// Create override configuration
		overrideContent := `[Service]
EnvironmentFile=-/etc/default/mariadb
Environment="MYSQLD_OPTS="
Environment="_WSREP_NEW_CLUSTER="
`
		overrideFile := filepath.Join(overrideDir, "override.conf")
		if err := os.WriteFile(overrideFile, []byte(overrideContent), 0644); err != nil {
			lg.Warn("Failed to create systemd override file", logger.Error(err))
		} else {
			lg.Info("Created systemd override configuration")
		}
	}

	// Reload systemd daemon
	lg.Info("Reloading systemd daemon")
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		lg.Warn("Failed to reload systemd daemon", logger.Error(err))
	}

	// Start MariaDB service
	lg.Info("Starting MariaDB service")
	cmd := exec.Command("systemctl", "start", "mariadb")
	if err := cmd.Run(); err != nil {
		// Try mysql service name
		cmd = exec.Command("systemctl", "start", "mysql")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start MariaDB service: %w", err)
		}
	}

	// Enable MariaDB service
	lg.Info("Enabling MariaDB service")
	cmd = exec.Command("systemctl", "enable", "mariadb")
	if err := cmd.Run(); err != nil {
		// Try mysql service name
		cmd = exec.Command("systemctl", "enable", "mysql")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to enable MariaDB service: %w", err)
		}
	}

	lg.Info("MariaDB service started and enabled")
	return nil
}
