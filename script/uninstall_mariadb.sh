#!/bin/bash
set -euo pipefail

# === MariaDB Uninstall Script ===
# This script provides an example of using the sfDBTools mariadb uninstall command
# and handles pre-uninstall backup recommendations

# === Logging ===
log() { echo -e "\e[32m[INFO]\e[0m $1"; }
warn() { echo -e "\e[33m[WARN]\e[0m $1"; }
err() { echo -e "\e[31m[ERROR]\e[0m $1" >&2; }

# === Configuration ===
SFDB_TOOLS="./sfDBTools"
BACKUP_DIR="./mariadb_backup_$(date +%Y%m%d_%H%M%S)"

# === Functions ===

# Check if sfDBTools binary exists
check_sfdb_tools() {
    if [[ ! -f "$SFDB_TOOLS" ]]; then
        err "sfDBTools binary not found at: $SFDB_TOOLS"
        err "Please build the project first: go build -o sfDBTools"
        exit 1
    fi
}

# Show backup recommendation
show_backup_recommendation() {
    warn "IMPORTANT: Before uninstalling MariaDB, consider backing up your databases!"
    echo
    echo "Recommended backup commands:"
    echo "  # Backup all databases:"
    echo "  $SFDB_TOOLS backup all --output-dir \"$BACKUP_DIR\""
    echo
    echo "  # Backup specific database:"
    echo "  $SFDB_TOOLS backup single --target_db mydb --output-dir \"$BACKUP_DIR\""
    echo
    echo "  # Backup with encryption:"
    echo "  $SFDB_TOOLS backup all --output-dir \"$BACKUP_DIR\" --encrypt"
    echo
}

# Interactive backup option
offer_backup() {
    read -p "Would you like to backup your databases before uninstalling? (y/n): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Starting database backup to: $BACKUP_DIR"
        
        # Create backup directory
        mkdir -p "$BACKUP_DIR"
        
        # Attempt backup
        if $SFDB_TOOLS backup all --output-dir "$BACKUP_DIR"; then
            log "Backup completed successfully at: $BACKUP_DIR"
            echo
        else
            warn "Backup failed. You may want to backup manually before proceeding."
            echo
        fi
    else
        warn "Skipping backup. Proceeding with uninstall..."
    fi
}

# Perform MariaDB uninstall
perform_uninstall() {
    log "Starting MariaDB uninstall process..."
    
    # Check if force flag should be used
    if [[ "${1:-}" == "--force" ]]; then
        log "Using force mode (no confirmation prompts)"
        $SFDB_TOOLS mariadb uninstall --force
    else
        log "Using interactive mode (with confirmation prompts)"
        $SFDB_TOOLS mariadb uninstall
    fi
}

# Show post-uninstall information
show_post_uninstall_info() {
    log "MariaDB uninstall process completed!"
    echo
    echo "📋 Post-uninstall checklist:"
    echo "  ✅ Verify no MariaDB processes are running: ps aux | grep -i maria"
    echo "  ✅ Check for remaining packages: dpkg -l | grep -i maria (Ubuntu) or rpm -qa | grep -i maria (CentOS)"
    echo "  ✅ Verify data directories are removed: ls -la /var/lib/mysql /etc/mysql"
    echo "  ✅ Check systemd services: systemctl list-units | grep -i maria"
    echo
    if [[ -d "$BACKUP_DIR" ]]; then
        echo "📁 Your backup is available at: $BACKUP_DIR"
        echo "   To restore later, use: $SFDB_TOOLS restore single --file <backup_file>"
    fi
}

# Show help
show_help() {
    echo "MariaDB Uninstall Helper Script"
    echo "==============================="
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --force     Skip backup recommendation and force uninstall"
    echo "  --help      Show this help message"
    echo
    echo "Examples:"
    echo "  $0                    # Interactive mode with backup option"
    echo "  $0 --force           # Force uninstall without backup prompts"
    echo
    echo "This script helps you safely uninstall MariaDB by:"
    echo "  1. Offering to backup your databases first"
    echo "  2. Running the sfDBTools mariadb uninstall command"
    echo "  3. Providing post-uninstall verification steps"
}

# === Main ===
main() {
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        --force)
            log "Force mode enabled - skipping backup recommendations"
            check_sfdb_tools
            perform_uninstall --force
            show_post_uninstall_info
            ;;
        "")
            log "Interactive mode - showing backup recommendations"
            check_sfdb_tools
            show_backup_recommendation
            offer_backup
            perform_uninstall
            show_post_uninstall_info
            ;;
        *)
            err "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
