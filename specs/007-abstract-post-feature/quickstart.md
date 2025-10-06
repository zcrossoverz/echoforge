# Quickstart Guide: Abstract Post System

**Feature**: Abstract Post System  
**Date**: October 5, 2025  
**Prerequisites**: Existing Echoforge installation with user authentication

## Overview

This quickstart guide validates the abstract post system implementation by walking through key user scenarios from the feature specification. Each scenario includes setup, execution steps, and expected outcomes.

## Setup Instructions

### 1. Database Setup
```bash
# Run migrations to create post system tables
migrate -path migrations -database "postgres://user:pass@localhost/echoforge_db?sslmode=disable" up

# Verify tables created
psql -d echoforge_db -c "\dt post*"
```

**Expected Tables**:
- `posts` - Core post entity
- `post_types` - Post type definitions  
- `post_categories` - Category hierarchy
- `post_tags` - Tag system
- `post_attachments` - File attachments
- `post_versions` - Version history
- `post_metadata` - Extensible attributes
- `post_category_assignments` - Many-to-many junction
- `post_tag_assignments` - Many-to-many junction

### 2. Default Data Seeding
```bash
# Seed default post types and system data
go run cmd/seed/main.go -feature=posts
```

**Expected Seed Data**:
- Post types: blog, manga, news
- Default category: "Uncategorized"
- System tags: "Draft", "Featured"

### 3. Authentication Setup
Ensure you have a valid JWT token for API requests:
```bash
# Register test user (if not exists)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass123"}'

# Login to get JWT token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass123"}' \
  | jq -r '.token')
```

## Scenario 1: Blog Site Extension

**User Story**: Site creator extends base system for blog functionality and publishes article.

### Step 1: Verify Blog Post Type
```bash
# Get available post types
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/post-types

# Expected: Blog post type with article-specific fields
```

**Expected Response**:
```json
{
  "postTypes": [
    {
      "id": "blog-uuid",
      "name": "blog",
      "displayName": "Blog Article",
      "fieldDefinitions": {
        "summary": {"type": "string", "maxLength": 500, "required": false},
        "tags": {"type": "array", "items": {"type": "string"}, "maxItems": 10}
      },
      "isActive": true,
      "requiresApproval": false,
      "allowsScheduling": true,
      "allowsAttachments": true
    }
  ]
}
```

### Step 2: Create Blog Article
```bash
# Create blog post with metadata
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Blog Post",
    "content": "This is the content of my blog article with **markdown** support.",
    "postTypeId": "blog-uuid",
    "status": "published",
    "metadata": {
      "summary": "A sample blog post demonstrating the system",
      "tags": ["tutorial", "getting-started"]
    }
  }'
```

**Expected Response**: 201 Created with complete post object including metadata.

### Step 3: Verify Article Display
```bash
# List published posts
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/posts?status=published"

# Expected: Article appears with proper formatting and metadata
```

**Success Criteria**:
- ✅ Blog post type exists with article-specific fields
- ✅ Article created with title, content, and metadata
- ✅ Article appears in published posts list
- ✅ Metadata fields are preserved and accessible

## Scenario 2: Manga Site Extension

**User Story**: Site creator uploads manga chapter with series information and images.

### Step 1: Verify Manga Post Type
```bash
# Get manga post type details
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/post-types/{manga-uuid}
```

**Expected Response**:
```json
{
  "id": "manga-uuid",
  "name": "manga",
  "displayName": "Manga Chapter",
  "fieldDefinitions": {
    "seriesName": {"type": "string", "maxLength": 200, "required": true},
    "chapterNumber": {"type": "number", "minimum": 1, "required": true},
    "pageCount": {"type": "integer", "minimum": 1, "required": true}
  },
  "allowsAttachments": true
}
```

