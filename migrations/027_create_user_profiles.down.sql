-- Rollback Migration 027: Drop user_profiles table

-- Drop indexes
DROP INDEX IF EXISTS idx_user_profiles_karma;
DROP INDEX IF EXISTS idx_user_profiles_updated_at;
DROP INDEX IF EXISTS idx_user_profiles_display_name;

-- Drop the user_profiles table
DROP TABLE IF EXISTS user_profiles CASCADE;
