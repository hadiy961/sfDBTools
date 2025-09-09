package file

import (
	"fmt"
	"os"
	"os/exec"
	"sfDBTools/internal/logger"
)

// SetFilePermissions sets file permissions and ownership
func SetFilePermissions(filePath string, mode os.FileMode, owner, group string) error {
	lg, _ := logger.Get()

	// Set permissions
	if err := os.Chmod(filePath, mode); err != nil {
		return err
	}

	// Change ownership if owner or group specified (allow either)
	if owner != "" || group != "" {
		ownerGroup := owner
		if group != "" {
			if owner != "" {
				ownerGroup = owner + ":" + group
			} else {
				ownerGroup = ":" + group
			}
		}

		cmd := exec.Command("chown", ownerGroup, filePath)
		if out, err := cmd.CombinedOutput(); err != nil {
			lg.Warn("Failed to set file ownership",
				logger.String("file", filePath),
				logger.String("owner", ownerGroup),
				logger.String("output", string(out)),
				logger.Error(err))
			return fmt.Errorf("chown failed: %w: %s", err, string(out))
		}
	}

	lg.Debug("File permissions set",
		logger.String("file", filePath),
		logger.String("mode", mode.String()),
		logger.String("owner", owner+":"+group))

	return nil
}

// SetDirectoryPermissions sets directory permissions and ownership recursively
func SetDirectoryPermissions(dirPath string, mode os.FileMode, owner, group string) error {
	lg, _ := logger.Get()

	// Change ownership (if requested)
	if owner != "" || group != "" {
		ownerGroup := owner
		if group != "" {
			if owner != "" {
				ownerGroup = owner + ":" + group
			} else {
				ownerGroup = ":" + group
			}
		}

		cmd := exec.Command("chown", "-R", ownerGroup, dirPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			lg.Warn("Failed to set directory ownership",
				logger.String("directory", dirPath),
				logger.String("owner", ownerGroup),
				logger.String("output", string(out)),
				logger.Error(err))
			return fmt.Errorf("chown failed: %w: %s", err, string(out))
		}
	}

	// Set permissions using numeric mode (chmod expects numeric or symbolic, not the FileMode string)
	modeArg := fmt.Sprintf("%o", mode.Perm())
	cmd := exec.Command("chmod", "-R", modeArg, dirPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		lg.Warn("Failed to set directory permissions",
			logger.String("directory", dirPath),
			logger.String("mode", modeArg),
			logger.String("output", string(out)),
			logger.Error(err))
		return fmt.Errorf("chmod failed: %w: %s", err, string(out))
	}

	lg.Debug("Directory permissions set",
		logger.String("directory", dirPath),
		logger.String("mode", mode.String()),
		logger.String("owner", owner+":"+group))

	return nil
}

// CreateDirectoryWithPermissions creates directory with specified permissions and ownership
func CreateDirectoryWithPermissions(dirPath string, mode os.FileMode, owner, group string) error {
	lg, _ := logger.Get()

	// Create directory
	if err := os.MkdirAll(dirPath, mode); err != nil {
		return err
	}

	// Set ownership if specified (allow owner or group)
	if owner != "" || group != "" {
		if err := SetDirectoryPermissions(dirPath, mode, owner, group); err != nil {
			return err
		}
	}

	lg.Info("Directory created with permissions",
		logger.String("path", dirPath),
		logger.String("mode", mode.String()),
		logger.String("owner", owner+":"+group))

	return nil
}
