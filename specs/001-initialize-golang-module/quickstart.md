# Quickstart: Initialize Golang Module for Echoforge Project

## Overview
This quickstart guide walks through initializing the Go module for echoforge, verifying the setup, and confirming all constitutional requirements are met.

## Prerequisites
- Go 1.25+ installed and in PATH
- Git installed and configured
- Network access for dependency downloads
- Command line access (bash/PowerShell)

## Step 1: Verify Environment

### Check Go Installation
```bash
# Verify Go version
go version
# Expected: go version go1.25+ ...

# Verify Go environment
go env GOPATH
go env GOROOT
```

### Check Git Installation
```bash
# Verify Git is available
git --version
# Expected: git version 2.x.x or higher
```

## Step 2: Initialize Module

### Create Project Directory
```bash
# Create and enter project directory
mkdir echoforge
cd echoforge
```

### Initialize Go Module
```bash
# Initialize with GitHub module path
go mod init github.com/zcrossoverz/echoforge

# Verify go.mod creation
cat go.mod
# Expected output:
# module github.com/zcrossoverz/echoforge
# 
# go 1.25
```

## Step 3: Add Core Dependencies

### Add HTTP Framework (Gin)
```bash
go get github.com/gin-gonic/gin@v1.10.0
```

### Add ORM and Database Driver
```bash
go get gorm.io/gorm@v1.25.12
go get gorm.io/driver/postgres@v1.5.9
```

### Add Configuration Management
```bash
go get github.com/spf13/viper@v1.19.0
```

### Add Structured Logging
```bash
go get go.uber.org/zap@v1.27.0
```

### Add Utilities
```bash
go get github.com/google/uuid@v1.6.0
go get golang.org/x/crypto@v0.42.0
```

### Add Dependency Injection
```bash
go get github.com/google/wire@v0.8.0
```

### Add Validation
```bash
go get github.com/go-playground/validator/v10@v10.27.0
```

### Add Testing Framework
```bash
go get github.com/stretchr/testify@v1.13.1
```

## Step 4: Create Project Structure

### Create Directory Structure
```bash
# Create core directories following hexagonal architecture
mkdir -p internal/domain
mkdir -p internal/usecase  
mkdir -p internal/adapters/http
mkdir -p internal/adapters/persistence
mkdir -p internal/adapters/logger

# Create application entry point
mkdir -p cmd/server

# Create public packages
mkdir -p pkg/auth
mkdir -p pkg/common

# Create configuration directory
mkdir -p configs

# Create testing directories
mkdir -p tests/unit
mkdir -p tests/integration
mkdir -p tests/contract

# Create migrations directory (future use)
mkdir -p migrations

# Create documentation directory
mkdir -p docs
```

### Create .gitignore File
```bash
cat > .gitignore << 'EOF'
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage/

# Go workspace file
go.work
go.work.sum

# Dependency directories
vendor/

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Environment files
.env
.env.local
.env.*.local

# Build artifacts
/bin/
/build/
/dist/

# Logs
*.log

# Temporary files
/tmp/
EOF
```

## Step 5: Clean Up Dependencies

### Run go mod tidy
```bash
go mod tidy

# Verify go.sum was created
ls -la go.sum
```

### Verify Module Structure
```bash
# Check final go.mod content
cat go.mod
# Should contain all dependencies with pinned versions

# Verify dependency count
go list -m all | wc -l
# Should show reasonable number of dependencies
```

## Step 6: Validation Tests

### Test Module Build
```bash
# Create a simple main.go for testing
cat > cmd/server/main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Echoforge module initialized successfully!")
}
EOF

# Build the module
go build -o bin/echoforge ./cmd/server

# Verify binary was created
ls -la bin/echoforge
```

### Check Binary Size
```bash
# Check binary size (should be under 20MB)
du -h bin/echoforge
# Expected: less than 20M

# Get exact size in bytes
stat -f%z bin/echoforge 2>/dev/null || stat -c%s bin/echoforge
# Should be less than 20971520 bytes (20MB)
```

