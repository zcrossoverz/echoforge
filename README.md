# Echoforge

> A reusable Golang backend core for multi-site content platforms (blog/manga/news) with hexagonal architecture.

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Hexagonal-green.svg)](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software))
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 📋 Overview

Echoforge is a modular monolith backend service designed to power multiple content-focused websites. Built with Go 1.25+ and following hexagonal (ports & adapters) architecture principles, it provides a lean, scalable foundation for multi-tenant applications.

### 🎯 Key Features

- **Modular Monolith**: Clean separation between domain, use cases, and adapters
- **Clone-and-Extend**: Each site runs as separate instance with dedicated database (v2.0+ refactor)
- **Hexagonal Architecture**: Domain-driven design with ports & adapters pattern
- **Lean Binary**: Optimized builds under 20MB
- **TDD Approach**: 80%+ test coverage requirement
- **Constitutional Compliance**: SemVer, reproducible builds, zero-downtime deployments

## 🛠 Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| **Language** | Go | 1.25+ |
| **HTTP Framework** | Gin | v1.11.0 |
| **Database ORM** | GORM | v1.31.0 |
| **Database Driver** | PostgreSQL | v1.6.0 |
| **Configuration** | Viper | v1.19.0 |
| **Logging** | Zap | v1.27.0 |
| **Dependency Injection** | Wire | v0.7.0 |
| **Validation** | Validator/v10 | v10.27.0 |
| **Testing** | Testify | v1.11.1 |
| **UUID Generation** | Google UUID | v1.6.0 |
| **Cryptography** | Go Crypto | v0.42.0 |

## 🏗 Project Structure

```
echoforge/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── domain/          # Business entities (pure, no dependencies)
│   ├── usecase/         # Business logic with DI
│   └── adapters/        # External concerns (HTTP, DB, etc.)
│       ├── http/        # Gin HTTP handlers
│       ├── persistence/ # GORM repositories
│       └── logger/      # Zap logging adapter
├── pkg/
│   ├── auth/           # JWT, bcrypt utilities
│   └── common/         # Shared utilities (logger config)
├── configs/            # Site configuration (clone-and-extend model)
├── tests/              # TDD tests (contract, integration, unit)
├── migrations/         # Database schema migrations
└── docs/              # Project documentation
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.25+** - [Download & Install](https://golang.org/dl/)
- **PostgreSQL 16+** - [Download & Install](https://www.postgresql.org/download/)
- **Git** - [Download & Install](https://git-scm.com/downloads)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/zcrossoverz/echoforge.git
   cd echoforge
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod verify
   ```

3. **Setup PostgreSQL database**
   ```bash
   # Create database (PostgreSQL command line)
   createdb echoforge
   
   # Or using psql
   psql -c "CREATE DATABASE echoforge;"
   ```

4. **Configure the application**
   ```bash
   # Copy example configuration
   cp configs/config.yaml.example configs/config.yaml
   
   # Edit configs/config.yaml with your database credentials
   ```

5. **Run database migrations**
   ```bash
   go run cmd/migrate/main.go up
   ```

6. **Start the development server**
   ```bash
   go run cmd/server/main.go
   ```

7. **Verify installation**
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Expected response: {"status":"ok"}
   ```

### Configuration

Create `configs/config.yaml`:

```yaml
# Server configuration
server:
  port: "8080"
  host: "localhost"
  read_timeout: "10s"
  write_timeout: "10s"
  shutdown_timeout: "5s"

# Logging configuration
logging:
  level: "info"           # debug, info, warn, error
  format: "json"          # json, console
  file: "logs/app.log"    # Optional: log file path

# Database configuration
database:
  dsn: "host=localhost user=postgres password=admin dbname=echoforge port=5432 sslmode=disable"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: "1h"

# Security & Authentication
auth:
  jwt_secret: "your-super-secret-jwt-key-change-this-in-production"
  jwt_expiration: "24h"
  password_min_length: 8
  rate_limit_requests: 100
  rate_limit_window: "1m"

# Site identification (for operational purposes)
site:
  id: "echoforge-blog"
  name: "Echoforge Blog"
  environment: "development"   # development, staging, production
```

### Environment Variables

For production deployment, you can override config values with environment variables:

```bash
# Server
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0

