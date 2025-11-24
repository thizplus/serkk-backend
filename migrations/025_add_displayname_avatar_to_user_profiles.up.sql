-- Migration: Add displayName and avatar to user_profiles table
-- Purpose: Move displayName and avatar from Auth Service to Backend Service
-- This makes profile updates simpler (1 request instead of 2)

-- Add displayName and avatar columns to user_profiles
ALTER TABLE user_profiles
ADD COLUMN display_name VARCHAR(100) NOT NULL DEFAULT '',
ADD COLUMN avatar VARCHAR(500) NOT NULL DEFAULT '';

-- Migrate existing data from users_cache to user_profiles
-- Copy displayName and avatar from users_cache (which is synced from Auth Service)
UPDATE user_profiles up
SET
  display_name = COALESCE(uc.display_name, ''),
  avatar = COALESCE(uc.avatar, '')
FROM users_cache uc
WHERE up.user_id = uc.id;

-- Create index for faster lookups by displayName
CREATE INDEX idx_user_profiles_display_name ON user_profiles(display_name);

-- Add comment to table
COMMENT ON COLUMN user_profiles.display_name IS 'User display name (moved from Auth Service for simpler profile updates)';
COMMENT ON COLUMN user_profiles.avatar IS 'User avatar URL (moved from Auth Service for simpler profile updates)';
