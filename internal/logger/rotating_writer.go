package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// rotatingFileWriter adalah implementasi writer log yang mendukung rotasi file
type rotatingFileWriter struct {
	mu            sync.Mutex
	dir           string
	appName       string
	timezone      string
	rotateDaily   bool
	retentionDays int
	currentDate   string
	file          *os.File
}

// newRotatingFileWriter membuat writer log dengan kemampuan rotasi
func newRotatingFileWriter(dir, appName, tz string, rotate bool, retention int) (*rotatingFileWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	w := &rotatingFileWriter{
		dir:           dir,
		appName:       appName,
		timezone:      tz,
		rotateDaily:   rotate,
		retentionDays: retention,
	}
	if err := w.rotate(); err != nil {
		return nil, err
	}
	w.cleanup()
	return w, nil
}

// now mengembalikan waktu saat ini dalam timezone yang dikonfigurasi
func (w *rotatingFileWriter) now() time.Time {
	if loc, err := time.LoadLocation(w.timezone); err == nil {
		return time.Now().In(loc)
	}
	return time.Now()
}

// filename mengembalikan nama file berdasarkan tanggal jika rotasi diaktifkan
func (w *rotatingFileWriter) filename(date string) string {
	if w.rotateDaily {
		return filepath.Join(w.dir, fmt.Sprintf("%s_%s.log", w.appName, date))
	}
	return filepath.Join(w.dir, fmt.Sprintf("%s.log", w.appName))
}

// rotate melakukan rotasi file log jika diperlukan
func (w *rotatingFileWriter) rotate() error {
	now := w.now()
	date := ""
	if w.rotateDaily {
		date = now.Format("2006_01_02")
		if date == w.currentDate && w.file != nil {
			return nil
		}
	} else if w.file != nil {
		return nil
	}

	if w.file != nil {
		w.file.Close()
	}

	f, err := os.OpenFile(w.filename(date), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.currentDate = date
	w.cleanup()
	return nil
}

// cleanup menghapus file log lama berdasarkan konfigurasi retensi
func (w *rotatingFileWriter) cleanup() {
	if w.retentionDays <= 0 {
		return
	}
	threshold := w.now().AddDate(0, 0, -w.retentionDays)
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, w.appName) || !strings.HasSuffix(name, ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(threshold) {
			os.Remove(filepath.Join(w.dir, name))
		}
	}
}

// Write mengimplementasikan io.Writer
func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.rotate(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

// Sync mengimplementasikan zapcore.WriteSyncer
func (w *rotatingFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}
