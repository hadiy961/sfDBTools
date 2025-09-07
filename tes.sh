#!/bin/bash

# Tentukan versi yang diinginkan (atau ambil yang terbaru)
MARIADB_VERSION="10.11.6"  # Ganti sesuai kebutuhan

# Deteksi OS
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    OS=$ID
    VERSION_ID=$VERSION_ID
fi

echo "Mengambil informasi download untuk MariaDB $MARIADB_VERSION..."

# Ambil informasi download dari API
API_RESPONSE=$(curl -s "https://downloads.mariadb.org/rest-api/mariadb/$MARIADB_VERSION/")

# Filter berdasarkan OS Anda
case $OS in
    "ubuntu"|"debian")
        DOWNLOAD_URL=$(echo $API_RESPONSE | jq -r '.releases[0].files[] | select(.os=="Linux" and (.file_name | contains("deb"))) | .file_download_url' | head -1)
        ;;
    "centos"|"rhel"|"fedora")
        DOWNLOAD_URL=$(echo $API_RESPONSE | jq -r '.releases[0].files[] | select(.os=="Linux" and (.file_name | contains("rpm"))) | .file_download_url' | head -1)
        ;;
    *)
        DOWNLOAD_URL=$(echo $API_RESPONSE | jq -r '.releases[0].files[] | select(.os=="Linux" and (.file_name | contains("tar.gz"))) | .file_download_url' | head -1)
        ;;
esac

echo "Download URL: $DOWNLOAD_URL"

# Download file
FILENAME=$(basename $DOWNLOAD_URL)
echo "Downloading $FILENAME..."
wget $DOWNLOAD_URL

# Install berdasarkan tipe file
case $FILENAME in
    *.deb)
        sudo dpkg -i $FILENAME
        sudo apt-get install -f  # Fix dependencies jika ada
        ;;
    *.rpm)
        sudo rpm -ivh $FILENAME
        ;;
    *.tar.gz)
        tar -xzf $FILENAME
        # Manual installation steps untuk source
        cd mariadb-*
        sudo ./scripts/mysql_install_db --user=mysql
        ;;
esac

echo "MariaDB installation completed!"