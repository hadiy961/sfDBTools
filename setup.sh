#!/bin/bash

# Client Setup Script for sfDBTools_new
# This script helps new users set up sfDBTools_new after installation

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if sfdbtools is installed
check_installation() {
    if ! command -v sfdbtools >/dev/null 2>&1; then
        log_error "sfdbtools is not installed or not in PATH"
        log_info "Please install sfdbtools first:"
        log_info "  curl -sSL https://raw.githubusercontent.com/hadiy961/sfDBTools_new/main/install.sh | bash"
        exit 1
    fi
    
    log_info "sfdbtools found: $(which sfdbtools)"
    log_info "Version: $(sfdbtools --version 2>/dev/null || sfdbtools version 2>/dev/null || echo 'Version unknown')"
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."
    
    local missing=()
    
    # Check for MariaDB/MySQL client
    if ! command -v mysql >/dev/null 2>&1 && ! command -v mariadb >/dev/null 2>&1; then
        missing+=("mysql-client or mariadb-client")
    fi
    
    # Check for rsync
    if ! command -v rsync >/dev/null 2>&1; then
        missing+=("rsync")
    fi
    
    # Check for systemctl (for service management)
    if ! command -v systemctl >/dev/null 2>&1; then
        log_warn "systemctl not found - some service management features may not work"
    fi
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing prerequisites: ${missing[*]}"
        log_info "Install them with:"
        log_info "  sudo apt update"
        log_info "  sudo apt install mysql-client rsync"
        log_info "  # or for MariaDB:"
        log_info "  sudo apt install mariadb-client rsync"
        exit 1
    fi
    
    log_info "All prerequisites are installed"
}

# Create config directory
setup_config_directory() {
    log_step "Setting up configuration directory..."
    
    local config_dir="$HOME/.config/sfdbtools"
    
    if [[ ! -d "$config_dir" ]]; then
        mkdir -p "$config_dir"
        log_info "Created config directory: $config_dir"
    else
        log_info "Config directory already exists: $config_dir"
    fi
    
    # Set appropriate permissions
    chmod 700 "$config_dir"
    
    echo "$config_dir"
}

# Generate initial config
generate_config() {
    local config_dir=$1
    local config_file="$config_dir/config.yaml"
    
    log_step "Generating initial configuration..."
    
    if [[ -f "$config_file" ]]; then
        log_warn "Config file already exists: $config_file"
        read -p "Overwrite existing config? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Keeping existing config"
            return
        fi
    fi
    
    # Use sfdbtools to generate config
    if sfdbtools config generate --output "$config_file" 2>/dev/null; then
        log_info "Config generated successfully: $config_file"
    else
        log_warn "Could not generate config automatically. Creating basic template..."
        
        cat > "$config_file" << 'EOF'
# sfDBTools_new Configuration
general:
  client_code: "YOUR_CLIENT_CODE"
  log_level: "info"
  log_file: "sfdbtools.log"

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: ""
  
backup:
  output_dir: "./backups"
  compression: true
  encryption: false
  
restore:
  temp_dir: "/tmp/sfdbtools_restore"
  
# Add more configuration as needed
EOF
        log_info "Basic config template created: $config_file"
    fi
    
    # Set appropriate permissions for config file
    chmod 600 "$config_file"
    log_warn "Config file permissions set to 600 (owner read/write only)"
}

# Show next steps
show_next_steps() {
    local config_dir=$1
    
    echo ""
    log_info "🎉 Setup completed successfully!"
    echo ""
    log_info "Next steps:"
    echo "  1. Edit your configuration:"
    echo "     nano $config_dir/config.yaml"
    echo ""
    echo "  2. Validate your configuration:"
    echo "     sfdbtools config validate"
    echo ""
    echo "  3. Test database connection:"
    echo "     sfdbtools database test"
    echo ""
    echo "  4. View available commands:"
    echo "     sfdbtools --help"
    echo ""
    echo "  5. Common operations:"
    echo "     sfdbtools backup --help"
    echo "     sfdbtools restore --help"
    echo "     sfdbtools mariadb --help"
    echo ""
    log_info "Configuration file: $config_dir/config.yaml"
    log_info "Log file location will be shown in config or set via --log-file flag"
    echo ""
}

# Show troubleshooting info
show_troubleshooting() {
    echo ""
    log_info "Troubleshooting:"
    echo "  • If commands fail, check the configuration file"
    echo "  • Ensure database credentials are correct"
    echo "  • Check log files for detailed error messages"
    echo "  • Use --verbose flag for more output"
    echo ""
    echo "  Common issues:"
    echo "    - Permission denied: Check file/directory permissions"
    echo "    - Connection failed: Verify database host/port/credentials"
    echo "    - Command not found: Ensure sfdbtools is in PATH"
    echo ""
    echo "  Support:"
    echo "    - Documentation: https://github.com/hadiy961/sfDBTools_new"
    echo "    - Issues: https://github.com/hadiy961/sfDBTools_new/issues"
    echo ""
}

# Main function
main() {
    log_info "sfDBTools_new Client Setup"
    log_info "====================="
    echo ""
    
    check_installation
    check_prerequisites
    
    local config_dir
    config_dir=$(setup_config_directory)
    
    generate_config "$config_dir"
    
    show_next_steps "$config_dir"
    show_troubleshooting
}

# Run setup if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
