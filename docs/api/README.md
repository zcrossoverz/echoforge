# Echoforge API Documentation

**Version**: 1.0.0  
**Base URL**: `http://localhost:8080`  
**API Version**: `v1`

## Overview

Echoforge provides a RESTful API for user management and authentication. All API endpoints are versioned and follow REST conventions.

### API Versioning

All API endpoints are prefixed with `/api/v1/` to ensure backward compatibility.

### Content Types

- **Request**: `application/json`
- **Response**: `application/json`

### Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

### Error Handling

All errors follow a consistent format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "validation error details"
  }
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | OK - Request successful |
| `201` | Created - Resource created successfully |
| `400` | Bad Request - Invalid request data |
| `401` | Unauthorized - Authentication required |
| `403` | Forbidden - Insufficient permissions |
| `404` | Not Found - Resource not found |
| `409` | Conflict - Resource already exists |
| `422` | Unprocessable Entity - Validation failed |
| `429` | Too Many Requests - Rate limit exceeded |
| `500` | Internal Server Error - Server error |

## Health & Status Endpoints

### Health Check

Check if the API is running and healthy.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "ok"
}
```

**Example**:
```bash
curl -X GET http://localhost:8080/health
```

### Detailed Status

Get detailed system status information.

**Endpoint**: `GET /api/v1/status`

**Headers**: None required

**Response**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "database": {
    "status": "connected",
    "connections": {
      "active": 5,
      "idle": 10,
      "max": 100
    }
  },
  "memory": {
    "allocated": "45MB",
    "system": "67MB"
  }
}
```

## Authentication Endpoints

### User Registration

Register a new user account.

**Endpoint**: `POST /api/v1/register`

**Headers**:
- `Content-Type: application/json`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Validation Rules**:
- `email`: Valid email format, max 255 characters, unique
- `password`: Minimum 8 characters, at least one uppercase, one lowercase, one number

**Success Response** (`201 Created`):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-02T12:00:00Z"
}
```

**Error Responses**:

`400 Bad Request` - Invalid input:
```json
{
  "error": "Validation failed",
  "details": {
    "email": "Invalid email format",
    "password": "Password must be at least 8 characters"
  }
}
```

`409 Conflict` - User already exists:
```json
{
  "error": "User already exists with this email"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecurePassword123!"
  }'
```

### User Login

Authenticate an existing user.

**Endpoint**: `POST /api/v1/login`

**Headers**:
- `Content-Type: application/json`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Success Response** (`200 OK`):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-02T12:00:00Z"
}
```

**Error Responses**:

`400 Bad Request` - Invalid input:
```json
{
  "error": "Email and password are required"
}
```

`401 Unauthorized` - Invalid credentials:
```json
{
  "error": "Invalid email or password"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

### User Logout

Logout the current user (invalidate JWT token).

**Endpoint**: `POST /api/v1/logout`

**Headers**:
- `Authorization: Bearer <jwt_token>`

**Request Body**: None

**Success Response** (`200 OK`):
```json
{
  "message": "Successfully logged out"
}
```

**Error Responses**:

`401 Unauthorized` - Invalid or missing token:
```json
{
  "error": "Invalid or expired token"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## User Management Endpoints

### Get User Profile

Get the current user's profile information.

**Endpoint**: `GET /api/v1/profile`

**Headers**:
- `Authorization: Bearer <jwt_token>`

**Success Response** (`200 OK`):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Responses**:

`401 Unauthorized` - Invalid or missing token:
```json
{
  "error": "Authentication required"
}
```

`404 Not Found` - User not found:
```json
{
  "error": "User not found"
}
```

**Example**:
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Update User Profile

Update the current user's profile information.

**Endpoint**: `PUT /api/v1/profile`

**Headers**:
- `Authorization: Bearer <jwt_token>`
- `Content-Type: application/json`

**Request Body**:
```json
{
  "email": "newemail@example.com"
}
```

**Success Response** (`200 OK`):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "newemail@example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:05:00Z"
  }
}
```

**Error Responses**:

`400 Bad Request` - Invalid input:
```json
{
  "error": "Invalid email format"
}
```

`409 Conflict` - Email already taken:
```json
{
  "error": "Email already in use"
}
```

**Example**:
```bash
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com"
  }'
```

### Change Password

Change the current user's password.

**Endpoint**: `PUT /api/v1/profile/password`

**Headers**:
- `Authorization: Bearer <jwt_token>`
- `Content-Type: application/json`

**Request Body**:
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewSecurePassword456!"
}
```

