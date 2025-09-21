package install

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sfDBTools/internal/logger"
	mariadb_config "sfDBTools/utils/mariadb/config"
	defaultsetup "sfDBTools/utils/mariadb/defaultSetup"
	"sfDBTools/utils/terminal"
)

// setupMariaDBRepository mengunduh dan menjalankan script setup repository
func setupMariaDBRepository(ctx context.Context, cfg *mariadb_config.MariaDBInstallConfig, deps *defaultsetup.Dependencies) error {
	lg, _ := logger.Get()
	terminal.PrintSubHeader("[Repository] Setup")
	lg.Info("[Repository] Setup Start")

	// 1) Detect existing repo files
	checkSpinner := terminal.NewInstallSpinner("Mengecek repository MariaDB yang ada...")
	checkSpinner.Start()
	found, err := detectExistingMariaDBRepo()
	if err != nil {
		checkSpinner.StopWithError("Gagal mengecek repository yang ada")
		lg.Debug("gagal mengecek repo yang ada", logger.Error(err))
		return fmt.Errorf("gagal mengecek repository yang ada: %w", err)
	}

	// If repo files exist, check whether they already match desired version
	normalized := normalizeVersionForRepo(cfg.Version)
	if len(found) > 0 {
		if repoFilesContainVersion(found, normalized) {
			checkSpinner.StopWithSuccess("Repository MariaDB sudah sesuai, tidak perlu setup")
			lg.Info("[Repository] Repo sudah sesuai versi yang akan diinstall, melewatkan setup")
			return nil
		}

		// Need to backup/cleanup existing repo files before proceeding
		checkSpinner.StopWithSuccess("Repository lama ditemukan — akan dibackup sebelum setup")
		backupSpinner := terminal.NewInstallSpinner("Membackup repository MariaDB yang lama...")
		backupSpinner.Start()
		backupDir, berr := backupRepoFiles(found)
		if berr != nil {
			backupSpinner.StopWithError("Gagal membackup repository lama")
			lg.Debug("gagal membackup repo", logger.Error(berr))
			return fmt.Errorf("gagal membackup repository lama: %w", berr)
		}
		backupSpinner.StopWithSuccess("Repository lama dibackup: " + backupDir)
		lg.Info("[Repository] Repo lama dibackup", logger.String("backup_dir", backupDir))
	} else {
		checkSpinner.StopWithSuccess("Tidak ditemukan repository MariaDB yang lama")
	}

	// 2) Check if mariadb_repo_setup script already exists on the system
	scriptPath := findExistingRepoSetupScript()
	if scriptPath != "" {
		lg.Info("Menemukan script mariadb_repo_setup yang sudah ada", logger.String("path", scriptPath))
	} else {
		// Download mariadb_repo_setup script (show spinner for the download)
		dlSpinner := terminal.NewDownloadSpinner("Mengunduh script setup repository...")
		dlSpinner.Start()
		scriptPath, err = downloadRepoSetupScript(ctx)
		if err != nil {
			dlSpinner.StopWithError("Gagal mengunduh script setup repository")
			lg.Debug("gagal mengunduh script setup repository", logger.Error(err))
			return fmt.Errorf("gagal mengunduh script setup repository: %w", err)
		}
		dlSpinner.StopWithSuccess("Script setup repository berhasil diunduh")
		defer os.Remove(scriptPath)

		// Buat permission executable
		if err := os.Chmod(scriptPath, 0755); err != nil {
			return fmt.Errorf("gagal mengubah permission script: %w", err)
		}
	}

	// 3) Jalankan script dengan parameter yang sesuai
	args := buildRepoSetupArgs(cfg)

	runSpinner := terminal.NewInstallSpinner("Menjalankan script setup repository...")
	runSpinner.Start()
	if err := deps.ProcessManager.ExecuteWithTimeout("bash", append([]string{scriptPath}, args...), 5*time.Minute); err != nil {
		runSpinner.StopWithError("Gagal menjalankan script setup repository")
		lg.Debug("[Repository] gagal menjalankan script setup", logger.Error(err))
		return fmt.Errorf("gagal menjalankan script setup repository: %w", err)
	}
	runSpinner.StopWithSuccess("Script setup repository berhasil dijalankan")
	lg.Info("[Repository] Setup selesai")

	return nil
}

// detectExistingMariaDBRepo mencari file repository MariaDB yang mungkin ada
func detectExistingMariaDBRepo() ([]string, error) {
	patterns := []string{
		"/etc/yum.repos.d/MariaDB.repo*",
		"/etc/yum.repos.d/mariadb.repo*",
		"/etc/apt/sources.list.d/mariadb.list*",
		"/etc/apt/sources.list.d/MariaDB.list*",
		"/etc/apt/trusted.gpg.d/mariadb*",
		"/etc/apt/trusted.gpg.d/MariaDB*",
	}

	var found []string
	for _, p := range patterns {
		matches, _ := filepath.Glob(p)
		for _, m := range matches {
			if fi, err := os.Stat(m); err == nil && fi.Mode().IsRegular() {
				found = append(found, m)
			}
		}
	}

	return found, nil
}

// backupRepoFiles memindahkan file repo yang ditemukan ke folder backup yang terpusat
func backupRepoFiles(files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	base := "backups/repo_backups"
	if err := os.MkdirAll(base, 0755); err != nil {
		// fallback ke /tmp jika tidak bisa membuat di /var/backups
		base = "/tmp/sfdbtools_backups"
		if err := os.MkdirAll(base, 0755); err != nil {
			return "", fmt.Errorf("gagal membuat direktori backup: %w", err)
		}
	}

	ts := time.Now().Format("20060102-150405")
	dst := filepath.Join(base, "mariadb-repo-"+ts)
	if err := os.MkdirAll(dst, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori backup: %w", err)
	}

	for _, f := range files {
		fname := filepath.Base(f)
		dstPath := filepath.Join(dst, fname)
		if err := os.Rename(f, dstPath); err != nil {
			// jika rename gagal (mis. cross-filesystem), coba copy
			in, err := os.Open(f)
			if err != nil {
				return dst, fmt.Errorf("gagal membuka file %s: %w", f, err)
			}
			out, err := os.Create(dstPath)
			if err != nil {
				in.Close()
				return dst, fmt.Errorf("gagal membuat backup file %s: %w", dstPath, err)
			}
			if _, err := io.Copy(out, in); err != nil {
				in.Close()
				out.Close()
				return dst, fmt.Errorf("gagal menyalin file %s ke %s: %w", f, dstPath, err)
			}
			in.Close()
			out.Close()
			if err := os.Remove(f); err != nil {
				return dst, fmt.Errorf("gagal menghapus sumber %s setelah backup: %w", f, err)
			}
		}
	}

	return dst, nil
}

// repoFilesContainVersion memeriksa apakah salah satu file repo berisi versi yang cocok
func repoFilesContainVersion(files []string, normalizedVersion string) bool {
	if normalizedVersion == "" {
		return false
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), normalizedVersion) {
			return true
		}
	}
	return false
}

// findExistingRepoSetupScript mencari apakah ada binary/script mariadb_repo_setup yang bisa dipakai
func findExistingRepoSetupScript() string {
	candidates := []string{"/usr/local/bin/mariadb_repo_setup", "/usr/bin/mariadb_repo_setup", "/bin/mariadb_repo_setup"}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.Mode().IsRegular() {
			return c
		}
	}
	return ""
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
func buildRepoSetupArgs(cfg *mariadb_config.MariaDBInstallConfig) []string {
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
