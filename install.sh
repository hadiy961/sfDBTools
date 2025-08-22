#!/bin/bash

# sfDBTools Installation Script
# This script automatically downloads and installs the latest version of sfDBTools

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="hadiy961"
REPO_NAME="sfDBTools"
BINARY_NAME="sfDBTools"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/sfdbtools"
LOG_DIR="/var/log/sfdbtools"

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root. Installation will be system-wide."
        INSTALL_DIR="/usr/local/bin"
        CONFIG_DIR="/etc/sfdbtools"
    else
        print_info "Running as regular user. Installing to user directory."
        INSTALL_DIR="$HOME/.local/bin"
        CONFIG_DIR="$HOME/.config/sfdbtools"
        LOG_DIR="$HOME/.local/share/sfdbtools/logs"
        
        # Create user bin directory if it doesn't exist
        mkdir -p "$INSTALL_DIR"
        
        # Add to PATH if not already there
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.bashrc"
            print_warning "Added $INSTALL_DIR to PATH. Please run 'source ~/.bashrc' or restart your terminal."
        fi
    fi
}

# Detect system architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            print_error "Supported architectures: x86_64 (amd64), aarch64/arm64"
            exit 1
            ;;
    esac
    print_info "Detected architecture: $ARCH"
}

# Get latest release version
get_latest_version() {
    print_info "Fetching latest release information..."
    
    if command -v curl >/dev/null 2>&1; then
        LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    elif command -v wget >/dev/null 2>&1; then
        LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [[ -z "$LATEST_VERSION" ]]; then
        print_error "Could not fetch latest version. Please check your internet connection."
        exit 1
    fi
    
    print_info "Latest version: $LATEST_VERSION"
}

# Download and install binary
install_binary() {
    local download_url="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/${REPO_NAME}_${LATEST_VERSION#v}_linux_${ARCH}.tar.gz"
    local temp_dir=$(mktemp -d)
    local archive_file="$temp_dir/sfdbtools.tar.gz"
    
    print_info "Downloading $REPO_NAME $LATEST_VERSION for linux_$ARCH..."
    print_info "Download URL: $download_url"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$archive_file" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$archive_file" "$download_url"
    fi
    
    if [[ ! -f "$archive_file" ]]; then
        print_error "Download failed. Please check the release exists for your architecture."
        exit 1
    fi
    
    print_info "Extracting archive..."
    tar -xzf "$archive_file" -C "$temp_dir"
    
    # Find the binary in the extracted files
    local binary_path
    if [[ -f "$temp_dir/$BINARY_NAME" ]]; then
        binary_path="$temp_dir/$BINARY_NAME"
    elif [[ -f "$temp_dir/sfdbtools" ]]; then
        binary_path="$temp_dir/sfdbtools"
    else
        print_error "Binary not found in archive"
        ls -la "$temp_dir"
        exit 1
    fi
    
    print_info "Installing binary to $INSTALL_DIR..."
    
    # Create install directory if it doesn't exist
    if [[ $EUID -eq 0 ]]; then
        mkdir -p "$INSTALL_DIR"
        cp "$binary_path" "$INSTALL_DIR/sfdbtools"
        chmod +x "$INSTALL_DIR/sfdbtools"
    else
        mkdir -p "$INSTALL_DIR"
        cp "$binary_path" "$INSTALL_DIR/sfdbtools"
        chmod +x "$INSTALL_DIR/sfdbtools"
    fi
    
    # Copy configuration files if they exist in the archive
    if [[ -d "$temp_dir/config" ]]; then
        print_info "Installing configuration files..."
        mkdir -p "$CONFIG_DIR"
        if [[ $EUID -eq 0 ]]; then
            cp -r "$temp_dir/config"/* "$CONFIG_DIR/"
            chmod -R 644 "$CONFIG_DIR"/*
        else
            cp -r "$temp_dir/config"/* "$CONFIG_DIR/"
        fi
    fi
    
    # Create log directory
    mkdir -p "$LOG_DIR"
    
    # Cleanup
    rm -rf "$temp_dir"
    
    print_success "Installation completed successfully!"
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    local binary_path
    if [[ $EUID -eq 0 ]]; then
        binary_path="/usr/local/bin/sfdbtools"
    else
        binary_path="$HOME/.local/bin/sfdbtools"
    fi
    
    if [[ -x "$binary_path" ]]; then
        local version_output
        version_output=$("$binary_path" --version 2>/dev/null || echo "Version command not available")
        print_success "sfDBTools installed at: $binary_path"
        print_info "Version: $version_output"
    else
        print_error "Installation verification failed. Binary not found or not executable."
        exit 1
    fi
}

# Main installation process
main() {
    echo -e "${BLUE}"
    echo "=================================="
    echo "    sfDBTools Installation"
    echo "=================================="
    echo -e "${NC}"
    
    # Check requirements
    if ! command -v tar >/dev/null 2>&1; then
        print_error "tar is required but not installed. Please install tar and try again."
        exit 1
    fi
    
    check_root
    detect_arch
    get_latest_version
    install_binary
    verify_installation
    
    echo -e "${GREEN}"
    echo "=================================="
    echo "    Installation Complete!"
    echo "=================================="
    echo -e "${NC}"
    
    print_info "Next steps:"
    print_info "1. Run 'sfdbtools --help' to see available commands"
    print_info "2. Run './setup.sh' to configure sfDBTools for first use"
    print_info "3. Edit configuration files in: $CONFIG_DIR"
    
    if [[ $EUID -ne 0 ]]; then
        print_warning "If 'sfdbtools' command is not found, run: source ~/.bashrc"
    fi
}

# Run main function
main "$@"
