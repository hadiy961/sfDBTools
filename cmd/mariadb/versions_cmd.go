package command_mariadb

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	mariadb_utils "sfDBTools/utils/mariadb"

	"github.com/spf13/cobra"
)

var VersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Show supported MariaDB versions",
	Long: `Display a comprehensive list of supported MariaDB versions organized by series and stability status.
	
The version data is dynamically fetched from MariaDB official sources and cached for performance.
Use --refresh flag to force update the version cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if refresh is requested
		refresh, _ := cmd.Flags().GetBool("refresh")
		if refresh {
			fmt.Println("🔄 Refreshing MariaDB version data...")
			mariadb_utils.RefreshVersionCache()
			fmt.Println("✅ Version data refreshed!")
			fmt.Println()
		}

		return showSupportedVersions()
	},
}

func init() {
	// Add refresh flag
	VersionsCmd.Flags().Bool("refresh", false, "Force refresh of version data from MariaDB sources")
}

// getCurrentInstalledVersion tries to detect currently installed MariaDB version
func getCurrentInstalledVersion() (string, error) {
	// Method 1: Try mysql command
	if output, err := exec.Command("mysql", "--version").Output(); err == nil {
		versionStr := string(output)
		// Extract version from output like: "mysql  Ver 15.1 Distrib 10.11.11-MariaDB"
		re := regexp.MustCompile(`(\d+\.\d+\.\d+)-MariaDB`)
		if matches := re.FindStringSubmatch(versionStr); len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Method 2: Try mariadb command
	if output, err := exec.Command("mariadb", "--version").Output(); err == nil {
		versionStr := string(output)
		re := regexp.MustCompile(`(\d+\.\d+\.\d+)-MariaDB`)
		if matches := re.FindStringSubmatch(versionStr); len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Method 3: Try mysqladmin version (if accessible without password)
	if output, err := exec.Command("mysqladmin", "--version").Output(); err == nil {
		versionStr := string(output)
		re := regexp.MustCompile(`(\d+\.\d+\.\d+)-MariaDB`)
		if matches := re.FindStringSubmatch(versionStr); len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Method 4: Check service status for version info
	if output, err := exec.Command("systemctl", "status", "mariadb").Output(); err == nil {
		statusStr := string(output)
		if strings.Contains(statusStr, "MariaDB") {
			re := regexp.MustCompile(`MariaDB (\d+\.\d+)`)
			if matches := re.FindStringSubmatch(statusStr); len(matches) > 1 {
				return matches[1] + ".x", nil
			}
		}
	}

	return "", fmt.Errorf("no installed MariaDB found")
}

// getInstallationSource tries to determine installation source
func getInstallationSource() string {
	// Check if external MariaDB repo exists
	if output, err := exec.Command("dnf", "repolist", "--enabled").Output(); err == nil {
		repoList := string(output)
		if strings.Contains(strings.ToLower(repoList), "mariadb") {
			return "External MariaDB Repository"
		}
	}

	// Check if installed via native packages
	if output, err := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}-%{VERSION}\\n", "mariadb-server").Output(); err == nil {
		pkgInfo := string(output)
		if strings.Contains(pkgInfo, "mariadb-server") && !strings.Contains(pkgInfo, "MariaDB") {
			return "Native CentOS Repository"
		}
	}

	return "Unknown"
}

func showSupportedVersions() error {
	fmt.Println()
	fmt.Println("📋 SUPPORTED MARIADB VERSIONS")
	fmt.Println("===============================")

	// Show currently installed version if available
	if currentVersion, err := getCurrentInstalledVersion(); err == nil {
		source := getInstallationSource()
		fmt.Printf("🔧 Currently Installed: MariaDB %s\n", currentVersion)
		fmt.Printf("📦 Installation Source: %s\n", source)
		fmt.Println()
	}

	// Step 1: Get versions from the new dynamic version management system (this will check connectivity and fetch from API)
	versions := mariadb_utils.GetSupportedVersions()

	// Show cache information
	source, lastUpdated, isExpired := mariadb_utils.GetCacheInfo()
	status := "✅ Valid"
	if isExpired {
		status = "⚠️ Expired"
	}
	fmt.Printf("📡 Version Data Source: %s (%s)\n", source, status)
	if !lastUpdated.IsZero() {
		fmt.Printf("🕐 Last Updated: %s\n", lastUpdated.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Stable versions (recommended)
	fmt.Println("✅ Stable Versions (Recommended):")
	for series, versionList := range versions.StableVersions {
		fmt.Printf("  %s: %v\n", series, versionList)
	}

	fmt.Println()

	// Other versions
	fmt.Println("🔄 Other Versions:")
	for series, versionList := range versions.OtherVersions {
		fmt.Printf("  %s: %v\n", series, versionList)
	}

	fmt.Println()

	// Step 2: Show latest and recommended versions (these will use cached data)
	latestVersion := mariadb_utils.GetLatestVersionWithCache(versions)
	fmt.Printf("💡 Latest Stable: %s\n", latestVersion)

	// Step 3: Detect OS locally (without connectivity check) for recommendation
	if osInfo, err := mariadb_utils.DetectOS(); err == nil {
		recommended := mariadb_utils.GetRecommendedVersion(osInfo)
		if recommended != latestVersion {
			fmt.Printf("💡 Recommended for %s %s: %s\n", osInfo.Name, osInfo.Version, recommended)
		}
	}

	fmt.Println()

	// Usage examples - make them dynamic
	fmt.Println("💡 Usage Examples:")
	fmt.Printf("  sfDBTools mariadb install --interactive\n")
	if latestVersion != "" {
		fmt.Printf("  sfDBTools mariadb install --version %s\n", latestVersion)
	}

	// Show examples from different series dynamically
	exampleCount := 0
	for _, versionList := range versions.StableVersions {
		if len(versionList) > 0 && exampleCount < 2 {
			fmt.Printf("  sfDBTools mariadb install --version %s\n", versionList[len(versionList)-1])
			exampleCount++
		}
	}
	fmt.Println()

	// Dynamic notes based on available versions
	fmt.Println("📌 Notes:")
	seriesInfo := mariadb_utils.GetVersionSeriesInfoWithCache(versions)

	// Show information for each stable series
	for series, info := range seriesInfo {
		if versionList, exists := versions.StableVersions[series]; exists && len(versionList) > 0 {
			fmt.Printf("  • %s series: %s\n", series, info)
		}
	}

	// OS-specific recommendations
	if osInfo, err := mariadb_utils.DetectOS(); err == nil {
		osRecommendations := mariadb_utils.GetOSSpecificRecommendation(osInfo)
		if primaryRec, exists := osRecommendations["primary"]; exists {
			fmt.Printf("  • For %s %s: %s recommended\n", osInfo.Name, osInfo.Version, primaryRec)
			if reason, exists := osRecommendations["reason"]; exists {
				fmt.Printf("    (%s)\n", reason)
			}
		}
	}
	fmt.Println()

	// Show refresh option
	fmt.Println("🔄 To refresh version data:")
	fmt.Println("  sfDBTools mariadb versions --refresh")

	return nil
}
