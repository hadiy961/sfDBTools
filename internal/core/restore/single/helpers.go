package single

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/logger"
	backup_utils "sfDBTools/utils/backup"
	"sfDBTools/utils/database/info"
)

// metadataPath returns the metadata JSON path for a backup file or empty string
func metadataPath(filePath string) string {
	base := strings.TrimSuffix(filePath, ".enc")
	ext := filepath.Ext(base)
	for _, e := range []string{".gz", ".zst", ".zlib", ".sql"} {
		if ext == e {
			base = strings.TrimSuffix(base, ext)
			ext = filepath.Ext(base)
		}
	}
	base = strings.TrimSuffix(base, ".sql")
	if base == filePath {
		return ""
	}
	return base + ".json"
}

func calculateChecksum(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// verifyChecksumIfPossible reads metadata and compares checksum if available
func verifyChecksumIfPossible(filePath string, lg *logger.Logger) {
	meta := metadataPath(filePath)
	if meta == "" {
		lg.Warn("Metadata file not found, skipping checksum verification", logger.String("file", filePath))
		return
	}

	data, err := os.ReadFile(meta)
	if err != nil {
		lg.Warn("Failed to read metadata file", logger.String("metadata", meta), logger.Error(err))
		return
	}

	var metaInfo backup_utils.BackupMetadata
	if err := json.Unmarshal(data, &metaInfo); err != nil {
		lg.Warn("Invalid metadata format", logger.String("metadata", meta), logger.Error(err))
		return
	}

	if metaInfo.Checksum == "" {
		lg.Warn("Checksum not found in metadata, skipping verification", logger.String("metadata", meta))
		return
	}

	sum, err := calculateChecksum(filePath)
	if err != nil {
		lg.Warn("Checksum calculation failed", logger.String("file", filePath), logger.Error(err))
		return
	}

	if strings.EqualFold(sum, metaInfo.Checksum) {
		lg.Info("Checksum verified successfully", logger.String("file", filePath))
	} else {
		lg.Error("Checksum mismatch", logger.String("file", filePath), logger.String("expected", metaInfo.Checksum), logger.String("got", sum))
	}
}

// ProcessMetadataAfterRestore reads metadata.json (if present) and compares with dbInfo
func ProcessMetadataAfterRestore(filePath string, dbInfo *info.DatabaseInfo, lg *logger.Logger) {
	meta := metadataPath(filePath)
	if meta == "" {
		lg.Debug("No metadata file found for backup", logger.String("file", filePath))
		return
	}

	data, err := os.ReadFile(meta)
	if err != nil {
		lg.Warn("Failed to read metadata file", logger.String("metadata", meta), logger.Error(err))
		return
	}

	var metaInfo backup_utils.BackupMetadata
	if err := json.Unmarshal(data, &metaInfo); err != nil {
		lg.Warn("Invalid metadata format", logger.String("metadata", meta), logger.Error(err))
		return
	}

	if dbInfo != nil {
		DisplayDatabaseComparison(metaInfo, *dbInfo)
	} else {
		lg.Warn("Skipping database comparison because database info was not collected")
	}
}
