-- Migration: Add client_post_id column to posts table for idempotency
-- Created: 2025-01-16
-- Purpose: Support frontend optimistic UI and prevent duplicate post creation

-- Add client_post_id column (nullable for backward compatibility)
ALTER TABLE posts
ADD COLUMN IF NOT EXISTS client_post_id VARCHAR(255);

-- Create unique constraint to prevent duplicate posts
-- Using DO block because PostgreSQL doesn't support IF NOT EXISTS for ADD CONSTRAINT
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'unique_client_post_id'
    ) THEN
        ALTER TABLE posts
        ADD CONSTRAINT unique_client_post_id UNIQUE (client_post_id);
    END IF;
END
$$;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_posts_client_post_id ON posts(client_post_id);

-- Add comment to explain the column
COMMENT ON COLUMN posts.client_post_id IS 'Client-generated unique ID for idempotency (format: client_post_{timestamp}_{random})';
