-- Simple Auto-Post Queue Table
-- เก็บ topics แบบง่ายๆ ทีละหัวข้อ

CREATE TABLE IF NOT EXISTS auto_post_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Topic & Tone
    topic TEXT NOT NULL,
    tone VARCHAR(50) DEFAULT 'neutral', -- neutral, casual, professional, humorous, controversial

    -- Status
    status VARCHAR(20) DEFAULT 'pending', -- pending, completed, failed

    -- Generated Post
    post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    generated_title TEXT,
    error_message TEXT,

    -- Metadata
    tokens_used INTEGER,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,

    -- Indexes
    INDEX idx_queue_status (status),
    INDEX idx_queue_bot_user (bot_user_id),
    INDEX idx_queue_created (created_at)
);

-- Comments
COMMENT ON TABLE auto_post_queue IS 'Simple queue for auto-post topics';
COMMENT ON COLUMN auto_post_queue.topic IS 'Topic for AI to generate content';
COMMENT ON COLUMN auto_post_queue.tone IS 'Tone style for content generation';
COMMENT ON COLUMN auto_post_queue.status IS 'pending = waiting, completed = done, failed = error';
