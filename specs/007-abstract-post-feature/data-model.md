# Data Model: Abstract Post System

**Feature**: Abstract Post System  
**Date**: October 5, 2025  
**Status**: Complete  

## Entity Overview

The abstract post system uses a flexible entity model enabling extensibility for different content types (blog, manga, news) while maintaining referential integrity and search capabilities.

## Core Entities

### Post (Base Entity)
Primary content entity supporting all post types with common attributes.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `Title` (string, max 255): Post title, required, indexed for search
- `Content` (text): Main post content, supports markdown/HTML
- `AuthorID` (UUID): Foreign key to User.ID from existing auth system
- `PostTypeID` (UUID): Foreign key to PostType.ID
- `Status` (enum): draft, scheduled, published, archived, pending_approval
- `ScheduledAt` (timestamp, nullable): Publication time for scheduled posts
- `CreatedAt` (timestamp): Auto-generated creation time
- `UpdatedAt` (timestamp): Auto-generated last modification time
- `PublishedAt` (timestamp, nullable): Actual publication time
- `ViewCount` (integer): Read-only view counter, default 0
- `IsApproved` (boolean): Approval status, default false
- `ApprovedBy` (UUID, nullable): Foreign key to User.ID for approver
- `ApprovedAt` (timestamp, nullable): Approval timestamp

**Validation Rules**:
- Title: Required, 1-255 characters, no HTML
- Content: Required for published posts, max 1MB
- AuthorID: Must exist in User table
- PostTypeID: Must exist in PostType table
- Status: Must be valid enum value
- ScheduledAt: Must be future time if status is 'scheduled'

**Indexes**:
- Primary: ID
- Search: (Title, Status, CreatedAt)
- Author: (AuthorID, Status, CreatedAt)
- Type: (PostTypeID, Status, CreatedAt)
- Publication: (PublishedAt DESC) for public listing

### PostType (Extension Definition)
Defines structure and validation rules for different content types.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `Name` (string, max 50): Type name (blog, manga, news), unique
- `DisplayName` (string, max 100): Human-readable name
- `Description` (text): Type description and usage guidelines
- `FieldDefinitions` (JSONB): Schema for custom fields and validation rules
- `IsActive` (boolean): Whether this type accepts new posts
- `RequiresApproval` (boolean): Whether posts need approval before publishing
- `AllowsScheduling` (boolean): Whether posts can be scheduled
- `AllowsAttachments` (boolean): Whether posts can have file attachments
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Name: Required, unique, lowercase, alphanumeric + underscore only
- DisplayName: Required, 1-100 characters
- FieldDefinitions: Valid JSON schema format

**Default Types**:
```json
{
  "blog": {
    "name": "blog",
    "displayName": "Blog Article",
    "fieldDefinitions": {
      "summary": {"type": "string", "maxLength": 500, "required": false},
      "tags": {"type": "array", "items": {"type": "string"}, "maxItems": 10}
    }
  },
  "manga": {
    "name": "manga",
    "displayName": "Manga Chapter", 
    "fieldDefinitions": {
      "seriesName": {"type": "string", "maxLength": 200, "required": true},
      "chapterNumber": {"type": "number", "minimum": 1, "required": true},
      "pageCount": {"type": "integer", "minimum": 1, "required": true}
    }
  },
  "news": {
    "name": "news",
    "displayName": "News Article",
    "fieldDefinitions": {
      "source": {"type": "string", "maxLength": 100, "required": false},
      "location": {"type": "string", "maxLength": 100, "required": false},
      "urgency": {"type": "string", "enum": ["low", "medium", "high"], "default": "medium"}
    }
  }
}
```

### PostCategory (Hierarchical Organization)
Hierarchical organization system for grouping posts within a site.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `Name` (string, max 100): Category name, required
- `Slug` (string, max 100): URL-friendly identifier, unique within parent
- `Description` (text, nullable): Category description
- `ParentID` (UUID, nullable): Self-referencing foreign key for hierarchy
- `SortOrder` (integer): Display order within same parent level
- `IsActive` (boolean): Whether category accepts new posts
- `PostCount` (integer): Cached count of posts in this category
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Name: Required, 1-100 characters
- Slug: Required, unique within same parent, URL-safe format
- ParentID: Must exist in PostCategory table if not null
- SortOrder: Non-negative integer

**Indexes**:
- Primary: ID
- Hierarchy: (ParentID, SortOrder)
- Lookup: (Slug, ParentID) unique constraint

