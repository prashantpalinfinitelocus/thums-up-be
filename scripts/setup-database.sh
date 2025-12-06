#!/bin/bash

set -e

echo "Creating PostgreSQL database and user for Thums Up Backend..."

DB_NAME="thums-up"
DB_USER="prashantpal"
DB_PASSWORD="password123"

psql -U postgres <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB_NAME}') THEN
            CREATE DATABASE "thums-up";
        END IF;
    END
    \$\$;

    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '${DB_USER}') THEN
            CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';
        END IF;
    END
    \$\$;

    GRANT ALL PRIVILEGES ON DATABASE "thums-up" TO ${DB_USER};
EOSQL

echo "Database setup completed successfully!"
echo "Database: thums-up"
echo "User: prashantpal"
echo "Password: password123"

