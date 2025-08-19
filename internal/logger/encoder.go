package logger

import (
	"fmt"
	"strings"
	"time"

	"sfDBTools/internal/config/model"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// createEncoder membuat dan mengkonfigurasi encoder berdasarkan konfigurasi yang diberikan
func createEncoder(cfg *model.LogConfig) zapcore.Encoder {
	encCfg := zap.NewProductionEncoderConfig()

	// Konfigurasi dasar
	encCfg.MessageKey = "message"
	encCfg.LevelKey = "level"
	encCfg.TimeKey = "time"
	encCfg.CallerKey = "caller"
	encCfg.NameKey = "logger"
	encCfg.EncodeName = zapcore.FullNameEncoder
	encCfg.EncodeDuration = zapcore.StringDurationEncoder
	encCfg.ConsoleSeparator = " "

	// Konfigurasi format: "[timestamp] [level] [caller] - message"
	encCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		loc, lerr := time.LoadLocation(cfg.Timezone)
		if lerr == nil {
			t = t.In(loc)
		}
		enc.AppendString(fmt.Sprintf("[%s]", t.Format("2006-01-02 15:04:05")))
	}

	encCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%s]", strings.ToUpper(l.String())))
	}

	// Caller hanya ditampilkan jika level debug
	encCfg.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		if strings.ToLower(cfg.Level) == "debug" || strings.ToLower(cfg.Level) == "trace" {
			enc.AppendString(fmt.Sprintf("[%s]", caller.TrimmedPath()))
		} else {
			// Kosong jika bukan level debug
			enc.AppendString("")
		}
	}

	// Pilih encoder berdasarkan format
	if strings.ToLower(cfg.Format) == "json" {
		return zapcore.NewJSONEncoder(encCfg)
	}

	return zapcore.NewConsoleEncoder(encCfg)
}
