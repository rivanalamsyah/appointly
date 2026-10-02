-- PostgreSQL initialization script for development
-- This runs only when the container is first created (fresh volume)

-- Ensure UTF-8 encoding
SET client_encoding = 'UTF8';

-- Create test database for integration tests
CREATE DATABASE appointly_test
    ENCODING 'UTF8'
    LC_COLLATE 'en_US.UTF-8'
    LC_CTYPE 'en_US.UTF-8'
    TEMPLATE template0;

GRANT ALL PRIVILEGES ON DATABASE appointly_test TO appointly;
