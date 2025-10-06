CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints for data integrity
    CONSTRAINT users_email_length CHECK (char_length(email) <= 255),
    CONSTRAINT users_password_hash_length CHECK (char_length(password_hash) >= 60),
    CONSTRAINT users_site_email_unique UNIQUE (site_id, email)
);

-- Indexes for performance optimization
CREATE INDEX idx_users_site_id ON users(site_id);
CREATE INDEX idx_users_site_email ON users(site_id, email);

-- Comments for documentation
COMMENT ON TABLE users IS 'User entities with multi-tenant isolation via site_id';
COMMENT ON COLUMN users.site_id IS 'Site identifier for multi-tenant isolation';
COMMENT ON COLUMN users.email IS 'User email address, unique within site';
COMMENT ON COLUMN users.password_hash IS 'bcrypt hashed password (60+ characters)';