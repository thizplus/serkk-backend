-- Migration 019: Update Auto-Post Tables with Advanced Features
-- Add new fields for title variations, batch generation, approval workflow, and metadata

-- Update auto_post_settings table
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS tone VARCHAR(50) DEFAULT 'neutral';
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS enable_variations BOOLEAN DEFAULT true;
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS variation_style JSONB;
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS require_approval BOOLEAN DEFAULT false;
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS sensitive_topics JSONB;
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS batch_size INTEGER DEFAULT 1;
ALTER TABLE auto_post_settings ADD COLUMN IF NOT EXISTS use_batch_mode BOOLEAN DEFAULT false;

-- Update auto_post_logs table
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS prompt_tokens INTEGER DEFAULT 0;
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS completion_tokens INTEGER DEFAULT 0;
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS title_variation_used VARCHAR(500);
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES users(id);
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE auto_post_logs ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

-- Create indexes for new fields
CREATE INDEX IF NOT EXISTS idx_auto_post_settings_tone ON auto_post_settings(tone);
CREATE INDEX IF NOT EXISTS idx_auto_post_settings_require_approval ON auto_post_settings(require_approval);
CREATE INDEX IF NOT EXISTS idx_auto_post_logs_approved_by ON auto_post_logs(approved_by);
CREATE INDEX IF NOT EXISTS idx_auto_post_logs_rejected_by ON auto_post_logs(rejected_by);

-- Add comments
COMMENT ON COLUMN auto_post_settings.tone IS 'Content tone: neutral, casual, professional, humorous, controversial';
COMMENT ON COLUMN auto_post_settings.enable_variations IS 'Enable title and content variations';
COMMENT ON COLUMN auto_post_settings.variation_style IS 'Settings for variations (emoji, punchlines, etc.)';
COMMENT ON COLUMN auto_post_settings.require_approval IS 'Require manual approval before posting';
COMMENT ON COLUMN auto_post_settings.sensitive_topics IS 'List of topics that require approval';
COMMENT ON COLUMN auto_post_settings.batch_size IS 'Number of posts to generate per batch';
COMMENT ON COLUMN auto_post_settings.use_batch_mode IS 'Enable batch generation mode';

COMMENT ON COLUMN auto_post_logs.metadata IS 'Additional metadata (tone, engagement predictions, etc.)';
COMMENT ON COLUMN auto_post_logs.title_variation_used IS 'Which title variation was selected';
COMMENT ON COLUMN auto_post_logs.approved_by IS 'User who approved the post';
COMMENT ON COLUMN auto_post_logs.rejected_by IS 'User who rejected the post';
