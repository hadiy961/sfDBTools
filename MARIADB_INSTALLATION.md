# MariaDB Installation Feature

## Overview
Fitur instalasi MariaDB yang lengkap dan interaktif untuk membantu user melakukan instalasi MariaDB dengan mudah.

## Features
- ✅ **Interactive Installation**: Proses instalasi sepenuhnya interaktif tanpa memerlukan flag apapun
- ✅ **Version Selection**: User dapat memilih versi MariaDB yang ingin diinstall dari daftar yang tersedia
- ✅ **OS Detection**: Otomatis mendeteksi sistem operasi dan arsitektur
- ✅ **Network Checks**: Memverifikasi konektivitas internet dan ketersediaan repository
- ✅ **Service Management**: Otomatis start dan enable MariaDB service setelah instalasi
- ✅ **Dry Run Mode**: Mode simulasi untuk testing tanpa instalasi aktual
- ✅ **Comprehensive Validation**: Validasi setiap step sebelum melanjutkan

## Command Usage

### Basic Installation
```bash
sfdbtools mariadb install
```

### Dry Run (Testing)
```bash
sfdbtools mariadb install --dry-run
```

## Installation Flow

### 1. Pre-installation Checks
- **Existing Service Check**: Mengecek service MariaDB/MySQL yang sudah ada
- **Internet Connectivity**: Memverifikasi koneksi internet untuk download
- **OS Detection**: Mendeteksi sistem operasi (CentOS, Ubuntu, dll)
- **Repository Availability**: Mengecek ketersediaan repository MariaDB

### 2. Version Selection
- **Fetch Available Versions**: Mengambil daftar versi dari repository resmi
- **Interactive Selection**: User memilih versi yang diinginkan
- **Version Validation**: Memastikan versi kompatibel dengan OS

### 3. Installation Process
- **Repository Setup**: Setup repository MariaDB oficial
- **Package Installation**: Install MariaDB server dan client
- **Service Configuration**: Start dan enable service
- **Installation Verification**: Verifikasi instalasi berhasil

## Supported Operating Systems
- ✅ **CentOS/RHEL** (rpm-based)
- ✅ **Ubuntu/Debian** (deb-based)
- ✅ **Other systemd-based distributions**

## Package Manager Support
- ✅ **YUM** (CentOS/RHEL)
- ✅ **APT** (Ubuntu/Debian)
- ✅ **DNF** (Fedora/newer RHEL)

## Error Handling
- **Input Validation**: Validasi input user dengan retry mechanism
- **Network Timeouts**: Timeout handling untuk operasi network
- **Permission Checks**: Handling untuk operasi yang memerlukan sudo
- **Rollback Capability**: Dapat membersihkan repository jika instalasi gagal

## Code Architecture

### Main Components
1. **`utils/mariadb/installer.go`**: Core installer logic
2. **`utils/mariadb/dry_run_installer.go`**: Dry run functionality
3. **`utils/system/service_manager.go`**: Service management (start/stop/enable)
4. **`utils/system/package_manager.go`**: Package management (install/remove)
5. **`utils/repository/manager.go`**: Repository setup and management
6. **`utils/common/network.go`**: Network connectivity checks
7. **`utils/common/os_detection.go`**: Operating system detection

### Integration Points
- **Version Checking**: Menggunakan existing `utils/mariadb/version.go` untuk fetch versions
- **Logging**: Terintegrasi dengan sistem logging internal
- **Error Handling**: Menggunakan pola error handling yang konsisten
- **Configuration**: Mendukung environment variables dan flags

## Examples

### Successful Installation Flow
```
🌐 Checking internet connectivity...
✅ Internet connectivity verified

🔍 Detecting operating system...
✅ Operating System detected:
   - Name: CentOS Stream
   - Version: 9
   - Architecture: x86_64
   - Package Type: rpm

📦 Checking MariaDB repository availability...
✅ MariaDB repository is accessible

📋 Fetching available MariaDB versions...
✅ Found 17 available versions

🔢 Available MariaDB versions for installation:
   1. MariaDB 10.1
   2. MariaDB 12.0.2
   3. MariaDB 12.1.1
   4. MariaDB 11.8.3
   5. MariaDB 11.4.8
   ...

Select version to install (1-10) [1]: 4
✅ Selected MariaDB 11.8.3 for installation

📦 Setting up MariaDB repository for version 11.8.3...
✅ Repository setup completed

⏳ Installing MariaDB server...
✅ MariaDB installation completed

🚀 Starting and enabling MariaDB service...
✅ Started mariadb service
✅ Enabled mariadb service

🔍 Verifying installation...
✅ MariaDB is running successfully

🎉 Installation process completed!
💡 Next steps:
   - Run 'mysql_secure_installation' to secure your installation
   - Configure root password and remove anonymous users
   - Create database users as needed
```

### Dry Run Example
```
🧪 Running in dry-run mode - no actual installation will be performed

🧪 MariaDB Installation Dry Run
=====================================
✅ Step 1: Checked existing MariaDB service (none found)
✅ Step 2: Internet connectivity verified
✅ Step 3: Operating System: CentOS Stream 9 (rpm)
✅ Step 4: MariaDB repository is accessible

🔢 Available MariaDB versions for installation:
   1. MariaDB 11.8.3
   2. MariaDB 11.4.8
   3. MariaDB 10.11.14
   4. MariaDB 10.6.23

Select version to install (1-4) [1]: 2
✅ Selected MariaDB 11.4.8 for installation
✅ Step 6: Would setup repository for MariaDB 11.4.8
✅ Step 7: Would install MariaDB 11.4.8 packages
✅ Step 8: Would start and enable MariaDB service
✅ Step 9: Would verify installation

🎉 Dry Run Completed Successfully!
📝 Summary: Would install MariaDB 11.4.8 on CentOS Stream 9 (rpm)
```

## Technical Details

### Dependencies
- No external dependencies required
- Uses existing utilities in the project
- Compatible with existing logging and configuration systems

### Security Considerations
- Repository setup requires sudo privileges
- Package installation requires sudo privileges
- Service management requires sudo privileges
- All network operations use secure HTTPS

### Performance
- Network operations have configurable timeouts
- Version fetching is optimized with multiple sources
- Minimal resource usage during installation

### Modularity
- Clean separation of concerns
- Reusable components
- Easy to extend for other database systems
- No code duplication with existing utilities

## Future Enhancements
- [ ] Custom configuration support
- [ ] Database initialization scripts
- [ ] Multi-version installation support
- [ ] Automated security setup
- [ ] Integration with backup/restore features
