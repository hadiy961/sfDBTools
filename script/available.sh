#!/bin/bash

# MariaDB Version Checker
# Script untuk mengecek versi MariaDB yang tersedia berdasarkan OS
# Author: Generated Script
# Version: 1.0

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Global variables
readonly SCRIPT_NAME=$(basename "$0")
readonly SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly LOG_FILE="/tmp/${SCRIPT_NAME%.*}.log"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE" >&2
}

log_debug() {
    if [[ "${DEBUG:-0}" == "1" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $*" | tee -a "$LOG_FILE"
    fi
}

# Error handling
cleanup() {
    log_debug "Cleaning up temporary files..."
    # Add cleanup logic here if needed
}

error_handler() {
    local line_number=$1
    log_error "Script failed at line $line_number"
    cleanup
    exit 1
}

trap 'error_handler $LINENO' ERR
trap cleanup EXIT

# OS Detection functions
detect_os() {
    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        source /etc/os-release
        echo "$ID"
    elif [[ -f /etc/redhat-release ]]; then
        echo "rhel"
    elif [[ -f /etc/debian_version ]]; then
        echo "debian"
    elif command -v sw_vers &> /dev/null; then
        echo "macos"
    else
        echo "unknown"
    fi
}

detect_os_version() {
    local os_id="$1"
    
    case "$os_id" in
        "ubuntu"|"debian")
            if [[ -f /etc/os-release ]]; then
                # shellcheck source=/dev/null
                source /etc/os-release
                echo "$VERSION_ID"
            else
                cat /etc/debian_version
            fi
            ;;
        "centos"|"rhel"|"fedora"|"rocky"|"almalinux")
            if [[ -f /etc/os-release ]]; then
                # shellcheck source=/dev/null
                source /etc/os-release
                echo "$VERSION_ID"
            else
                rpm -q --queryformat '%{VERSION}' centos-release 2>/dev/null || echo "unknown"
            fi
            ;;
        "opensuse"|"sles")
            if [[ -f /etc/os-release ]]; then
                # shellcheck source=/dev/null
                source /etc/os-release
                echo "$VERSION_ID"
            fi
            ;;
        "macos")
            sw_vers -productVersion
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

detect_architecture() {
    case "$(uname -m)" in
        "x86_64"|"amd64")
            echo "amd64"
            ;;
        "aarch64"|"arm64")
            echo "arm64"
            ;;
        "armv7l")
            echo "armhf"
            ;;
        "i386"|"i686")
            echo "i386"
            ;;
        *)
            echo "$(uname -m)"
            ;;
    esac
}

# MariaDB version checking functions
check_mariadb_ubuntu() {
    local version="$1"
    local arch="$2"
    
    log_info "Checking MariaDB versions for Ubuntu $version ($arch)..."
    
    # Update package list
    if ! apt-cache policy mariadb-server &>/dev/null; then
        log_warn "MariaDB repository not found. Adding official MariaDB repository..."
        echo "You may need to add MariaDB repository manually:"
        echo "curl -sS https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | sudo bash"
        return 1
    fi
    
    log_info "Available MariaDB versions:"
    apt-cache madison mariadb-server | awk '{print $3}' | sort -V | uniq
}

check_mariadb_centos() {
    local version="$1"
    local arch="$2"
    
    log_info "Checking MariaDB versions for CentOS/RHEL $version ($arch)..."
    
    if ! yum list available mariadb-server &>/dev/null && ! dnf list available mariadb-server &>/dev/null; then
        log_warn "MariaDB not found in default repositories. Adding MariaDB repository..."
        echo "You may need to add MariaDB repository manually:"
        echo "Check: https://downloads.mariadb.org/mariadb/repositories/"
        return 1
    fi
    
    log_info "Available MariaDB versions:"
    if command -v dnf &>/dev/null; then
        dnf list available mariadb-server --showduplicates | grep mariadb-server | awk '{print $2}' | sort -V
    else
        yum list available mariadb-server --showduplicates | grep mariadb-server | awk '{print $2}' | sort -V
    fi
}

check_mariadb_fedora() {
    local version="$1"
    local arch="$2"
    
    log_info "Checking MariaDB versions for Fedora $version ($arch)..."
    
    log_info "Available MariaDB versions:"
    dnf list available mariadb-server --showduplicates | grep mariadb-server | awk '{print $2}' | sort -V
}

check_mariadb_opensuse() {
    local version="$1"
    local arch="$2"
    
    log_info "Checking MariaDB versions for openSUSE $version ($arch)..."
    
    log_info "Available MariaDB versions:"
    zypper search -s mariadb | grep mariadb | awk '{print $4}' | sort -V | uniq
}

check_mariadb_macos() {
    local version="$1"
    local arch="$2"
    
    log_info "Checking MariaDB versions for macOS $version ($arch)..."
    
    if command -v brew &>/dev/null; then
        log_info "Available MariaDB versions via Homebrew:"
        brew search mariadb
    else
        log_warn "Homebrew not installed. Please install Homebrew first:"
        echo "/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
    fi
}

