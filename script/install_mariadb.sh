#!/bin/bash

# Pastikan script dijalankan sebagai root
if [ "$EUID" -ne 0 ]; then
  echo "Script ini harus dijalankan sebagai root (gunakan sudo)."
  exit 1
fi

echo "---"
echo "📦 Memulai instalasi MariaDB dengan pemilihan versi dinamis..."

## Pengecekan Awal: Periksa keberadaan instalasi MariaDB yang sudah ada
if command -v mariadb &> /dev/null; then
    CURRENT_VERSION=$(mariadb --version | grep -oP 'Ver \K\S+')
    echo "⚠️  MariaDB versi $CURRENT_VERSION sudah terinstal di sistem ini."
    echo "⛔ Skrip dihentikan untuk menghindari konflik."
    exit 0
elif command -v mysql &> /dev/null; then
    CURRENT_VERSION=$(mysql --version | grep -oP 'Ver \K\S+')
    echo "⚠️  MySQL versi $CURRENT_VERSION sudah terinstal di sistem ini."
    echo "⛔ Skrip dihentikan untuk menghindari konflik."
    exit 0
fi

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

# 1. Deteksi sistem operasi dan instal paket yang dibutuhkan
echo "---"
echo "✅ Mendeteksi sistem dan memastikan paket prasyarat terinstal..."
if command -v apt-get &> /dev/null; then
    apt-get update
    install_if_missing "wget"
    install_if_missing "curl"
    # install_if_missing "gnupg"
    install_if_missing "software-properties-common"
elif command -v yum &> /dev/null || command -v dnf &> /dev/null; then
    install_if_missing "wget"
    install_if_missing "curl"
    # install_if_missing "gnupg"
else
    echo "❌ Manajer paket tidak dikenal. Script ini hanya mendukung Debian, Ubuntu, CentOS, dan Fedora."
    exit 1
fi

# 2. Unduh skrip mariadb-repo-setup untuk mendapatkan daftar versi
echo "---"
echo "⬇️ Mengunduh skrip mariadb-repo-setup..."

TEMP_REPO_SETUP="/tmp/mariadb_repo_setup"
REPO_URL="https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

if ! curl -LsS "$REPO_URL" -o "$TEMP_REPO_SETUP"; then
    echo "❌ Gagal mengunduh skrip dari $REPO_URL. Periksa koneksi internet atau firewall Anda."
    exit 1
fi

chmod +x "$TEMP_REPO_SETUP"

# Jalankan skrip dengan opsi --check untuk mendapatkan daftar versi
echo "---"
echo "🔍 Mengambil daftar versi MariaDB yang tersedia..."
VERSION_OUTPUT=$("$TEMP_REPO_SETUP" --check 2>&1)
AVAILABLE_VERSIONS=$(echo "$VERSION_OUTPUT" | grep -oP '(?<=\s\s\s\s)\d{2}\.\d{1,2}(?=\s|\sRC|\sGA)')

# Periksa apakah kita berhasil mendapatkan versi
if [ -z "$AVAILABLE_VERSIONS" ]; then
    echo "❌ Gagal mendapatkan daftar versi MariaDB. Menggunakan versi LTS sebagai fallback."
    AVAILABLE_VERSIONS="11.8 11.4 10.11 10.6"
fi

# 3. Tampilkan pilihan versi MariaDB dan minta input
echo "---"
echo "⬇️ Silakan pilih versi MariaDB yang ingin diinstal:"
echo ""
echo "$AVAILABLE_VERSIONS" | sort -rV | nl -w2 -s') '
echo ""

PS3="Masukkan nomor pilihan Anda: "
select version_choice in $AVAILABLE_VERSIONS
do
    if [ -n "$version_choice" ]; then
        MARIADB_VERSION="$version_choice"
        break
    else
        echo "Pilihan tidak valid. Silakan coba lagi."
    fi
done

echo "---"
echo "✅ Anda memilih versi MariaDB: $MARIADB_VERSION"