# Database
export DATABASE_DSN="host=db user=echoforge password=secure_password dbname=echoforge sslmode=require"

# Security
export AUTH_JWT_SECRET="your-production-jwt-secret"
export AUTH_RATE_LIMIT_REQUESTS=1000

# Logging
export LOGGING_LEVEL=warn
export LOGGING_FORMAT=json
```

## 🔌 API Endpoints

### Health & Status

| Method | Endpoint | Description | Response |
|--------|----------|-------------|----------|
| `GET` | `/health` | Health check | `{"status":"ok"}` |
| `GET` | `/api/v1/status` | Detailed status | System information |

### Authentication & User Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/register` | User registration | ❌ |
| `POST` | `/api/v1/login` | User authentication | ❌ |
| `POST` | `/api/v1/logout` | User logout | ✅ |
| `GET` | `/api/v1/profile` | Get user profile | ✅ |
| `PUT` | `/api/v1/profile` | Update user profile | ✅ |

### Example API Usage

#### User Registration
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

Response:
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### User Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

#### Authenticated Request
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage report
go test -cover ./...

# Run with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test categories
go test ./tests/unit/domain/...      # Domain unit tests
go test ./tests/unit/auth/...        # Authentication tests
go test ./tests/unit/usecase/...     # Use case tests
go test ./tests/performance/...      # Performance tests
go test ./tests/contract/...         # Contract tests
go test ./tests/integration/...      # Integration tests

# Run tests with race detection
go test -race ./...

# Benchmark tests
go test -bench=. ./tests/performance/
```

### Test Categories

| Category | Directory | Purpose | Coverage Target |
|----------|-----------|---------|----------------|
| **Unit Tests** | `tests/unit/` | Individual component testing | 80%+ |
| **Integration Tests** | `tests/integration/` | Multi-component interactions | 70%+ |
| **Contract Tests** | `tests/contract/` | Interface compliance | 100% |
| **Performance Tests** | `tests/performance/` | Response time & throughput | <500ms |

### Test-Driven Development (TDD)

This project follows strict TDD principles:

1. **🔴 Red**: Write failing tests first
2. **🟢 Green**: Write minimal code to pass tests  
3. **🔵 Refactor**: Improve code while maintaining tests

### Performance Testing

Performance tests ensure response times under 500ms:

```bash
# Run performance tests with timing
go test -v ./tests/performance/ -timeout=60s

# Expected results:
# - Health check: <10ms
# - User registration: <100ms  
# - User lookup: <50ms
# - Concurrent operations: <500ms average
```

## 📦 Building

### Development Build

```bash
go build -o echoforge cmd/server/main.go
```

### Production Build (Optimized)

```bash
go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go
```

### Cross-Platform Builds

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o echoforge-linux-amd64 cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o echoforge-darwin-amd64 cmd/server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o echoforge-windows-amd64.exe cmd/server/main.go
```

## 🐳 Docker

Build and run with Docker:

```dockerfile
# Dockerfile (example)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/echoforge .
COPY --from=builder /app/configs ./configs
CMD ["./echoforge"]
```

## 🏛 Architecture Principles

### Hexagonal Architecture (Ports & Adapters)

- **Domain Layer** (`internal/domain/`): Pure business entities with no external dependencies
- **Use Case Layer** (`internal/usecase/`): Business logic orchestration with dependency injection
- **Adapter Layer** (`internal/adapters/`): External concerns (HTTP, database, logging)

### Clone-and-Extend Architecture

Each site clones the core repository and runs with its own database:

```go
// Example: Simple user query (no site filtering needed)
users := []User{}
db.Find(&users)  // Each site has its own database
```

**Benefits**:
- Natural data isolation per site
- Simplified queries (no site_id joins)  
- Independent scaling and deployment
- Core updates via `go get` (SemVer)

## 📊 Performance Targets

- **Concurrent Users**: 1000+ per site
- **Binary Size**: <20MB unoptimized, <15MB optimized
- **Build Time**: <10 seconds for basic build
- **Memory Usage**: <100MB baseline per site
- **Response Time**: <100ms for typical API calls

## 🔒 Security

