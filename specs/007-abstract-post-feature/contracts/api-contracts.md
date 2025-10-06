# API Contracts: Abstract Post System

**Version**: v1  
**Base Path**: `/api/v1`  
**Date**: October 5, 2025  

## Authentication
All endpoints require JWT authentication via `Authorization: Bearer <token>` header unless specified otherwise.

## Content-Type
- Request: `application/json`
- Response: `application/json`
- File Upload: `multipart/form-data`

## Post Management Endpoints

### Create Post
**Endpoint**: `POST /api/v1/posts`  
**Description**: Create a new post with specified type and content.

**Request Body**:
```json
{
  "title": "string (required, max 255)",
  "content": "string (required, max 1MB)",
  "postTypeId": "uuid (required)",
  "status": "enum: draft|scheduled|published (default: draft)",
  "scheduledAt": "string (ISO 8601, optional)",
  "categoryIds": ["uuid (optional, array)"],
  "tagIds": ["uuid (optional, array)"],
  "metadata": {
    "key": "value (optional, site-specific fields)"
  }
}
```

**Response** (201 Created):
```json
{
  "id": "uuid",
  "title": "string",
  "content": "string",
  "authorId": "uuid",
  "postTypeId": "uuid",
  "status": "string",
  "scheduledAt": "string (nullable)",
  "createdAt": "string (ISO 8601)",
  "updatedAt": "string (ISO 8601)",
  "publishedAt": "string (nullable, ISO 8601)",
  "viewCount": 0,
  "isApproved": false,
  "categories": ["CategoryObject"],
  "tags": ["TagObject"],
  "metadata": {}
}
```

**Error Responses**:
- 400: Invalid request data, validation errors
- 401: Authentication required
- 403: Insufficient permissions
- 422: Business rule violations (e.g., invalid post type)

### Get Post
**Endpoint**: `GET /api/v1/posts/{id}`  
**Description**: Retrieve a single post by ID with full details.

**Path Parameters**:
- `id`: UUID of the post

**Response** (200 OK):
```json
{
  "id": "uuid",
  "title": "string",
  "content": "string",
  "authorId": "uuid", 
  "author": {
    "id": "uuid",
    "email": "string"
  },
  "postType": {
    "id": "uuid",
    "name": "string",
    "displayName": "string"
  },
  "status": "string",
  "scheduledAt": "string (nullable)",
  "createdAt": "string (ISO 8601)",
  "updatedAt": "string (ISO 8601)", 
  "publishedAt": "string (nullable)",
  "viewCount": 0,
  "isApproved": false,
  "categories": ["CategoryObject"],
  "tags": ["TagObject"],
  "attachments": ["AttachmentObject"],
  "metadata": {}
}
```

**Error Responses**:
- 401: Authentication required
- 403: Access denied (private post)
- 404: Post not found

### Update Post
**Endpoint**: `PUT /api/v1/posts/{id}`  
**Description**: Update an existing post (creates new version).

**Path Parameters**:
- `id`: UUID of the post

**Request Body**: Same as Create Post

**Response** (200 OK): Same as Get Post

**Error Responses**:
- 400: Invalid request data
- 401: Authentication required
- 403: Not post author or insufficient permissions
- 404: Post not found
- 409: Concurrent modification conflict

### Delete Post
**Endpoint**: `DELETE /api/v1/posts/{id}`  
**Description**: Soft delete a post (sets status to archived).

**Path Parameters**:
- `id`: UUID of the post

**Response** (204 No Content)

**Error Responses**:
- 401: Authentication required
- 403: Not post author or insufficient permissions
- 404: Post not found

### List Posts
**Endpoint**: `GET /api/v1/posts`  
**Description**: Retrieve paginated list of posts with filtering and sorting.

