// pkg/auth/jwt.go (stub - full JWT ở Sprint 2)
package auth

import "fmt"

func GenerateJWT(userID, role string) (string, error) {
	// Stub: Return fake token, no real signing
	return fmt.Sprintf("jwt-token.%s.%s", userID[:8], role), nil
}
