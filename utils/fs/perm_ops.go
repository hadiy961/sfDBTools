package fs

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"syscall"

	"sfDBTools/internal/logger"

	"github.com/spf13/afero"
)

// permissionManager mengimplementasikan PermissionManager interface
type permissionManager struct {
	fs     afero.Fs
	logger *logger.Logger
}

// newPermissionManager membuat instance permission manager baru
func newPermissionManager(fs afero.Fs, logger *logger.Logger) PermissionManager {
	return &permissionManager{
		fs:     fs,
		logger: logger,
	}
}

// SetFilePerms mengatur permission dan ownership untuk file
func (p *permissionManager) SetFilePerms(path string, mode os.FileMode, owner, group string) error {
	// Set permission menggunakan filesystem
	if err := p.fs.Chmod(path, mode); err != nil {
		// Fallback ke os.Chmod
		if err2 := os.Chmod(path, mode); err2 != nil {
			return fmt.Errorf("gagal set permission file: %w", err2)
		}
	}

	// Set ownership jika diperlukan
	if owner != "" || group != "" {
		if err := p.setOwnership(path, owner, group); err != nil {
			p.logger.Warn("Gagal set ownership file",
				logger.String("path", path),
				logger.String("owner", owner),
				logger.String("group", group),
				logger.Error(err))
			return err
		}
	}

	p.logger.Debug("Permission file berhasil diset",
		logger.String("path", path),
		logger.String("mode", mode.String()))

	return nil
}

// SetDirPerms mengatur permission dan ownership untuk direktori
func (p *permissionManager) SetDirPerms(path string, mode os.FileMode, owner, group string) error {
	// Set permission menggunakan filesystem
	if err := p.fs.Chmod(path, mode); err != nil {
		// Fallback ke os.Chmod
		if err2 := os.Chmod(path, mode); err2 != nil {
			return fmt.Errorf("gagal set permission direktori: %w", err2)
		}
	}

	// Set ownership jika diperlukan
	if owner != "" || group != "" {
		if err := p.setOwnership(path, owner, group); err != nil {
			p.logger.Warn("Gagal set ownership direktori",
				logger.String("path", path),
				logger.String("owner", owner),
				logger.String("group", group),
				logger.Error(err))
			return err
		}
	}

	p.logger.Debug("Permission direktori berhasil diset",
		logger.String("path", path),
		logger.String("mode", mode.String()))

	return nil
}

// PreserveOwnership mempertahankan ownership dari FileInfo yang diberikan
func (p *permissionManager) PreserveOwnership(path string, info os.FileInfo) error {
	if info == nil {
		return nil
	}

	if statT, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := os.Chown(path, int(statT.Uid), int(statT.Gid)); err != nil {
			p.logger.Debug("Gagal preserve ownership",
				logger.String("path", path),
				logger.Error(err))
			return err
		}
	}

	return nil
}

// setOwnership mengatur ownership menggunakan metode yang sesuai platform
func (p *permissionManager) setOwnership(path, owner, group string) error {
	if runtime.GOOS == "windows" {
		p.logger.Debug("Ownership setting tidak didukung di Windows")
		return nil
	}

	return p.setUnixOwnership(path, owner, group)
}

// setUnixOwnership mengatur ownership untuk Unix-like systems
func (p *permissionManager) setUnixOwnership(path, owner, group string) error {
	if owner == "" && group == "" {
		return nil
	}

	var uid, gid int = -1, -1

	// Resolve owner ke UID
	if owner != "" {
		if u, err := user.Lookup(owner); err == nil {
			if parsed, err := strconv.Atoi(u.Uid); err == nil {
				uid = parsed
			}
		} else {
			// Coba parse sebagai numeric UID
			if parsed, err := strconv.Atoi(owner); err == nil {
				uid = parsed
			} else {
				return fmt.Errorf("user tidak ditemukan: %s", owner)
			}
		}
	}

	// Resolve group ke GID
	if group != "" {
		if g, err := user.LookupGroup(group); err == nil {
			if parsed, err := strconv.Atoi(g.Gid); err == nil {
				gid = parsed
			}
		} else {
			// Coba parse sebagai numeric GID
			if parsed, err := strconv.Atoi(group); err == nil {
				gid = parsed
			} else {
				return fmt.Errorf("group tidak ditemukan: %s", group)
			}
		}
	}

	// Set ownership menggunakan syscall
	if err := os.Chown(path, uid, gid); err != nil {
		// Fallback ke chown command
		return p.setOwnershipViaCommand(path, owner, group)
	}

	p.logger.Debug("Ownership berhasil diset",
		logger.String("path", path),
		logger.String("owner", owner),
		logger.String("group", group))

	return nil
}

// setOwnershipViaCommand menggunakan chown command sebagai fallback
func (p *permissionManager) setOwnershipViaCommand(path, owner, group string) error {
	ownerGroup := owner
	if group != "" {
		if owner != "" {
			ownerGroup = owner + ":" + group
		} else {
			ownerGroup = ":" + group
		}
	}

	cmd := exec.Command("chown", ownerGroup, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("chown command gagal: %w: %s", err, string(out))
	}

	p.logger.Debug("Ownership set via chown command",
		logger.String("path", path),
		logger.String("owner_group", ownerGroup))

	return nil
}
