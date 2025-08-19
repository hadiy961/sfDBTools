#!/bin/bash
set -euo pipefail

# === MariaDB Install Script ===
# This script provides an example of using the sfDBTools mariadb install command
# and handles pre-install checks and recommendations

# === Logging ===
log() { echo -e "\e[32m[INFO]\e[0m $1"; }
warn() { echo -e "\e[33m[WARN]\e[0m $1"; }
err() { echo -e "\e[31m[ERROR]\e[0m $1" >&2; }

# === Configuration ===
SFDB_TOOLS="./sfDBTools"
DEFAULT_VERSION="10.6.22"
DEFAULT_PORT="3306"
DEFAULT_DATA_DIR="/var/lib/mysql"
DEFAULT_LOG_DIR="/var/lib/mysql" 
DEFAULT_BINLOG_DIR="/var/lib/mysqlbinlogs"

# === Functions ===

# Check if sfDBTools binary exists
check_sfdb_tools() {
    if [[ ! -f "$SFDB_TOOLS" ]]; then
        err "sfDBTools binary not found at: $SFDB_TOOLS"
        err "Please build the project first: go build -o sfDBTools"
        exit 1
    fi
}

# Show installation information
show_install_info() {
    echo "🔧 MariaDB Installation Helper"
    echo "=============================="
    echo
    echo "This script helps you install MariaDB with custom configurations."
    echo "It supports:"
    echo "  • CentOS/RHEL/AlmaLinux/Rocky Linux"
    echo "  • Ubuntu/Debian"
    echo "  • Custom ports and directories"
    echo "  • Version selection (10.6.22 to 12.1.11)"
    echo "  • Automatic configuration from template"
    echo
}

# Check system requirements
check_requirements() {
    log "Checking system requirements..."
    
    # Check if running as root or with sudo access
    if ! sudo -v &>/dev/null; then
        err "This script requires sudo privileges for package installation"
        exit 1
    fi

    # Check available disk space
    available_space=$(df / | awk 'NR==2 {print $4}')
    if [[ $available_space -lt 1048576 ]]; then # Less than 1GB
        warn "Low disk space available: $(df -h / | awk 'NR==2 {print $4}')"
        read -p "Continue anyway? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    log "System requirements check passed"
}

# Get installation configuration from user
get_install_config() {
    echo
    echo "📋 Installation Configuration"
    echo "============================"
    
    # Version
    read -p "MariaDB version [$DEFAULT_VERSION]: " VERSION
    VERSION=${VERSION:-$DEFAULT_VERSION}
    
    # Port
    read -p "Port [$DEFAULT_PORT]: " PORT
    PORT=${PORT:-$DEFAULT_PORT}
    
    # Ask about custom paths
    read -p "Use custom directories? (y/n) [n]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Data directory [$DEFAULT_DATA_DIR]: " DATA_DIR
        DATA_DIR=${DATA_DIR:-$DEFAULT_DATA_DIR}
        
        read -p "Log directory [$DEFAULT_LOG_DIR]: " LOG_DIR
        LOG_DIR=${LOG_DIR:-$DEFAULT_LOG_DIR}
        
        read -p "Binlog directory [$DEFAULT_BINLOG_DIR]: " BINLOG_DIR
        BINLOG_DIR=${BINLOG_DIR:-$DEFAULT_BINLOG_DIR}
        
        CUSTOM_PATHS="--custom-paths --data-dir \"$DATA_DIR\" --log-dir \"$LOG_DIR\" --binlog-dir \"$BINLOG_DIR\""
    else
        CUSTOM_PATHS=""
    fi
    
    # Ask about key file
    read -p "Specify encryption key file path (optional): " KEY_FILE
    if [[ -n "$KEY_FILE" ]]; then
        KEY_FILE_ARG="--key-file \"$KEY_FILE\""
    else
        KEY_FILE_ARG=""
    fi
}

