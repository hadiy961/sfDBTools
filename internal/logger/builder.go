package logger

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/config/model"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// buildLogger membuat dan mengkonfigurasi instance Logger dari konfigurasi yang diberikan
func buildLogger(cfg *model.Config) (*Logger, error) {
	c := cfg.Log
	level := parseLevel(c.Level)
	// Buat encoder berdasarkan konfigurasi
	enc := createEncoder(&c)

	// Siapkan semua cores yang dibutuhkan
	cores := []zapcore.Core{}

	if c.Output.Console {
		cores = append(cores, zapcore.NewCore(enc, zapcore.Lock(os.Stdout), level))
	}

	if c.Output.File {
		writer, err := newRotatingFileWriter(
			c.File.Dir,
			cfg.General.AppName,
			c.Timezone,
			c.File.RotateDaily,
			c.File.RetentionDays,
		)
		if err != nil {
			return nil, fmt.Errorf("init file writer: %w", err)
		}
		cores = append(cores, zapcore.NewCore(enc, writer, level))
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no log output configured")
	}

	// Gabungkan semua cores
	core := zapcore.NewTee(cores...)

	// Tambahkan wrapper core untuk format output
	if strings.ToLower(c.Format) != "json" {
		core = &messageHyphenCore{core}
	}

	// Siapkan opsi logger
	opts := []zap.Option{}

	// Tambahkan caller hanya jika level debug
	if strings.ToLower(c.Level) == "debug" || strings.ToLower(c.Level) == "trace" {
		opts = append(opts, zap.AddCaller())
		opts = append(opts, zap.AddCallerSkip(1))
	}

	// Buat logger
	z := zap.New(core, opts...)
	return &Logger{z}, nil
}

// parseLevel mengkonversi string level log ke zapcore.Level
func parseLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "trace", "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
