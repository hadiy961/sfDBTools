package system

import (
	"fmt"
	"os"
	"os/user"
)

// checkPrivileges memeriksa apakah user memiliki privilege sudo/root
func CheckPrivileges() error {
	// Cek apakah running sebagai root
	if os.Geteuid() == 0 {
		return nil
	}

	// Jika bukan root, cek apakah ada sudo access
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Cek apakah user ada di grup sudo/wheel
	groups, err := currentUser.GroupIds()
	if err != nil {
		return fmt.Errorf("failed to get user groups: %w", err)
	}

	// Cek grup sudo (Ubuntu/Debian) atau wheel (CentOS/RHEL)
	hasSudo := false
	for _, gid := range groups {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			continue
		}
		if group.Name == "sudo" || group.Name == "wheel" || group.Name == "admin" {
			hasSudo = true
			break
		}
	}

	if !hasSudo {
		return fmt.Errorf("user %s does not have sudo privileges. Please run with sudo or as root", currentUser.Username)
	}

	return nil
}