# Show configuration summary
show_config_summary() {
    echo
    echo "📋 Installation Summary"
    echo "======================"
    echo "  Version: $VERSION"
    echo "  Port: $PORT"
    if [[ -n "$CUSTOM_PATHS" ]]; then
        echo "  Data Directory: $DATA_DIR"
        echo "  Log Directory: $LOG_DIR"
        echo "  Binlog Directory: $BINLOG_DIR"
    else
        echo "  Using default directories"
    fi
    if [[ -n "$KEY_FILE" ]]; then
        echo "  Key File: $KEY_FILE"
    fi
    echo
}

# Perform MariaDB installation
perform_install() {
    log "Starting MariaDB installation process..."
    
    # Build command
    CMD="$SFDB_TOOLS mariadb install --version \"$VERSION\" --port \"$PORT\""
    
    if [[ -n "$CUSTOM_PATHS" ]]; then
        CMD="$CMD $CUSTOM_PATHS"
    fi
    
    if [[ -n "$KEY_FILE_ARG" ]]; then
        CMD="$CMD $KEY_FILE_ARG"
    fi
    
    # Check if force flag should be used
    if [[ "${1:-}" == "--force" ]]; then
        log "Using force mode (no confirmation prompts)"
        CMD="$CMD --force"
    fi
    
    log "Executing: $CMD"
    eval "$CMD"
}

# Show post-install information
show_post_install_info() {
    log "MariaDB installation process completed!"
    echo
    echo "📋 Post-installation checklist:"
    echo "  ✅ Check service status: systemctl status mariadb"
    echo "  ✅ Connect to MariaDB: mysql -u root -P $PORT"
    echo "  ✅ Secure installation: mysql_secure_installation"
    echo "  ✅ Create databases: CREATE DATABASE mydb;"
    echo "  ✅ Create users: CREATE USER 'myuser'@'%' IDENTIFIED BY 'password';"
    echo "  ✅ Grant privileges: GRANT ALL PRIVILEGES ON mydb.* TO 'myuser'@'%';"
    echo
    echo "📁 Configuration files:"
    echo "  • Main config: /etc/my.cnf.d/server.cnf"
    echo "  • Error log: Check the configured log directory"
    echo
    echo "🔧 Management commands:"
    echo "  • Start: sudo systemctl start mariadb"
    echo "  • Stop: sudo systemctl stop mariadb"
    echo "  • Restart: sudo systemctl restart mariadb"
    echo "  • Status: sudo systemctl status mariadb"
}

# Show help
show_help() {
    echo "MariaDB Install Helper Script"
    echo "============================"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --force     Skip confirmation prompts and force install"
    echo "  --help      Show this help message"
    echo
    echo "Examples:"
    echo "  $0                    # Interactive installation with prompts"
    echo "  $0 --force           # Force installation with default settings"
    echo
    echo "This script helps you install MariaDB by:"
    echo "  1. Checking system requirements"
    echo "  2. Gathering configuration preferences"
    echo "  3. Running the sfDBTools mariadb install command"
    echo "  4. Providing post-installation guidance"
    echo
    echo "Supported versions: 10.6.22 to 12.1.11"
    echo "Supported OS: CentOS/RHEL/AlmaLinux/Rocky/Ubuntu/Debian"
}

# === Main ===
main() {
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        --force)
            log "Force mode enabled - using defaults where possible"
            check_sfdb_tools
            show_install_info
            check_requirements
            # Use default values for force mode
            VERSION="$DEFAULT_VERSION"
            PORT="$DEFAULT_PORT"
            CUSTOM_PATHS=""
            KEY_FILE_ARG=""
            show_config_summary
            perform_install --force
            show_post_install_info
            ;;
        "")
            log "Interactive mode - gathering configuration"
            check_sfdb_tools
            show_install_info
            check_requirements
            get_install_config
            show_config_summary
            read -p "Proceed with installation? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                perform_install
                show_post_install_info
            else
                log "Installation cancelled by user"
            fi
            ;;
        *)
            err "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
