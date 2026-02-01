-- Create user 'booklog' if it doesn't exist
DO
$$
BEGIN
  IF NOT EXISTS (
    SELECT FROM pg_catalog.pg_roles
    WHERE rolname = 'booklog'
  ) THEN
    CREATE ROLE booklog WITH LOGIN PASSWORD 'password';
  END IF;
END
$$;

-- Create development database if it doesn't exist
SELECT 'CREATE DATABASE booklog_dev'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'booklog_dev')\gexec

-- Create test database if it doesn't exist
SELECT 'CREATE DATABASE booklog_test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'booklog_test')\gexec
