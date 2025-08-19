#!/bin/bash

# Load environment variables
source "$(dirname "$0")/set_env.sh"

# Set default values if not set
DB_USER=${sfDBTools_SOURCE_USER:-"dbaDO"}
DB_PASSWORD=${sfDBTools_SOURCE_PASSWORD:-"DataOn24!!"}
DB_HOST=${sfDBTools_SOURCE_HOST:-"localhost"}
DB_PORT=${sfDBTools_SOURCE_PORT:-"3306"}
DB_NAME=${1:-"dbsf_nbc_aaa"}

echo "⚙️ Checking and fixing privileges for $DB_USER on $DB_NAME..."

# Create MySQL command with credentials
MYSQL_CMD="mysql -u root -h $DB_HOST -P $DB_PORT -p"

cat << EOF
Please run these commands as a MySQL admin user to grant proper privileges:

# Connect to MySQL
$MYSQL_CMD

# Run these commands inside MySQL:
SHOW GRANTS FOR '$DB_USER'@'%';

# Grant all privileges on the specific database
GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'%';

# Or grant specific privileges if full access is not appropriate
GRANT SELECT, LOCK TABLES, CREATE, ALTER, DROP, INSERT, UPDATE, DELETE, 
REFERENCES, INDEX, CREATE VIEW, SHOW VIEW, TRIGGER ON \`$DB_NAME\`.* TO '$DB_USER'@'%';

# Apply changes
FLUSH PRIVILEGES;

# Verify grants again
SHOW GRANTS FOR '$DB_USER'@'%';
EOF

echo
echo "ℹ️ After fixing the privileges, try running the backup again."
echo "ℹ️ You may need to update user host pattern if it's not using '%'"
echo
