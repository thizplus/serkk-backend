-- ⚠️ DEPRECATED: This file has been replaced by 022_create_user_profiles.up.sql
-- Please use the numbered migration file instead
-- This file is kept for reference only

-- Create user_profiles table for social-specific data
-- This separates auth data (users_identity) from social data (user_profiles)

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY,

    -- Profile Info
    bio TEXT,
    location VARCHAR(100),
    website VARCHAR(255),

    -- Social Stats
    karma INTEGER DEFAULT 0,
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Foreign Key to users_identity (same DB)
    CONSTRAINT fk_user_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users_identity(id)
        ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_profiles_karma ON user_profiles(karma DESC);
CREATE INDEX IF NOT EXISTS idx_user_profiles_updated_at ON user_profiles(updated_at);

-- Migrate existing data from old users table (if exists)
-- This will copy bio, location, website, karma, followers_count, following_count
-- from users table to user_profiles

-- Note: Run this only if you have old users table with data
-- INSERT INTO user_profiles (user_id, bio, location, website, karma, followers_count, following_count, created_at, updated_at)
-- SELECT id, bio, location, website, karma, followers_count, following_count, created_at, updated_at
-- FROM users
-- WHERE EXISTS (SELECT 1 FROM users_cache WHERE users_cache.id = users.id)
-- ON CONFLICT (user_id) DO UPDATE SET
--     bio = EXCLUDED.bio,
--     location = EXCLUDED.location,
--     website = EXCLUDED.website,
--     karma = EXCLUDED.karma,
--     followers_count = EXCLUDED.followers_count,
--     following_count = EXCLUDED.following_count,
--     updated_at = EXCLUDED.updated_at;

-- Auto-create user_profiles for all users_identity entries (with default values)
INSERT INTO user_profiles (user_id, created_at, updated_at)
SELECT id, created_at, updated_at
FROM users_identity
WHERE NOT EXISTS (
    SELECT 1 FROM user_profiles WHERE user_profiles.user_id = users_identity.id
);

COMMENT ON TABLE user_profiles IS 'Social-specific user data managed by Backend Service';
COMMENT ON COLUMN user_profiles.user_id IS 'Foreign key to users_identity.id (synced from Auth Service)';
COMMENT ON COLUMN user_profiles.karma IS 'User reputation score from upvotes/downvotes';
COMMENT ON COLUMN user_profiles.followers_count IS 'Number of users following this user';
COMMENT ON COLUMN user_profiles.following_count IS 'Number of users this user follows';
