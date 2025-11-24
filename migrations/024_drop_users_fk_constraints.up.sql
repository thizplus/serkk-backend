-- Migration: Drop all Foreign Key constraints that reference users table
-- Reason: Auth Service now uses separate database (gofiber_auth)
--         PostgreSQL doesn't support cross-database foreign keys
--         We'll use GORM virtual relations instead

-- Drop FK constraints from posts table
ALTER TABLE IF EXISTS posts
    DROP CONSTRAINT IF EXISTS fk_posts_author,
    DROP CONSTRAINT IF EXISTS posts_author_id_fkey;

-- Drop FK constraints from comments table
ALTER TABLE IF EXISTS comments
    DROP CONSTRAINT IF EXISTS fk_comments_author,
    DROP CONSTRAINT IF EXISTS comments_author_id_fkey;

-- Drop FK constraints from follows table
ALTER TABLE IF EXISTS follows
    DROP CONSTRAINT IF EXISTS fk_follows_follower,
    DROP CONSTRAINT IF EXISTS follows_follower_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_follows_following,
    DROP CONSTRAINT IF EXISTS follows_following_id_fkey;

-- Drop FK constraints from notifications table
ALTER TABLE IF EXISTS notifications
    DROP CONSTRAINT IF EXISTS fk_notifications_user,
    DROP CONSTRAINT IF EXISTS notifications_user_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_notifications_sender,
    DROP CONSTRAINT IF EXISTS notifications_sender_id_fkey;

-- Drop FK constraints from votes table
ALTER TABLE IF EXISTS votes
    DROP CONSTRAINT IF EXISTS fk_votes_user,
    DROP CONSTRAINT IF EXISTS votes_user_id_fkey;

-- Drop FK constraints from messages table
ALTER TABLE IF EXISTS messages
    DROP CONSTRAINT IF EXISTS fk_messages_sender,
    DROP CONSTRAINT IF EXISTS messages_sender_id_fkey;

-- Drop FK constraints from saved_posts table
ALTER TABLE IF EXISTS saved_posts
    DROP CONSTRAINT IF EXISTS fk_saved_posts_user,
    DROP CONSTRAINT IF EXISTS saved_posts_user_id_fkey;

-- Drop FK constraints from blocks table
ALTER TABLE IF EXISTS blocks
    DROP CONSTRAINT IF EXISTS fk_blocks_blocker,
    DROP CONSTRAINT IF EXISTS blocks_blocker_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_blocks_blocked,
    DROP CONSTRAINT IF EXISTS blocks_blocked_id_fkey;

-- Note: GORM relations in code will still work via virtual relations
-- The database won't enforce referential integrity, but we'll handle it at application level
