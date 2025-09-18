package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FormatSize memformat ukuran byte menjadi string yang mudah dibaca
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsHidden mengecek apakah file/direktori hidden berdasarkan platform
func IsHidden(name string) bool {
	if runtime.GOOS == "windows" {
		// Di Windows, file hidden dimulai dengan '.' (simplified)
		return strings.HasPrefix(name, ".")
	}
	// Di Unix-like systems, file hidden dimulai dengan '.'
	return strings.HasPrefix(name, ".")
}

// ValidatePath melakukan validasi dasar path
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}

	if strings.Contains(path, "..") {
		return fmt.Errorf("path tidak boleh mengandung '..'")
	}

	if len(path) > 4096 {
		return fmt.Errorf("path terlalu panjang")
	}

	return nil
}

// NormalizePath menormalisasi path untuk cross-platform compatibility
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// IsSpecialFile mengecek apakah file adalah special file (socket, device, pipe)
func IsSpecialFile(mode os.FileMode) bool {
	return mode&(os.ModeSocket|os.ModeDevice|os.ModeNamedPipe) != 0
}

// GenerateUniqueFileName membuat nama file unik dengan timestamp
func GenerateUniqueFileName(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", prefix, timestamp)
}

// SplitPath memisahkan path menjadi directory dan filename
func SplitPath(path string) (dir, file string) {
	return filepath.Split(path)
}

// JoinPath menggabungkan path elements dengan separator yang benar
func JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// GetFileExtension mendapatkan ekstensi file (dengan dot)
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// StripExtension menghilangkan ekstensi dari filename
func StripExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return filename
	}
	return filename[:len(filename)-len(ext)]
}

// IsAbsolutePath mengecek apakah path adalah absolute path
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// GetAbsolutePath mengconvert relative path ke absolute path
func GetAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

// GetRelativePath mendapatkan relative path dari base ke target
func GetRelativePath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// SafeJoin melakukan path join dengan validasi keamanan
func SafeJoin(base, path string) (string, error) {
	if err := ValidatePath(path); err != nil {
		return "", err
	}

	joined := filepath.Join(base, path)

	// Pastikan hasil masih dalam base directory
	cleanBase := filepath.Clean(base)
	cleanJoined := filepath.Clean(joined)

	if !strings.HasPrefix(cleanJoined, cleanBase) {
		return "", fmt.Errorf("path keluar dari base directory")
	}

	return cleanJoined, nil
}
