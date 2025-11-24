-- Rollback: Drop users_cache table

DROP INDEX IF EXISTS idx_users_cache_is_active;
DROP INDEX IF EXISTS idx_users_cache_synced_at;
DROP INDEX IF EXISTS idx_users_cache_username;
DROP INDEX IF EXISTS idx_users_cache_email;

DROP TABLE IF EXISTS users_cache;
