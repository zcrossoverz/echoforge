package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetadataType represents the type of metadata
type MetadataType string

const (
	MetadataTypeSEO       MetadataType = "seo"
	MetadataTypeSocial    MetadataType = "social"
	MetadataTypeCustom    MetadataType = "custom"
	MetadataTypeAnalytics MetadataType = "analytics"
	MetadataTypeSystem    MetadataType = "system"
	MetadataTypeExtension MetadataType = "extension"
)

// PostMetadata represents flexible metadata for posts using JSONB
// Provides extensible key-value storage with type safety and validation
type PostMetadata struct {
	ID        uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID    `gorm:"type:uuid;not null;index" json:"post_id" validate:"required"`
	Type      MetadataType `gorm:"size:20;not null;index" json:"type" validate:"required"`
	Key       string       `gorm:"size:100;not null;index" json:"key" validate:"required,max=100"`
	Value     string       `gorm:"type:jsonb" json:"value" validate:"required"`
	IsPublic  bool         `gorm:"default:true" json:"is_public"`
	IsSystem  bool         `gorm:"default:false" json:"is_system"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Post *Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
}

// BeforeCreate GORM hook - validate before creating
func (pm *PostMetadata) BeforeCreate(tx *gorm.DB) error {
	return pm.Validate()
}

// BeforeUpdate GORM hook - validate before updating
func (pm *PostMetadata) BeforeUpdate(tx *gorm.DB) error {
	return pm.Validate()
}

// Validate performs comprehensive validation on the PostMetadata
func (pm *PostMetadata) Validate() error {
	if err := pm.validateRequired(); err != nil {
		return err
	}
	if err := pm.validateType(); err != nil {
		return err
	}
	if err := pm.validateKey(); err != nil {
		return err
	}
	if err := pm.validateValue(); err != nil {
		return err
	}
	return nil
}

// validateRequired validates required fields
func (pm *PostMetadata) validateRequired() error {
	if pm.PostID == uuid.Nil {
		return errors.New("post_id is required")
	}
	if strings.TrimSpace(pm.Key) == "" {
		return errors.New("key is required")
	}
	if strings.TrimSpace(pm.Value) == "" {
		return errors.New("value is required")
	}
	return nil
}

// validateType validates the metadata type
func (pm *PostMetadata) validateType() error {
	validTypes := []MetadataType{
		MetadataTypeSEO, MetadataTypeSocial, MetadataTypeCustom,
		MetadataTypeAnalytics, MetadataTypeSystem, MetadataTypeExtension,
	}

	for _, validType := range validTypes {
		if pm.Type == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid metadata type: %s", pm.Type)
}

// validateKey validates the metadata key
func (pm *PostMetadata) validateKey() error {
	key := strings.TrimSpace(pm.Key)
	if len(key) > 100 {
		return errors.New("key cannot exceed 100 characters")
	}

	// Key should follow snake_case or kebab-case convention
	if strings.ContainsAny(key, " <>\"'&@#$%^*()+=[]{}|\\:;?.,") {
		return errors.New("key contains invalid characters")
	}

	return nil
}

// validateValue validates the metadata value (should be valid JSON)
func (pm *PostMetadata) validateValue() error {
	value := strings.TrimSpace(pm.Value)
	if len(value) == 0 {
		return errors.New("value cannot be empty")
	}

	// Basic JSON validation - should start with { or [ or be a quoted string
	firstChar := string(value[0])
	lastChar := string(value[len(value)-1])

	switch firstChar {
	case "{":
		if lastChar != "}" {
			return errors.New("invalid JSON object format")
		}
	case "[":
		if lastChar != "]" {
			return errors.New("invalid JSON array format")
		}
	case "\"":
		if lastChar != "\"" {
			return errors.New("invalid JSON string format")
		}
	case "t", "f": // true/false
		if value != "true" && value != "false" {
			return errors.New("invalid JSON boolean format")
		}
	default:
		// Check if it's a number
		if !isValidJSONNumber(value) {
			return errors.New("value must be valid JSON")
		}
	}

	// Check maximum size - 64KB limit for JSONB
	if len(value) > 65536 {
		return errors.New("value size exceeds 64KB limit")
	}

	return nil
}

// isValidJSONNumber checks if a string is a valid JSON number
func isValidJSONNumber(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Simple number validation - just check for digits, minus, plus, decimal point
	validChars := "0123456789.-+"
	for _, char := range s {
		found := false
		for _, validChar := range validChars {
			if char == validChar {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// IsSEO checks if this is SEO metadata
func (pm *PostMetadata) IsSEO() bool {
	return pm.Type == MetadataTypeSEO
}

// IsSocial checks if this is social media metadata
func (pm *PostMetadata) IsSocial() bool {
	return pm.Type == MetadataTypeSocial
}

// IsAnalytics checks if this is analytics metadata
func (pm *PostMetadata) IsAnalytics() bool {
	return pm.Type == MetadataTypeAnalytics
}

// IsExtension checks if this is extension-specific metadata
func (pm *PostMetadata) IsExtension() bool {
	return pm.Type == MetadataTypeExtension
}

// CanModify checks if the metadata can be modified
func (pm *PostMetadata) CanModify() error {
	if pm.IsSystem {
		return errors.New("system metadata cannot be modified")
	}
	return nil
}

// CanDelete checks if the metadata can be deleted
func (pm *PostMetadata) CanDelete() error {
	if pm.IsSystem {
		return errors.New("system metadata cannot be deleted")
	}
	return nil
}

// SetStringValue sets a string value (JSON-encoded)
func (pm *PostMetadata) SetStringValue(value string) error {
	pm.Value = fmt.Sprintf(`"%s"`, strings.ReplaceAll(value, `"`, `\"`))
	return pm.validateValue()
}

