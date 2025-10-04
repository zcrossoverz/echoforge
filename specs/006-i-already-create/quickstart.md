# Quickstart Guide: Database Connection and Authentication APIs

**Feature**: Database Connection and Authentication APIs  
**Date**: October 4, 2025  
**Prerequisites**: PostgreSQL 16+, Go 1.25+

## Quick Setup

### 1. Database Setup

Create PostgreSQL database and user:
```sql
-- Connect as postgres superuser
CREATE DATABASE bloggo;
CREATE USER postgres WITH PASSWORD 'admin';
GRANT ALL PRIVILEGES ON DATABASE bloggo TO postgres;
```

### 2. Environment Configuration

Update `configs/config.yaml`:
```yaml
# Database Configuration
DB_DSN: "postgres://postgres:admin@localhost:5432/bloggo?sslmode=disable"

# JWT Configuration  
JWT_SECRET: "your-super-secure-jwt-secret-key-at-least-32-characters-long"

# Logging Configuration
LOG_LEVEL: "info"
ENABLE_HOT_RELOAD: false
```

Or use environment variables:
```bash
export DB_DSN="postgres://postgres:admin@localhost:5432/bloggo?sslmode=disable"
export JWT_SECRET="your-super-secure-jwt-secret-key-at-least-32-characters-long"
export LOG_LEVEL="info"
```

### 3. Run Database Migrations

```bash
# Install golang-migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migrations directory
mkdir -p migrations

# Run migrations (will be created by implementation)
migrate -path ./migrations -database $DB_DSN up
```

### 4. Start the Server

```bash
# Install dependencies
go mod tidy

# Run server
go run cmd/server/main.go
```

Expected output:
```json
{"level":"info","timestamp":"2025-10-04T10:30:00.000Z","caller":"server/main.go:45","message":"app starting","server_id":"uuid-here","log_level":"info","hot_reload":false}
{"level":"info","timestamp":"2025-10-04T10:30:00.001Z","caller":"server/main.go:52","message":"Database connected successfully"}
{"level":"info","timestamp":"2025-10-04T10:30:00.002Z","caller":"server/main.go:95","message":"HTTP server starting","port":"8080"}
```

## API Usage Examples

### Check System Health

```bash
curl -X GET http://localhost:8080/api/v1/health
```

Expected response:
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

### User Registration

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securePassword123"
  }'
```

Expected response:
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "john.doe@example.com",
      "created_at": "2025-10-04T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-05T10:30:00Z"
  }
}
```

### User Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securePassword123"
  }'
```

Expected response:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "john.doe@example.com",
      "created_at": "2025-10-04T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-05T10:30:00Z"
  }
}
```

### Get User Profile (Protected)

```bash
# Save token from login response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

Expected response:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "john.doe@example.com",
      "created_at": "2025-10-04T10:30:00Z",
      "updated_at": "2025-10-04T10:30:00Z"
    }
  }
}
```

### User Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

Expected response:
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

## Testing the Implementation

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Expected coverage: 80%+
```

### Integration Tests

```bash
# Run integration tests (requires database)
go test -tags=integration ./tests/integration/...

# Expected: All authentication flows working
```

### Load Testing (Optional)

```bash
# Install k6 load testing tool
# Test registration endpoint
k6 run --vus 100 --duration 30s - <<EOF
import http from 'k6/http';
import { check } from 'k6';

export default function() {
  const payload = JSON.stringify({
    email: \`user-\${__VU}-\${__ITER}@example.com\`,
    password: 'securePassword123'
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const response = http.post('http://localhost:8080/api/v1/auth/register', payload, params);
  check(response, {
    'status is 201': (r) => r.status === 201,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
}
EOF
```

## Troubleshooting

### Database Connection Issues

1. **Error**: `failed to connect to database`
   - Check PostgreSQL is running: `pg_ctl status`
   - Verify connection string in config
   - Test connection: `psql postgres://postgres:admin@localhost:5432/bloggo`

2. **Error**: `database "bloggo" does not exist`
   - Create database: `createdb bloggo`
   - Or use SQL: `CREATE DATABASE bloggo;`

### Authentication Issues

1. **Error**: `Invalid or expired token`
   - Check JWT_SECRET is set and consistent
   - Verify token hasn't expired (24-hour limit)
   - Ensure token format: `Bearer <token>`

2. **Error**: `Too many requests`
   - Rate limiting active (5 requests/minute)
   - Wait for rate limit window to reset
   - Check for proper IP forwarding in production

### Performance Issues

1. **Slow authentication**: 
   - Check bcrypt cost factor (should be 12)
   - Monitor database connection pool
   - Verify email index exists

2. **High memory usage**:
   - Check connection pool settings
   - Monitor blacklist token cleanup
   - Review JWT token size

## Production Considerations

### Security Checklist

- [ ] Change default JWT_SECRET to strong secret
- [ ] Enable HTTPS/TLS in production
- [ ] Set up proper firewall rules
- [ ] Configure rate limiting behind reverse proxy
- [ ] Enable security headers (CORS, CSP, etc.)
- [ ] Set up log monitoring and alerting
- [ ] Regular security updates and patches

### Performance Optimization

- [ ] Configure database connection pooling
- [ ] Set up database read replicas (if needed)
- [ ] Implement Redis for blacklist token caching
- [ ] Enable gzip compression for API responses
- [ ] Set up CDN for static assets (future)
- [ ] Monitor and optimize database queries

### Monitoring Setup

- [ ] Application health checks
- [ ] Database connectivity monitoring
- [ ] API response time monitoring
- [ ] Error rate tracking
- [ ] Security event alerting
- [ ] Performance metrics collection

## Development Workflow

### Adding New Features

1. Update data model if needed
2. Create new API contracts
3. Write failing tests (TDD)
4. Implement business logic
5. Add HTTP handlers
6. Update documentation
7. Verify 80%+ test coverage

### Database Changes

1. Create new migration files
2. Test migration up/down
3. Update data model documentation  
4. Verify backward compatibility
5. Test with existing data

**Quickstart Status**: COMPLETE - Ready for implementation