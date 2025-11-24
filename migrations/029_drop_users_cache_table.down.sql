-- Rollback Migration 028: Recreate users_cache table
-- This rollback recreates the users_cache table structure for downgrade purposes

CREATE TABLE IF NOT EXISTS users_cache (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    avatar VARCHAR(500),
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    synced_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_cache_email ON users_cache(email);
CREATE INDEX IF NOT EXISTS idx_users_cache_username ON users_cache(username);