**Query Parameters**:
- `page`: integer (default: 1, min: 1)
- `limit`: integer (default: 20, max: 100)
- `status`: enum filter (published, draft, archived)
- `postTypeId`: UUID filter for post type
- `authorId`: UUID filter for author
- `categoryId`: UUID filter for category
- `tagId`: UUID filter for tag
- `search`: string for full-text search
- `sortBy`: enum (createdAt, updatedAt, publishedAt, title, viewCount)
- `sortOrder`: enum (asc, desc, default: desc)

**Response** (200 OK):
```json
{
  "posts": ["PostSummaryObject"],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5,
    "hasNext": true,
    "hasPrev": false
  },
  "filters": {
    "status": "published",
    "postTypeId": "uuid",
    "search": "query"
  }
}
```

**PostSummaryObject**:
```json
{
  "id": "uuid",
  "title": "string",
  "content": "string (truncated to 200 chars)",
  "authorId": "uuid",
  "author": {"id": "uuid", "email": "string"},
  "postType": {"id": "uuid", "name": "string", "displayName": "string"},
  "status": "string",
  "createdAt": "string",
  "publishedAt": "string (nullable)",
  "viewCount": 0,
  "categoryCount": 2,
  "tagCount": 3
}
```

## Post Type Management Endpoints

### List Post Types
**Endpoint**: `GET /api/v1/post-types`  
**Description**: Retrieve all available post types.

**Response** (200 OK):
```json
{
  "postTypes": [
    {
      "id": "uuid",
      "name": "string",
      "displayName": "string", 
      "description": "string",
      "fieldDefinitions": {},
      "isActive": true,
      "requiresApproval": false,
      "allowsScheduling": true,
      "allowsAttachments": true,
      "postCount": 25
    }
  ]
}
```

### Get Post Type
**Endpoint**: `GET /api/v1/post-types/{id}`  
**Description**: Retrieve detailed post type information including field schema.

**Response** (200 OK): Single PostType object from list response.

## Category Management Endpoints

### List Categories
**Endpoint**: `GET /api/v1/categories`  
**Description**: Retrieve hierarchical category tree.

**Query Parameters**:
- `parentId`: UUID filter for children of specific parent (optional)
- `includeEmpty`: boolean to include categories with no posts (default: true)

**Response** (200 OK):
```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "string",
      "slug": "string",
      "description": "string",
      "parentId": "uuid (nullable)",
      "sortOrder": 0,
      "isActive": true,
      "postCount": 10,
      "children": ["CategoryObject (recursive)"]
    }
  ]
}
```

### Create Category
**Endpoint**: `POST /api/v1/categories`  
**Description**: Create a new category.

**Request Body**:
```json
{
  "name": "string (required, max 100)",
  "slug": "string (optional, auto-generated if empty)",
  "description": "string (optional)",
  "parentId": "uuid (optional)",
  "sortOrder": 0
}
```

**Response** (201 Created): Single CategoryObject

## Tag Management Endpoints

### List Tags
**Endpoint**: `GET /api/v1/tags`  
**Description**: Retrieve all tags with usage statistics.

**Query Parameters**:
- `popular`: boolean to sort by usage count (default: false)
- `limit`: integer max results (default: 100, max: 500)
- `search`: string for tag name search

**Response** (200 OK):
```json
{
  "tags": [
    {
      "id": "uuid",
      "name": "string", 
      "slug": "string",
      "color": "#color",
      "description": "string",
      "usageCount": 15
    }
  ]
}
```

### Create Tag
**Endpoint**: `POST /api/v1/tags`  
**Description**: Create a new tag.

**Request Body**:
```json
{
  "name": "string (required, max 50)",
  "slug": "string (optional, auto-generated)",
  "color": "string (optional, hex color)",
  "description": "string (optional)"
}
```

**Response** (201 Created): Single TagObject

## Search Endpoints

### Global Search
**Endpoint**: `GET /api/v1/search`  
**Description**: Full-text search across posts with advanced filtering.

