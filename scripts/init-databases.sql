-- Create databases if they don't exist
SELECT 'CREATE DATABASE "thums-up"'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'thums-up')\gexec

SELECT 'CREATE DATABASE "strapi-cms"'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'strapi-cms')\gexec

-- Grant all privileges
GRANT ALL PRIVILEGES ON DATABASE "thums-up" TO prashantpal;
GRANT ALL PRIVILEGES ON DATABASE "strapi-cms" TO prashantpal;

