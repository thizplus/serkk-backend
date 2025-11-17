-- Migration Rollback: Remove cursor pagination indexes
-- Date: 2025-01-14
-- Description: Drop all indexes created for cursor-based pagination
--
-- Note: Use CONCURRENTLY to avoid blocking production

-- Drop Index 6: Posts for Feed (following)
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_feed_following;

-- Drop Index 5: Posts for Tag Join
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_for_tag_join;

-- Drop Index 4: Posts by Author
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_by_author_cursor;

-- Drop Index 3: Hot Feed
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_feed_hot;

-- Drop Index 2: Top Feed
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_feed_top;

-- Drop Index 1: New Feed
DROP INDEX CONCURRENTLY IF EXISTS idx_posts_feed_new;

-- Verification: Check that indexes are removed
-- SELECT indexname FROM pg_indexes WHERE tablename = 'posts' AND indexname LIKE 'idx_posts_%cursor%';
-- Should return 0 rows
