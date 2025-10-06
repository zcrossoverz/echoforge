# Documentation API Contracts

## Documentation Generation Endpoints

### GET /api/v1/docs/generate
**Purpose**: Generate complete documentation suite
**Request**: 
```json
{
  "format": ["html", "pdf", "json"],
  "include_diagrams": true,
  "include_postman": true,
  "site_types": ["manga", "blog", "portfolio"]
}
```
**Response**:
```json
{
  "status": "success",
  "generated_at": "2025-10-06T10:30:00Z",
  "assets": {
    "documentation": "/generated/docs/index.html",
    "api_spec": "/generated/api/openapi.yaml",
    "postman_collection": "/generated/postman/collection.json",
    "diagrams": ["/generated/diagrams/architecture.svg"]
  },
  "validation_results": {
    "total_checks": 45,
    "passed": 43,
    "failed": 2,
    "warnings": 3
  }
}
```

### GET /api/v1/docs/validate
**Purpose**: Validate documentation completeness and accuracy
**Request**:
```json
{
  "check_links": true,
  "validate_examples": true,
  "check_diagrams": true
}
```
**Response**:
```json
{
  "status": "success",
  "validation_summary": {
    "documentation_coverage": 95.5,
    "broken_links": 0,
    "failing_examples": 2,
    "invalid_diagrams": 0
  },
  "issues": [
    {
      "type": "warning",
      "file": "guides/manga-setup.md",
      "line": 45,
      "message": "Code example may be outdated"
    }
  ]
}
```

## Site Configuration Management Endpoints

### GET /api/v1/sites/templates
**Purpose**: Retrieve available site type templates
**Response**:
```json
{
  "templates": [
    {
      "type": "manga",
      "name": "Manga Reader Site",
      "description": "Multi-chapter manga reading platform",
      "config_template": "/templates/manga-site.yaml",
      "example_site": "https://manga.example.com"
    },
    {
      "type": "blog",
      "name": "Blog Platform",
      "description": "Multi-author blogging platform",
      "config_template": "/templates/blog-site.yaml",
      "example_site": "https://blog.example.com"
    }
  ]
}
```

### POST /api/v1/sites/validate-config
**Purpose**: Validate site configuration before deployment
**Request**:
```json
{
  "site_type": "manga",
  "config": {
    "site_id": "manga-001",
    "db_dsn": "postgres://user:pass@localhost/manga_db",
    "features": {
      "comments": true,
      "ratings": true,
      "bookmarks": true
    }
  }
}
```
**Response**:
```json
{
  "valid": true,
  "warnings": [
    "Consider enabling SSL for production database connection"
  ],
  "suggestions": [
    "Add rate limiting configuration for comment endpoints"
  ]
}
```

## Postman Collection Management

### GET /api/v1/postman/collection
**Purpose**: Generate or retrieve current Postman collection
**Query Parameters**:
- `include_auth`: boolean (default: true)
- `environment`: string (dev, staging, prod)
- `format`: string (json, yaml)

**Response**:
```json
{
  "info": {
    "name": "Echoforge API Collection",
    "version": "1.0.0",
    "description": "Complete API testing suite for Echoforge"
  },
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "User Registration",
          "request": {
            "method": "POST",
            "header": [],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"test@example.com\",\n  \"password\": \"securepassword\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/api/v1/auth/register",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "auth", "register"]
            }
          }
        }
      ]
    }
  ]
}
```

### GET /api/v1/postman/environments
**Purpose**: Get environment configurations for Postman
**Response**:
```json
{
  "environments": [
    {
      "name": "Development",
      "values": [
        {
          "key": "base_url",
          "value": "http://localhost:8080",
          "enabled": true
        },
        {
          "key": "auth_token",
          "value": "",
          "enabled": true
        }
      ]
    },
    {
      "name": "Production",
      "values": [
        {
          "key": "base_url", 
          "value": "https://api.echoforge.com",
          "enabled": true
        }
      ]
    }
  ]
}
```

## Architecture Visualization Endpoints

### GET /api/v1/diagrams/list
**Purpose**: List available architecture diagrams
**Response**:
```json
{
  "diagrams": [
    {
      "id": "hexagonal-architecture",
      "title": "Hexagonal Architecture Overview",
      "type": "mermaid",
      "source": "/diagrams/architecture.mmd",
      "formats": ["svg", "png", "pdf"]
    },
    {
      "id": "data-flow",
      "title": "Multi-Tenant Data Flow",
      "type": "mermaid",
      "source": "/diagrams/data-flow.mmd",
      "formats": ["svg", "png"]
    }
  ]
}
```

### GET /api/v1/diagrams/{id}/render
**Purpose**: Render diagram in specified format
**Path Parameters**:
- `id`: diagram identifier
**Query Parameters**:
- `format`: svg, png, pdf (default: svg)
- `theme`: light, dark (default: light)

**Response**: Binary content with appropriate Content-Type header

## Error Response Schema

All endpoints follow consistent error response format:
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Configuration validation failed",
    "details": [
      {
        "field": "db_dsn",
        "issue": "Invalid database connection string format"
      }
    ],
    "timestamp": "2025-10-06T10:30:00Z",
    "request_id": "req-123456789"
  }
}
```

## Authentication Requirements

All documentation API endpoints require:
- Valid JWT token in Authorization header
- Appropriate permissions based on operation
- Rate limiting: 100 requests/minute per authenticated user
- Documentation generation endpoints: additional "docs:generate" permission