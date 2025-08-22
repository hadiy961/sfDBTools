#!/bin/bash

# sfDBTools Auto Installer for Linux
# This script downloads and installs the latest release of sfDBTools

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="hadiy961/sfDBTools"  # GitHub repository
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="sfdbtools"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# Detect OS
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $os in
        linux)
            echo "linux"
            ;;
        *)
            log_error "Unsupported OS: $os. This tool only supports Linux."
            exit 1
            ;;
    esac
}

# Check if running as root for system-wide installation
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        log_info "Running as root, installing system-wide to $INSTALL_DIR"
        return 0
    else
        log_warn "Not running as root. Installing to $HOME/.local/bin"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        
        # Add to PATH if not already there
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
            log_info "Added $INSTALL_DIR to PATH in .bashrc"
            log_warn "Please run 'source ~/.bashrc' or restart your terminal"
        fi
    fi
}

# Get latest release info
get_latest_release() {
    log_info "Fetching latest release information..."
    
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed."
        exit 1
    fi
    
    local latest_url="https://api.github.com/repos/$REPO/releases/latest"
    local release_info
    
    log_info "Fetching from: $latest_url"
    release_info=$(curl -s "$latest_url")
    if [[ $? -ne 0 ]]; then
        log_error "Failed to fetch release information"
        exit 1
    fi
    
    if [[ -z "$release_info" ]]; then
        log_error "Empty response from GitHub API"
        exit 1
    fi
    
    echo "$release_info"
}

# Download and install
install_binary() {
    local os arch release_info download_url version
    
    os=$(detect_os)
    arch=$(detect_arch)
    release_info=$(get_latest_release)
    
    # Debug: show release info
    # echo "DEBUG: Release info: $release_info" >&2
    
    # Extract version
    version=$(echo "$release_info" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "$version" ]]; then
        log_error "Could not determine latest version"
        log_error "Release info received: $release_info"
        exit 1
    fi
    
    log_info "Latest version: $version"
    log_info "Detected OS: $os, Architecture: $arch"
    
    # Remove 'v' prefix from version if present
    version_number=${version#v}
    
    # Construct download URL - match GoReleaser naming pattern
    local filename="sfDBTools_${version_number}_linux_${arch}.tar.gz"
    download_url=$(echo "$release_info" | grep "browser_download_url.*$filename" | cut -d '"' -f 4)
    
    if [[ -z "$download_url" ]]; then
        log_error "Could not find download URL for $os $arch"
        log_error "Available files:"
        echo "$release_info" | grep "browser_download_url" | cut -d '"' -f 4
        exit 1
    fi
    
    log_info "Downloading from: $download_url"
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Download
    if ! curl -L -o "$temp_dir/$filename" "$download_url"; then
        log_error "Failed to download $filename"
        exit 1
    fi
    
    # Extract
    log_info "Extracting $filename..."
    if ! tar -xzf "$temp_dir/$filename" -C "$temp_dir"; then
        log_error "Failed to extract $filename"
        exit 1
    fi
    
    # Install
    log_info "Installing to $INSTALL_DIR/$BINARY_NAME..."
    # Binary name in tar file is 'sfDBTools', but we want to install as 'sfdbtools'
    if ! cp "$temp_dir/sfDBTools" "$INSTALL_DIR/$BINARY_NAME"; then
        log_error "Failed to copy binary to $INSTALL_DIR"
        exit 1
    fi
    
    # Make executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    log_info "Successfully installed $BINARY_NAME $version"
    
    # Setup configuration
    setup_config "$temp_dir"
    
    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log_info "Installation verified. Version:"
        "$BINARY_NAME" --version 2>/dev/null || "$BINARY_NAME" version || echo "Version command not available"
    else
        log_warn "Binary installed but not found in PATH. You may need to:"
        log_warn "  1. Restart your terminal, or"
        log_warn "  2. Run: source ~/.bashrc"
        log_warn "  3. Or use full path: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Setup configuration files
setup_config() {
    local temp_dir=$1
    local config_dir
    
    # Determine config directory based on user
    if [[ $EUID -eq 0 ]]; then
        config_dir="/etc/sfdbtools"
    else
        config_dir="$HOME/.config/sfdbtools"
    fi
    
    log_info "Setting up configuration in $config_dir..."
    
    # Create config directory
    if ! mkdir -p "$config_dir"; then
        log_warn "Failed to create config directory $config_dir"
        return 1
    fi
    
    # Copy example config if it exists in the archive
    if [[ -f "$temp_dir/config/config.example.yaml" ]]; then
        if ! cp "$temp_dir/config/config.example.yaml" "$config_dir/"; then
            log_warn "Failed to copy example config file"
        else
            log_info "Example config copied to $config_dir/config.example.yaml"
        fi
    fi
    
    # Copy ready-to-use config if it exists in the archive
    if [[ -f "$temp_dir/config/config.yaml" ]]; then
        if [[ ! -f "$config_dir/config.yaml" ]]; then
            if ! cp "$temp_dir/config/config.yaml" "$config_dir/"; then
                log_warn "Failed to copy config file"
            else
                log_info "Default config copied to $config_dir/config.yaml"
                log_info "Config is ready to use! You may customize it as needed."
            fi
        else
            log_info "Config file already exists at $config_dir/config.yaml (keeping existing)"
        fi
    else
        # Create default config from example if main config doesn't exist
        if [[ ! -f "$config_dir/config.yaml" ]]; then
            if [[ -f "$config_dir/config.example.yaml" ]]; then
                if cp "$config_dir/config.example.yaml" "$config_dir/config.yaml"; then
                    log_info "Created default config at $config_dir/config.yaml"
                    log_info "Please edit $config_dir/config.yaml to configure your database settings"
                else
                    log_warn "Failed to create default config file"
                fi
            else
                log_warn "No config files found in archive, you can generate it with:"
                log_warn "  $BINARY_NAME config generate"
            fi
        else
            log_info "Config file already exists at $config_dir/config.yaml"
        fi
    fi
}

# Main execution
main() {
    log_info "sfDBTools Auto Installer"
    log_info "========================"
    
    # Check prerequisites
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed. Please install curl first."
        exit 1
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        log_error "tar is required but not installed. Please install tar first."
        exit 1
    fi
    
    check_permissions
    install_binary
    
    # Determine config directory for final message
    local config_dir
    if [[ $EUID -eq 0 ]]; then
        config_dir="/etc/sfdbtools"
    else
        config_dir="$HOME/.config/sfdbtools"
    fi
    
    log_info ""
    log_info "Installation complete! 🎉"
    log_info ""
    log_info "Configuration:"
    log_info "  Config directory: $config_dir"
    if [[ -f "$config_dir/config.yaml" ]]; then
        log_info "  Edit config: $config_dir/config.yaml"
    else
        log_info "  Generate config: $BINARY_NAME config generate"
    fi
    log_info ""
    log_info "Quick start:"
    log_info "  $BINARY_NAME --help"
    log_info "  $BINARY_NAME config show"
    log_info "  $BINARY_NAME mariadb versions"
    log_info ""
}

# Run main function
main "$@"
