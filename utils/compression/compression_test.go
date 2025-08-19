package compression

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompressionTypes(t *testing.T) {
	testData := []byte("This is test data for compression testing. It should be long enough to see compression benefits. " +
		"Let's add more text to make it more compressible. Compression algorithms work better with larger datasets. " +
		"This text is repeated multiple times to increase the size and demonstrate compression efficiency.")

	// Repeat the test data to make it larger
	testData = bytes.Repeat(testData, 10)

	compressionTypes := []CompressionType{
		CompressionNone,
		CompressionGzip,
		CompressionPgzip,
		CompressionZlib,
		CompressionZstd,
	}

	for _, ct := range compressionTypes {
		t.Run(string(ct), func(t *testing.T) {
			// Create a buffer to write to
			var buf bytes.Buffer

			// Create compression config
			config := CompressionConfig{
				Type:  ct,
				Level: LevelDefault,
			}

			// Create compressing writer
			cw, err := NewCompressingWriter(&buf, config)
			if err != nil {
				t.Fatalf("Failed to create compressing writer: %v", err)
			}

			// Write test data
			n, err := cw.Write(testData)
			if err != nil {
				t.Fatalf("Failed to write data: %v", err)
			}

			if n != len(testData) {
				t.Fatalf("Expected to write %d bytes, wrote %d", len(testData), n)
			}

			// Close the writer
			if err := cw.Close(); err != nil {
				t.Fatalf("Failed to close writer: %v", err)
			}

			// Check results
			compressedSize := buf.Len()
			originalSize := len(testData)

			t.Logf("Compression type: %s", ct)
			t.Logf("Original size: %d bytes", originalSize)
			t.Logf("Compressed size: %d bytes", compressedSize)

			if ct == CompressionNone {
				if compressedSize != originalSize {
					t.Errorf("No compression should result in same size, got %d expected %d", compressedSize, originalSize)
				}
			} else {
				if compressedSize >= originalSize {
					t.Logf("Warning: Compressed size (%d) is not smaller than original (%d) for %s", compressedSize, originalSize, ct)
				}
			}
		})
	}
}

func TestCompressionLevels(t *testing.T) {
	testData := bytes.Repeat([]byte("Test data for compression levels. "), 1000)

	levels := []CompressionLevel{
		LevelBestSpeed,
		LevelFast,
		LevelDefault,
		LevelBetter,
		LevelBest,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			var buf bytes.Buffer

			config := CompressionConfig{
				Type:  CompressionGzip,
				Level: level,
			}

			cw, err := NewCompressingWriter(&buf, config)
			if err != nil {
				t.Fatalf("Failed to create compressing writer: %v", err)
			}

			if _, err := cw.Write(testData); err != nil {
				t.Fatalf("Failed to write data: %v", err)
			}

			if err := cw.Close(); err != nil {
				t.Fatalf("Failed to close writer: %v", err)
			}

			t.Logf("Level: %s, Compressed size: %d bytes", level, buf.Len())
		})
	}
}

func TestFileCompression(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create input file
	inputPath := filepath.Join(tempDir, "test_input.txt")
	testData := []byte(strings.Repeat("This is test data for file compression testing.\n", 1000))

	if err := os.WriteFile(inputPath, testData, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Test different compression types
	compressionTypes := []CompressionType{
		CompressionGzip,
		CompressionZstd,
	}

	for _, ct := range compressionTypes {
		t.Run(string(ct), func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "test_output"+GetFileExtension(ct))

			config := CompressionConfig{
				Type:  ct,
				Level: LevelDefault,
			}

			// Compress file
			if err := CompressFile(inputPath, outputPath, config); err != nil {
				t.Fatalf("Failed to compress file: %v", err)
			}

			// Check if output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Fatalf("Output file was not created")
			}

			// Check file sizes
			inputStat, _ := os.Stat(inputPath)
			outputStat, _ := os.Stat(outputPath)

			t.Logf("Input size: %d bytes", inputStat.Size())
			t.Logf("Output size: %d bytes", outputStat.Size())
			t.Logf("Compression ratio: %.2f%%", float64(outputStat.Size())/float64(inputStat.Size())*100)
		})
	}
}

func TestValidateCompressionType(t *testing.T) {
	validTypes := []string{"none", "gzip", "pgzip", "zlib", "zstd"}
	invalidTypes := []string{"invalid", "bzip2", "lz4", ""}

	for _, validType := range validTypes {
		t.Run("valid_"+validType, func(t *testing.T) {
			_, err := ValidateCompressionType(validType)
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", validType, err)
			}
		})
	}

	for _, invalidType := range invalidTypes {
		t.Run("invalid_"+invalidType, func(t *testing.T) {
			_, err := ValidateCompressionType(invalidType)
			if err == nil {
				t.Errorf("Expected %s to be invalid, but got no error", invalidType)
			}
		})
	}
}

func TestValidateCompressionLevel(t *testing.T) {
	validLevels := []string{"best_speed", "fast", "default", "better", "best"}
	invalidLevels := []string{"invalid", "maximum", "minimum", ""}

	for _, validLevel := range validLevels {
		t.Run("valid_"+validLevel, func(t *testing.T) {
			_, err := ValidateCompressionLevel(validLevel)
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", validLevel, err)
			}
		})
	}

	for _, invalidLevel := range invalidLevels {
		t.Run("invalid_"+invalidLevel, func(t *testing.T) {
			_, err := ValidateCompressionLevel(invalidLevel)
			if err == nil {
				t.Errorf("Expected %s to be invalid, but got no error", invalidLevel)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	expectations := map[CompressionType]string{
		CompressionNone:  "",
		CompressionGzip:  ".gz",
		CompressionPgzip: ".gz",
		CompressionZlib:  ".zlib",
		CompressionZstd:  ".zst",
	}

	for ct, expectedExt := range expectations {
		t.Run(string(ct), func(t *testing.T) {
			ext := GetFileExtension(ct)
			if ext != expectedExt {
				t.Errorf("Expected extension %s for %s, got %s", expectedExt, ct, ext)
			}
		})
	}
}

func BenchmarkCompression(b *testing.B) {
	testData := bytes.Repeat([]byte("Benchmark data for compression testing. "), 10000)

	compressionTypes := []CompressionType{
		CompressionGzip,
		CompressionPgzip,
		CompressionZlib,
		CompressionZstd,
	}

	for _, ct := range compressionTypes {
		b.Run(string(ct), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				config := CompressionConfig{
					Type:  ct,
					Level: LevelDefault,
				}

				cw, err := NewCompressingWriter(&buf, config)
				if err != nil {
					b.Fatalf("Failed to create compressing writer: %v", err)
				}

				if _, err := cw.Write(testData); err != nil {
					b.Fatalf("Failed to write data: %v", err)
				}

				if err := cw.Close(); err != nil {
					b.Fatalf("Failed to close writer: %v", err)
				}
			}
		})
	}
}
