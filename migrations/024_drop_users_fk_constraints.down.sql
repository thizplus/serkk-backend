-- Rollback: Restore Foreign Key constraints that reference users table
-- WARNING: This rollback assumes users table exists in the same database
-- If Auth Service is already separated, this rollback will FAIL

-- Restore FK constraints to posts table
ALTER TABLE IF EXISTS posts
    ADD CONSTRAINT fk_posts_author
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore FK constraints to comments table
ALTER TABLE IF EXISTS comments
    ADD CONSTRAINT fk_comments_author
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore FK constraints to follows table
ALTER TABLE IF EXISTS follows
    ADD CONSTRAINT fk_follows_follower
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_follows_following
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore FK constraints to notifications table
ALTER TABLE IF EXISTS notifications
    ADD CONSTRAINT fk_notifications_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_notifications_sender
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL;

-- Restore FK constraints to votes table
ALTER TABLE IF EXISTS votes
    ADD CONSTRAINT fk_votes_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore FK constraints to messages table
ALTER TABLE IF EXISTS messages
    ADD CONSTRAINT fk_messages_sender
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore FK constraints to saved_posts table
ALTER TABLE IF EXISTS saved_posts
    ADD CONSTRAINT fk_saved_posts_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restore FK constraints to blocks table
ALTER TABLE IF EXISTS blocks
    ADD CONSTRAINT fk_blocks_blocker
    FOREIGN KEY (blocker_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_blocks_blocked
    FOREIGN KEY (blocked_id) REFERENCES users(id) ON DELETE CASCADE;

-- Note: This rollback is only for emergency use
-- Once Auth Service is separated to different database, these constraints cannot be restored
