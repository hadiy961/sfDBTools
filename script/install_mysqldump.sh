#!/bin/bash
set -euo pipefail

# === Logging ===
log()  { echo -e "\e[32m[INFO]\e[0m $1"; }
err()  { echo -e "\e[31m[ERROR]\e[0m $1" >&2; }

# === OS Detection ===
detect_os() {
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    OS_ID="$ID"
    OS_VERSION_ID="$VERSION_ID"
  else
    err "Tidak bisa mendeteksi OS"
    exit 1
  fi
}

# === Install mysqldump ===
install_mysqldump() {
  case "$OS_ID" in
    centos|rhel|almalinux|rocky)
      log "Detected: $OS_ID $OS_VERSION_ID (RedHat-based)"
      log "Menginstall mysqldump dari mariadb..."
      sudo dnf install -y mariadb
      ;;
    ubuntu|debian)
      log "Detected: $OS_ID $OS_VERSION_ID (Debian-based)"
      log "Menginstall mysqldump dari mariadb-client..."
      sudo apt update
      sudo apt install -y mariadb-client
      ;;
    *)
      err "OS $OS_ID belum didukung"
      exit 1
      ;;
  esac
}

# === Verifikasi ===
verify_install() {
  if command -v mysqldump &>/dev/null; then
    log "mysqldump berhasil diinstall: $(mysqldump --version)"
  else
    err "mysqldump tidak ditemukan setelah instalasi!"
    exit 1
  fi
}

# === Main ===
main() {
  detect_os
  install_mysqldump
  verify_install
  log "✅ mysqldump siap digunakan."
}

main "$@"
