#!/bin/bash
set -euo pipefail

# === CONFIG ===
MYDUMPER_REPO="https://github.com/mydumper/mydumper.git"
BUILD_DIR="/tmp/build_mydumper"

# === LOGGING ===
log() { echo -e "\e[32m[INFO]\e[0m $1"; }
err() { echo -e "\e[31m[ERROR]\e[0m $1" >&2; }

# === DETEKSI OS ===
detect_os() {
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    OS_ID="$ID"
    OS_VERSION_ID="$VERSION_ID"
  else
    err "Tidak dapat mendeteksi OS"
    exit 1
  fi
}

# === INSTALL DEPENDENSI BERDASARKAN OS ===
install_deps() {
  case "$OS_ID" in
    centos|rhel|almalinux|rocky)
      log "Detected: $OS_ID $OS_VERSION_ID (RedHat-based)"
      sudo dnf install -y \
        git cmake gcc gcc-c++ make glib2-devel zlib-devel pcre2-devel \
        mariadb-connector-c-devel mariadb-devel openssl-devel
      ;;
    ubuntu|debian)
      log "Detected: $OS_ID $OS_VERSION_ID (Debian-based)"
      sudo apt update
      sudo apt install -y \
        git cmake gcc g++ make libglib2.0-dev zlib1g-dev libpcre2-dev \
        libmariadb-dev libssl-dev
      ;;
    *)
      err "OS $OS_ID tidak didukung"
      exit 1
      ;;
  esac
}


# === CLONE DAN BUILD ===
build_mydumper() {
  log "Cloning MyDumper repo..."
  rm -rf "$BUILD_DIR"
  git clone --depth=1 "$MYDUMPER_REPO" "$BUILD_DIR"

  log "Building MyDumper..."
  cd "$BUILD_DIR"
  mkdir -p build && cd build
  cmake ..
  make -j"$(nproc)"
  sudo make install
}

# === VERIFIKASI INSTALL ===
verify_install() {
  if command -v mydumper &>/dev/null; then
    log "MyDumper berhasil diinstall: $(mydumper --version)"
  else
    err "Install selesai, tapi binary tidak ditemukan!"
    exit 1
  fi
}

# === MAIN ===
main() {
  detect_os
  install_deps
  build_mydumper
  verify_install
  log "Cleanup..."
  rm -rf "$BUILD_DIR"
  log "✅ Installasi MyDumper selesai."
}

main "$@"
