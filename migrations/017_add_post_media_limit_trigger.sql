-- Migration: Add trigger to limit media per post to 10
-- Created: 2025-01-16
-- Purpose: Prevent posts from having more than 10 media files

-- ============================================
-- Create function to check media limit
-- ============================================

CREATE OR REPLACE FUNCTION check_post_media_limit()
RETURNS TRIGGER AS $$
DECLARE
    media_count INTEGER;
BEGIN
    -- Count existing media for this post
    SELECT COUNT(*)
    INTO media_count
    FROM post_media
    WHERE post_id = NEW.post_id;

    -- Check if adding this media would exceed the limit
    IF media_count >= 10 THEN
        RAISE EXCEPTION 'Post cannot have more than 10 media files. Current count: %', media_count
            USING ERRCODE = '23514'; -- check_violation
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Create trigger
-- ============================================

DROP TRIGGER IF EXISTS enforce_post_media_limit ON post_media;

CREATE TRIGGER enforce_post_media_limit
    BEFORE INSERT ON post_media
    FOR EACH ROW
    EXECUTE FUNCTION check_post_media_limit();

-- ============================================
-- Add comment
-- ============================================

COMMENT ON FUNCTION check_post_media_limit() IS 'Prevents posts from having more than 10 media files';
COMMENT ON TRIGGER enforce_post_media_limit ON post_media IS 'Enforces maximum 10 media files per post';