### Step 2: Create Manga Chapter
```bash
# Create manga chapter post
MANGA_POST_ID=$(curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "One Piece Chapter 1000",
    "content": "Epic chapter with amazing artwork!",
    "postTypeId": "manga-uuid",
    "status": "published",
    "metadata": {
      "seriesName": "One Piece",
      "chapterNumber": 1000,
      "pageCount": 20
    }
  }' | jq -r '.id')
```

### Step 3: Upload Chapter Images
```bash
# Upload manga page image
curl -X POST http://localhost:8080/api/v1/posts/$MANGA_POST_ID/attachments \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@sample-manga-page.jpg" \
  -F "altText=One Piece Chapter 1000 Page 1" \
  -F "sortOrder=1"
```

**Expected Response**: 201 Created with attachment details including download URL.

### Step 4: Verify Chapter with Images
```bash
# Get manga chapter with attachments
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/posts/$MANGA_POST_ID
```

**Success Criteria**:
- ✅ Manga post type supports series-specific fields
- ✅ Chapter created with series name, chapter number, page count
- ✅ Image attachments uploaded successfully (up to 100MB limit)
- ✅ Chapter displays with images in correct order

## Scenario 3: Multi-Type Search and Filtering

**User Story**: End user searches across different content types and filters by type.

### Step 1: Create Mixed Content
```bash
# Create news article
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Breaking News: Tech Update",
    "content": "Important technology news article...",
    "postTypeId": "news-uuid",
    "status": "published",
    "metadata": {
      "source": "Tech Today",
      "location": "San Francisco",
      "urgency": "high"
    }
  }'
```

### Step 2: Search Across All Types
```bash
# Global search across all post types
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/search?q=technology"
```

**Expected Response**:
```json
{
  "results": ["PostSummaryObject"],
  "facets": {
    "postTypes": [
      {"id": "blog-uuid", "name": "blog", "count": 2},
      {"id": "news-uuid", "name": "news", "count": 1}
    ]
  },
  "pagination": {...}
}
```

### Step 3: Filter by Post Type
```bash
# Filter search results by manga type
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/posts?postTypeId=manga-uuid"
```

**Success Criteria**:
- ✅ Search finds content across all post types
- ✅ Results can be filtered by specific post type
- ✅ Faceted search shows available post types with counts
- ✅ Each post type maintains its specific metadata

## Scenario 4: Scheduling and Approval Workflow

**User Story**: Content creator schedules post publication with approval workflow.

### Step 1: Create Scheduled Post
```bash
# Create post scheduled for future publication
SCHEDULED_TIME=$(date -d "+1 hour" -Iseconds)
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"title\": \"Scheduled Blog Post\",
    \"content\": \"This post will be published in one hour.\",
    \"postTypeId\": \"blog-uuid\",
    \"status\": \"scheduled\",
    \"scheduledAt\": \"$SCHEDULED_TIME\"
  }"
```

**Expected Response**: Post created with status "scheduled" and future scheduledAt time.

### Step 2: Verify Scheduling Behavior
```bash
# List scheduled posts (should not appear in published list)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/posts?status=scheduled"

# List published posts (should not include scheduled post)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/posts?status=published"
```

### Step 3: Test Approval Workflow (if configured)
```bash
# Create post requiring approval (depends on site configuration)
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Post Requiring Approval",
    "content": "This post needs admin approval before publishing.",
    "postTypeId": "news-uuid",
    "status": "published"
  }'
```

**Expected Response**: Post created with status "pending_approval" if workflow is configured.

**Success Criteria**:
- ✅ Posts can be scheduled with hourly precision
- ✅ Scheduled posts don't appear in published listings
- ✅ Approval workflow (if configured) prevents direct publishing
- ✅ Post status correctly reflects scheduling and approval state

## Scenario 5: Bulk Operations

**User Story**: Site administrator performs bulk operations on multiple posts.

### Step 1: Create Multiple Test Posts
```bash
# Create several test posts for bulk operations
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/v1/posts \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"Bulk Test Post $i\",
      \"content\": \"Test content for bulk operations\",
      \"postTypeId\": \"blog-uuid\",
      \"status\": \"draft\"
    }"
done
```

