package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostTag represents a tag that can be applied to posts
// Provides flexible tagging system with usage tracking and metadata
type PostTag struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name" validate:"required,min=1,max=100"`
	Slug        string    `gorm:"size:120;not null;uniqueIndex" json:"slug"`
	Description string    `gorm:"size:500" json:"description" validate:"max=500"`
	Color       string    `gorm:"size:7;default:'#6B7280'" json:"color" validate:"hexcolor"`
	UsageCount  int       `gorm:"default:0;index" json:"usage_count"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	Metadata    string    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Posts []Post `gorm:"many2many:post_tag_assignments;" json:"posts,omitempty"`
}

// BeforeCreate GORM hook - auto-generate slug and validate
func (pt *PostTag) BeforeCreate(tx *gorm.DB) error {
	if err := pt.GenerateSlug(); err != nil {
		return err
	}
	return pt.Validate()
}

// BeforeUpdate GORM hook - validate and update slug if name changed
func (pt *PostTag) BeforeUpdate(tx *gorm.DB) error {
	if err := pt.GenerateSlug(); err != nil {
		return err
	}
	return pt.Validate()
}

// Validate performs comprehensive validation on the PostTag
func (pt *PostTag) Validate() error {
	if err := pt.validateName(); err != nil {
		return err
	}
	if err := pt.validateColor(); err != nil {
		return err
	}
	if err := pt.validateMetadata(); err != nil {
		return err
	}
	return nil
}

// validateName validates the tag name
func (pt *PostTag) validateName() error {
	if strings.TrimSpace(pt.Name) == "" {
		return errors.New("tag name cannot be empty")
	}

	if len(pt.Name) > 100 {
		return errors.New("tag name cannot exceed 100 characters")
	}

	// Check for invalid characters
	if strings.ContainsAny(pt.Name, "<>\"'&") {
		return errors.New("tag name contains invalid characters")
	}

	return nil
}

// validateColor validates the color hex code
func (pt *PostTag) validateColor() error {
	if pt.Color == "" {
		pt.Color = "#6B7280" // Default gray color
		return nil
	}

	hexColorRegex := regexp.MustCompile(`^#[A-Fa-f0-9]{6}$`)
	if !hexColorRegex.MatchString(pt.Color) {
		return errors.New("color must be a valid hex color code (e.g., #FF5733)")
	}

	return nil
}

// validateMetadata validates the JSON metadata
func (pt *PostTag) validateMetadata() error {
	if pt.Metadata == "" {
		return nil
	}

	// Basic JSON structure validation
	if !strings.HasPrefix(pt.Metadata, "{") || !strings.HasSuffix(pt.Metadata, "}") {
		return errors.New("metadata must be valid JSON object")
	}

	return nil
}

// GenerateSlug creates a URL-safe slug from the tag name
func (pt *PostTag) GenerateSlug() error {
	if pt.Name == "" {
		return errors.New("cannot generate slug: tag name is empty")
	}

	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(strings.TrimSpace(pt.Name))
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return errors.New("cannot generate valid slug from tag name")
	}

	if len(slug) > 120 {
		slug = slug[:120]
		slug = strings.Trim(slug, "-")
	}

	pt.Slug = slug
	return nil
}

// IncrementUsage increments the usage count for this tag
func (pt *PostTag) IncrementUsage() {
	pt.UsageCount++
}

// DecrementUsage decrements the usage count for this tag
func (pt *PostTag) DecrementUsage() {
	if pt.UsageCount > 0 {
		pt.UsageCount--
	}
}

// CanDelete checks if the tag can be deleted
func (pt *PostTag) CanDelete() error {
	if pt.IsSystem {
		return errors.New("system tags cannot be deleted")
	}
	return nil
}

// CanModify checks if the tag can be modified
func (pt *PostTag) CanModify() error {
	if pt.IsSystem {
		return errors.New("system tags cannot be modified")
	}
	return nil
}

// GetUsageStats returns usage statistics for the tag
func (pt *PostTag) GetUsageStats() map[string]interface{} {
	return map[string]interface{}{
		"usage_count": pt.UsageCount,
		"is_system":   pt.IsSystem,
		"created_at":  pt.CreatedAt,
		"slug":        pt.Slug,
	}
}

// SetMetadata sets JSON metadata for the tag
func (pt *PostTag) SetMetadata(metadata map[string]interface{}) error {
	if len(metadata) == 0 {
		pt.Metadata = ""
		return nil
	}

	// Convert map to JSON string (simplified for demo)
	var jsonParts []string
	for key, value := range metadata {
		jsonParts = append(jsonParts, fmt.Sprintf(`"%s":"%v"`, key, value))
	}
	pt.Metadata = "{" + strings.Join(jsonParts, ",") + "}"

	return pt.validateMetadata()
}

// GetMetadata parses and returns the JSON metadata as a map
func (pt *PostTag) GetMetadata() map[string]interface{} {
	// Simplified JSON parsing for demo - in production use json.Unmarshal
	metadata := make(map[string]interface{})
	if pt.Metadata == "" {
		return metadata
	}

	// This is a simplified parser - in real implementation use proper JSON parsing
	metadata["raw"] = pt.Metadata
	return metadata
}

// NewPostTag creates a new PostTag with validation
func NewPostTag(name, description, color string) (*PostTag, error) {
	tag := &PostTag{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Color:       color,
		UsageCount:  0,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tag.GenerateSlug(); err != nil {
		return nil, err
	}

	if err := tag.Validate(); err != nil {
		return nil, err
	}

	return tag, nil
}

// NewSystemPostTag creates a new system PostTag that cannot be deleted
func NewSystemPostTag(name, description, color string) (*PostTag, error) {
	tag, err := NewPostTag(name, description, color)
	if err != nil {
		return nil, err
	}

	tag.IsSystem = true
	return tag, nil
}

// TableName returns the database table name for GORM
func (PostTag) TableName() string {
	return "post_tags"
}
