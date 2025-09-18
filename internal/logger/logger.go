package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
	// showCaller controls whether we scan the call stack for the external
	// caller and attach file:line info. It should only be true for debug
	// level logging to avoid adding caller overhead in normal runs.
	showCaller bool
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

// Time returns a Field containing time.Time value for structured logging
func Time(key string, t time.Time) Field {
	return Field{Key: key, Value: t}
}

// Convert our Fields to logrus.Fields
func fieldsToLogrusFields(fields []Field) logrus.Fields {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}
	return logrusFields
}

// writerHook is a logrus Hook that writes formatted entries to an io.Writer.
// We use this to write file output with a JSON formatter while keeping
// console/syslog formats separate.
type writerHook struct {
	Writer    io.Writer
	Formatter logrus.Formatter
	LevelsVal []logrus.Level
}

func (h *writerHook) Fire(entry *logrus.Entry) error {
	// Let the configured formatter format the entry (it may include caller info
	// when log.ReportCaller is true).
	b, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.Writer.Write(b)
	return err
}

func (h *writerHook) Levels() []logrus.Level {
	if h.LevelsVal != nil {
		return h.LevelsVal
	}
	return logrus.AllLevels
}

// PrettyJSONFormatter is a custom logrus formatter that emits human-friendly
// JSON. It orders common fields first (time, level, msg, file) and then the
// remaining fields sorted by key. It can omit the function name when empty.
type PrettyJSONFormatter struct {
	TimestampFormat  string
	PrettyPrint      bool
	CallerPrettyfier func(*runtime.Frame) (string, string)
}

func (f *PrettyJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// copy data to avoid mutating original
	data := make(map[string]interface{}, len(entry.Data))
	for k, v := range entry.Data {
		data[k] = v
	}

	// Prepare caller info: prefer explicit data["file"] if present (injected by
	// wrappers); otherwise try CallerPrettyfier/entry.Caller; if still empty,
	// fallback to scanning the runtime stack for the external caller.
	if _, ok := data["file"]; !ok {
		if entry.HasCaller() && f.CallerPrettyfier != nil {
			fn, file := f.CallerPrettyfier(entry.Caller)
			if file != "" {
				data["file"] = file
			}
			if fn != "" {
				data["func"] = fn
			}
		} else if entry.HasCaller() {
			data["file"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		} else {
			if cf, ok := getExternalCaller(); ok {
				data["file"] = cf
			}
		}
	}

	// Build ordered output: time, level, msg, file, then other keys sorted
	out := make(map[string]interface{})
	timestamp := entry.Time.Format(f.TimestampFormat)
	out["time"] = timestamp
	out["level"] = entry.Level.String()
	out["msg"] = entry.Message
	if v, ok := data["file"]; ok {
		out["file"] = v
		delete(data, "file")
	}
	// Remove func from map if empty or not desired
	if v, ok := data["func"]; ok {
		// keep it only if non-empty string
		if s, ok2 := v.(string); ok2 && s != "" {
			out["func"] = s
		}
		delete(data, "func")
	}

	// Sort remaining keys for stable output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out[k] = data[k]
	}

	var b []byte
	var err error
	if f.PrettyPrint {
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(out)
		b = buf.Bytes()
	} else {
		b, err = json.Marshal(out)
		if err == nil {
			b = append(b, '\n')
		}
	}
	return b, err
}

// ConsoleFormatter formats logs to a concise single-line human-readable form:
// [timestamp][LEVEL][file:line] - Message ( {k=v}, {k2=v2} )
type ConsoleFormatter struct {
	TimestampFormat string
}