**Success Response** (`200 OK`):
```json
{
  "message": "Password updated successfully"
}
```

**Error Responses**:

`400 Bad Request` - Invalid input:
```json
{
  "error": "New password must be at least 8 characters"
}
```

`401 Unauthorized` - Wrong current password:
```json
{
  "error": "Current password is incorrect"
}
```

**Example**:
```bash
curl -X PUT http://localhost:8080/api/v1/profile/password \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "OldPassword123!",
    "new_password": "NewSecurePassword456!"
  }'
```

## Utility Endpoints

### Check Email Availability

Check if an email address is available for registration.

**Endpoint**: `GET /api/v1/check-email/{email}`

**Parameters**:
- `email` (path): Email address to check

**Success Response** (`200 OK`):
```json
{
  "email": "test@example.com",
  "available": true
}
```

**Error Responses**:

`400 Bad Request` - Invalid email format:
```json
{
  "error": "Invalid email format"
}
```

**Example**:
```bash
curl -X GET http://localhost:8080/api/v1/check-email/test@example.com
```

## Rate Limiting

API endpoints are protected by rate limiting:

- **Default**: 100 requests per minute per IP
- **Authentication endpoints**: 10 requests per minute per IP
- **Registration**: 5 requests per minute per IP

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1640995200
```

When rate limited, the API returns `429 Too Many Requests`:

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

## JWT Token Details

### Token Structure

JWT tokens contain the following claims:

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",  // User ID
  "iat": 1640995200,  // Issued at
  "exp": 1641081600   // Expires at
}
```

### Token Expiration

- **Default**: 24 hours
- **Configurable**: Via `AUTH_JWT_EXPIRATION` environment variable
- **Refresh**: Re-login required after expiration

### Token Validation

Tokens are validated on each request:

1. **Signature verification**: Using HMAC-SHA256
2. **Expiration check**: Token must not be expired
3. **Blacklist check**: Token must not be blacklisted (after logout)

## Performance Characteristics

Based on performance testing:

| Endpoint | Average Response Time | Throughput |
|----------|----------------------|------------|
| `GET /health` | <10µs | 100,000+ req/s |
| `POST /api/v1/register` | ~50ms | 1,000+ req/s |
| `POST /api/v1/login` | ~45ms | 1,200+ req/s |
| `GET /api/v1/profile` | ~10µs | 100,000+ req/s |
| `GET /api/v1/check-email/*` | <1ms | 50,000+ req/s |

### Concurrency

- **Concurrent users**: 1,000+ supported
- **Concurrent registrations**: 300+ simultaneous
- **Database connections**: Pooled (max 100)

## Security Considerations

### Password Security

- **Hashing**: bcrypt with default cost (10)
- **Minimum length**: 8 characters
- **Complexity**: Recommended (uppercase, lowercase, number, special char)

### JWT Security

- **Algorithm**: HMAC-SHA256
- **Secret**: Configurable via environment variable
- **Expiration**: 24 hours default
- **Blacklisting**: Supported for logout functionality

### Input Validation

- **Email**: RFC 5322 compliant validation
- **Sanitization**: XSS protection for all inputs
- **Length limits**: Enforced on all fields

### HTTPS

- **Production**: HTTPS required
- **Development**: HTTP acceptable
- **Headers**: Security headers included (HSTS, CSP, etc.)

## Examples

### Complete Registration Flow

```bash
# 1. Check email availability
curl -X GET http://localhost:8080/api/v1/check-email/newuser@example.com

# 2. Register user
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecurePassword123!"
  }'

# 3. Use the returned token for authenticated requests
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 4. Get user profile
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $TOKEN"
```

### Login and Profile Update Flow

```bash
# 1. Login
RESPONSE=$(curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }')

# 2. Extract token
TOKEN=$(echo $RESPONSE | jq -r '.token')

# 3. Update profile
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com"
  }'

# 4. Logout
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer $TOKEN"
```

### Error Handling Example

```bash
# Try to register with invalid data
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "password": "123"
  }' \
  -w "HTTP Status: %{http_code}\n"

# Expected response:
# HTTP Status: 400
# {
#   "error": "Validation failed",
#   "details": {
#     "email": "Invalid email format",
#     "password": "Password must be at least 8 characters"
#   }
# }
```

## Support

For API support and questions:

- **Documentation**: [GitHub Repository](https://github.com/zcrossoverz/echoforge)
- **Issues**: [GitHub Issues](https://github.com/zcrossoverz/echoforge/issues)
- **API Version**: Check `/api/v1/status` for current version information