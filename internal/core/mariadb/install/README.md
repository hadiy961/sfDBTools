# MariaDB Installation Module

This module provides multi-distribution MariaDB installation functionality for sfDBTools.

## Supported Operating Systems

- **CentOS**: 7, 8, 9
- **Ubuntu**: 18.04, 20.04, 22.04, 24.04
- **RHEL**: 7, 8, 9
- **Rocky Linux**: 8, 9
- **AlmaLinux**: 8, 9

## Architecture

The installation module follows the modular architecture pattern:

```
internal/core/mariadb/install/
├── types.go              # Type definitions and configurations
├── os_detector.go        # Operating system detection and validation
├── package_manager.go    # Package manager implementations (YUM/APT)
├── repository.go         # Repository configuration management
├── version_selector.go   # Version selection interface
└── runner.go            # Main installation orchestrator
```

## Installation Flow

1. **OS Compatibility Check**
   - Detects operating system from `/etc/os-release`
   - Validates OS and version compatibility
   - Determines package manager type (RPM/DEB)

2. **Internet Connectivity Verification**
   - Checks internet connection using `utils/common/network.go`
   - Validates access to MariaDB download servers

3. **Version Discovery**
   - Fetches available MariaDB versions using `check_version` module
   - Filters versions based on minimum requirements (10.6+)

4. **Version Selection**
   - Interactive version selection with table display
   - Auto-confirmation mode for automated deployments
   - Version validation and EOL information display

5. **Existing Installation Check**
   - Detects existing MariaDB installations
   - Optional removal of existing installations

6. **Repository Configuration**
   - Adds official MariaDB repositories
   - Configures GPG keys and repository priorities
   - OS-specific repository URL generation

7. **Package Installation**
   - Installs MariaDB server packages
   - Handles dependencies automatically

8. **Post-Installation Setup**
   - Starts MariaDB service
   - Enables service on boot
   - Provides security setup guidance

## Usage Examples

### Basic Interactive Installation
```bash
sfdbtools mariadb install
```

### Automated Installation
```bash
# Install specific version automatically
sfdbtools mariadb install --version 10.11 --auto-confirm

# Install with custom options
sfdbtools mariadb install \
    --version 10.6 \
    --auto-confirm \
    --data-dir /var/lib/mysql-custom \
    --remove-existing
```

### Available Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--version, -v` | MariaDB major version (e.g., 10.11) | Interactive selection |
| `--auto-confirm, -y` | Skip all confirmation prompts | `false` |
| `--data-dir` | Custom data directory path | `/var/lib/mysql` |
| `--config-file` | Custom configuration file path | Auto-detected |
| `--remove-existing` | Remove existing installation | `false` |
| `--enable-security` | Enable security guidance | `true` |
| `--start-service` | Start service after install | `true` |

## Components

### OSDetector
- Reads `/etc/os-release` for OS identification
- Validates OS version compatibility
- Determines package manager type (RPM vs DEB)

### PackageManager Interface
- **YumPackageManager**: Handles RPM-based systems (CentOS, RHEL, Rocky, AlmaLinux)
- **AptPackageManager**: Handles DEB-based systems (Ubuntu, Debian)

### RepositoryManager
- Generates OS-specific repository URLs
- Manages GPG key configuration
- Handles repository file creation

### VersionSelector
- Displays available versions in formatted table
- Handles interactive selection
- Supports auto-confirmation mode

### InstallRunner
- Orchestrates the complete installation process
- Provides progress feedback with spinners
- Handles error recovery and cleanup

## Configuration

### InstallConfig Structure
```go
type InstallConfig struct {
    Version           string        // MariaDB version to install
    AutoConfirm       bool          // Skip confirmations
    DataDir          string        // Data directory path
    ConfigFile       string        // Configuration file path
    RemoveExisting   bool          // Remove existing installation
    EnableSecurity   bool          // Enable security setup
    StartService     bool          // Start service after install
}
```

## Error Handling

The module provides comprehensive error handling:

- **OS Compatibility**: Clear messages for unsupported systems
- **Network Issues**: Detailed connectivity error messages
- **Repository Errors**: Helpful repository configuration guidance
- **Installation Failures**: Package manager error output included
- **Service Issues**: Service start/enable failure handling

## Integration

The installation module integrates with:

- `internal/core/mariadb/check_version`: Version discovery
- `utils/common/network`: Connectivity checking
- `utils/terminal`: User interface and progress display
- `internal/logger`: Structured logging

## Testing

Run installation in test environments:

```bash
# Test OS detection
go run main.go mariadb install --version 10.11 --auto-confirm

# Verify repository configuration
cat /etc/yum.repos.d/mariadb-*.repo  # RPM systems
cat /etc/apt/sources.list.d/mariadb*  # DEB systems
```

## Security Considerations

1. **GPG Verification**: All repositories use GPG key verification
2. **Official Sources**: Only uses official MariaDB repositories
3. **Root Privileges**: Installation requires root/sudo access
4. **Service Security**: Recommends running `mysql_secure_installation`

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   sudo sfdbtools mariadb install
   ```

2. **Repository Not Found**
   - Check internet connectivity
   - Verify OS version is supported

3. **Package Conflicts**
   - Use `--remove-existing` flag
   - Manually remove conflicting packages

4. **Service Start Failure**
   - Check system logs: `journalctl -u mariadb`
   - Verify port 3306 availability

### Logs

Installation logs are available in:
- Application logs: `logs/sfDBTools.log`
- System logs: `journalctl -u mariadb`

## Future Enhancements

- [ ] Configuration file templating
- [ ] Custom root password setup
- [ ] Database initialization scripts
- [ ] Performance tuning options
- [ ] SSL/TLS configuration
- [ ] Backup configuration integration
