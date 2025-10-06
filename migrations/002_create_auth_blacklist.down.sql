-- Drop auth_blacklist table and associated objects
-- This migration safely removes all JWT blacklist-related database objects

-- Drop the cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_blacklist_tokens();

-- Drop indexes (they will be dropped automatically with the table, but explicit for clarity)
DROP INDEX IF EXISTS idx_auth_blacklist_token_hash;
DROP INDEX IF EXISTS idx_auth_blacklist_expires_at;
DROP INDEX IF EXISTS idx_auth_blacklist_user_id;
DROP INDEX IF EXISTS idx_auth_blacklist_token_expires;

-- Drop the auth_blacklist table
DROP TABLE IF EXISTS auth_blacklist;