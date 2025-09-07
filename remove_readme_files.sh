#!/bin/bash

# Script untuk menghapus semua file README.md dalam folder ini dan subfolder
# Script to remove all README.md files in this folder and subfolders

set -e  # Exit on any error

# Warna untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Script Penghapusan File README.md ===${NC}"
echo "Mencari file README.md dalam folder ini dan semua subfolder..."
echo

# Cari semua file README.md (case-insensitive)
readarray -t readme_files < <(find . -type f -iname "readme.md" 2>/dev/null)

if [ ${#readme_files[@]} -eq 0 ]; then
    echo -e "${YELLOW}Tidak ada file README.md yang ditemukan.${NC}"
    exit 0
fi

echo -e "${YELLOW}File README.md yang ditemukan:${NC}"
for file in "${readme_files[@]}"; do
    echo "  - $file"
done
echo

# Konfirmasi sebelum menghapus
read -p "Apakah Anda yakin ingin menghapus ${#readme_files[@]} file README.md? (y/N): " confirm

if [[ $confirm =~ ^[Yy]$ ]]; then
    echo
    echo "Menghapus file README.md..."
    
    deleted_count=0
    for file in "${readme_files[@]}"; do
        if [ -f "$file" ]; then
            rm -f "$file"
            echo -e "  ${GREEN}✓${NC} Menghapus: $file"
            ((deleted_count++))
        else
            echo -e "  ${RED}✗${NC} File tidak ditemukan: $file"
        fi
    done
    
    echo
    echo -e "${GREEN}Selesai! ${deleted_count} file README.md telah dihapus.${NC}"
else
    echo -e "${YELLOW}Operasi dibatalkan.${NC}"
fi
