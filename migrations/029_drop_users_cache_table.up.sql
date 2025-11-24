-- Migration 029: Drop users_cache table (replaced by users_identity)
-- After migrating to V2 architecture with users_identity, users_cache is no longer needed
-- Must run AFTER migration 028 (migrate_users_cache_to_identity)

-- Drop FK constraint from user_profiles first (if it still references users_cache)
-- This is needed if create_user_profiles.sql was run before the FK was updated
ALTER TABLE IF EXISTS user_profiles
    DROP CONSTRAINT IF EXISTS fk_user_profiles_user_cache;

-- Drop the users_cache table
DROP TABLE IF EXISTS users_cache CASCADE;
