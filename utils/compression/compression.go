package compression

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
)

// CompressionType represents the type of compression algorithm
type CompressionType string

const (
	CompressionNone  CompressionType = "none"
	CompressionGzip  CompressionType = "gzip"
	CompressionPgzip CompressionType = "pgzip" // Parallel gzip
	CompressionZlib  CompressionType = "zlib"
	CompressionZstd  CompressionType = "zstd" // Zstandard
)

// CompressionLevel represents the compression level
type CompressionLevel string

const (
	LevelBestSpeed CompressionLevel = "best_speed"
	LevelFast      CompressionLevel = "fast"
	LevelDefault   CompressionLevel = "default"
	LevelBetter    CompressionLevel = "better"
	LevelBest      CompressionLevel = "best"
)

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Type  CompressionType
	Level CompressionLevel
}

// CompressingWriter wraps an io.Writer with compression
type CompressingWriter struct {
	baseWriter      io.Writer
	compressor      io.WriteCloser
	compressionType CompressionType
}

// NewCompressingWriter creates a new compressing writer
func NewCompressingWriter(baseWriter io.Writer, config CompressionConfig) (*CompressingWriter, error) {
	if config.Type == CompressionNone {
		return &CompressingWriter{
			baseWriter:      baseWriter,
			compressor:      nil,
			compressionType: CompressionNone,
		}, nil
	}

	var compressor io.WriteCloser
	var err error

	switch config.Type {
	case CompressionGzip:
		compressor, err = createGzipWriter(baseWriter, config.Level)
	case CompressionPgzip:
		compressor, err = createPgzipWriter(baseWriter, config.Level)
	case CompressionZlib:
		compressor, err = createZlibWriter(baseWriter, config.Level)
	case CompressionZstd:
		compressor, err = createZstdWriter(baseWriter, config.Level)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create compressor: %w", err)
	}

	return &CompressingWriter{
		baseWriter:      baseWriter,
		compressor:      compressor,
		compressionType: config.Type,
	}, nil
}

// Write writes data through the compressor
func (cw *CompressingWriter) Write(p []byte) (n int, err error) {
	if cw.compressor == nil {
		return cw.baseWriter.Write(p)
	}
	return cw.compressor.Write(p)
}

// Close closes the compressor
func (cw *CompressingWriter) Close() error {
	if cw.compressor != nil {
		return cw.compressor.Close()
	}
	return nil
}

// createGzipWriter creates a gzip writer with specified level
func createGzipWriter(w io.Writer, level CompressionLevel) (*gzip.Writer, error) {
	var gzipLevel int
	switch level {
	case LevelBestSpeed:
		gzipLevel = gzip.BestSpeed
	case LevelFast:
		gzipLevel = gzip.BestSpeed
	case LevelDefault:
		gzipLevel = gzip.DefaultCompression
	case LevelBetter:
		gzipLevel = gzip.BestCompression
	case LevelBest:
		gzipLevel = gzip.BestCompression
	default:
		gzipLevel = gzip.DefaultCompression
	}

	return gzip.NewWriterLevel(w, gzipLevel)
}

// createPgzipWriter creates a parallel gzip writer with specified level
func createPgzipWriter(w io.Writer, level CompressionLevel) (*pgzip.Writer, error) {
	var gzipLevel int
	switch level {
	case LevelBestSpeed:
		gzipLevel = pgzip.BestSpeed
	case LevelFast:
		gzipLevel = pgzip.BestSpeed
	case LevelDefault:
		gzipLevel = pgzip.DefaultCompression
	case LevelBetter:
		gzipLevel = pgzip.BestCompression
	case LevelBest:
		gzipLevel = pgzip.BestCompression
	default:
		gzipLevel = pgzip.DefaultCompression
	}

	return pgzip.NewWriterLevel(w, gzipLevel)
}

// createZlibWriter creates a zlib writer with specified level
func createZlibWriter(w io.Writer, level CompressionLevel) (*zlib.Writer, error) {
	var zlibLevel int
	switch level {
	case LevelBestSpeed:
		zlibLevel = zlib.BestSpeed
	case LevelFast:
		zlibLevel = zlib.BestSpeed
	case LevelDefault:
		zlibLevel = zlib.DefaultCompression
	case LevelBetter:
		zlibLevel = zlib.BestCompression
	case LevelBest:
		zlibLevel = zlib.BestCompression
	default:
		zlibLevel = zlib.DefaultCompression
	}

	return zlib.NewWriterLevel(w, zlibLevel)
}

// createZstdWriter creates a zstandard writer with specified level
func createZstdWriter(w io.Writer, level CompressionLevel) (*zstd.Encoder, error) {
	var zstdLevel zstd.EncoderLevel
	switch level {
	case LevelBestSpeed:
		zstdLevel = zstd.SpeedFastest
	case LevelFast:
		zstdLevel = zstd.SpeedDefault
	case LevelDefault:
		zstdLevel = zstd.SpeedDefault
	case LevelBetter:
		zstdLevel = zstd.SpeedBetterCompression
	case LevelBest:
		zstdLevel = zstd.SpeedBestCompression
	default:
		zstdLevel = zstd.SpeedDefault
	}

	return zstd.NewWriter(w, zstd.WithEncoderLevel(zstdLevel))
}

// GetFileExtension returns the appropriate file extension for the compression type
func GetFileExtension(compressionType CompressionType) string {
	switch compressionType {
	case CompressionGzip, CompressionPgzip:
		return ".gz"
	case CompressionZlib:
		return ".zlib"
	case CompressionZstd:
		return ".zst"
	default:
		return ""
	}
}

// CompressFile compresses an existing file and creates a new compressed file
func CompressFile(inputPath, outputPath string, config CompressionConfig) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create compressing writer
	compressingWriter, err := NewCompressingWriter(outputFile, config)
	if err != nil {
		return fmt.Errorf("failed to create compressing writer: %w", err)
	}
	defer compressingWriter.Close()

	// Copy data through compressor
	if _, err := io.Copy(compressingWriter, inputFile); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	return nil
}

// ValidateCompressionType validates if the compression type is supported
func ValidateCompressionType(compressionType string) (CompressionType, error) {
	ct := CompressionType(strings.ToLower(compressionType))
	switch ct {
	case CompressionNone, CompressionGzip, CompressionPgzip, CompressionZlib, CompressionZstd:
		return ct, nil
	default:
		return CompressionNone, fmt.Errorf("unsupported compression type: %s. Supported types: none, gzip, pgzip, zlib, zstd", compressionType)
	}
}

// ValidateCompressionLevel validates if the compression level is supported
func ValidateCompressionLevel(compressionLevel string) (CompressionLevel, error) {
	cl := CompressionLevel(strings.ToLower(compressionLevel))
	switch cl {
	case LevelBestSpeed, LevelFast, LevelDefault, LevelBetter, LevelBest:
		return cl, nil
	default:
		return LevelDefault, fmt.Errorf("unsupported compression level: %s. Supported levels: best_speed, fast, default, better, best", compressionLevel)
	}
}

// GetCompressionInfo returns information about available compression types
func GetCompressionInfo() map[CompressionType]string {
	return map[CompressionType]string{
		CompressionNone:  "No compression",
		CompressionGzip:  "Standard gzip compression",
		CompressionPgzip: "Parallel gzip compression (faster for large files)",
		CompressionZlib:  "Zlib compression (good compression ratio)",
		CompressionZstd:  "Zstandard compression (fast and good ratio)",
	}
}
