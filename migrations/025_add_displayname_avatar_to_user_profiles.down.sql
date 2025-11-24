-- Rollback: Remove displayName and avatar from user_profiles

-- Drop index
DROP INDEX IF EXISTS idx_user_profiles_display_name;

-- Remove columns
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS display_name,
DROP COLUMN IF EXISTS avatar;
