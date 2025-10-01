# Quickstart: Register and Login Authentication Usecases

**Feature**: Register and Login Authentication  
**Estimated Time**: 15 minutes  
**Prerequisites**: Completed Task 1.2 (User Domain Entity + Repository)

## 🚀 Quick Setup

### 1. Add JWT Dependency
```bash
cd echoforge
go get github.com/golang-jwt/jwt/v5@latest
go mod tidy
```

### 2. Configure JWT Secret
```bash
# Add to environment or .env file
export JWT_SECRET="your-super-secure-jwt-secret-key-at-least-32-chars"
```

### 3. Verify Existing Foundation
```bash
# Ensure user domain and repository are available
go test ./tests/ -run "User" -v
```

## 📝 Implementation Validation

### Register Usecase Test
Create and run this validation test:

```go
// File: internal/usecase/user/register_test.go
func TestRegisterUsecase_QuickValidation(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    registerUC := NewRegisterUsecase(mockRepo, "test-secret")
    
    ctx := context.Background()
    input := RegisterInput{
        SiteID:   uuid.New(),
        Email:    "test@example.com", 
        Password: "securepass123",
    }
    
    // Execute
    user, err := registerUC.Execute(ctx, input)
    
    // Validate
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, input.Email, user.Email)
    assert.NotEmpty(t, user.PasswordHash)
    assert.True(t, len(user.PasswordHash) >= 60) // bcrypt hash length
}
```

### Login Usecase Test
```go
// File: internal/usecase/user/login_test.go  
func TestLoginUsecase_QuickValidation(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    loginUC := NewLoginUsecase(mockRepo, "test-secret")
    
    // Pre-create user with known password
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), 12)
    existingUser := &domain.User{
        ID:           uuid.New(),
        SiteID:       uuid.New(),
        Email:        "existing@example.com",
        PasswordHash: string(hashedPassword),
    }
    mockRepo.users = []*domain.User{existingUser}
    
    ctx := context.Background()
    input := LoginInput{
        SiteID:   existingUser.SiteID,
        Email:    existingUser.Email,
        Password: "testpass123",
    }
    
    // Execute
    result, err := loginUC.Execute(ctx, input)
    
    // Validate
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.Token)
    assert.True(t, result.ExpiresAt.After(time.Now()))
}
```

### Run Validation Tests
```bash
go test ./internal/usecase/user/ -v -run "QuickValidation"
```

## 🔧 Integration Test

### End-to-End Authentication Flow
```go
func TestAuthenticationFlow_Integration(t *testing.T) {
    // Setup with real repository (requires database)
    db := setupTestDB(t)
    userRepo := persistence.NewUserRepository(db)
    
    registerUC := NewRegisterUsecase(userRepo, "integration-test-secret") 
    loginUC := NewLoginUsecase(userRepo, "integration-test-secret")
    
    ctx := context.Background()
    siteID := uuid.New()
    
    // Step 1: Register new user
    registerInput := RegisterInput{
        SiteID:   siteID,
        Email:    "integration@example.com",
        Password: "integrationpass123",
    }
    
    user, err := registerUC.Execute(ctx, registerInput)
    require.NoError(t, err)
    require.NotNil(t, user)
    
    // Step 2: Login with registered credentials
    loginInput := LoginInput{
        SiteID:   siteID,
        Email:    "integration@example.com", 
        Password: "integrationpass123",
    }
    
    authResult, err := loginUC.Execute(ctx, loginInput)
    require.NoError(t, err)
    require.NotNil(t, authResult)
    
    // Step 3: Validate JWT token
    token, err := jwt.Parse(authResult.Token, func(token *jwt.Token) (interface{}, error) {
        return []byte("integration-test-secret"), nil
    })
    require.NoError(t, err)
    require.True(t, token.Valid)
    
    // Step 4: Verify claims
    claims := token.Claims.(jwt.MapClaims)
    assert.Equal(t, user.ID.String(), claims["sub"])
    assert.Equal(t, siteID.String(), claims["site_id"])
}
```

## 🔒 Security Validation

