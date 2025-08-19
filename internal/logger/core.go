package logger

import (
	"go.uber.org/zap/zapcore"
)

// messageHyphenCore adalah wrapper untuk zapcore.Core yang menambahkan tanda hubung sebelum pesan
type messageHyphenCore struct {
	zapcore.Core
}

// With mewarisi fields dari core induk
func (c *messageHyphenCore) With(fields []zapcore.Field) zapcore.Core {
	return &messageHyphenCore{c.Core.With(fields)}
}

// Check menambahkan entri dengan pesan yang diubah
func (c *messageHyphenCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	// Salin entri dan tambahkan tanda hubung ke pesan
	modifiedEntry := entry
	modifiedEntry.Message = "- " + entry.Message

	if ce = c.Core.Check(modifiedEntry, ce); ce != nil {
		return ce
	}
	return ce
}

// Write menulis entri dengan pesan yang telah diubah
func (c *messageHyphenCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Salin entri dan tambahkan tanda hubung ke pesan
	modifiedEntry := entry
	modifiedEntry.Message = "- " + entry.Message

	return c.Core.Write(modifiedEntry, fields)
}
