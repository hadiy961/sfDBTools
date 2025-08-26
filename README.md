# sfDBTools

A comprehensive MariaDB/MySQL database management tool written in Go, designed for automated database operations including backup, restore, configuration, and health checks.

## Features

- **Database Configuration**: Automated MariaDB/MySQL server setup and configuration
- **Backup Management**: Create compressed and encrypted database backups
- **Restore Operations**: Restore databases from backup files
- **Health Checks**: Monitor database status and performance
- **Logging**: Structured logging with configurable levels
- **CLI Interface**: User-friendly command-line interface built with Cobra

## Installation

### Quick Install (Recommended)

Download the latest pre-compiled binary for Linux:

```bash
# Download the latest release
curl -s https://api.github.com/repos/YOURUSERNAME/sfDBTools/releases/latest \
| grep "browser_download_url.*linux_amd64" \
| cut -d : -f 2,3 \
| tr -d \" \
| wget -qi -

# Extract the binary
tar -xzf sfdbtools_*_Linux_amd64.tar.gz

# Make it executable and move to PATH
chmod +x sfdbtools
sudo mv sfdbtools /usr/local/bin/

# Verify installation
sfdbtools --version
```

### ARM64 Linux

```bash
# For ARM64 systems (like Raspberry Pi, AWS Graviton, etc.)
curl -s https://api.github.com/repos/YOURUSERNAME/sfDBTools/releases/latest \
| grep "browser_download_url.*linux_arm64" \
| cut -d : -f 2,3 \
| tr -d \" \
| wget -qi -

tar -xzf sfdbtools_*_Linux_arm64.tar.gz
chmod +x sfdbtools
sudo mv sfdbtools /usr/local/bin/
```

### Prerequisites

- MariaDB/MySQL server
- Required system tools: `rsync`, `mysql_install_db`, `mariadb-install-db`, `systemctl`

### Build from Source

```bash
git clone https://github.com/YOURUSERNAME/sfDBTools.git
cd sfDBTools
go build -o sfdbtools main.go
```

## Configuration

Create a `config.yaml` file in the config directory. See `config/example*.json` for sample configurations.

```yaml
general:
  client_code: "YOUR_CLIENT_CODE"
  # other configuration options
```

## Usage

### Basic Commands

```bash
# Basic setup after installation
sfdbtools config generate

# Check database status
sfdbtools mariadb check

# Configure MariaDB server
sfdbtools mariadb configure

# Create database backup
sfdbtools backup user <username>

# Restore database from backup
sfdbtools restore user <username>
```

## Quick Start for New Users

After installation, run the setup script to get started quickly:

```bash
# Download and run setup script
curl -sSL https://raw.githubusercontent.com/YOURUSERNAME/sfDBTools/main/setup.sh | bash

# Or if you cloned the repository
./setup.sh
```

This will:
- Check prerequisites
- Create configuration directory
- Generate initial config file
- Show next steps

### Available Commands

- `config` - Configuration management (generate, edit, validate)
- `mariadb` - MariaDB server management (check_version, install, remove, check_config)
- `backup` - Database backup operations (user, selection, all)
- `restore` - Database restore operations (user, single, all)
- `migrate` - Database migration tools
- `database` - Database operations (drop, test connection)

### Command Examples

```bash
# Configuration
sfdbtools config generate          # Generate config file
sfdbtools config validate          # Validate configuration
sfdbtools config show              # Show current config

# MariaDB Management
sfdbtools mariadb check_version     # Check available MariaDB versions
sfdbtools mariadb install          # Install MariaDB (coming soon)
sfdbtools mariadb remove           # Remove MariaDB (coming soon)
sfdbtools mariadb check_config     # Check MariaDB configuration (coming soon)

# Backup Operations
sfdbtools backup user myuser       # Backup specific user's databases
sfdbtools backup all               # Backup all databases
sfdbtools backup selection         # Interactive database selection

# Restore Operations
sfdbtools restore user myuser      # Restore user's databases
sfdbtools restore single mydb      # Restore single database
sfdbtools restore all              # Restore all databases from backup
```

## For Developers

### Creating Releases

This project uses automated releases via GitHub Actions. To create a new release:

```bash
# Using the release script (recommended)
./release.sh 1.2.3

# Or manually
git tag v1.2.3
git push origin v1.2.3
```

The GitHub Actions workflow will automatically:
- Build binaries for Linux (amd64 and arm64)
- Create compressed archives
- Upload to GitHub Releases
- Generate changelog

### Build from Source

```bash
git clone https://github.com/YOURUSERNAME/sfDBTools.git
cd sfDBTools
go mod download
go build -o sfdbtools main.go
```

## Project Structure

```
sfDBTools/
├── cmd/                    # CLI commands
│   └── mariadb/           # MariaDB-specific commands
├── internal/              # Core business logic
│   ├── config/           # Configuration management
│   ├── core/             # Domain logic
│   └── logger/           # Logging utilities
├── utils/                 # Reusable utilities
│   ├── common/           # Common helpers
│   ├── compression/      # Compression utilities
│   ├── crypto/           # Encryption utilities
│   └── database/         # Database connection helpers
├── config/               # Configuration files
└── logs/                 # Runtime logs
```

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Running with Logs

```bash
go run main.go mariadb check
tail -f logs/sfDBTools.log
```

## Database Connection

The tool supports multiple connection methods:
- TCP connection
- Unix socket connection
- Passwordless root access (default)

Connection helpers automatically try different methods and fallback as needed.

## Logging

Structured logging is available throughout the application:

```go
lg := logger.Get()
lg.Info("Operation completed successfully")
lg.Error("Error occurred", "error", err)
```

Logs are saved to `logs/sfDBTools.log` for debugging and monitoring.

## Security Features

- Database backup encryption
- Secure file permissions management
- Configuration validation
- Error handling with proper logging

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the project conventions
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Coding Conventions

- Use structured logging via `internal/logger`
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Use configuration helpers: `config.Get()`
- Follow existing patterns in `cmd/mariadb/` for new commands

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For questions and support:
- Open an issue on GitHub
- Check the logs in `logs/sfDBTools.log` for debugging
- Review configuration examples in `config/`

## Roadmap

- [ ] Enhanced backup scheduling
- [ ] Multi-database support improvements
- [ ] Web interface for monitoring
- [ ] Docker containerization
- [ ] Additional database engine support
