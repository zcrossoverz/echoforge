# API Contracts: Authentication Endpoints

**Feature**: Database Connection and Authentication APIs  
**Date**: October 4, 2025  
**Base URL**: `/api/v1`

## Authentication Endpoints

### 1. User Registration

**Endpoint**: `POST /api/v1/auth/register`  
**Purpose**: Create a new user account  
**Authentication**: None required  
**Rate Limit**: 5 requests per minute per IP

#### Request

**Headers**:
```
Content-Type: application/json
```

**Body**:
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Validation Rules**:
- `email`: required, valid email format, max 320 characters, unique
- `password`: required, min 8 characters, at least one letter and number

#### Response

**Success (201 Created)**:
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "created_at": "2025-10-04T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-05T10:30:00Z"
  }
}
```

**Error (400 Bad Request)**:
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "Email is required"
    },
    {
      "field": "password", 
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

**Error (409 Conflict)**:
```json
{
  "success": false,
  "message": "Email already exists"
}
```

**Error (429 Too Many Requests)**:
```json
{
  "success": false,
  "message": "Too many registration attempts. Please try again later.",
  "retry_after": 60
}
```

### 2. User Login

**Endpoint**: `POST /api/v1/auth/login`  
**Purpose**: Authenticate user and return JWT token  
**Authentication**: None required  
**Rate Limit**: 5 requests per minute per IP

#### Request

**Headers**:
```
Content-Type: application/json
```

**Body**:
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Validation Rules**:
- `email`: required, valid email format
- `password`: required, min 1 character

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "created_at": "2025-10-04T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-05T10:30:00Z"
  }
}
```

**Error (401 Unauthorized)**:
```json
{
  "success": false,
  "message": "Invalid email or password"
}
```

**Error (429 Too Many Requests)**:
```json
{
  "success": false,
  "message": "Too many login attempts. Please try again later.",
  "retry_after": 60
}
```

### 3. User Logout

**Endpoint**: `POST /api/v1/auth/logout`  
**Purpose**: Invalidate current JWT token  
**Authentication**: Bearer token required  
**Rate Limit**: 10 requests per minute per user

#### Request

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Body**: Empty `{}`

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Error (401 Unauthorized)**:
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

### 4. Get User Profile

**Endpoint**: `GET /api/v1/auth/profile`  
**Purpose**: Get authenticated user's profile information  
**Authentication**: Bearer token required  
**Rate Limit**: 30 requests per minute per user

#### Request

**Headers**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "created_at": "2025-10-04T10:30:00Z",
      "updated_at": "2025-10-04T10:30:00Z"
    }
  }
}
```

**Error (401 Unauthorized)**:
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

## Health Check Endpoint

### System Health

**Endpoint**: `GET /api/v1/health`  
**Purpose**: Check system and database connectivity  
**Authentication**: None required  
**Rate Limit**: 60 requests per minute per IP

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "status": "healthy",
  "timestamp": "2025-10-04T10:30:00Z",
  "services": {
    "database": "connected",
    "auth": "operational"
  }
}
```

**Error (503 Service Unavailable)**:
```json
{
  "success": false,
  "status": "unhealthy",
  "timestamp": "2025-10-04T10:30:00Z",
  "services": {
    "database": "disconnected",
    "auth": "operational"
  }
}
```

## Common Error Responses

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Internal server error",
  "error_id": "req_123e4567-e89b-12d3-a456-426614174000"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Endpoint not found"
}
```

### 405 Method Not Allowed
```json
{
  "success": false,
  "message": "Method not allowed"
}
```

## Request/Response Standards

### Headers
- All requests: `Content-Type: application/json`
- Authentication: `Authorization: Bearer <token>`
- Rate limiting: `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers

### Status Codes
- `200` OK - Successful operation
- `201` Created - Resource created successfully
- `400` Bad Request - Invalid input data
- `401` Unauthorized - Authentication required or failed
- `409` Conflict - Resource already exists
- `429` Too Many Requests - Rate limit exceeded
- `500` Internal Server Error - Server error
- `503` Service Unavailable - Service temporarily unavailable

### Response Format
All responses follow consistent structure:
- `success` (boolean): Operation success status
- `message` (string): Human-readable message
- `data` (object): Response data (success only)
- `errors` (array): Validation errors (error only)
- `error_id` (string): Error tracking ID (500 errors only)

### JWT Token Format
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "iss": "bloggo",
  "iat": 1696431000,
  "exp": 1696517400
}
```

**Contract Status**: COMPLETE - Ready for test generation