#!/bin/bash
set -e

# Create databases and users for both Circles.DIY and Docmost

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create Circles.DIY database and user
    CREATE USER circles WITH PASSWORD '${CIRCLES_DB_PASSWORD:-circles_secure_password}';
    CREATE DATABASE circles_diy OWNER circles;
    GRANT ALL PRIVILEGES ON DATABASE circles_diy TO circles;

    -- Create Docmost database and user
    CREATE USER docmost WITH PASSWORD '${DOCMOST_DB_PASSWORD:-docmost_password}';
    CREATE DATABASE docmost OWNER docmost;
    GRANT ALL PRIVILEGES ON DATABASE docmost TO docmost;
EOSQL

echo "Databases created successfully"
