#!/bin/bash

# Pastikan script dijalankan sebagai root untuk menginstal paket prasyarat
if [ "$EUID" -ne 0 ]; then
  echo "Script ini harus dijalankan sebagai root (gunakan sudo)."
  exit 1
fi

echo "---"
echo "🔍 Memulai pengecekan versi MariaDB yang tersedia..."

# Fungsi untuk memeriksa dan menginstal paket jika belum terinstal
install_if_missing() {
    local PKG_NAME="$1"
    if ! command -v "$PKG_NAME" &> /dev/null; then
        echo "  -> Paket '$PKG_NAME' tidak ditemukan, menginstal..."
        if command -v apt-get &> /dev/null; then
            apt-get install -y "$PKG_NAME"
        elif command -v yum &> /dev/null; then
            yum install -y "$PKG_NAME"
        elif command -v dnf &> /dev/null; then
            dnf install -y "$PKG_NAME"
        fi
    else
        echo "  -> Paket '$PKG_NAME' sudah terinstal."
    fi
}

# 1. Deteksi manajer paket dan pastikan prasyarat terinstal
echo "---"
echo "✅ Mendeteksi sistem dan memastikan paket prasyarat terinstal..."
if command -v apt-get &> /dev/null; then
    apt-get update
    install_if_missing "curl"
elif command -v yum &> /dev/null || command -v dnf &> /dev/null; then
    install_if_missing "curl"
else
    echo "❌ Manajer paket tidak dikenal. Script ini hanya mendukung Debian, Ubuntu, CentOS, dan Fedora."
    exit 1
fi

# 2. Unduh skrip mariadb-repo-setup
echo "---"
echo "⬇️ Mengunduh skrip mariadb-repo-setup untuk mengecek versi yang tersedia..."

TEMP_REPO_SETUP="/tmp/mariadb_repo_setup"
REPO_URL="https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

if ! curl -LsS "$REPO_URL" -o "$TEMP_REPO_SETUP"; then
    echo "❌ Gagal mengunduh skrip dari $REPO_URL. Periksa koneksi internet atau firewall Anda."
    exit 1
fi

chmod +x "$TEMP_REPO_SETUP"

# 3. Jalankan skrip dengan opsi --check dan tampilkan versi yang tersedia
echo "---"
echo "📋 Daftar versi MariaDB yang tersedia:"
VERSION_OUTPUT=$("$TEMP_REPO_SETUP" --check 2>&1)
AVAILABLE_VERSIONS=$(echo "$VERSION_OUTPUT" | grep -oP '(?<=\s\s\s\s)\d{2}\.\d{1,2}(?=\s|\sRC|\sGA)')

if [ -z "$AVAILABLE_VERSIONS" ]; then
    echo "❌ Gagal mendapatkan daftar versi MariaDB. Silakan coba lagi nanti."
else
    echo "$AVAILABLE_VERSIONS" | sort -rV | nl -w2 -s') '
fi

# 4. Pembersihan file sementara
echo "---"
echo "🧹 Membersihkan file sementara..."
rm -f "$TEMP_REPO_SETUP"
echo "✅ Selesai! File sementara berhasil dihapus."