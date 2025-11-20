-- Migration 018: Create Auto-Post Tables for AI-powered automatic posting
-- Create auto_post_settings table
CREATE TABLE IF NOT EXISTS auto_post_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    cron_schedule VARCHAR(100) NOT NULL DEFAULT '0 * * * *',
    model VARCHAR(50) NOT NULL DEFAULT 'gpt-4o-mini',
    topics JSONB NOT NULL DEFAULT '[]'::jsonb,
    max_tokens INTEGER NOT NULL DEFAULT 1500,
    temperature VARCHAR(10) NOT NULL DEFAULT '0.8',
    total_posts_generated INTEGER NOT NULL DEFAULT 0,
    last_generated_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create auto_post_logs table
CREATE TABLE IF NOT EXISTS auto_post_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_id UUID NOT NULL REFERENCES auto_post_settings(id) ON DELETE CASCADE,
    post_id UUID REFERENCES posts(id) ON DELETE SET NULL,
    topic VARCHAR(500) NOT NULL,
    generated_title VARCHAR(300),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_auto_post_settings_bot_user_id ON auto_post_settings(bot_user_id);
CREATE INDEX IF NOT EXISTS idx_auto_post_settings_is_enabled ON auto_post_settings(is_enabled);
CREATE INDEX IF NOT EXISTS idx_auto_post_settings_last_generated_at ON auto_post_settings(last_generated_at);

CREATE INDEX IF NOT EXISTS idx_auto_post_logs_setting_id ON auto_post_logs(setting_id);
CREATE INDEX IF NOT EXISTS idx_auto_post_logs_post_id ON auto_post_logs(post_id);
CREATE INDEX IF NOT EXISTS idx_auto_post_logs_status ON auto_post_logs(status);
CREATE INDEX IF NOT EXISTS idx_auto_post_logs_created_at ON auto_post_logs(created_at);

-- Create unique constraint: one setting per bot user
CREATE UNIQUE INDEX IF NOT EXISTS idx_auto_post_bot_user ON auto_post_settings(bot_user_id);

-- Add comments for documentation
COMMENT ON TABLE auto_post_settings IS 'Stores configuration for AI-powered auto-posting';
COMMENT ON TABLE auto_post_logs IS 'Stores history of auto-generated posts';

COMMENT ON COLUMN auto_post_settings.bot_user_id IS 'The bot user account that creates auto-posts';
COMMENT ON COLUMN auto_post_settings.cron_schedule IS 'Cron expression for scheduling posts (e.g., "0 * * * *" for hourly)';
COMMENT ON COLUMN auto_post_settings.topics IS 'JSON array of topics/prompts for content generation';
COMMENT ON COLUMN auto_post_settings.model IS 'OpenAI model to use (gpt-4, gpt-4o, gpt-4o-mini, etc.)';
COMMENT ON COLUMN auto_post_logs.status IS 'pending, success, or failed';