### PostTag (Flexible Labeling)
Flexible labeling system for cross-cutting post classification.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `Name` (string, max 50): Tag name, unique
- `Slug` (string, max 50): URL-friendly identifier, unique
- `Color` (string, max 7): Hex color code for UI display
- `Description` (text, nullable): Tag description
- `UsageCount` (integer): Cached count of posts using this tag
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Name: Required, unique, 1-50 characters
- Slug: Required, unique, URL-safe format
- Color: Valid hex color format (#RRGGBB)

**Indexes**:
- Primary: ID
- Lookup: Name unique, Slug unique
- Usage: (UsageCount DESC) for popular tags

### PostAttachment (Multimedia Content)
Media files and documents associated with posts.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `PostID` (UUID): Foreign key to Post.ID
- `FileName` (string, max 255): Original filename
- `FileSize` (bigint): File size in bytes, max 100MB
- `MimeType` (string, max 100): File MIME type
- `StoragePath` (string, max 500): Internal storage path
- `AltText` (string, max 255): Accessibility description
- `SortOrder` (integer): Display order within post
- `UploadedAt` (timestamp): Upload completion time
- `CreatedAt` (timestamp): Record creation time

**Validation Rules**:
- PostID: Must exist in Post table
- FileName: Required, valid filename format
- FileSize: Required, max 104,857,600 bytes (100MB)
- MimeType: Required, valid MIME type
- StoragePath: Required, unique

**Indexes**:
- Primary: ID
- Post: (PostID, SortOrder)
- Storage: StoragePath unique

### PostVersion (Content History)
Historical snapshots of post content for change tracking.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `PostID` (UUID): Foreign key to Post.ID
- `VersionNumber` (integer): Sequential version number within post
- `Title` (string, max 255): Title at this version
- `Content` (text): Content at this version
- `ChangeReason` (string, max 255): Optional reason for change
- `CreatedBy` (UUID): Foreign key to User.ID for editor
- `CreatedAt` (timestamp): Version creation time

**Validation Rules**:
- PostID: Must exist in Post table
- VersionNumber: Required, positive integer
- Title: Required, 1-255 characters
- Content: Required for published versions
- CreatedBy: Must exist in User table

**Cleanup Policy**:
- Automatic cleanup when version count > 5 per post
- Cleanup preserves first version and last 4 versions
- Cleanup triggered on new version creation

**Indexes**:
- Primary: ID
- Post: (PostID, VersionNumber DESC)
- Cleanup: (PostID, CreatedAt)

### PostMetadata (Extensible Attributes)
Extensible key-value storage for site-specific post attributes.

**Attributes**:
- `ID` (UUID): Unique identifier, primary key
- `PostID` (UUID): Foreign key to Post.ID
- `MetaKey` (string, max 100): Metadata key name
- `MetaValue` (text): Metadata value (JSON supported)
- `DataType` (enum): string, integer, boolean, json, date
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- PostID: Must exist in Post table
- MetaKey: Required, alphanumeric + underscore, unique within post
- MetaValue: Required, format validated by DataType
- DataType: Must be valid enum value

**Indexes**:
- Primary: ID
- Post: (PostID, MetaKey) unique constraint
- Search: (MetaKey, MetaValue) for cross-post queries

## Relationship Mappings

### Many-to-Many: Posts and Categories
**Junction Table**: `post_categories`
- `PostID` (UUID): Foreign key to Post.ID
- `CategoryID` (UUID): Foreign key to PostCategory.ID
- `CreatedAt` (timestamp): Assignment time

**Constraints**:
- Primary key: (PostID, CategoryID)
- Both foreign keys required and must exist

### Many-to-Many: Posts and Tags
**Junction Table**: `post_tags`
- `PostID` (UUID): Foreign key to Post.ID
- `TagID` (UUID): Foreign key to PostTag.ID
- `CreatedAt` (timestamp): Assignment time

**Constraints**:
- Primary key: (PostID, TagID)
- Both foreign keys required and must exist

## Database Schema Summary

```sql
-- Core post entity
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id),
    post_type_id UUID NOT NULL REFERENCES post_types(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('draft', 'scheduled', 'published', 'archived', 'pending_approval')),
    scheduled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP,
    view_count INTEGER NOT NULL DEFAULT 0,
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP
);

-- Additional tables follow similar pattern with appropriate constraints and indexes
```

## Migration Strategy

**Order of Creation**:
1. `post_types` (foundation for post validation)
2. `post_categories` (organizational structure)
3. `post_tags` (labeling system)
4. `posts` (core entity with foreign keys)
5. `post_attachments` (dependent on posts)
6. `post_versions` (dependent on posts)
7. `post_metadata` (dependent on posts)
8. Junction tables: `post_categories`, `post_tags`

**Data Seeding**:
- Default post types (blog, manga, news)
- System categories (Uncategorized)
- System tags (Draft, Featured)

**Indexing Strategy**:
- Primary keys with UUID
- Foreign key constraints with indexes
- Composite indexes for common query patterns
- Full-text search indexes on content fields

This data model provides the foundation for extensible post management while maintaining referential integrity and supporting the performance requirements specified in the functional requirements.