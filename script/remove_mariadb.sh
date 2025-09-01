#!/bin/bash

# Pastikan script dijalankan sebagai root
if [ "$EUID" -ne 0 ]; then
  echo "Script ini harus dijalankan sebagai root (gunakan sudo)."
  exit 1
fi

echo "⚠️  Peringatan: Script ini akan menghapus MariaDB secara total, termasuk semua database dan konfigurasi. Pastikan Anda sudah mem-backup data Anda."
read -p "Apakah Anda yakin ingin melanjutkan? (y/n): " confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Operasi dibatalkan."
  exit 0
fi

# 1. Hentikan dan nonaktifkan layanan mariadb
echo "---"
echo "✅ Menghentikan layanan MariaDB..."
systemctl stop mariadb 2>/dev/null
systemctl disable mariadb 2>/dev/null

# 2. Hapus paket-paket MariaDB
echo "---"
echo "✅ Menghapus paket MariaDB..."
if command -v apt-get &> /dev/null; then
    # Untuk Debian/Ubuntu
    apt-get purge --auto-remove '^mariadb.*' '^mysql.*'
elif command -v yum &> /dev/null; then
    # Untuk CentOS/RHEL
    yum remove -y mariadb-server mariadb-client mariadb
elif command -v dnf &> /dev/null; then
    # Untuk Fedora/CentOS 8+
    dnf remove -y mariadb-server mariadb-client mariadb
else
    echo "⚠️  Manajer paket tidak dikenal. Silakan hapus paket secara manual."
fi

# 3. Hapus direktori data dan konfigurasi
echo "---"
echo "✅ Mencari dan menghapus direktori yang tersisa..."

# Direktori default MariaDB
DEFAULT_DIRS=(
  "/var/lib/mysql"
  "/etc/mysql"
  "/etc/my.cnf"
  "/etc/mysql/mariadb.conf.d"
  "/usr/lib/systemd/system/mariadb.service"
  "/var/log/mysql"
  "/var/run/mysqld"
)

for dir in "${DEFAULT_DIRS[@]}"; do
  if [ -e "$dir" ]; then
    echo "  - Menghapus direktori/file: $dir"
    rm -rf "$dir"
  fi
done

# 4. Cari direktori data di luar lokasi default
echo "---"
echo "🔍 Mencari direktori data MariaDB kustom..."
CUSTOM_DATA_DIRS=()
# Cari file my.cnf di lokasi non-standar
CUSTOM_MYCNF_FILES=$(find / -name my.cnf -type f 2>/dev/null | grep -v "/etc/my.cnf")

for file in $CUSTOM_MYCNF_FILES; do
  DATA_DIR=$(grep -oP 'datadir\s*=\s*\K[^\s]+' "$file" 2>/dev/null)
  if [ -n "$DATA_DIR" ] && [ -d "$DATA_DIR" ]; then
    CUSTOM_DATA_DIRS+=("$DATA_DIR")
  fi
done

# Hapus duplikasi jika ada
CUSTOM_DATA_DIRS=($(printf "%s\n" "${CUSTOM_DATA_DIRS[@]}" | sort -u))

if [ ${#CUSTOM_DATA_DIRS[@]} -gt 0 ]; then
  echo "⚠️  Direktori data kustom ditemukan:"
  for dir in "${CUSTOM_DATA_DIRS[@]}"; do
    echo "  - $dir"
  done
  read -p "Apakah Anda ingin menghapus direktori data kustom ini? (y/n): " confirm_custom
  if [[ "$confirm_custom" =~ ^[Yy]$ ]]; then
    for dir in "${CUSTOM_DATA_DIRS[@]}"; do
      echo "    - Menghapus: $dir"
      rm -rf "$dir"
    done
  else
    echo "    - Direktori kustom tidak dihapus."
  fi
fi

echo "---"
echo "✅ Pembersihan MariaDB selesai."
echo "Anda mungkin perlu me-reboot server untuk memastikan semua service telah dibersihkan."