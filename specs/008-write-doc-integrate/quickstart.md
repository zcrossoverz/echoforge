# Quickstart: Documentation Integration and Site Extension Guide

## Overview
This quickstart guide validates the documentation system by walking through creating a new manga site type, generating documentation, and testing the API integration.

## Prerequisites
- Go 1.25+ installed
- PostgreSQL 16+ running
- Git configured
- Postman or similar API client (optional)

## Quick Validation Steps

### Step 1: Clone and Setup (5 minutes)
```bash
# Clone the Echoforge repository
git clone https://github.com/zcrossoverz/echoforge.git
cd echoforge

# Install dependencies
go mod download

# Run basic tests to ensure system works
go test ./...
```

**Expected Result**: All tests pass, confirming base system functionality.

### Step 2: Generate Documentation (10 minutes)
```bash
# Generate complete documentation suite
go run scripts/generate-docs.go --all

# Validate documentation completeness
go run scripts/validate-examples.go
```

**Expected Result**: 
- Documentation generated in `docs/` directory
- All code examples validate successfully
- No broken links or missing references

### Step 3: Create Manga Site Configuration (15 minutes)
```bash
# Copy manga site template
cp docs/site-configs/manga-site.yaml mysite-config.yaml

# Edit configuration for your site
# Update site_id, database connection, and features
nano mysite-config.yaml
```

**Example Configuration**:
```yaml
site:
  id: "manga-reader-001"
  name: "My Manga Site"
  description: "A manga reading platform"

database:
  dsn: "postgres://user:pass@localhost/manga_db?sslmode=disable"

features:
  comments: true
  ratings: true
  bookmarks: true
  notifications: false

customization:
  theme: "dark"
  language: "en"
  timezone: "UTC"
```

**Expected Result**: Valid configuration file ready for deployment.

### Step 4: Test API Documentation (10 minutes)
```bash
# Start the development server
go run cmd/server/main.go --config mysite-config.yaml

# In another terminal, test API endpoints
curl -X GET http://localhost:8080/api/v1/health
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"securepass123"}'
```

**Expected Result**: 
- Server starts without errors
- Health check returns 200 OK
- User registration creates account successfully

### Step 5: Import Postman Collection (5 minutes)
```bash
# Generate Postman collection
go run scripts/build-postman.go --output echoforge-collection.json

# Import into Postman:
# 1. Open Postman
# 2. Click Import
# 3. Select echoforge-collection.json
# 4. Import environments from docs/postman/environments/
```

**Expected Result**: 
- Collection imports successfully
- All endpoints are available and organized
- Environment variables are properly configured

### Step 6: Validate Site Extension Process (20 minutes)
```bash
# Follow the site extension guide
open docs/guides/site-extension/manga-site-setup.md

# Create custom manga-specific features
mkdir -p internal/domain/manga
mkdir -p adapters/http/manga

# Copy example implementations
cp docs/examples/manga/* internal/domain/manga/
cp docs/examples/handlers/manga/* adapters/http/manga/
```

**Expected Result**: 
- Site-specific code integrates cleanly
- No conflicts with core multi-tenant architecture
- Custom features work alongside base functionality

## Validation Checklist

### Documentation Quality ✅
- [ ] All guides are clear and actionable
- [ ] Code examples execute without errors
- [ ] Visual diagrams render correctly
- [ ] Links and references are valid

### API Integration ✅
- [ ] OpenAPI specification is complete
- [ ] Postman collection includes all endpoints
- [ ] Authentication flows work correctly
- [ ] Error responses are properly documented

### Site Extension ✅
- [ ] Manga site template creates working site
- [ ] Multi-tenant isolation is maintained
- [ ] Custom features integrate smoothly
- [ ] Configuration validation prevents errors

### Performance ✅
- [ ] Documentation generation completes in <30 seconds
- [ ] Site startup time is <5 seconds
- [ ] API responses are <200ms for simple operations
- [ ] System supports concurrent users

## Troubleshooting Common Issues

### Issue: Documentation Generation Fails
**Symptoms**: Build errors when running `generate-docs.go`
**Solution**: 
1. Check Go version: `go version` (must be 1.25+)
2. Verify all dependencies: `go mod tidy`
3. Check file permissions on docs/ directory

### Issue: Database Connection Errors
**Symptoms**: "connection refused" or "invalid DSN" errors
**Solution**:
1. Verify PostgreSQL is running: `pg_ctl status`
2. Check connection string format in config
3. Ensure database exists: `createdb manga_db`

### Issue: Postman Collection Import Fails
**Symptoms**: "Invalid collection format" error
**Solution**:
1. Regenerate collection: `go run scripts/build-postman.go --force`
2. Verify JSON format: `jsonlint echoforge-collection.json`
3. Check Postman version compatibility

### Issue: Site-Specific Features Don't Work
**Symptoms**: Custom manga features cause errors
**Solution**:
1. Check multi-tenant configuration: ensure `site_id` is properly set
2. Verify database migrations: `migrate -path migrations up`
3. Review logs for specific error messages

## Next Steps

After completing this quickstart:

1. **Explore Advanced Customization**: 
   - Review `docs/guides/customization/` for advanced patterns
   - Try creating a blog site using `blog-site.yaml` template
   - Experiment with custom authentication providers

2. **Production Deployment**:
   - Follow `docs/guides/deployment/docker-setup.md`
   - Configure environment-specific settings
   - Set up monitoring and logging

3. **API Integration**:
   - Use Postman collection for API exploration
   - Build client applications using OpenAPI specification
   - Implement custom integrations with third-party services

4. **Contribute to Documentation**:
   - Report issues or improvements needed
   - Submit pull requests for additional site types
   - Share your custom patterns with the community

## Success Criteria

This quickstart is successful if:
- ✅ Complete setup and validation takes <60 minutes
- ✅ All steps execute without manual intervention
- ✅ Resulting manga site is fully functional
- ✅ Documentation is accessible and accurate
- ✅ API integration works with provided tools

## Feedback and Support

If you encounter issues not covered in troubleshooting:
1. Check the full troubleshooting guide: `docs/troubleshooting/`
2. Review architecture documentation: `docs/architecture/`
3. Consult API documentation: `docs/api/openapi.yaml`
4. Submit issues via GitHub repository issue tracker