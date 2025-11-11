-- Migration: Add status column to posts table for draft/published posts
-- Created: 2025-11-09

-- Add status column with default value 'published'
ALTER TABLE posts
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'published' NOT NULL;

-- Create index on status for better query performance
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);

-- Update existing posts to 'published' status
UPDATE posts SET status = 'published' WHERE status IS NULL OR status = '';

-- Add comment to explain the column
COMMENT ON COLUMN posts.status IS 'Post status: draft or published. Draft posts are only visible to the author until videos finish encoding.';
