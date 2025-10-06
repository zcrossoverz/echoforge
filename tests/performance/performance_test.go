package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/zcrossoverz/echoforge/internal/config"
	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
	"github.com/zcrossoverz/echoforge/pkg/auth"
)

// MockUserRepository for performance testing
type MockUserRepository struct {
	users map[string]*domain.User
	mutex sync.RWMutex
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.users[user.Email]; exists {
		return domain.ErrUserAlreadyExists
	}
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, user := range m.users {
		if user.ID.String() == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for email, user := range m.users {
		if user.ID.String() == id {
			delete(m.users, email)
			return nil
		}
	}
	return nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.users[email]
	return exists, nil
}

func (m *MockUserRepository) Count() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.users)
}

// Performance test configuration
const (
	ResponseTimeTarget     = 500 * time.Millisecond // Target response time
	ConcurrentUsers        = 100                    // Number of concurrent users
	RequestsPerUser        = 10                     // Requests per user
	PerformanceTestTimeout = 30 * time.Second       // Total test timeout
)

func setupTestRouter() (*gin.Engine, *MockUserRepository) {
	gin.SetMode(gin.TestMode)

	mockRepo := NewMockUserRepository()
	userUseCase := usecase.NewUserUseCase(mockRepo)

	cfg := &config.Config{
		JWTSecret: "test-performance-secret-key",
	}
	jwtService := auth.NewJWTService(cfg)

	router := gin.New()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// User registration endpoint
	router.POST("/api/v1/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Create user
		user, err := userUseCase.CreateUser(c.Request.Context(), req.Email, string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		// Generate JWT token
		token, _, err := jwtService.GenerateToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
			},
			"token": token,
		})
	})

	// User profile endpoint
	router.GET("/api/v1/profile/:email", func(c *gin.Context) {
		email := c.Param("email")

		user, err := userUseCase.GetUserByEmail(c.Request.Context(), email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"created_at": user.CreatedAt,
			},
		})
	})

	// Email availability check endpoint
	router.GET("/api/v1/check-email/:email", func(c *gin.Context) {
		email := c.Param("email")

		available, err := userUseCase.IsEmailAvailable(c.Request.Context(), email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"email":     email,
			"available": available,
		})
	})

	return router, mockRepo
}

func TestPerformance_HealthCheck(t *testing.T) {
	router, _ := setupTestRouter()

	// Benchmark health check endpoint
	start := time.Now()

	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	duration := time.Since(start)
	avgResponseTime := duration / 1000

	t.Logf("Health check average response time: %v", avgResponseTime)
	assert.Less(t, avgResponseTime, 10*time.Millisecond, "Health check should respond in less than 10ms")
}

