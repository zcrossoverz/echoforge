-- Create post system tables
-- Migration: 006_create_post_tables.up.sql
-- Feature: Abstract Post System

BEGIN;

-- Create enum types
CREATE TYPE post_status AS ENUM ('draft', 'scheduled', 'published', 'archived', 'pending_approval');
CREATE TYPE metadata_type AS ENUM ('string', 'integer', 'boolean', 'json', 'date');

-- Create post_types table (foundation for post validation)
CREATE TABLE post_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    field_definitions JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    allows_scheduling BOOLEAN NOT NULL DEFAULT TRUE,
    allows_attachments BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create post_categories table (organizational structure)
CREATE TABLE post_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES post_categories(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    post_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_slug_per_parent UNIQUE (slug, parent_id)
);

-- Create post_tags table (labeling system)
CREATE TABLE post_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    color VARCHAR(7) NOT NULL DEFAULT '#007bff',
    description TEXT,
    usage_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create posts table (core entity)
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_type_id UUID NOT NULL REFERENCES post_types(id) ON DELETE RESTRICT,
    status post_status NOT NULL DEFAULT 'draft',
    scheduled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP,
    view_count INTEGER NOT NULL DEFAULT 0,
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    CONSTRAINT valid_scheduled_time CHECK (
        (status != 'scheduled') OR (scheduled_at IS NOT NULL AND scheduled_at > NOW())
    )
);

-- Create post_attachments table (multimedia content)
CREATE TABLE post_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0 AND file_size <= 104857600), -- 100MB max
    mime_type VARCHAR(100) NOT NULL,
    storage_path VARCHAR(500) NOT NULL UNIQUE,
    alt_text VARCHAR(255),
    sort_order INTEGER NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create post_versions table (content history)
CREATE TABLE post_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    change_reason VARCHAR(255),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_version_per_post UNIQUE (post_id, version_number)
);

-- Create post_metadata table (extensible attributes)
CREATE TABLE post_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    meta_key VARCHAR(100) NOT NULL,
    meta_value TEXT NOT NULL,
    data_type metadata_type NOT NULL DEFAULT 'string',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_meta_key_per_post UNIQUE (post_id, meta_key)
);

-- Create junction tables for many-to-many relationships
CREATE TABLE post_category_assignments (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES post_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, category_id)
);

CREATE TABLE post_tag_assignments (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES post_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, tag_id)
);

-- Create indexes for performance optimization
-- Posts table indexes
CREATE INDEX idx_posts_author_status_created ON posts(author_id, status, created_at DESC);
CREATE INDEX idx_posts_type_status_created ON posts(post_type_id, status, created_at DESC);
CREATE INDEX idx_posts_published_at ON posts(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_posts_status_scheduled ON posts(status, scheduled_at) WHERE status = 'scheduled';
CREATE INDEX idx_posts_title_search ON posts USING gin(to_tsvector('english', title));
CREATE INDEX idx_posts_content_search ON posts USING gin(to_tsvector('english', content));

-- Categories table indexes  
CREATE INDEX idx_categories_parent_sort ON post_categories(parent_id, sort_order);
CREATE INDEX idx_categories_active ON post_categories(is_active) WHERE is_active = TRUE;

-- Tags table indexes
CREATE INDEX idx_tags_usage_count ON post_tags(usage_count DESC);

-- Attachments table indexes
CREATE INDEX idx_attachments_post_sort ON post_attachments(post_id, sort_order);

-- Versions table indexes
CREATE INDEX idx_versions_post_version ON post_versions(post_id, version_number DESC);
CREATE INDEX idx_versions_cleanup ON post_versions(post_id, created_at);

-- Metadata table indexes
CREATE INDEX idx_metadata_key_value ON post_metadata(meta_key, meta_value);

-- Junction table indexes (covered by primary keys, but useful for reverse lookups)
CREATE INDEX idx_post_categories_category ON post_category_assignments(category_id);
CREATE INDEX idx_post_tags_tag ON post_tag_assignments(tag_id);

-- Insert default post types
INSERT INTO post_types (name, display_name, description, field_definitions) VALUES
('blog', 'Blog Article', 'Standard blog post with article content', 
 '{"summary": {"type": "string", "maxLength": 500, "required": false}, "tags": {"type": "array", "items": {"type": "string"}, "maxItems": 10}}'),
('manga', 'Manga Chapter', 'Manga chapter with series information and page count',
 '{"seriesName": {"type": "string", "maxLength": 200, "required": true}, "chapterNumber": {"type": "number", "minimum": 1, "required": true}, "pageCount": {"type": "integer", "minimum": 1, "required": true}}'),
('news', 'News Article', 'News article with source and urgency information',
 '{"source": {"type": "string", "maxLength": 100, "required": false}, "location": {"type": "string", "maxLength": 100, "required": false}, "urgency": {"type": "string", "enum": ["low", "medium", "high"], "default": "medium"}}');

-- Insert default category
INSERT INTO post_categories (name, slug, description) VALUES
('Uncategorized', 'uncategorized', 'Default category for uncategorized posts');

-- Insert default tags
INSERT INTO post_tags (name, slug, color, description) VALUES
('Draft', 'draft', '#6c757d', 'Posts in draft status'),
('Featured', 'featured', '#28a745', 'Featured posts for homepage');

COMMIT;