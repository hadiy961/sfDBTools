package common

import (
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

	// Change ownership if owner/group specified
	if owner != "" && group != "" {
		cmd := exec.Command("chown", owner+":"+group, filePath)
		if err := cmd.Run(); err != nil {
			lg.Warn("Failed to set file ownership",
				logger.String("file", filePath),
				logger.String("owner", owner+":"+group),
				logger.Error(err))
			return err
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

	// Change ownership
	if owner != "" && group != "" {
		cmd := exec.Command("chown", "-R", owner+":"+group, dirPath)
		if err := cmd.Run(); err != nil {
			lg.Warn("Failed to set directory ownership",
				logger.String("directory", dirPath),
				logger.String("owner", owner+":"+group),
				logger.Error(err))
			return err
		}
	}

	// Set permissions
	cmd := exec.Command("chmod", "-R", mode.String(), dirPath)
	if err := cmd.Run(); err != nil {
		lg.Warn("Failed to set directory permissions",
			logger.String("directory", dirPath),
			logger.String("mode", mode.String()),
			logger.Error(err))
		return err
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

	// Set ownership if specified
	if owner != "" && group != "" {
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