# 4. Jalankan skrip konfigurasi repositori dengan versi yang dipilih
echo "---"
echo "🚀 Menjalankan skrip konfigurasi repositori untuk versi $MARIADB_VERSION..."
if ! "$TEMP_REPO_SETUP" --mariadb-server-version="$MARIADB_VERSION"; then
    echo "❌ Gagal menjalankan skrip konfigurasi repositori. Periksa output di atas untuk detail."
    exit 1
fi

# Hapus file skrip sementara setelah digunakan
rm -f "$TEMP_REPO_SETUP"

# 5. Instal MariaDB Server
echo "---"
echo "✅ Repositori berhasil dikonfigurasi. Sekarang menginstal MariaDB Server..."
if command -v apt-get &> /dev/null; then
    apt-get update
    apt-get install -y mariadb-server mariadb-client
elif command -v yum &> /dev/null || command -v dnf &> /dev/null; then
    if command -v yum &> /dev/null; then
        yum install -y MariaDB-server MariaDB-client
    else
        dnf install -y MariaDB-server MariaDB-client
    fi
fi

# 6. Mengaktifkan dan memulai layanan MariaDB
echo "---"
echo "⚙️ Mengaktifkan dan memulai layanan MariaDB..."
systemctl enable mariadb.service
systemctl start mariadb.service

# 7. Amankan instalasi MariaDB
echo "---"
echo "🛡️ Instalasi selesai!"
echo "⚠️  Selanjutnya, Anda harus mengamankan instalasi MariaDB dengan menjalankan:"
echo "    sudo mysql_secure_installation"
echo ""
echo "Mengecek status layanan MariaDB..."
if systemctl is-active --quiet mariadb; then
    echo "Layanan MariaDB sedang berjalan."
else
    echo "Layanan MariaDB tidak berjalan. Coba cek log atau status layanan."
fi

# 8. Hapus repositori MariaDB dan kunci GPG untuk kebersihan sistem
echo "---"
echo "🧹 Membersihkan repositori MariaDB dan kunci GPG..."

# Hapus file repo yang dibuat oleh skrip
REPO_FILE_APT="/etc/apt/sources.list.d/mariadb.list"
REPO_FILE_YUM_DNF="/etc/yum.repos.d/mariadb.repo"

if [ -f "$REPO_FILE_APT" ]; then
    rm -f "$REPO_FILE_APT"
    echo "  - File repo $REPO_FILE_APT berhasil dihapus."
    apt-get update # Perbarui cache paket setelah menghapus repo
elif [ -f "$REPO_FILE_YUM_DNF" ]; then
    # Hapus file repo dan file backupnya
    rm -f "$REPO_FILE_YUM_DNF"
    rm -f "$REPO_FILE_YUM_DNF.old_"* # Hapus juga file backup yang dibuat mariadb_repo_setup
    echo "  - File repo $REPO_FILE_YUM_DNF dan backupnya berhasil dihapus."
    if command -v yum &> /dev/null; then
        yum clean all # Bersihkan cache
    elif command -v dnf &> /dev/null; then
        dnf clean all # Bersihkan cache
    fi
fi

# Hapus kunci GPG
if command -v apt-key &> /dev/null; then
    # Mencari dan menghapus kunci GPG MariaDB di Debian/Ubuntu
    KEY_ID=$(apt-key list 2>/dev/null | grep -i "MariaDB Signing Key" | awk '{print $NF}')
    if [ -n "$KEY_ID" ]; then
        apt-key del "$KEY_ID"
        echo "  - Kunci GPG MariaDB berhasil dihapus."
    fi
elif command -v rpm &> /dev/null; then
    # Mencari dan menghapus kunci GPG MariaDB di CentOS/RHEL/Fedora
    KEY_ID=$(rpm -qa gpg-pubkey* | xargs -r rpm -qi 2>/dev/null | grep -A2 "MariaDB" | grep "V4" | awk '{print $NF}')
    if [ -n "$KEY_ID" ]; then
        rpm -e "$KEY_ID"
        echo "  - Kunci GPG MariaDB berhasil dihapus."
    fi
fi