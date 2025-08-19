package validate

import (
	"fmt"
	"os"
	"time"
)

// Cek apakah item ada dalam slice
func InSlice(val string, list []string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

// Validasi timezone
func IsValidTimezone(name string) error {
	if _, err := time.LoadLocation(name); err != nil {
		return err
	}
	return nil
}

// Cek apakah direktori valid dan writable
func DirExistsAndWritable(path string) error {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		if mkErr := os.MkdirAll(path, 0o755); mkErr != nil {
			return fmt.Errorf("tidak ditemukan atau tidak bisa diakses: %w", mkErr)
		}
		stat, err = os.Stat(path)
	}
	if err != nil {
		return fmt.Errorf("tidak ditemukan atau tidak bisa diakses: %w", err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("bukan direktori: %s", path)
	}

	// Cek izin tulis
	testFile := path + "/.writetest"
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("tidak bisa tulis ke direktori: %w", err)
	}
	defer os.Remove(testFile)
	defer f.Close()

	return nil
}
