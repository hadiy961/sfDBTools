#!/bin/bash

# Simple MariaDB Version Checker
# Purpose: Check available MariaDB versions only
# OS: CentOS 9 Stream

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "======================================="
echo "    MariaDB Available Versions"
echo "======================================="
echo

# Check versions from default CentOS repository
echo -e "${BLUE}From CentOS 9 Stream Repository:${NC}"
echo "-----------------------------------"
# Try multiple approaches to find MariaDB packages
if dnf list available mariadb-server 2>/dev/null | grep -q mariadb-server; then
    dnf list available mariadb-server 2>/dev/null | grep mariadb-server | head -5
elif dnf list available mysql-server 2>/dev/null | grep -q mysql-server; then
    echo "MySQL Server available instead:"
    dnf list available mysql-server 2>/dev/null | grep mysql-server | head -3
else
    echo "No MariaDB/MySQL server found in default repos"
    # Check what database servers are available
    echo "Available database servers:"
    dnf list available "*sql-server" 2>/dev/null | grep -E "(mysql|maria)" || echo "None found"
fi
echo

# Check if MariaDB official repo exists and show versions
echo -e "${BLUE}From MariaDB Official Repository:${NC}"
echo "-----------------------------------"
if ls /etc/yum.repos.d/*[Mm]aria* >/dev/null 2>&1; then
    echo "MariaDB repository found:"
    ls /etc/yum.repos.d/*[Mm]aria* 2>/dev/null
    echo
    # Check if repo is enabled
    if dnf repolist enabled | grep -qi maria; then
        echo "Repository is enabled, checking packages:"
        dnf --showduplicates list available mariadb-server 2>/dev/null | grep mariadb-server | head -10 || echo "No mariadb-server packages found"
    else
        echo "Repository exists but may be disabled"
        dnf repolist disabled | grep -i maria || echo "Not found in disabled repos either"
    fi
else
    echo "MariaDB Official Repository: Not configured"
    echo "To add MariaDB repo manually:"
    echo "  sudo dnf install -y curl"
    echo "  curl -sS https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | sudo bash"
fi
echo

# More comprehensive package search
echo -e "${BLUE}Available Database Packages:${NC}"
echo "-----------------------------------"
# Search for any database-related packages
echo "Searching for MariaDB packages:"
dnf search mariadb 2>/dev/null | grep -E "^mariadb" | head -5 || echo "No MariaDB packages found via search"
echo
echo "Searching for MySQL packages:"
dnf search mysql 2>/dev/null | grep -E "^mysql" | head -5 || echo "No MySQL packages found via search"
echo

# Check what repositories are available
echo -e "${BLUE}Active Repositories:${NC}"
echo "-----------------------------------"
dnf repolist enabled | grep -E "(maria|mysql|database)" || echo "No database-specific repositories found"

echo
echo -e "${BLUE}Current Installation:${NC}"
echo "-----------------------------------"
if command -v mariadb >/dev/null 2>&1; then
    mariadb --version
elif command -v mysql >/dev/null 2>&1; then
    mysql --version
else
    echo "MariaDB/MySQL not installed"
fi

echo
echo "======================================="