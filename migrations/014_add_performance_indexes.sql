-- Performance Indexes Migration
-- Creates composite indexes for frequently queried columns

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email, is_active);
CREATE INDEX IF NOT EXISTS idx_users_username_active ON users(username, is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Posts table indexes
CREATE INDEX IF NOT EXISTS idx_posts_author_status ON posts(author_id, status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_posts_created_status ON posts(created_at DESC, status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_posts_votes_status ON posts(votes DESC, status) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_posts_hot_score ON posts(((votes::float / POWER(EXTRACT(EPOCH FROM (NOW() - created_at))/3600 + 2, 1.5))) DESC) WHERE is_deleted = false AND status = 'published';

-- Comments table indexes
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments(post_id, created_at DESC) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_comments_author ON comments(author_id, created_at DESC);

-- Votes table indexes
CREATE INDEX IF NOT EXISTS idx_votes_target_type ON votes(target_id, target_type, vote_type);
CREATE INDEX IF NOT EXISTS idx_votes_user_target ON votes(user_id, target_id, target_type);

-- Follows table indexes
CREATE INDEX IF NOT EXISTS idx_follows_follower ON follows(follower_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_follows_following ON follows(following_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_follows_lookup ON follows(follower_id, following_id);

-- Saved Posts table indexes
CREATE INDEX IF NOT EXISTS idx_saved_posts_user_created ON saved_posts(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_saved_posts_post ON saved_posts(post_id);

-- Notifications table indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- Messages table indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation_created ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(receiver_id, is_read, created_at DESC) WHERE is_read = false;

-- Conversations table indexes
CREATE INDEX IF NOT EXISTS idx_conversations_last_message ON conversations(last_message_at DESC);

-- Media table indexes
CREATE INDEX IF NOT EXISTS idx_media_user_created ON media(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_type ON media(type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_unused ON media(usage_count, created_at) WHERE usage_count = 0;

-- Post-Tag junction table indexes
CREATE INDEX IF NOT EXISTS idx_post_tags_post ON post_tags(post_id);
CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag_id);

-- Post-Media junction table indexes
CREATE INDEX IF NOT EXISTS idx_post_media_post_order ON post_media(post_id, display_order);
CREATE INDEX IF NOT EXISTS idx_post_media_media ON post_media(media_id);

-- Tags table indexes
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_post_count ON tags(post_count DESC);

-- Full-text search indexes (if using PostgreSQL)
CREATE INDEX IF NOT EXISTS idx_posts_search ON posts USING gin(to_tsvector('english', title || ' ' || content));
CREATE INDEX IF NOT EXISTS idx_users_search ON users USING gin(to_tsvector('english', username || ' ' || display_name));
CREATE INDEX IF NOT EXISTS idx_tags_search ON tags USING gin(to_tsvector('english', name));

-- Partial indexes for common filters
CREATE INDEX IF NOT EXISTS idx_posts_published ON posts(created_at DESC) WHERE status = 'published' AND is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_posts_draft ON posts(created_at DESC) WHERE status = 'draft';

COMMENT ON INDEX idx_users_email_active IS 'Composite index for user login queries';
COMMENT ON INDEX idx_posts_hot_score IS 'Index for hot/trending posts algorithm';
COMMENT ON INDEX idx_posts_search IS 'Full-text search index for posts';
COMMENT ON INDEX idx_messages_unread IS 'Partial index for unread messages';
