package logger

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"sfDBTools/internal/config"
	"sfDBTools/internal/config/model"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *logrus.Logger
	once   sync.Once
)

// Logger wraps logrus.Logger untuk konsistensi interface
type Logger struct {
	*logrus.Logger
}

// Field represents a structured log field
type Field struct {
	Key   string
	Value interface{}
}

// Common field constructors
func String(key, val string) Field            { return Field{Key: key, Value: val} }
func Strings(key string, vals []string) Field { return Field{Key: key, Value: vals} }
func Int(key string, val int) Field           { return Field{Key: key, Value: val} }
func Int64(key string, val int64) Field       { return Field{Key: key, Value: val} }
func Float64(key string, val float64) Field   { return Field{Key: key, Value: val} }
func Bool(key string, val bool) Field         { return Field{Key: key, Value: val} }
func Error(err error) Field                   { return Field{Key: "error", Value: err} }

// Convert our Fields to logrus.Fields
func fieldsToLogrusFields(fields []Field) logrus.Fields {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}
	return logrusFields
}

// Logging methods
func (l *Logger) Debug(msg string, fields ...Field) {
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Debug(msg)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Info(msg)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Warn(msg)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Error(msg)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Fatal(msg)
}

func (l *Logger) Sync() error {
	// Logrus doesn't have explicit sync, but we can flush if needed
	return nil
}

// Get returns a singleton Logger configured using config package
func Get() (*Logger, error) {
	var err error
	once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("logger initialization panicked: %v", r)
				// Create a minimal logger as fallback
				logger = logrus.New()
				logger.SetLevel(logrus.InfoLevel)
			}
		}()

		var cfg *model.Config
		cfg, err = config.LoadConfig()
		if err != nil {
			// Create fallback logger if config loading fails
			logger = logrus.New()
			logger.SetLevel(logrus.InfoLevel)
			logger.Warn("Failed to load config, using default logger settings")
			return
		}

		logger, err = buildLogger(cfg)
	})

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &Logger{Logger: logger}, err
}

// buildLogger creates and configures a logrus logger from configuration
func buildLogger(cfg *model.Config) (*logrus.Logger, error) {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(strings.ToLower(cfg.Log.Level))
	if err != nil {
		level = logrus.InfoLevel
		log.Warnf("Invalid log level '%s', using 'info'", cfg.Log.Level)
	}
	log.SetLevel(level)

	// Set formatter based on format config
	switch strings.ToLower(cfg.Log.Format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text", "":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Setup outputs
	var writers []io.Writer

	// Console output
	if cfg.Log.Output.Console.Enabled {
		writers = append(writers, os.Stdout)
	}

	// File output
	if cfg.Log.Output.File.Enabled {
		fileWriter, err := setupFileOutput(cfg)
		if err != nil {
			log.Warnf("Failed to setup file output: %v", err)
		} else {
			writers = append(writers, fileWriter)
		}
	}

	// Syslog output
	if cfg.Log.Output.Syslog.Enabled {
		syslogWriter, err := setupSyslogOutput(cfg)
		if err != nil {
			log.Warnf("Failed to setup syslog output: %v", err)
		} else {
			writers = append(writers, syslogWriter)
		}
	}

	// Set output writers
	if len(writers) > 0 {
		log.SetOutput(io.MultiWriter(writers...))
	} else {
		// Default to stdout if no outputs configured
		log.SetOutput(os.Stdout)
	}

	return log, nil
}

// setupFileOutput configures file-based logging with rotation
func setupFileOutput(cfg *model.Config) (io.Writer, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.Log.Output.File.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Parse max size
	maxSize := parseSize(cfg.Log.Output.File.Rotation.MaxSize)

	// Generate filename from pattern
	filename := generateFilename(cfg.Log.Output.File.FilenamePattern)
	fullPath := filepath.Join(cfg.Log.Output.File.Dir, filename)

	// Setup lumberjack for file rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fullPath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: cfg.Log.Output.File.Rotation.RetentionDays,
		MaxAge:     cfg.Log.Output.File.Rotation.RetentionDays, // days
		Compress:   cfg.Log.Output.File.Rotation.CompressOld,
	}

	return lumberjackLogger, nil
}

// setupSyslogOutput configures syslog output
func setupSyslogOutput(cfg *model.Config) (io.Writer, error) {
	priority := syslog.LOG_INFO
	facility := parseSyslogFacility(cfg.Log.Output.Syslog.Facility)

	syslogWriter, err := syslog.New(facility|priority, cfg.Log.Output.Syslog.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}

	return syslogWriter, nil
}

// Helper functions
func parseSize(sizeStr string) int {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	if sizeStr == "" {
		return 100 // default 100MB
	}

	// Extract number and unit
	var numStr string
	var unit string

	for i, char := range sizeStr {
		if char >= '0' && char <= '9' || char == '.' {
			numStr += string(char)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	size, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 100 // default on error
	}

	switch unit {
	case "KB":
		return int(size / 1024) // Convert to MB
	case "MB", "":
		return int(size)
	case "GB":
		return int(size * 1024)
	default:
		return int(size) // Assume MB
	}
}

func generateFilename(pattern string) string {
	now := time.Now()

	filename := pattern
	filename = strings.ReplaceAll(filename, "{date}", now.Format("2006-01-02"))
	filename = strings.ReplaceAll(filename, "{datetime}", now.Format("2006-01-02_15-04-05"))
	filename = strings.ReplaceAll(filename, "{timestamp}", strconv.FormatInt(now.Unix(), 10))

	// If no pattern variables found, use default
	if filename == pattern && !strings.Contains(pattern, ".log") {
		filename = "sfDBTools_" + now.Format("2006-01-02") + ".log"
	}

	return filename
}

func parseSyslogFacility(facility string) syslog.Priority {
	switch strings.ToLower(facility) {
	case "kern":
		return syslog.LOG_KERN
	case "user":
		return syslog.LOG_USER
	case "mail":
		return syslog.LOG_MAIL
	case "daemon":
		return syslog.LOG_DAEMON
	case "auth":
		return syslog.LOG_AUTH
	case "syslog":
		return syslog.LOG_SYSLOG
	case "lpr":
		return syslog.LOG_LPR
	case "news":
		return syslog.LOG_NEWS
	case "uucp":
		return syslog.LOG_UUCP
	case "cron":
		return syslog.LOG_CRON
	case "authpriv":
		return syslog.LOG_AUTHPRIV
	case "ftp":
		return syslog.LOG_FTP
	case "local0":
		return syslog.LOG_LOCAL0
	case "local1":
		return syslog.LOG_LOCAL1
	case "local2":
		return syslog.LOG_LOCAL2
	case "local3":
		return syslog.LOG_LOCAL3
	case "local4":
		return syslog.LOG_LOCAL4
	case "local5":
		return syslog.LOG_LOCAL5
	case "local6":
		return syslog.LOG_LOCAL6
	case "local7":
		return syslog.LOG_LOCAL7
	default:
		return syslog.LOG_LOCAL0
	}
}