# API-based version checking (fallback method)
check_mariadb_api() {
    local os_id="$1"
    local os_version="$2"
    local arch="$3"
    
    log_info "Fetching MariaDB versions from API for $os_id $os_version ($arch)..."
    
    # MariaDB Foundation API endpoint
    local api_url="https://downloads.mariadb.org/rest-api/mariadb/"
    
    if command -v curl &>/dev/null; then
        local versions
        versions=$(curl -s "$api_url" | grep -o '"release_id":"[^"]*"' | cut -d'"' -f4 | sort -V)
        
        if [[ -n "$versions" ]]; then
            log_info "Available MariaDB versions from official API:"
            echo "$versions"
        else
            log_error "Failed to fetch versions from API"
            return 1
        fi
    else
        log_error "curl not available. Cannot fetch from API."
        return 1
    fi
}

# Package manager detection
detect_package_manager() {
    local os_id="$1"
    
    case "$os_id" in
        "ubuntu"|"debian")
            echo "apt"
            ;;
        "centos"|"rhel"|"rocky"|"almalinux")
            if command -v dnf &>/dev/null; then
                echo "dnf"
            else
                echo "yum"
            fi
            ;;
        "fedora")
            echo "dnf"
            ;;
        "opensuse"|"sles")
            echo "zypper"
            ;;
        "macos")
            echo "brew"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Repository URL generator
get_mariadb_repo_url() {
    local os_id="$1"
    local os_version="$2"
    local arch="$3"
    
    local base_url="https://downloads.mariadb.org/mariadb/repositories"
    
    case "$os_id" in
        "ubuntu")
            echo "$base_url/#mirror=digitalocean-nyc&distro=Ubuntu&distro_release=${os_version}&version=latest"
            ;;
        "debian")
            echo "$base_url/#mirror=digitalocean-nyc&distro=Debian&distro_release=${os_version}&version=latest"
            ;;
        "centos"|"rhel")
            echo "$base_url/#mirror=digitalocean-nyc&distro=CentOS&distro_release=${os_version}&version=latest"
            ;;
        "fedora")
            echo "$base_url/#mirror=digitalocean-nyc&distro=Fedora&distro_release=${os_version}&version=latest"
            ;;
        *)
            echo "$base_url/"
            ;;
    esac
}

# Main version checking function
check_mariadb_versions() {
    local os_id="$1"
    local os_version="$2"
    local arch="$3"
    
    log_info "System Information:"
    log_info "  OS: $os_id $os_version"
    log_info "  Architecture: $arch"
    log_info "  Package Manager: $(detect_package_manager "$os_id")"
    echo
    
    case "$os_id" in
        "ubuntu"|"debian")
            check_mariadb_ubuntu "$os_version" "$arch"
            ;;
        "centos"|"rhel"|"rocky"|"almalinux")
            check_mariadb_centos "$os_version" "$arch"
            ;;
        "fedora")
            check_mariadb_fedora "$os_version" "$arch"
            ;;
        "opensuse"|"sles")
            check_mariadb_opensuse "$os_version" "$arch"
            ;;
        "macos")
            check_mariadb_macos "$os_version" "$arch"
            ;;
        *)
            log_warn "Unsupported OS: $os_id"
            log_info "Trying API-based method..."
            check_mariadb_api "$os_id" "$os_version" "$arch"
            ;;
    esac
    
    echo
    log_info "Repository setup URL: $(get_mariadb_repo_url "$os_id" "$os_version" "$arch")"
}

# Help function
show_help() {
    cat << EOF
Usage: $SCRIPT_NAME [OPTIONS]

MariaDB Version Checker - Check available MariaDB versions for your OS

OPTIONS:
    -h, --help          Show this help message
    -d, --debug         Enable debug output
    -v, --verbose       Enable verbose output
    -l, --log FILE      Specify log file (default: $LOG_FILE)
    --os OS             Override OS detection
    --version VERSION   Override OS version detection
    --arch ARCH         Override architecture detection

EXAMPLES:
    $SCRIPT_NAME                    # Check for current system
    $SCRIPT_NAME --debug            # Enable debug output
    $SCRIPT_NAME --os ubuntu --version 20.04 --arch amd64

SUPPORTED OS:
    - Ubuntu
    - Debian
    - CentOS/RHEL
    - Fedora
    - Rocky Linux
    - AlmaLinux
    - openSUSE
    - SLES
    - macOS

EOF
}

# Main function
main() {
    local os_override=""
    local version_override=""
    local arch_override=""
    local verbose=0
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -d|--debug)
                export DEBUG=1
                log_debug "Debug mode enabled"
                shift
                ;;
            -v|--verbose)
                verbose=1
                shift
                ;;
            -l|--log)
                readonly LOG_FILE="$2"
                shift 2
                ;;
            --os)
                os_override="$2"
                shift 2
                ;;
            --version)
                version_override="$2"
                shift 2
                ;;
            --arch)
                arch_override="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Initialize log file
    echo "=== MariaDB Version Checker Log - $(date) ===" > "$LOG_FILE"
    
    log_info "Starting MariaDB version check..."
    
    # Detect system information
    local os_id="${os_override:-$(detect_os)}"
    local os_version="${version_override:-$(detect_os_version "$os_id")}"
    local arch="${arch_override:-$(detect_architecture)}"
    
    log_debug "Detected OS: $os_id"
    log_debug "Detected Version: $os_version"
    log_debug "Detected Architecture: $arch"
    
    # Check MariaDB versions
    check_mariadb_versions "$os_id" "$os_version" "$arch"
    
    log_info "MariaDB version check completed successfully!"
    log_info "Log file: $LOG_FILE"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi