package auth

import (
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Password validation constants as per specification
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
	BcryptCost        = 12 // As per specification: bcrypt cost 12
)

// Password validation patterns
var (
	// At least one uppercase letter
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	// At least one lowercase letter
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	// At least one digit
	digitRegex = regexp.MustCompile(`[0-9]`)
	// At least one special character
	specialCharRegex = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
)

// Password validation errors
var (
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong       = errors.New("password must not exceed 128 characters")
	ErrPasswordNoUppercase   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial     = errors.New("password must contain at least one special character")
	ErrPasswordEmpty         = errors.New("password cannot be empty")
	ErrPasswordHashingFailed = errors.New("failed to hash password")
)

// HashPassword generates a bcrypt hash from a plain text password using cost factor 12
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrPasswordEmpty
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrPasswordHashingFailed, err)
	}

	return string(hash), nil
}

// VerifyPassword compares a plain text password with a bcrypt hash
func VerifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordEmpty
	}

	// Check length constraints
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	// Check for required character types
	if !uppercaseRegex.MatchString(password) {
		return ErrPasswordNoUppercase
	}
	if !lowercaseRegex.MatchString(password) {
		return ErrPasswordNoLowercase
	}
	if !digitRegex.MatchString(password) {
		return ErrPasswordNoDigit
	}
	if !specialCharRegex.MatchString(password) {
		return ErrPasswordNoSpecial
	}

	return nil
}

// ValidateAndHashPassword validates a password and returns its hash if valid
func ValidateAndHashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	return HashPassword(password)
}

// PasswordStrength calculates password strength score (0-100)
func PasswordStrength(password string) int {
	if password == "" {
		return 0
	}

	score := 0

	// Length scoring
	length := len(password)
	if length >= 8 {
		score += 20
	}
	if length >= 12 {
		score += 10
	}
	if length >= 16 {
		score += 10
	}

	// Character type scoring
	if uppercaseRegex.MatchString(password) {
		score += 15
	}
	if lowercaseRegex.MatchString(password) {
		score += 15
	}
	if digitRegex.MatchString(password) {
		score += 15
	}
	if specialCharRegex.MatchString(password) {
		score += 15
	}

	// Complexity bonus for variety
	uniqueChars := make(map[rune]bool)
	for _, char := range password {
		uniqueChars[char] = true
	}
	if len(uniqueChars) >= length/2 {
		score += 10 // Good character diversity
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// GeneratePasswordRequirements returns a human-readable list of password requirements
func GeneratePasswordRequirements() []string {
	return []string{
		fmt.Sprintf("At least %d characters long", MinPasswordLength),
		fmt.Sprintf("No more than %d characters", MaxPasswordLength),
		"At least one uppercase letter (A-Z)",
		"At least one lowercase letter (a-z)",
		"At least one digit (0-9)",
		"At least one special character (!@#$%^&*()_+-=[]{}|;':\",./<>?~`)",
	}
}

// SanitizePasswordError returns a user-friendly error message for password validation
func SanitizePasswordError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, ErrPasswordEmpty):
		return "Password is required"
	case errors.Is(err, ErrPasswordTooShort):
		return fmt.Sprintf("Password must be at least %d characters long", MinPasswordLength)
	case errors.Is(err, ErrPasswordTooLong):
		return fmt.Sprintf("Password must not exceed %d characters", MaxPasswordLength)
	case errors.Is(err, ErrPasswordNoUppercase):
		return "Password must contain at least one uppercase letter"
	case errors.Is(err, ErrPasswordNoLowercase):
		return "Password must contain at least one lowercase letter"
	case errors.Is(err, ErrPasswordNoDigit):
		return "Password must contain at least one number"
	case errors.Is(err, ErrPasswordNoSpecial):
		return "Password must contain at least one special character"
	case errors.Is(err, ErrPasswordHashingFailed):
		return "Password processing failed"
	default:
		// Don't expose internal errors to users
		return "Password validation failed"
	}
}

// PasswordService implements the domain's PasswordService interface
type PasswordService struct{}

// NewPasswordService creates a new password service
func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// HashPassword implements domain.PasswordService
func (s *PasswordService) HashPassword(password string) (string, error) {
	return HashPassword(password)
}

// VerifyPassword implements domain.PasswordService
func (s *PasswordService) VerifyPassword(password, hash string) bool {
	return VerifyPassword(password, hash)
}

// ValidatePassword implements domain.PasswordService
func (s *PasswordService) ValidatePassword(password string) error {
	return ValidatePassword(password)
}
