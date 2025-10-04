package user

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test input validation for RegisterInput and LoginInput
func TestRegisterInput_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		input   RegisterInput
		wantErr bool
		errTag  string
	}{
		{
			name: "valid input",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "securepass123",
			},
			wantErr: false,
		},
		{
			name: "empty site ID",
			input: RegisterInput{
				SiteID:   uuid.Nil,
				Email:    "test@example.com",
				Password: "securepass123",
			},
			wantErr: true,
			errTag:  "required",
		},
		{
			name: "empty email",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "",
				Password: "securepass123",
			},
			wantErr: true,
			errTag:  "required",
		},
		{
			name: "invalid email format",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "invalid-email",
				Password: "securepass123",
			},
			wantErr: true,
			errTag:  "email",
		},
		{
			name: "email too long",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "verylongemailaddressthatexceedsthemaximumlengthof320characters@verylongdomainnamethatexceedsthemaximumlengthallowedforemailaddressesaccordingtorfc5321whichspecifiesthatthemaximumlengthofanemailaddressis320charactersincludingthelocalpartanddomainpartandtheatcharacterthatconnectsthem.com",
				Password: "securepass123",
			},
			wantErr: true,
			errTag:  "max",
		},
		{
			name: "empty password",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
			errTag:  "required",
		},
		{
			name: "password too short",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "short",
			},
			wantErr: true,
			errTag:  "min",
		},
		{
			name: "password too long",
			input: RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "thispasswordisextremelylongandexceedsthemaximumlengthof128characterswhichisspecifiedinthevalidationtagsforthepasswordfield",
			},
			wantErr: true,
			errTag:  "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTag != "" {
					validationErrors := err.(validator.ValidationErrors)
					assert.True(t, len(validationErrors) > 0)

					// Check if any validation error has the expected tag
					found := false
					for _, validationErr := range validationErrors {
						if validationErr.Tag() == tt.errTag {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected validation error with tag '%s'", tt.errTag)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoginInput_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		input   LoginInput
		wantErr bool
		errTag  string
	}{
		{
			name: "valid input",
			input: LoginInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty site ID",
			input: LoginInput{
				SiteID:   uuid.Nil,
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: true,
			errTag:  "required",
		},
		{
			name: "empty email",
			input: LoginInput{
				SiteID:   uuid.New(),
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
			errTag:  "required",
		},
		{
			name: "invalid email format",
			input: LoginInput{
				SiteID:   uuid.New(),
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
			errTag:  "email",
		},
		{
			name: "email too long",
			input: LoginInput{
				SiteID:   uuid.New(),
				Email:    "verylongemailaddressthatexceedsthemaximumlengthof320characters@verylongdomainnamethatexceedsthemaximumlengthallowedforemailaddressesaccordingtorfc5321whichspecifiesthatthemaximumlengthofanemailaddressis320charactersincludingthelocalpartanddomainpartandtheatcharacterthatconnectsthem.com",
				Password: "password123",
			},
			wantErr: true,
			errTag:  "max",
		},
		{
			name: "empty password",
			input: LoginInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
			errTag:  "required",
		},
		{
			name: "password too long",
			input: LoginInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: "thispasswordisextremelylongandexceedsthemaximumlengthof128characterswhichisspecifiedinthevalidationtagsforthepasswordfield",
			},
			wantErr: true,
			errTag:  "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTag != "" {
					validationErrors := err.(validator.ValidationErrors)
					assert.True(t, len(validationErrors) > 0)

					// Check if any validation error has the expected tag
					found := false
					for _, validationErr := range validationErrors {
						if validationErr.Tag() == tt.errTag {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected validation error with tag '%s'", tt.errTag)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSiteIDValidation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		siteID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "valid UUID",
			siteID:  uuid.New(),
			wantErr: false,
		},
		{
			name:    "nil UUID",
			siteID:  uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RegisterInput{
				SiteID:   tt.siteID,
				Email:    "test@example.com",
				Password: "securepass123",
			}

			err := validate.Struct(input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailValidation_EdgeCases(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "simple valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "email with plus",
			email:   "test+tag@example.com",
			wantErr: false,
		},
		{
			name:    "email with dots",
			email:   "test.user@example.com",
			wantErr: false,
		},
		{
			name:    "subdomain email",
			email:   "test@mail.example.com",
			wantErr: false,
		},
		{
			name:    "no @ symbol",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "multiple @ symbols",
			email:   "test@@example.com",
			wantErr: true,
		},
		{
			name:    "no domain",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "no local part",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "spaces in email",
			email:   "test user@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RegisterInput{
				SiteID:   uuid.New(),
				Email:    tt.email,
				Password: "securepass123",
			}

			err := validate.Struct(input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordValidation_EdgeCases(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "minimum length password",
			password: "12345678", // 8 characters - minimum
			wantErr:  false,
		},
		{
			name:     "maximum length password",
			password: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", // 128 characters
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "password!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "password with unicode",
			password: "пароль123",
			wantErr:  false,
		},
		{
			name:     "7 characters - too short",
			password: "1234567",
			wantErr:  true,
		},
		{
			name:     "129 characters - too long",
			password: "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901", // 129 characters
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RegisterInput{
				SiteID:   uuid.New(),
				Email:    "test@example.com",
				Password: tt.password,
			}

			err := validate.Struct(input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
