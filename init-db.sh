#!/bin/bash
set -e

# Create databases and users for circles.diy

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create circles.diy database and user
    CREATE USER circles WITH PASSWORD '${CIRCLES_DB_PASSWORD:-circles_secure_password}';
    CREATE DATABASE circles_diy OWNER circles;
    GRANT ALL PRIVILEGES ON DATABASE circles_diy TO circles;
EOSQL

echo "Databases created successfully"
