-- Migration: Create users_cache table for Auth Service separation
-- This table stores a cached copy of user data from gofiber_auth database
-- to support cross-database foreign key references

CREATE TABLE IF NOT EXISTS users_cache (
    -- Core Identity (synced from Auth Service)
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,

    -- Display Info (for displaying in posts, comments, etc.)
    display_name VARCHAR(255) NOT NULL,
    avatar VARCHAR(500),

    -- Role & Status
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,

    -- Sync Metadata
    synced_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for common queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_cache_email ON users_cache(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_cache_username ON users_cache(username);
CREATE INDEX IF NOT EXISTS idx_users_cache_synced_at ON users_cache(synced_at);
CREATE INDEX IF NOT EXISTS idx_users_cache_is_active ON users_cache(is_active);

-- Comment
COMMENT ON TABLE users_cache IS 'Cached copy of users from Auth Service (gofiber_auth DB)';
COMMENT ON COLUMN users_cache.id IS 'User ID - same UUID as in Auth Service';
COMMENT ON COLUMN users_cache.synced_at IS 'Last sync timestamp from Auth Service';
