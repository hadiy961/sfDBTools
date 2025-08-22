package logger

import (
	"fmt"
	"sync"

	"sfDBTools/internal/config"
	"sfDBTools/internal/config/model"

	"go.uber.org/zap"
)

var (
	lg   *Logger
	once sync.Once
)

// Logger wraps zap.Logger so other packages do not depend on zap directly.
type Logger struct{ *zap.Logger }

// Field represents a structured log field.
type Field = zap.Field

// Common field constructors.
func String(key, val string) Field            { return zap.String(key, val) }
func Strings(key string, vals []string) Field { return zap.Strings(key, vals) }
func Int(key string, val int) Field           { return zap.Int(key, val) }
func Int64(key string, val int64) Field       { return zap.Int64(key, val) }
func Float64(key string, val float64) Field   { return zap.Float64(key, val) }
func Bool(key string, val bool) Field         { return zap.Bool(key, val) }
func Error(err error) Field                   { return zap.Error(err) }

// Logging helper methods.
func (l *Logger) Debug(msg string, fields ...Field) { l.Logger.Debug(msg, fields...) }
func (l *Logger) Info(msg string, fields ...Field)  { l.Logger.Info(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...Field)  { l.Logger.Warn(msg, fields...) }
func (l *Logger) Error(msg string, fields ...Field) { l.Logger.Error(msg, fields...) }
func (l *Logger) Fatal(msg string, fields ...Field) { l.Logger.Fatal(msg, fields...) }
func (l *Logger) Sync() error                       { return l.Logger.Sync() }

// Get returns a singleton Logger configured using config package.
func Get() (*Logger, error) {
	var err error
	once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("logger initialization panicked: %v", r)
				// Create a minimal logger as fallback
				lg = &Logger{zap.NewNop()}
			}
		}()
		
		var cfg *model.Config
		cfg, err = config.LoadConfig()
		if err != nil {
			return
		}
		lg, err = buildLogger(cfg)
	})
	return lg, err
}