### Multi-Tenant Isolation Test
```go
func TestMultiTenantIsolation_Security(t *testing.T) {
    userRepo := setupMockRepository()
    loginUC := NewLoginUsecase(userRepo, "security-test-secret")
    
    siteA := uuid.New()
    siteB := uuid.New()
    
    // Create user in site A
    userA := createTestUser(siteA, "user@example.com", "password123")
    userRepo.AddUser(userA)
    
    // Try to login to site B with site A credentials (should fail)
    loginInput := LoginInput{
        SiteID:   siteB, // Wrong site!
        Email:    "user@example.com",
        Password: "password123", 
    }
    
    _, err := loginUC.Execute(context.Background(), loginInput)
    assert.Error(t, err)
    assert.Equal(t, ErrAuthenticationFailed, err)
}
```

### Password Security Test
```go
func TestPasswordSecurity_Validation(t *testing.T) {
    registerUC := NewRegisterUsecase(&MockUserRepository{}, "security-test")
    
    weakPasswords := []string{
        "123",      // Too short
        "password", // Too weak
        "",         // Empty
    }
    
    for _, weakPass := range weakPasswords {
        input := RegisterInput{
            SiteID:   uuid.New(),
            Email:    "test@example.com",
            Password: weakPass,
        }
        
        _, err := registerUC.Execute(context.Background(), input)
        assert.Error(t, err, "Should reject weak password: %s", weakPass)
    }
}
```

## ✅ Success Criteria

Run this checklist to verify implementation:

```bash
# 1. Unit tests pass with >80% coverage
go test ./internal/usecase/user/ -cover -v
# Expected: PASS with coverage >80%

# 2. Integration tests pass (requires database)
go test ./internal/usecase/user/ -tags=integration -v
# Expected: PASS for end-to-end flow

# 3. Security tests pass
go test ./internal/usecase/user/ -run "Security" -v  
# Expected: PASS for multi-tenant isolation

# 4. Performance validation
go test ./internal/usecase/user/ -bench=. -benchtime=10s
# Expected: <100ms per operation under load

# 5. JWT validation
echo "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." | base64 -d
# Expected: Valid JWT structure with correct claims
```

## 🎯 Expected Outcomes

After successful implementation:

### ✅ Register Usecase
- [x] Validates input according to struct tags
- [x] Prevents duplicate email registration within site
- [x] Allows same email across different sites
- [x] Hashes passwords with bcrypt cost ≥12
- [x] Creates user entity with generated UUID
- [x] Persists to database via repository
- [x] Returns created user (without password hash)

### ✅ Login Usecase  
- [x] Validates input credentials
- [x] Finds user by email within site scope
- [x] Verifies password against bcrypt hash
- [x] Generates JWT token with user + site claims
- [x] Sets 24-hour token expiration
- [x] Returns authentication result with token
- [x] Prevents user enumeration via generic errors

### ✅ Security Features
- [x] Multi-tenant isolation enforced
- [x] OWASP-compliant password handling
- [x] Secure JWT token generation
- [x] Input validation prevents injection
- [x] Generic error messages protect privacy
- [x] Context support for timeout/cancellation

### ✅ Testing Coverage
- [x] >80% unit test coverage achieved
- [x] Integration tests with real database
- [x] Security tests for multi-tenancy
- [x] Performance tests for concurrency
- [x] Error handling and edge cases covered

## 🚧 Common Issues & Solutions

### Issue: JWT Secret Not Configured
**Error**: `panic: JWT secret not provided`  
**Solution**: Ensure `JWT_SECRET` environment variable is set

### Issue: Weak Password Rejected
**Error**: `password validation failed: too weak`  
**Solution**: Use passwords ≥8 characters with reasonable complexity

### Issue: Email Already Exists
**Error**: `email address already registered`  
**Solution**: Check for existing users in the same site, or use different site

### Issue: Multi-Tenant Isolation Failure
**Error**: Authentication succeeds across sites  
**Solution**: Verify all repository calls include `siteID` parameter

---

**Quickstart Status**: READY  
**Estimated Implementation Time**: 4-6 hours with TDD approach  
**Dependencies**: User domain + repository from Task 1.2  
**Next Steps**: Run `/tasks` command to generate detailed implementation tasks