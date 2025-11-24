-- Migration 030: Drop users table
-- All auth data moved to Auth Service (gofiber_auth)
-- All identity data now in users_identity table (migration 026)
-- All profile data now in user_profiles table (migration 027)
-- Must run AFTER migration 029 (drop_users_cache_table)

-- Drop remaining FK constraints that reference users table
-- These should have been dropped in migration 024, but we'll ensure they're gone

-- Drop FK constraints from auto_post_logs table
ALTER TABLE IF EXISTS auto_post_logs
    DROP CONSTRAINT IF EXISTS auto_post_logs_approved_by_fkey,
    DROP CONSTRAINT IF EXISTS auto_post_logs_rejected_by_fkey;

-- Drop FK constraints from auto_post_settings table
ALTER TABLE IF EXISTS auto_post_settings
    DROP CONSTRAINT IF EXISTS auto_post_settings_bot_user_id_fkey;

-- Drop FK constraints from conversations table
ALTER TABLE IF EXISTS conversations
    DROP CONSTRAINT IF EXISTS conversations_user1_id_fkey,
    DROP CONSTRAINT IF EXISTS conversations_user2_id_fkey;

-- Drop FK constraints from files table
ALTER TABLE IF EXISTS files
    DROP CONSTRAINT IF EXISTS files_user_id_fkey;

-- Drop FK constraints from media table
ALTER TABLE IF EXISTS media
    DROP CONSTRAINT IF EXISTS media_user_id_fkey;

-- Drop FK constraints from messages table (receiver_id)
ALTER TABLE IF EXISTS messages
    DROP CONSTRAINT IF EXISTS messages_receiver_id_fkey;

-- Drop FK constraints from notification_settings table
ALTER TABLE IF EXISTS notification_settings
    DROP CONSTRAINT IF EXISTS notification_settings_user_id_fkey;

-- Drop FK constraints from push_subscriptions table
ALTER TABLE IF EXISTS push_subscriptions
    DROP CONSTRAINT IF EXISTS push_subscriptions_user_id_fkey;

-- Drop FK constraints from search_histories table
ALTER TABLE IF EXISTS search_histories
    DROP CONSTRAINT IF EXISTS search_histories_user_id_fkey;

-- Drop FK constraints from tasks table
ALTER TABLE IF EXISTS tasks
    DROP CONSTRAINT IF EXISTS tasks_user_id_fkey;

-- Drop the users table
DROP TABLE IF EXISTS users CASCADE;

-- Add comment
COMMENT ON TABLE users_identity IS 'Replaces old users table - Identity data from Auth Service';
