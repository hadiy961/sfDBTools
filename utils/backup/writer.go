package backup_utils

import (
	"io"
	"sfDBTools/internal/config"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/compression"
	"sfDBTools/utils/crypto"
)

// BuildWriterChain sets up the writer chain for compression and encryption
func BuildWriterChain(base io.WriteCloser, options BackupOptions, lg *logger.Logger) (io.WriteCloser, []io.Closer, error) {
	var closers []io.Closer
	var writer io.WriteCloser = base

	// Encryption (outer - closest to file)
	if options.Encrypt {
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, nil, err
		}
		key, err := crypto.DeriveKeyFromAppConfig(cfg.General.AppName, cfg.General.ClientCode, cfg.General.Version, cfg.General.Author)
		if err != nil {
			return nil, nil, err
		}
		lg.Debug("Creating encryption writer", logger.Int("key_length", len(key)))
		ew, err := crypto.NewGCMEncryptingWriter(writer, key)
		if err != nil {
			lg.Error("Failed to create encryption writer", logger.Error(err))
			return nil, nil, err
		}
		closers = append(closers, ew)
		writer = ew
		lg.Info("Encryption configured", logger.String("method", "AES-GCM"))
		lg.Debug("Encryption writer chain setup complete")
	}

	// Compression (inner - closest to source)
	if options.Compress {
		compressionType, err := compression.ValidateCompressionType(options.Compression)
		if err != nil {
			lg.Warn("Invalid compression type, using gzip", logger.String("requested", options.Compression), logger.Error(err))
			compressionType = compression.CompressionGzip
		}
		compressionLevel, err := compression.ValidateCompressionLevel(options.CompressionLevel)
		if err != nil {
			lg.Warn("Invalid compression level, using default", logger.String("requested", options.CompressionLevel), logger.Error(err))
			compressionLevel = compression.LevelDefault
		}
		compressionConfig := compression.CompressionConfig{Type: compressionType, Level: compressionLevel}
		cw, err := compression.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return nil, nil, err
		}
		closers = append(closers, cw)
		writer = cw
		lg.Info("Compression configured", logger.String("type", string(compressionType)), logger.String("level", string(compressionLevel)))
	}

	return writer, closers, nil
}