func (f *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	ts := entry.Time.Format(f.TimestampFormat)
	lvl := strings.ToUpper(entry.Level.String())

	// file:line if available
	fileInfo := ""
	if v, ok := entry.Data["file"]; ok {
		if s, ok2 := v.(string); ok2 {
			fileInfo = s
		}
	} else if entry.HasCaller() {
		fileInfo = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
	} else if cf, ok := getExternalCaller(); ok {
		fileInfo = cf
	}

	// Build message
	var b strings.Builder
	b.WriteString("[")
	b.WriteString(ts)
	b.WriteString("][")
	b.WriteString(lvl)
	b.WriteString("]")
	if fileInfo != "" {
		b.WriteString("[")
		b.WriteString(fileInfo)
		b.WriteString("]")
	}
	b.WriteString(" - ")
	b.WriteString(entry.Message)

	// Append details as ( {k=v}, {k2=v2} ) if there are fields
	if len(entry.Data) > 0 {
		// Collect keys sorted for stable output
		keys := make([]string, 0, len(entry.Data))
		for k := range entry.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		b.WriteString(" [")
		first := true
		for _, k := range keys {
			if !first {
				b.WriteString(", ")
			}
			first = false
			v := entry.Data[k]
			// Format value: quote string values that contain spaces
			switch vv := v.(type) {
			case string:
				if strings.ContainsAny(vv, " \t\n") {
					b.WriteString(fmt.Sprintf("{%s=\"%s\"}", k, vv))
				} else {
					b.WriteString(fmt.Sprintf("{%s=%s}", k, vv))
				}
			default:
				b.WriteString(fmt.Sprintf("{%s=%v}", k, vv))
			}
		}
		b.WriteString("]")
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// Logging methods
func (l *Logger) Debug(msg string, fields ...Field) {
	// Ensure caller file:line is attached (skip if provided)
	if !hasField(fields, "file") {
		if cf, ok := findCallerField(); ok {
			fields = append([]Field{cf}, fields...)
		}
	}
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Debug(msg)
}

func (l *Logger) Info(msg string, fields ...Field) {
	if !hasField(fields, "file") {
		if cf, ok := findCallerField(); ok {
			fields = append([]Field{cf}, fields...)
		}
	}
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Info(msg)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	if !hasField(fields, "file") {
		if cf, ok := findCallerField(); ok {
			fields = append([]Field{cf}, fields...)
		}
	}
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Warn(msg)
}

func (l *Logger) Error(msg string, fields ...Field) {
	if !hasField(fields, "file") {
		if cf, ok := findCallerField(); ok {
			fields = append([]Field{cf}, fields...)
		}
	}
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Error(msg)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	if !hasField(fields, "file") {
		if cf, ok := findCallerField(); ok {
			fields = append([]Field{cf}, fields...)
		}
	}
	l.Logger.WithFields(fieldsToLogrusFields(fields)).Fatal(msg)
}

// hasField checks if provided fields contain a key
func hasField(fields []Field, key string) bool {
	for _, f := range fields {
		if f.Key == key {
			return true
		}
	}
	return false
}

// findCallerField walks the stack to find the first caller outside this package
// and returns a Field with key "file" and value "filename:line".
func findCallerField() (Field, bool) {
	// If caller info is disabled, return immediately.
	if !showCaller {
		return Field{}, false
	}
	// start at 3 to skip runtime.Callers -> our helper -> logger wrapper
	for i := 3; i < 16; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			continue
		}
		// skip frames inside the logger package
		if strings.Contains(file, string(filepath.Separator)+"internal"+string(filepath.Separator)+"logger") {
			continue
		}
		_ = pc
		return String("file", fmt.Sprintf("%s:%d", filepath.Base(file), line)), true
	}
	return Field{}, false
}

// getExternalCaller scans the call stack to find the first frame outside the
// logger package and returns a "filename:line" string.
func getExternalCaller() (string, bool) {
	// If caller info is disabled, don't scan the stack.
	if !showCaller {
		return "", false
	}
	// skip a few frames from runtime.Callers -> our helper -> logger wrappers
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, string(filepath.Separator)+"internal"+string(filepath.Separator)+"logger") {
			return fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line), true
		}
		if !more {
			break
		}
	}
	return "", false
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

	// Only enable caller info scanning when in debug level to avoid the
	// overhead of walking the stack in normal runs.
	showCaller = (level == logrus.DebugLevel)

	// Set formatter for console/syslog based on format config. File output will
	// be forced to JSON via a hook so file logs are always JSON.
	var consoleFormatter logrus.Formatter
	switch strings.ToLower(cfg.Log.Format) {
	case "json":
		consoleFormatter = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		}
	case "text", "":
		// Use custom ConsoleFormatter for the terminal human-friendly format
		consoleFormatter = &ConsoleFormatter{TimestampFormat: "2006-01-02 15:04:05"}
	default:
		consoleFormatter = &ConsoleFormatter{TimestampFormat: "2006-01-02 15:04:05"}
	}
	// Ensure console caller prettyfier hides function name and prints file:line.
	switch cf := consoleFormatter.(type) {
	case *logrus.TextFormatter:
		cf.CallerPrettyfier = func(frame *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
		}
	case *logrus.JSONFormatter:
		cf.CallerPrettyfier = func(frame *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
		}
	}
	log.SetFormatter(consoleFormatter)

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

	// Enable logrus's ReportCaller when debug to include richer caller info
	// (file:line) in entries. We still guard our own stack scanning helpers
	// with showCaller so non-debug runs avoid the extra work.
	if level == logrus.DebugLevel {
		log.SetReportCaller(true)
	} else {
		log.SetReportCaller(false)
	}

	// Attach file writer as a hook with a JSON formatter so file output is
	// JSON regardless of console format. Other writers (console/syslog) will
	// use the logger's SetOutput target.
	for _, w := range writers {
		if w == os.Stdout {
			continue
		}
		// Use our PrettyJSONFormatter for file outputs; compact by default.
		pfmt := &PrettyJSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			PrettyPrint:     false, // compact single-line JSON for file logs
			CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
				// Return empty function name to avoid function in output, but
				// keep file:line.
				return "", fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
			},
		}
		log.AddHook(&writerHook{Writer: w, Formatter: pfmt})
	}

	// If no writers other than file hooks were configured, default console output
	// to stdout so logs still appear on console.
	if len(writers) == 0 || (len(writers) == 1 && writers[0] != os.Stdout) {
		log.SetOutput(os.Stdout)
	} else {
		// Keep stdout as the main output for console and syslog; file outputs are
		// handled via hooks.
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
