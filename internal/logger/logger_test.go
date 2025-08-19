package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func writeConfig(dir, logDir string, rotate bool) error {
	cfg := fmt.Sprintf(`general:
  client_code: "123"
  app_name: "sfDBTools"
  version: "1.0.0"
  author: "Hadiyatna Muflihun"
log:
  level: "info"
  format: "json"
  timezone: "UTC"
  output:
    console: false
    file: true
    syslog: false
  file:
    dir: "%s"
    rotate_daily: %t
    retention_days: 1
`, logDir, rotate)
	return os.WriteFile(filepath.Join(dir, "config", "config.yaml"), []byte(cfg), 0o644)
}

func reset() {
	lg = nil
	once = sync.Once{}
}

func TestFileRotation(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "config"), 0o755)
	logDir := filepath.Join(dir, "logs")
	os.Mkdir(logDir, 0o755)

	if err := writeConfig(dir, logDir, true); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldTime := time.Now().AddDate(0, 0, -2)
	oldName := filepath.Join(logDir, fmt.Sprintf("sfDBTools_%s.log", oldTime.Format("2006_01_02")))
	os.WriteFile(oldName, []byte("old"), 0o644)
	os.Chtimes(oldName, oldTime, oldTime)

	oldwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldwd)

	reset()
	lg, err := Get()
	if err != nil {
		t.Fatalf("get logger: %v", err)
	}
	lg.Info("test")
	lg.Sync()

	todayFile := filepath.Join(logDir, fmt.Sprintf("sfDBTools_%s.log", time.Now().UTC().Format("2006_01_02")))
	if _, err := os.Stat(todayFile); err != nil {
		t.Fatalf("expected log file %s: %v", todayFile, err)
	}
	if _, err := os.Stat(oldName); !os.IsNotExist(err) {
		t.Fatalf("old file not removed")
	}
}