### Step 2: Get Post IDs for Bulk Operation
```bash
# Get draft posts for bulk operation
DRAFT_POSTS=$(curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/posts?status=draft" \
  | jq -r '.posts[].id' | head -3 | tr '\n' ',' | sed 's/,$//')
```

### Step 3: Perform Bulk Status Update
```bash
# Bulk publish multiple posts
curl -X POST http://localhost:8080/api/v1/posts/bulk \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"postIds\": [\"$(echo $DRAFT_POSTS | tr ',' '", "')\"],
    \"operation\": \"update_status\",
    \"data\": {
      \"status\": \"published\"
    },
    \"applyApprovalWorkflow\": true
  }"
```

**Expected Response**:
```json
{
  "processedCount": 3,
  "failedCount": 0,
  "results": [
    {"postId": "uuid1", "success": true},
    {"postId": "uuid2", "success": true},
    {"postId": "uuid3", "success": true}
  ],
  "approvalRequired": false
}
```

**Success Criteria**:
- ✅ Multiple posts can be updated in single operation
- ✅ Bulk operations respect approval workflow settings
- ✅ Operation results show success/failure per post
- ✅ Post count limits are enforced (max 100 posts per operation)

## Performance Validation

### Concurrent User Test
```bash
# Test concurrent post creation (simulate 10 concurrent users)
for i in {1..10}; do
  (curl -X POST http://localhost:8080/api/v1/posts \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"Concurrent Post $i\",
      \"content\": \"Testing concurrent creation\",
      \"postTypeId\": \"blog-uuid\",
      \"status\": \"published\"
    }" &)
done
wait
```

### Response Time Test
```bash
# Test response time for post listing (should be < 500ms)
time curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/posts?limit=50"
```

**Performance Criteria**:
- ✅ System handles 1000+ concurrent users per site (load testing required)
- ✅ Post retrieval operations complete within 500ms
- ✅ Search operations maintain sub-500ms response times
- ✅ File upload handles 100MB attachments without timeout

## Cleanup

### Remove Test Data
```bash
# Clean up test posts (optional)
curl -X DELETE http://localhost:8080/api/v1/posts/bulk \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"postIds": ["test-post-ids"], "operation": "archive"}'
```

## Validation Checklist

**Feature Requirements Validated**:
- ✅ FR-001: Base post entity extensible for different content types
- ✅ FR-002: Custom post types with specific fields and behaviors  
- ✅ FR-003: Common post operations (CRUD) across all types
- ✅ FR-004: Content creators can publish per post type definition
- ✅ FR-005: Post metadata maintained (dates, author, status, visibility)
- ✅ FR-006: Post categorization and tagging for organization
- ✅ FR-007: Search and filtering across different post types
- ✅ FR-008: Site-specific access controls and visibility rules
- ✅ FR-009: Referential integrity maintained across post types
- ✅ FR-010: Post versioning with 5-version cleanup policy
- ✅ FR-011: Multimedia attachments (any file type, 100MB max)
- ✅ FR-012: Post scheduling with hourly precision
- ✅ FR-013: Post status workflow with configurable approval
- ✅ FR-014: Bulk operations with approval workflow integration
- ✅ FR-015: Audit trail for post modifications

**Non-Functional Requirements Validated**:
- ✅ NFR-001: 1000+ concurrent users per site support
- ✅ NFR-002: Post operations complete within 500ms
- ✅ NFR-003: 99.9% uptime capability
- ✅ NFR-004: OWASP Top 10 security compliance
- ✅ NFR-005: Horizontal scaling support
- ✅ NFR-006: Backward compatibility maintained

**Success Indicator**: All scenarios execute successfully with expected responses and performance within specified limits.

This quickstart guide serves as both validation tool and onboarding documentation for the abstract post system implementation.