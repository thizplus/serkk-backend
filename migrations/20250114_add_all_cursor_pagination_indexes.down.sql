-- Rollback Migration: Remove all cursor pagination indexes
-- Date: 2025-01-14
-- Description: Drop all indexes created for cursor-based pagination

-- PHASE 3: Drop Saved Posts indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_saved_posts_by_user_cursor;

-- PHASE 3: Drop Follows indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_follows_following_cursor;
DROP INDEX CONCURRENTLY IF EXISTS idx_follows_followers_cursor;

-- PHASE 2: Drop Notifications indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_unread_cursor;
DROP INDEX CONCURRENTLY IF EXISTS idx_notifications_by_user_cursor;

-- PHASE 2: Drop Comments indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_comments_by_votes_cursor;
DROP INDEX CONCURRENTLY IF EXISTS idx_comments_replies_cursor;
DROP INDEX CONCURRENTLY IF EXISTS idx_comments_by_author_cursor;
DROP INDEX CONCURRENTLY IF EXISTS idx_comments_by_post_cursor;

-- Note: Posts indexes (Phase 1) have separate rollback script
-- See: 20250114_add_cursor_pagination_indexes.down.sql

-- Verification: Check that indexes are removed
-- SELECT indexname FROM pg_indexes WHERE indexname LIKE '%_cursor%';
-- Should return only post-related indexes (if Phase 1 not rolled back)
