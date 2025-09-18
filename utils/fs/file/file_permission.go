package file

import (
	"fmt"
	"os"
	"os/exec"
	"sfDBTools/internal/logger"
	"sfDBTools/utils/fs/dir"
)

// SetFilePermissions sets file permissions and ownership (non-recursive)
func (m *Manager) SetFilePermissions(filePath string, mode os.FileMode, owner, group string) error {

	// Set permission using FS where possible
	if err := m.fs.Chmod(filePath, mode); err != nil {
		// fallback to os.Chmod
		if err2 := os.Chmod(filePath, mode); err2 != nil {
			return err2
		}
	}

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
			m.logger.Warn("Failed to set file ownership",
				logger.String("file", filePath),
				logger.String("owner", ownerGroup),
				logger.String("output", string(out)),
				logger.Error(err))
			return fmt.Errorf("chown failed: %w: %s", err, string(out))
		}
	}

	m.logger.Debug("File permissions set",
		logger.String("file", filePath),
		logger.String("mode", mode.String()),
		logger.String("owner", owner+":"+group))

	return nil
}

// EnsureDirectoryCreated is a thin helper that delegates to dir package.
func (m *Manager) EnsureDirectoryCreated(dirPath string, mode os.FileMode, owner, group string) error {
	dm := dir.NewManager()
	return dm.CreateWithPermissions(dirPath, mode, owner, group)
}
