package discovery

import (
	"sfDBTools/internal/logger"
)

// DiscoverMariaDBInstallation mendeteksi instalasi MariaDB di sistem
func DiscoverMariaDBInstallation() (*MariaDBInstallation, error) {
	lg, _ := logger.Get()
	lg.Debug("Memulai discovery instalasi MariaDB")

	installation := &MariaDBInstallation{ConfigPaths: []string{}}

	if err := detectMariaDBBinary(installation); err != nil {
		lg.Warn("Gagal mendeteksi binary MariaDB", logger.Error(err))
	}
	if installation.BinaryPath != "" {
		if err := detectMariaDBVersion(installation); err != nil {
			lg.Debug("Gagal mendeteksi versi MariaDB", logger.Error(err))
		}
	}
	if err := detectConfigFiles(installation); err != nil {
		lg.Debug("Gagal mendeteksi file konfigurasi MariaDB", logger.Error(err))
	}
	if err := detectMariaDBService(installation); err != nil {
		lg.Debug("Gagal mendeteksi service MariaDB", logger.Error(err))
	}
	if err := detectDataDirAndSocket(installation); err != nil {
		lg.Debug("Gagal mendeteksi data directory dan socket", logger.Error(err))
	}

	return installation, nil
}
