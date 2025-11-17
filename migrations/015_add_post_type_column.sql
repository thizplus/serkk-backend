-- Migration: Add type column to posts table for categorizing posts
-- Created: 2025-01-15

-- Add type column with default value 'text'
ALTER TABLE posts
ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'text' NOT NULL;

-- Create index on type for better query performance
CREATE INDEX IF NOT EXISTS idx_posts_type ON posts(type);

-- Update existing posts to determine type based on media
-- This will set the correct type for existing posts based on their media
UPDATE posts p
SET type = CASE
    WHEN EXISTS (
        SELECT 1 FROM post_media pm
        JOIN media m ON pm.media_id = m.id
        WHERE pm.post_id = p.id AND m.type = 'video'
    ) THEN 'video'
    WHEN (
        SELECT COUNT(*) FROM post_media pm
        JOIN media m ON pm.media_id = m.id
        WHERE pm.post_id = p.id AND m.type = 'image'
    ) > 1 THEN 'gallery'
    WHEN EXISTS (
        SELECT 1 FROM post_media pm
        JOIN media m ON pm.media_id = m.id
        WHERE pm.post_id = p.id AND m.type = 'image'
    ) THEN 'image'
    ELSE 'text'
END
WHERE type = 'text';

-- Add comment to explain the column
COMMENT ON COLUMN posts.type IS 'Post type: text (no media), image (single image), gallery (multiple images), video (has video content)';
