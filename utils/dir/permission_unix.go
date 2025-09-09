//go:build !windows
// +build !windows

package dir

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"

	"sfDBTools/internal/logger"
)

// setUnixOwnership sets directory ownership pada Unix-like systems
func (m *Manager) setUnixOwnership(path, owner, group string) error {
	if owner == "" && group == "" {
		return nil
	}

	var uid, gid int = -1, -1

	// Resolve owner ke UID
	if owner != "" {
		if u, err := user.Lookup(owner); err == nil {
			if uid, err = strconv.Atoi(u.Uid); err != nil {
				return fmt.Errorf("invalid UID untuk user %s: %w", owner, err)
			}
		} else {
			// Coba parse sebagai numeric UID
			if uid, err = strconv.Atoi(owner); err != nil {
				return fmt.Errorf("user tidak ditemukan dan bukan numeric UID: %s", owner)
			}
		}
	}

	// Resolve group ke GID
	if group != "" {
		if g, err := user.LookupGroup(group); err == nil {
			if gid, err = strconv.Atoi(g.Gid); err != nil {
				return fmt.Errorf("invalid GID untuk group %s: %w", group, err)
			}
		} else {
			// Coba parse sebagai numeric GID
			if gid, err = strconv.Atoi(group); err != nil {
				return fmt.Errorf("group tidak ditemukan dan bukan numeric GID: %s", group)
			}
		}
	}

	// Set ownership menggunakan syscall (lebih reliable dari exec)
	if err := syscall.Chown(path, uid, gid); err != nil {
		m.logger.Warn("Gagal set ownership dengan syscall, mencoba chown command",
			logger.String("path", path),
			logger.String("owner", owner),
			logger.String("group", group),
			logger.Error(err))

		// Fallback ke chown command
		return m.setOwnershipWithChown(path, owner, group)
	}

	m.logger.Debug("Ownership berhasil diset",
		logger.String("path", path),
		logger.Int("uid", uid),
		logger.Int("gid", gid))

	return nil
}

// setOwnershipWithChown menggunakan chown command sebagai fallback
func (m *Manager) setOwnershipWithChown(path, owner, group string) error {
	ownerGroup := owner
	if group != "" {
		if owner != "" {
			ownerGroup = owner + ":" + group
		} else {
			ownerGroup = ":" + group
		}
	}

	cmd := exec.Command("chown", "-R", ownerGroup, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		m.logger.Error("chown command gagal",
			logger.String("path", path),
			logger.String("owner_group", ownerGroup),
			logger.String("output", string(output)),
			logger.Error(err))
		return fmt.Errorf("chown gagal untuk '%s': %w: %s", path, err, string(output))
	}

	m.logger.Debug("Ownership set dengan chown command",
		logger.String("path", path),
		logger.String("owner_group", ownerGroup))

	return nil
}

// SetPermissions sets directory permissions dan ownership untuk Unix systems
func (m *Manager) SetPermissions(path string, mode os.FileMode, owner, group string) error {
	normalizedPath := filepath.Clean(path)

	if !m.Exists(normalizedPath) {
		return fmt.Errorf("direktori tidak ada: %s", normalizedPath)
	}

	// Set file mode permission
	if err := m.fs.Chmod(normalizedPath, mode); err != nil {
		return fmt.Errorf("gagal set permission untuk '%s': %w", normalizedPath, err)
	}

	// Set ownership jika diminta
	if owner != "" || group != "" {
		if err := m.setUnixOwnership(normalizedPath, owner, group); err != nil {
			return err
		}
	}

	m.logger.Info("Permission direktori berhasil diset",
		logger.String("path", normalizedPath),
		logger.String("mode", mode.String()),
		logger.String("owner", owner),
		logger.String("group", group))

	return nil
}

// GetPermissions mendapatkan informasi permission direktori
func (m *Manager) GetPermissions(path string) (*PermissionInfo, error) {
	normalizedPath := filepath.Clean(path)

	info, err := m.fs.Stat(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan info file '%s': %w", normalizedPath, err)
	}

	permInfo := &PermissionInfo{
		Path: normalizedPath,
		Mode: info.Mode(),
	}

	// Get ownership info dari syscall
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		permInfo.UID = int(stat.Uid)
		permInfo.GID = int(stat.Gid)

		// Resolve UID ke username
		if u, err := user.LookupId(strconv.Itoa(permInfo.UID)); err == nil {
			permInfo.Owner = u.Username
		}

		// Resolve GID ke group name
		if g, err := user.LookupGroupId(strconv.Itoa(permInfo.GID)); err == nil {
			permInfo.Group = g.Name
		}
	}

	return permInfo, nil
}

// CanRead mengecek apakah current user dapat membaca direktori
func (m *Manager) CanRead(path string) bool {
	return unix.Access(path, unix.R_OK) == nil
}

// CanWrite mengecek apakah current user dapat menulis ke direktori
func (m *Manager) CanWrite(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

// CanExecute mengecek apakah current user dapat execute (traverse) direktori
func (m *Manager) CanExecute(path string) bool {
	return unix.Access(path, unix.X_OK) == nil
}

// GetCurrentUser mendapatkan informasi user yang sedang menjalankan program
func GetCurrentUser() (*UserInfo, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan current user: %w", err)
	}

	uid, _ := strconv.Atoi(currentUser.Uid)
	gid, _ := strconv.Atoi(currentUser.Gid)

	return &UserInfo{
		UID:      uid,
		GID:      gid,
		Username: currentUser.Username,
		Name:     currentUser.Name,
		HomeDir:  currentUser.HomeDir,
	}, nil
}

// IsRoot mengecek apakah program berjalan sebagai root
func IsRoot() bool {
	return os.Geteuid() == 0
}
