package install

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/mariadb"
	"sfDBTools/utils/terminal"
)

// setupMariaDBRepository mengunduh dan menjalankan script setup repository
func setupMariaDBRepository(ctx context.Context, cfg *mariadb.MariaDBInstallConfig, deps *Dependencies) error {
	lg, _ := logger.Get()

	terminal.SafePrintln("📦 Menyiapkan repository MariaDB...")

	// Download mariadb_repo_setup script
	scriptPath, err := downloadRepoSetupScript(ctx)
	if err != nil {
		return fmt.Errorf("gagal mengunduh script setup repository: %w", err)
	}
	defer os.Remove(scriptPath)

	// Buat permission executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("gagal mengubah permission script: %w", err)
	}

	// Jalankan script dengan parameter yang sesuai
	args := buildRepoSetupArgs(cfg)

	lg.Info("Menjalankan mariadb_repo_setup script", logger.Strings("args", args))

	if err := deps.ProcessManager.ExecuteWithTimeout("bash", append([]string{scriptPath}, args...), 5*time.Minute); err != nil {
		return fmt.Errorf("gagal menjalankan script setup repository: %w", err)
	}

	lg.Info("Repository MariaDB berhasil dikonfigurasi")
	return nil
}

// downloadRepoSetupScript mengunduh script setup repository ke file temporary
func downloadRepoSetupScript(ctx context.Context) (string, error) {
	url := "https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal mengunduh script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gagal mengunduh script, status code: %d", resp.StatusCode)
	}

	// Simpan ke file temporary
	tmpFile, err := os.CreateTemp("", "mariadb_repo_setup_*.sh")
	if err != nil {
		return "", fmt.Errorf("gagal membuat file temporary: %w", err)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("gagal menyimpan script: %w", err)
	}

	return tmpFile.Name(), nil
}

// buildRepoSetupArgs membangun argumen untuk script setup repository
func buildRepoSetupArgs(cfg *mariadb.MariaDBInstallConfig) []string {
	args := []string{}

	// Tambahkan versi MariaDB (normalisasi ke major.minor karena skrip repo tidak selalu
	// menerima patch version seperti 10.6.23 — skrip biasanya menerima 10.6 atau 11.4)
	normalized := normalizeVersionForRepo(cfg.Version)
	args = append(args, "--mariadb-server-version="+normalized)

	// Skip MaxScale (tidak diperlukan untuk instalasi dasar)
	args = append(args, "--skip-maxscale")

	return args
}

// normalizeVersionForRepo mengubah versi lengkap (mis. 10.6.23) menjadi major.minor (10.6)
// Skrip resmi `mariadb_repo_setup` terkadang menolak patch-level releases; gunakan format
// mayor.minor saat memanggil script.
func normalizeVersionForRepo(version string) string {
	if version == "" {
		return version
	}

	// Split into at most 3 parts: major, minor, patch
	parts := strings.SplitN(version, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	// if version doesn't contain a minor part, return as-is
	return version
}
