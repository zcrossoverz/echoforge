-- Drop post system tables
-- Migration: 006_create_post_tables.down.sql
-- Feature: Abstract Post System

BEGIN;

-- Drop junction tables first (to avoid foreign key constraints)
DROP TABLE IF EXISTS post_tag_assignments;
DROP TABLE IF EXISTS post_category_assignments;

-- Drop dependent tables
DROP TABLE IF EXISTS post_metadata;
DROP TABLE IF EXISTS post_versions;
DROP TABLE IF EXISTS post_attachments;
DROP TABLE IF EXISTS posts;

-- Drop independent tables
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS post_categories;
DROP TABLE IF EXISTS post_types;

-- Drop enum types
DROP TYPE IF EXISTS metadata_type;
DROP TYPE IF EXISTS post_status;

COMMIT;