- **OWASP Top 10 Compliance**: Input validation, authentication, authorization
- **bcrypt Password Hashing**: Default cost for production security
- **JWT Tokens**: Stateless authentication with configurable expiration
- **Rate Limiting**: Built-in protection against abuse
- **Input Validation**: Comprehensive validation using go-playground/validator

## ⚡ Performance & Monitoring

### Performance Benchmarks

Based on performance test results:

| Operation | Average Response Time | Throughput |
|-----------|----------------------|------------|
| Health Check | <10µs | 100,000+ req/s |
| User Registration | ~50ms | 1,000+ req/s |
| User Lookup | ~10µs | 100,000+ req/s |
| Email Availability | <1ms | 50,000+ req/s |
| Concurrent Registration | ~300ms | 300+ concurrent |

### Monitoring Endpoints

- **Health**: `GET /health` - Basic health check
- **Metrics**: `GET /metrics` - Prometheus-compatible metrics
- **Debug**: `GET /debug/pprof/` - Go profiling (development only)

## 🚀 Deployment

### Production Deployment

1. **Build optimized binary**
   ```bash
   go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go
   ```

2. **Setup production database**
   ```bash
   # Run migrations
   ./echoforge migrate up
   
   # Verify database connection
   ./echoforge health-check
   ```

3. **Configure production settings**
   ```yaml
   # configs/config.yaml
   site:
     environment: "production"
   
   logging:
     level: "warn"
     format: "json"
   
   database:
     max_open_conns: 100
     max_idle_conns: 25
   ```

### Docker Deployment

```dockerfile
# Multi-stage Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w" -trimpath -o echoforge cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/echoforge .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./echoforge"]
```

```bash
# Build and run
docker build -t echoforge:latest .
docker run -p 8080:8080 -e DATABASE_DSN="..." echoforge:latest
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_DSN=host=db user=echoforge password=echoforge dbname=echoforge sslmode=disable
      - AUTH_JWT_SECRET=your-production-secret
      - LOGGING_LEVEL=info
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=echoforge
      - POSTGRES_USER=echoforge
      - POSTGRES_PASSWORD=echoforge
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
```

### Zero-Downtime Deployment

1. **Blue-Green Deployment**: Run two identical production environments
2. **Health Checks**: `/health` endpoint for load balancer monitoring
3. **Graceful Shutdown**: SIGTERM handling with configurable timeout
4. **Rolling Updates**: Kubernetes/Docker Swarm support

## 🛠 Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Test database connectivity
psql -h localhost -U postgres -d echoforge -c "SELECT version();"

# Check application database connection
curl http://localhost:8080/health
```

#### JWT Token Issues
```bash
# Regenerate JWT secret
openssl rand -base64 32

# Update config or environment variable
export AUTH_JWT_SECRET="new-secret-here"
```

#### Performance Issues
```bash
# Enable debug logging
export LOGGING_LEVEL=debug

# Run performance tests
go test -v ./tests/performance/

# Check memory usage
curl http://localhost:8080/debug/pprof/heap
```

#### Build Issues
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy

# Verify Go version
go version  # Should be 1.25+
```

### Debugging

#### Enable Debug Mode
```yaml
# configs/config.yaml
logging:
  level: "debug"
  format: "console"  # More readable in development

site:
  environment: "development"
```

#### Using pprof
```bash
# Start server with pprof
go run cmd/server/main.go

# In another terminal, analyze CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile

# Analyze memory usage
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Logs Analysis

```bash
# Follow application logs
tail -f logs/app.log

# Filter error logs (JSON format)
jq 'select(.level=="error")' logs/app.log

# Performance monitoring
grep "response_time" logs/app.log | jq '.response_time'
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests first (TDD approach)
4. Implement the feature
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Maintain 80%+ test coverage
- Use `gofmt` for code formatting
- Follow hexagonal architecture patterns
- Write clear, self-documenting code

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/zcrossoverz/echoforge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/zcrossoverz/echoforge/discussions)

## 🗺 Roadmap

- [ ] Authentication & Authorization MVP
- [ ] Database Migration System
- [ ] Admin Dashboard API
- [ ] Content Management System
- [ ] Multi-language Support
- [ ] Caching Layer (Redis)
- [ ] Message Queue Integration
- [ ] Monitoring & Observability

---

**Built with ❤️ using Go and hexagonal architecture principles.**