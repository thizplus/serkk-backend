-- Rollback Migration 029: Restore users table
-- Warning: This rollback cannot fully restore the users table
-- You must restore from backup or re-sync from Auth Service

-- This migration cannot be automatically rolled back
-- Please restore the users table from backup if needed
SELECT 'Migration 029 cannot be rolled back automatically. Please restore users table from backup or Auth Service.' as warning;

-- If you need to restore, you should:
-- 1. Restore from database backup
-- 2. OR re-sync from Auth Service (gofiber_auth)
-- 3. OR manually create users table and copy data from users_identity + user_profiles
