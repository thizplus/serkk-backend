-- Migration: Add display_order to post_media junction table
-- Purpose: Support ordering of media in posts (Facebook-style)
-- Date: 2025-01-11

-- Add display_order column to post_media
ALTER TABLE post_media
ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

-- Add created_at for tracking (if not exists)
ALTER TABLE post_media
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Create index for efficient ordering queries
CREATE INDEX IF NOT EXISTS idx_post_media_order
ON post_media(post_id, display_order);

-- Update existing records: set order based on existing data
-- This ensures existing posts maintain their current media order
UPDATE post_media pm
SET display_order = sub.row_num - 1
FROM (
    SELECT
        post_id,
        media_id,
        ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY media_id) as row_num
    FROM post_media
) sub
WHERE pm.post_id = sub.post_id
  AND pm.media_id = sub.media_id
  AND pm.display_order = 0;  -- Only update records that haven't been set

-- Add comment to table
COMMENT ON COLUMN post_media.display_order IS 'Order of media display in post (0-indexed)';
