-- Essential Feed Indexes Migration
-- Only creates indexes for post feed optimization

-- ============================================
-- POSTS TABLE INDEXES (Critical for feed)
-- ============================================

-- Composite index for published posts (covers most feed queries)
CREATE INDEX IF NOT EXISTS idx_posts_feed_composite
ON posts(status, is_deleted, created_at DESC)
WHERE status = 'published' AND is_deleted = false;

-- Index for sorting by votes
CREATE INDEX IF NOT EXISTS idx_posts_votes_desc
ON posts(votes DESC)
WHERE status = 'published' AND is_deleted = false;

-- Index for author queries
CREATE INDEX IF NOT EXISTS idx_posts_author_feed
ON posts(author_id, created_at DESC)
WHERE is_deleted = false;

-- ============================================
-- POST-MEDIA JUNCTION TABLE (Critical)
-- ============================================

-- Composite index for batch loading media with order
CREATE INDEX IF NOT EXISTS idx_post_media_batch
ON post_media(post_id, display_order ASC);

-- Reverse lookup
CREATE INDEX IF NOT EXISTS idx_post_media_media_lookup
ON post_media(media_id);

-- ============================================
-- POST-TAGS JUNCTION TABLE (Critical)
-- ============================================

-- Index for batch loading tags
CREATE INDEX IF NOT EXISTS idx_post_tags_batch
ON post_tags(post_id);

-- Index for tag lookup
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_lookup
ON post_tags(tag_id);

-- ============================================
-- TAGS TABLE
-- ============================================

-- Case-insensitive tag name lookup
CREATE INDEX IF NOT EXISTS idx_tags_name_lower
ON tags(LOWER(name));

-- Popular tags
CREATE INDEX IF NOT EXISTS idx_tags_popular
ON tags(post_count DESC);

-- ============================================
-- MEDIA TABLE (for batch loading)
-- ============================================

-- User's media
CREATE INDEX IF NOT EXISTS idx_media_user
ON media(user_id);

-- ============================================
-- USERS TABLE (for author batch loading)
-- ============================================

-- User lookup (already exists as primary key, but add composite for common fields)
CREATE INDEX IF NOT EXISTS idx_users_active
ON users(id)
WHERE is_active = true;

-- Comments
COMMENT ON INDEX idx_posts_feed_composite IS 'Main index for feed queries with sorting by created_at';
COMMENT ON INDEX idx_post_media_batch IS 'Batch loading media for multiple posts with display order';
COMMENT ON INDEX idx_post_tags_batch IS 'Batch loading tags for multiple posts';
