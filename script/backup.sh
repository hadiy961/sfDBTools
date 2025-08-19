#!/bin/bash

# Database Backup Script Only
# Created: 2025-06-18
# This script performs database backups from the specified MySQL server

set -e

SCRIPT_DIR=${SCRIPT_DIR:-"$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"}

if [[ "${output_dir:-}" == /* ]]; then
    OUTPUT_DIR="${output_dir:-"$SCRIPT_DIR/backup_$(date +%Y%m%d)"}"
else
    OUTPUT_DIR="${output_dir:+$SCRIPT_DIR/$output_dir}"
    OUTPUT_DIR="${OUTPUT_DIR:-"$SCRIPT_DIR/backup_$(date +%Y%m%d)"}"
fi
mkdir -p "$OUTPUT_DIR"

TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
LOG_FILE="$OUTPUT_DIR/db_backup_$TIMESTAMP.log"

if command -v pigz &>/dev/null; then
    COMPRESS_CMD="pigz"
    echo -e "\e[36m[INFO]\e[0m Using pigz for compression"
else
    COMPRESS_CMD="gzip"
    echo -e "\e[33m[INFO]\e[0m Using gzip for compression (pigz not found)"
fi

separator() {
    echo -e "\e[90m-------------------------------------------------------------\e[0m"
    echo "-------------------------------------------------------------" >> "$LOG_FILE"
}
section() {
    local title="$1"
    separator
    echo -e "\e[1;34m$title\e[0m"
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] [SECTION] - $title" >> "$LOG_FILE"
    separator
}

log_message() {
    local type="${2:-INFO}"
    local timestamp
    timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    local color_reset="\e[0m"
    local color_info="\e[36m"
    local color_warn="\e[33m"
    local color_error="\e[31m"
    local color_done="\e[32m"
    local color_step="\e[1;34m"

    case "$type" in
        INFO)
            color="$color_info";;
        WARN)
            color="$color_warn";;
        ERROR)
            color="$color_error";;
        DONE)
            color="$color_done";;
        STEP)
            color="$color_step";;
        *)
            color="$color_info";;
    esac

    local log_plain="[$timestamp] [$type] - $1"
    echo -e "$color$1$color_reset"
    echo -e "$log_plain" >> "$LOG_FILE"
}

perform_backup() {
    local user=$1
    local password=$2
    local host=$3
    local port=$4
    local dbname=$5
    local output_file=$6

    log_message "Starting backup of \e[1m$dbname\e[0m from $host..." STEP
    local start_time
    start_time=$(date +%s)

    mysqldump -u"$user" -p"$password" -h"$host" -P"$port" -CfQq \
        --max-allowed-packet=1G --hex-blob --order-by-primary \
        --single-transaction --routines=true --triggers=true \
        --no-data=false --opt "$dbname" | $COMPRESS_CMD -c > "$output_file"

    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local file_size
    file_size=$(du -h "$output_file" | cut -f1)

    log_message "  Backup of \e[1m$dbname\e[0m completed in \e[32m$duration seconds\e[0m. File size: \e[36m$file_size\e[0m" DONE
}

main() {
    section "INPUT VALIDATION"
    if [ -z "${source_user:-}" ] || [ -z "${source_password:-}" ] || [ -z "${source_host:-}" ] || [ -z "${source_port:-}" ] || [ -z "${source_dbname:-}" ]; then
        log_message "ERROR: Missing required parameters. Please set all required variables (source_user, source_password, source_host, source_port, source_dbname)." ERROR
        exit 1
    fi

    source_dbname_dmart="${source_dbname}_dmart"
    sanitized_source_ip=$(echo "$source_host" | tr -cd '0-9A-Za-z._-')

    section "BACKUP SOURCE DATABASES"
    timestamp=$(date +"%Y_%m_%d")
    source_db_backup="$OUTPUT_DIR/${source_dbname}_src_${sanitized_source_ip}_${timestamp}.sql.gz"
    source_dmart_backup="$OUTPUT_DIR/${source_dbname_dmart}_src_${sanitized_source_ip}_${timestamp}.sql.gz"

    perform_backup "$source_user" "$source_password" "$source_host" "$source_port" "$source_dbname" "$source_db_backup"
    perform_backup "$source_user" "$source_password" "$source_host" "$source_port" "$source_dbname_dmart" "$source_dmart_backup"

    section "DONE"
    log_message "Database backup process completed successfully." DONE
}

if [[ "${1:-}" == "--background" ]]; then
    log_message "Running in background mode..." INFO
    nohup "$0" --run > /dev/null 2>&1 &
    echo "Process started in background. Check $LOG_FILE for progress."
    exit 0
elif [[ "${1:-}" == "--run" ]]; then
    main
else
    main
fi