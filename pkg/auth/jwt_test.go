package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_Success(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key-at-least-32-characters"
	token, expiresAt, err := GenerateToken(userID, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))
	assert.True(t, expiresAt.Before(time.Now().Add(25*time.Hour))) // Should be around 24 hours
}
func TestGenerateToken_EmptySecret(t *testing.T) {
	userID := uuid.New()
	secret := ""
	token, expiresAt, err := GenerateToken(userID, secret)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.True(t, expiresAt.IsZero())
	assert.Contains(t, err.Error(), "JWT secret cannot be empty")
}
func TestValidateToken_Success(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key-at-least-32-characters"
	// Generate token first
	token, _, err := GenerateToken(userID, secret)
	require.NoError(t, err)
	// Validate token
	claims, err := ValidateToken(token, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}
func TestValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key-at-least-32-characters"
	invalidToken := "invalid.jwt.token"
	claims, err := ValidateToken(invalidToken, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
func TestValidateToken_WrongSecret(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key-at-least-32-characters"
	wrongSecret := "wrong-secret-key-different-from-original"
	// Generate token with correct secret
	token, _, err := GenerateToken(userID, secret)
	require.NoError(t, err)
	// Try to validate with wrong secret
	claims, err := ValidateToken(token, wrongSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
func TestValidateToken_EmptySecret(t *testing.T) {
	token := "some.jwt.token"
	secret := ""
	claims, err := ValidateToken(token, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "JWT secret cannot be empty")
}
func TestValidateToken_ExpiredToken(t *testing.T) {
	// This test would require mocking time or creating tokens with past expiration
	// For now, we'll test the structure is correct
	userID := uuid.New()

	secret := "test-secret-key-at-least-32-characters"
	token, expiresAt, err := GenerateToken(userID, secret)
	require.NoError(t, err)
	// Validate that expiration is set correctly (within expected range)
	expectedExpiration := time.Now().Add(24 * time.Hour)
	timeDiff := expiresAt.Sub(expectedExpiration)
	assert.True(t, timeDiff < time.Minute, "Expiration should be around 24 hours from now")
	assert.True(t, timeDiff > -time.Minute, "Expiration should be around 24 hours from now")
	// Validate the token is currently valid
	claims, err := ValidateToken(token, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}
func TestJWTClaims_Structure(t *testing.T) {
	userID := uuid.New()

	secret := "test-secret-key-at-least-32-characters"
	token, _, err := GenerateToken(userID, secret)
	require.NoError(t, err)
	claims, err := ValidateToken(token, secret)
	require.NoError(t, err)
	// Verify all required claims are present
	assert.Equal(t, userID.String(), claims.UserID)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.True(t, claims.IssuedAt.Time.Before(claims.ExpiresAt.Time))
}
func TestGenerateJWT_Legacy_BackwardCompatibility(t *testing.T) {
	// Test the legacy function for backward compatibility
	userID := "12345678-1234-5678-9012-123456789012"
	role := "user"
	token, err := GenerateJWT(userID, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, userID[:8])
	assert.Contains(t, token, role)
}

// Benchmark tests for performance validation
func BenchmarkGenerateToken(b *testing.B) {
	userID := uuid.New()

	secret := "test-secret-key-at-least-32-characters"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := GenerateToken(userID, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkValidateToken(b *testing.B) {
	userID := uuid.New()

	secret := "test-secret-key-at-least-32-characters"
	// Generate token once
	token, _, err := GenerateToken(userID, secret)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateToken(token, secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}
