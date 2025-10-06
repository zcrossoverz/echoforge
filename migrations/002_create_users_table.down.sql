-- Drop indexes first
DROP INDEX IF EXISTS idx_users_site_email;
DROP INDEX IF EXISTS idx_users_site_id;

-- Drop table
DROP TABLE IF EXISTS users;