func TestPerformance_UserRegistration(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Test single user registration performance
	requestBody := map[string]string{
		"email":    "perf@example.com",
		"password": "SecurePassword123!",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	start := time.Now()

	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	duration := time.Since(start)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Less(t, duration, ResponseTimeTarget, "User registration should complete in less than 500ms")
	assert.Equal(t, 1, mockRepo.Count())

	t.Logf("User registration response time: %v", duration)
}

func TestPerformance_ConcurrentUserRegistration(t *testing.T) {
	router, mockRepo := setupTestRouter()

	var wg sync.WaitGroup
	results := make(chan time.Duration, ConcurrentUsers)
	errors := make(chan error, ConcurrentUsers)

	// Concurrent user registration test
	start := time.Now()

	for i := 0; i < ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			requestStart := time.Now()

			requestBody := map[string]string{
				"email":    fmt.Sprintf("user%d@example.com", userID),
				"password": "SecurePassword123!",
			}

			jsonBody, err := json.Marshal(requestBody)
			if err != nil {
				errors <- err
				return
			}

			req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			requestDuration := time.Since(requestStart)
			results <- requestDuration

			if w.Code != http.StatusCreated {
				errors <- fmt.Errorf("user %d registration failed with status %d", userID, w.Code)
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	totalDuration := time.Since(start)

	// Collect results
	var responseTimes []time.Duration
	for responseTime := range results {
		responseTimes = append(responseTimes, responseTime)
	}

	// Check for errors
	var errorCount int
	for err := range errors {
		if err != nil {
			t.Logf("Error: %v", err)
			errorCount++
		}
	}

	// Calculate statistics
	var totalResponseTime time.Duration
	var maxResponseTime time.Duration
	var minResponseTime time.Duration = time.Hour // Initialize to high value

	for _, rt := range responseTimes {
		totalResponseTime += rt
		if rt > maxResponseTime {
			maxResponseTime = rt
		}
		if rt < minResponseTime {
			minResponseTime = rt
		}
	}

	avgResponseTime := totalResponseTime / time.Duration(len(responseTimes))

	t.Logf("Concurrent registration test results:")
	t.Logf("  Total users: %d", ConcurrentUsers)
	t.Logf("  Successful registrations: %d", len(responseTimes))
	t.Logf("  Failed registrations: %d", errorCount)
	t.Logf("  Total duration: %v", totalDuration)
	t.Logf("  Average response time: %v", avgResponseTime)
	t.Logf("  Min response time: %v", minResponseTime)
	t.Logf("  Max response time: %v", maxResponseTime)
	t.Logf("  Users created in repository: %d", mockRepo.Count())

	// Assertions
	assert.Equal(t, 0, errorCount, "All registrations should succeed")
	assert.Equal(t, ConcurrentUsers, mockRepo.Count(), "All users should be created")
	assert.Less(t, avgResponseTime, ResponseTimeTarget, "Average response time should be less than 500ms")
	assert.Less(t, maxResponseTime, 2*ResponseTimeTarget, "Max response time should be less than 1000ms")
}

func TestPerformance_UserLookup(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Pre-populate with test users
	testEmails := make([]string, 50)
	for i := 0; i < 50; i++ {
		email := fmt.Sprintf("lookup%d@example.com", i)
		testEmails[i] = email

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
		user, _ := domain.NewUser(email, string(hashedPassword))
		mockRepo.Create(context.Background(), user)
	}

	// Test user lookup performance
	var totalDuration time.Duration
	lookupCount := 100

	for i := 0; i < lookupCount; i++ {
		email := testEmails[i%len(testEmails)]

		start := time.Now()

		req := httptest.NewRequest("GET", "/api/v1/profile/"+email, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		duration := time.Since(start)
		totalDuration += duration

		assert.Equal(t, http.StatusOK, w.Code)
	}

	avgResponseTime := totalDuration / time.Duration(lookupCount)

	t.Logf("User lookup average response time: %v", avgResponseTime)
	assert.Less(t, avgResponseTime, ResponseTimeTarget, "User lookup should complete in less than 500ms")
}

func TestPerformance_EmailAvailabilityCheck(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Pre-populate with some users
	for i := 0; i < 25; i++ {
		email := fmt.Sprintf("existing%d@example.com", i)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("SecurePassword123!"), bcrypt.DefaultCost)
		user, _ := domain.NewUser(email, string(hashedPassword))
		mockRepo.Create(context.Background(), user)
	}

	// Test email availability check performance
	var totalDuration time.Duration
	checkCount := 100

	for i := 0; i < checkCount; i++ {
		var email string
		if i%2 == 0 {
			// Check existing email
			email = fmt.Sprintf("existing%d@example.com", i%25)
		} else {
			// Check non-existing email
			email = fmt.Sprintf("available%d@example.com", i)
		}

		start := time.Now()

		req := httptest.NewRequest("GET", "/api/v1/check-email/"+email, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		duration := time.Since(start)
		totalDuration += duration

		assert.Equal(t, http.StatusOK, w.Code)
	}

	avgResponseTime := totalDuration / time.Duration(checkCount)

	t.Logf("Email availability check average response time: %v", avgResponseTime)
	assert.Less(t, avgResponseTime, ResponseTimeTarget, "Email availability check should complete in less than 500ms")
}

func TestPerformance_MixedWorkload(t *testing.T) {
	router, mockRepo := setupTestRouter()

	var wg sync.WaitGroup
	results := make(chan struct {
		operation string
		duration  time.Duration
		success   bool
	}, 200)

	ctx, cancel := context.WithTimeout(context.Background(), PerformanceTestTimeout)
	defer cancel()

	// Mixed workload test
	start := time.Now()

	// Registration workload (30%)
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			requestStart := time.Now()

			requestBody := map[string]string{
				"email":    fmt.Sprintf("mixed%d@example.com", userID),
				"password": "SecurePassword123!",
			}

			jsonBody, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			results <- struct {
				operation string
				duration  time.Duration
				success   bool
			}{
				operation: "registration",
				duration:  time.Since(requestStart),
				success:   w.Code == http.StatusCreated,
			}
		}(i)
	}

	// Profile lookup workload (50%)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			requestStart := time.Now()

			email := fmt.Sprintf("mixed%d@example.com", userID%30)
			req := httptest.NewRequest("GET", "/api/v1/profile/"+email, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			results <- struct {
				operation string
				duration  time.Duration
				success   bool
			}{
				operation: "lookup",
				duration:  time.Since(requestStart),
				success:   w.Code == http.StatusOK || w.Code == http.StatusNotFound,
			}
		}(i)
	}

	// Email availability check workload (20%)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			requestStart := time.Now()

			email := fmt.Sprintf("check%d@example.com", userID)
			req := httptest.NewRequest("GET", "/api/v1/check-email/"+email, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			results <- struct {
				operation string
				duration  time.Duration
				success   bool
			}{
				operation: "availability",
				duration:  time.Since(requestStart),
				success:   w.Code == http.StatusOK,
			}
		}(i)
	}

	wg.Wait()
	close(results)

	totalDuration := time.Since(start)

	// Analyze results
	stats := make(map[string]struct {
		count        int
		totalTime    time.Duration
		successCount int
		maxTime      time.Duration
	})

	for result := range results {
		stat := stats[result.operation]
		stat.count++
		stat.totalTime += result.duration
		if result.success {
			stat.successCount++
		}
		if result.duration > stat.maxTime {
			stat.maxTime = result.duration
		}
		stats[result.operation] = stat
	}

	t.Logf("Mixed workload test results:")
	t.Logf("  Total duration: %v", totalDuration)
	t.Logf("  Users created: %d", mockRepo.Count())

	for operation, stat := range stats {
		avgTime := stat.totalTime / time.Duration(stat.count)
		successRate := float64(stat.successCount) / float64(stat.count) * 100

		t.Logf("  %s operations:", operation)
		t.Logf("    Count: %d", stat.count)
		t.Logf("    Success rate: %.1f%%", successRate)
		t.Logf("    Average time: %v", avgTime)
		t.Logf("    Max time: %v", stat.maxTime)

		assert.Greater(t, successRate, 95.0, "Success rate should be > 95%")
		assert.Less(t, avgTime, ResponseTimeTarget, "Average response time should be < 500ms")
	}
}
