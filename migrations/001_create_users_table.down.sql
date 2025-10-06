-- Drop users table and associated objects
-- This migration safely removes all user-related database objects

-- Drop trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (they will be dropped automatically with the table, but explicit for clarity)
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop the users table
DROP TABLE IF EXISTS users;

-- Note: We don't drop the uuid-ossp extension as it might be used by other tables