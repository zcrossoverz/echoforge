-- Create auth_blacklist table for JWT token blacklisting (logout functionality)
-- This table stores blacklisted JWT tokens to prevent their reuse after logout

CREATE TABLE auth_blacklist (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    blacklisted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE
);

-- Create unique index on token_hash for fast lookups and prevent duplicates
CREATE UNIQUE INDEX idx_auth_blacklist_token_hash ON auth_blacklist(token_hash);

-- Create index on expires_at for efficient cleanup of expired tokens
CREATE INDEX idx_auth_blacklist_expires_at ON auth_blacklist(expires_at);

-- Create index on user_id for user-specific token management
CREATE INDEX idx_auth_blacklist_user_id ON auth_blacklist(user_id);

-- Create composite index for common queries (token lookup with expiration check)
CREATE INDEX idx_auth_blacklist_token_expires ON auth_blacklist(token_hash, expires_at);

-- Add comments for documentation
COMMENT ON TABLE auth_blacklist IS 'Blacklisted JWT tokens to prevent reuse after logout';
COMMENT ON COLUMN auth_blacklist.id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN auth_blacklist.token_hash IS 'SHA-256 hash of the blacklisted JWT token';
COMMENT ON COLUMN auth_blacklist.expires_at IS 'Original expiration time of the JWT token';
COMMENT ON COLUMN auth_blacklist.blacklisted_at IS 'Timestamp when token was blacklisted (logout time)';
COMMENT ON COLUMN auth_blacklist.user_id IS 'User who owned the blacklisted token (nullable for cleanup purposes)';

-- Create function to automatically clean up expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_blacklist_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM auth_blacklist 
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_blacklist_tokens() IS 'Removes expired tokens from blacklist and returns count of deleted rows';