// SetBoolValue sets a boolean value
func (pm *PostMetadata) SetBoolValue(value bool) error {
	if value {
		pm.Value = "true"
	} else {
		pm.Value = "false"
	}
	return pm.validateValue()
}

// SetNumberValue sets a number value
func (pm *PostMetadata) SetNumberValue(value interface{}) error {
	pm.Value = fmt.Sprintf("%v", value)
	return pm.validateValue()
}

// SetObjectValue sets an object value (simplified JSON)
func (pm *PostMetadata) SetObjectValue(obj map[string]interface{}) error {
	if len(obj) == 0 {
		pm.Value = "{}"
		return nil
	}

	// Convert map to JSON string (simplified for demo)
	var jsonParts []string
	for key, value := range obj {
		jsonParts = append(jsonParts, fmt.Sprintf(`"%s":"%v"`, key, value))
	}
	pm.Value = "{" + strings.Join(jsonParts, ",") + "}"

	return pm.validateValue()
}

// SetArrayValue sets an array value (simplified JSON)
func (pm *PostMetadata) SetArrayValue(arr []interface{}) error {
	if len(arr) == 0 {
		pm.Value = "[]"
		return nil
	}

	// Convert array to JSON string (simplified for demo)
	var jsonParts []string
	for _, value := range arr {
		jsonParts = append(jsonParts, fmt.Sprintf(`"%v"`, value))
	}
	pm.Value = "[" + strings.Join(jsonParts, ",") + "]"

	return pm.validateValue()
}

// GetStringValue parses the value as a string
func (pm *PostMetadata) GetStringValue() (string, error) {
	if len(pm.Value) < 2 || pm.Value[0] != '"' || pm.Value[len(pm.Value)-1] != '"' {
		return "", errors.New("value is not a JSON string")
	}

	// Remove quotes and unescape
	return strings.ReplaceAll(pm.Value[1:len(pm.Value)-1], `\"`, `"`), nil
}

// GetBoolValue parses the value as a boolean
func (pm *PostMetadata) GetBoolValue() (bool, error) {
	switch pm.Value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("value is not a JSON boolean")
	}
}

// GetRawValue returns the raw JSON value
func (pm *PostMetadata) GetRawValue() string {
	return pm.Value
}

// GetValueType returns the detected JSON value type
func (pm *PostMetadata) GetValueType() string {
	if len(pm.Value) == 0 {
		return "empty"
	}

	switch pm.Value[0] {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f':
		return "boolean"
	default:
		if isValidJSONNumber(pm.Value) {
			return "number"
		}
		return "unknown"
	}
}

// GetMetadataInfo returns metadata information as a map
func (pm *PostMetadata) GetMetadataInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":       pm.Type,
		"key":        pm.Key,
		"value_type": pm.GetValueType(),
		"is_public":  pm.IsPublic,
		"is_system":  pm.IsSystem,
		"created_at": pm.CreatedAt,
		"updated_at": pm.UpdatedAt,
	}
}

// NewPostMetadata creates a new PostMetadata with validation
func NewPostMetadata(postID uuid.UUID, metadataType MetadataType, key, value string) (*PostMetadata, error) {
	metadata := &PostMetadata{
		ID:        uuid.New(),
		PostID:    postID,
		Type:      metadataType,
		Key:       strings.TrimSpace(key),
		Value:     strings.TrimSpace(value),
		IsPublic:  true,
		IsSystem:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	return metadata, nil
}

// NewSystemPostMetadata creates a new system PostMetadata that cannot be deleted
func NewSystemPostMetadata(postID uuid.UUID, metadataType MetadataType, key, value string) (*PostMetadata, error) {
	metadata, err := NewPostMetadata(postID, metadataType, key, value)
	if err != nil {
		return nil, err
	}

	metadata.IsSystem = true
	return metadata, nil
}

// NewSEOMetadata creates SEO-specific metadata
func NewSEOMetadata(postID uuid.UUID, key, value string) (*PostMetadata, error) {
	return NewPostMetadata(postID, MetadataTypeSEO, key, value)
}

// NewSocialMetadata creates social media-specific metadata
func NewSocialMetadata(postID uuid.UUID, key, value string) (*PostMetadata, error) {
	return NewPostMetadata(postID, MetadataTypeSocial, key, value)
}

// NewAnalyticsMetadata creates analytics-specific metadata
func NewAnalyticsMetadata(postID uuid.UUID, key, value string) (*PostMetadata, error) {
	return NewPostMetadata(postID, MetadataTypeAnalytics, key, value)
}

// TableName returns the database table name for GORM
func (PostMetadata) TableName() string {
	return "post_metadata"
}
