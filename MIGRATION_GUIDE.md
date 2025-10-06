# Migration Guide: Multi-Tenant to Clone-and-Extend Model

This guide helps you migrate from the previous multi-tenant site_id model to the new clone-and-extend architecture.

## Overview

**Before**: Single deployment serving multiple sites with `site_id` isolation
**After**: Each site clones core repository and runs separate database instance

## Migration Steps

### 1. Data Migration

For each site, extract their data to a separate database:

```sql
-- For each site, create a new database
CREATE DATABASE blog_site;
CREATE DATABASE manga_site;
CREATE DATABASE news_site;

-- Export site-specific data
\c original_database
COPY (SELECT * FROM users WHERE site_id = 'blog-site-uuid') TO '/tmp/blog_users.csv' CSV HEADER;

-- Import to new site database
\c blog_site
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
COPY users FROM '/tmp/blog_users.csv' CSV HEADER;
```

### 2. Code Migration

#### Update User Entity References
```go
// Before: Multi-tenant
user, err := domain.NewUser(siteID, email, passwordHash)
user, err := repo.FindByEmail(ctx, siteID, email)

// After: Clone-and-extend  
user, err := domain.NewUser(email, passwordHash)
user, err := repo.FindByEmail(ctx, email)
```

#### Update Use Case Calls
```go
// Before: Site-scoped operations
result, err := registerUC.Execute(ctx, RegisterInput{
    SiteID:   siteID,
    Email:    email,
    Password: password,
})

// After: Simplified operations
result, err := registerUC.Execute(ctx, RegisterInput{
    Email:    email,
    Password: password,
})
```

#### Update JWT Tokens
```go
// Before: Site + User claims
token, err := auth.GenerateToken(userID, siteID, secret)

// After: User-only claims  
token, err := auth.GenerateToken(userID, secret)
```

### 3. Configuration Migration

#### Database Configuration
```yaml
# Before: Single database
database:
  dsn: "host=localhost user=app password=secret dbname=multi_tenant port=5432"

# After: Site-specific database
database:
  dsn: "host=localhost user=app password=secret dbname=blog_site port=5432"
```

#### Site Identification (Optional)
```yaml
# Keep for operational purposes only
app:
  site_id: "blog-site"  # For logging and monitoring
```

### 4. Deployment Migration

#### Before: Single Multi-Tenant Service
```
production-server
├── app (serves all sites)
├── database (multi-tenant with site_id)
└── config (site_id switching)
```

#### After: Separate Site Instances
```
blog-site-server
├── cloned echoforge core
├── blog_site database
└── blog-specific config

manga-site-server  
├── cloned echoforge core
├── manga_site database
└── manga-specific config
```

### 5. Testing Migration

Update your tests to remove site_id parameters:

```go
// Before
func TestUserRegistration(t *testing.T) {
    siteID := uuid.New()
    user, err := usecase.Register(ctx, siteID, email, password)
    // ...
}

// After
func TestUserRegistration(t *testing.T) {
    user, err := usecase.Register(ctx, email, password)  
    // ...
}
```

## Benefits After Migration

1. **Simplified Code**: No more site_id parameters throughout codebase
2. **Better Performance**: No site_id JOIN conditions in queries
3. **Natural Isolation**: Database-level separation instead of application-level filtering
4. **Independent Scaling**: Each site can scale independently
5. **Easier Maintenance**: Core updates via `go get`, site-specific customizations separate

## Rollback Plan

If you need to rollback:

1. Keep the old multi-tenant deployment running during migration
2. Test the new clone-and-extend deployments thoroughly
3. Use feature flags to route traffic between old and new systems
4. Maintain data sync between old and new systems during transition period

## Support

For migration issues:
- Check the test files in `/tests` for examples of the new API
- Review the quickstart guide in `/specs/005-update-specifications-for/quickstart.md`
- Ensure all tests pass: `go test ./...`