### Test Dependency Import
```bash
# Create a test file to verify imports work
cat > test_imports.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/spf13/viper"
    "go.uber.org/zap"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
)

func main() {
    // Test that all imports are available
    _ = gin.New()
    _ = &gorm.DB{}
    _ = viper.New()
    _ = zap.NewProduction
    _ = uuid.New()
    _ = bcrypt.GenerateFromPassword
    _ = validator.New()
    _ = assert.True
    
    println("All dependencies imported successfully!")
}
EOF

# Build and run the test
go run test_imports.go

# Clean up test file
rm test_imports.go
```

## Step 7: Verify Constitutional Compliance

### Check Go Version Compliance
```bash
# Verify go.mod specifies correct version
grep "go 1.25" go.mod
# Should return: go 1.25
```

### Verify SemVer Compliance
```bash
# Check that all dependencies use semantic versioning
go list -m all | grep -E "v[0-9]+\.[0-9]+\.[0-9]+"
# All dependencies should have proper version tags
```

### Verify Architecture Preparation
```bash
# Check directory structure follows hexagonal architecture
tree internal/
# Should show domain, usecase, and adapters directories

# Verify adapters are properly separated
ls -la internal/adapters/
# Should show http, persistence, logger subdirectories
```

### Verify Lean Binary Requirement
```bash
# Final binary size check with optimized build
go build -ldflags="-s -w" -o bin/echoforge-optimized ./cmd/server
du -h bin/echoforge-optimized
# Should be even smaller with optimization flags
```

## Step 8: Documentation

### Create Basic README
```bash
cat > README.md << 'EOF'
# Echoforge

A reusable Golang backend core for multi-site content platforms (blog/manga/news).

## Architecture
- Modular monolith with hexagonal (ports & adapters) pattern
- Clean separation of domain, use cases, and adapters
- Multi-tenant support via site_id configuration

## Tech Stack
- Go 1.25+
- Gin v1.10.0 (HTTP API)
- GORM v1.25.12 (ORM)
- PostgreSQL 16+ (Database)
- Zap v1.27.0 (Logging)
- Viper v1.19.0 (Configuration)
- Wire v0.8.0 (Dependency Injection)

## Getting Started
```bash
go mod download
go build ./cmd/server
```

## Testing
```bash
go test ./...
```
EOF
```

## Expected Results

After completing this quickstart, you should have:

1. ✅ Valid go.mod with `github.com/zcrossoverz/echoforge` module path
2. ✅ Go 1.25+ specified as minimum version
3. ✅ All required dependencies with pinned versions:
   - Gin v1.10.0
   - GORM v1.25.12
   - Postgres driver v1.5.9
   - Viper v1.19.0
   - Zap v1.27.0
   - UUID v1.6.0
   - Crypto v0.42.0
   - Wire v0.8.0
   - Validator v10.27.0
   - Testify v1.13.1
4. ✅ Complete directory structure following hexagonal architecture
5. ✅ Proper .gitignore excluding go.sum and binaries
6. ✅ Binary size under 20MB limit
7. ✅ Module builds successfully without errors
8. ✅ All dependencies import correctly

## Troubleshooting

### Common Issues

**"go: cannot find main module"**
- Ensure you're in the project directory
- Verify go.mod exists in current directory

**"dependency version conflicts"**
- Run `go mod tidy` to resolve conflicts
- Check for pre-release versions that may conflict

**"binary size too large"**
- Use build flags: `go build -ldflags="-s -w"`
- Consider removing unused dependencies

**"import not found"**
- Run `go mod download` to fetch dependencies
- Verify module path matches go.mod

### Getting Help
- Check Go documentation: https://golang.org/doc/
- Review dependency documentation
- Consult echoforge project documentation

## Next Steps
1. Run `/tasks` command to generate implementation tasks
2. Set up development environment with IDE
3. Begin implementing core domain entities
4. Set up CI/CD pipeline for automated builds
5. Configure Docker for containerized deployment