**Query Parameters**:
- `q`: string search query (required)
- `type`: enum (posts, categories, tags) - default: posts
- `postTypeId`: UUID filter for post type
- `categoryId`: UUID filter for category
- `tagId`: UUID filter for tag
- `dateFrom`: ISO date filter (posts after)
- `dateTo`: ISO date filter (posts before)
- `page`: integer pagination
- `limit`: integer pagination

**Response** (200 OK):
```json
{
  "results": ["PostSummaryObject"],
  "facets": {
    "postTypes": [{"id": "uuid", "name": "string", "count": 5}],
    "categories": [{"id": "uuid", "name": "string", "count": 3}],
    "tags": [{"id": "uuid", "name": "string", "count": 8}]
  },
  "pagination": "PaginationObject",
  "query": {
    "q": "search term",
    "filters": {}
  }
}
```

## Attachment Endpoints

### Upload Attachment
**Endpoint**: `POST /api/v1/posts/{postId}/attachments`  
**Description**: Upload file attachment to post.

**Content-Type**: `multipart/form-data`

**Form Fields**:
- `file`: File (required, max 100MB)
- `altText`: string (optional, max 255)
- `sortOrder`: integer (optional)

**Response** (201 Created):
```json
{
  "id": "uuid",
  "postId": "uuid", 
  "fileName": "string",
  "fileSize": 1024,
  "mimeType": "string",
  "altText": "string",
  "sortOrder": 0,
  "uploadedAt": "string (ISO 8601)",
  "url": "string (download URL)"
}
```

### Download Attachment
**Endpoint**: `GET /api/v1/attachments/{id}`  
**Description**: Download attachment file.

**Response**: File content with appropriate headers

### List Post Attachments
**Endpoint**: `GET /api/v1/posts/{postId}/attachments`  
**Description**: List all attachments for a post.

**Response** (200 OK):
```json
{
  "attachments": ["AttachmentObject"]
}
```

## Bulk Operations Endpoints

### Bulk Update Posts
**Endpoint**: `POST /api/v1/posts/bulk`  
**Description**: Perform bulk operations on multiple posts.

**Request Body**:
```json
{
  "postIds": ["uuid (required, array, max 100)"],
  "operation": "enum: update_status|add_categories|remove_categories|add_tags|remove_tags|archive",
  "data": {
    "status": "string (for update_status)",
    "categoryIds": ["uuid (for category operations)"],
    "tagIds": ["uuid (for tag operations)"]
  },
  "applyApprovalWorkflow": true
}
```

**Response** (200 OK):
```json
{
  "processedCount": 25,
  "failedCount": 2, 
  "results": [
    {
      "postId": "uuid",
      "success": true,
      "error": "string (if failed)"
    }
  ],
  "approvalRequired": true,
  "approvalRequestId": "uuid"
}
```

## Error Response Format

All error responses follow consistent format:

```json
{
  "error": {
    "code": "string (machine-readable)",
    "message": "string (human-readable)",
    "details": {},
    "timestamp": "string (ISO 8601)",
    "requestId": "uuid"
  }
}
```

**Common Error Codes**:
- `VALIDATION_ERROR`: Request validation failed
- `AUTHENTICATION_REQUIRED`: Missing or invalid JWT token
- `AUTHORIZATION_DENIED`: Insufficient permissions
- `RESOURCE_NOT_FOUND`: Requested resource doesn't exist
- `DUPLICATE_RESOURCE`: Resource already exists (unique constraint)
- `BUSINESS_RULE_VIOLATION`: Operation violates business rules
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_SERVER_ERROR`: Unexpected server error

## Rate Limiting

**Limits**:
- Authenticated users: 1000 requests/hour
- File uploads: 10 uploads/hour
- Bulk operations: 5 operations/hour

**Headers**:
- `X-RateLimit-Limit`: Total limit
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset timestamp

## Versioning

API uses URL versioning (`/api/v1/`). Breaking changes will increment major version. Non-breaking changes (new optional fields, new endpoints) will not change version.