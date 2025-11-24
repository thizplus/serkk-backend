-- สร้างตาราง users_identity (V2 - Minimal Identity Event)
-- เก็บเฉพาะ identity data จาก Auth Service

CREATE TABLE IF NOT EXISTS users_identity (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    synced_at TIMESTAMP,              -- เวลาที่ sync จาก Auth Service event
    deleted_at TIMESTAMP               -- Soft delete
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_identity_email ON users_identity(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_identity_username ON users_identity(username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_identity_deleted_at ON users_identity(deleted_at);

-- Comments
COMMENT ON TABLE users_identity IS 'Identity data from Auth Service (V2 - Minimal Identity Event)';
COMMENT ON COLUMN users_identity.id IS 'User ID from Auth Service';
COMMENT ON COLUMN users_identity.email IS 'Email address (unique)';
COMMENT ON COLUMN users_identity.username IS 'Username (unique)';
COMMENT ON COLUMN users_identity.synced_at IS 'Last sync time from Auth Service event';
COMMENT ON COLUMN users_identity.deleted_at IS 'Soft delete timestamp';
