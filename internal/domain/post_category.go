package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PostCategory represents a hierarchical category entity
type PostCategory struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string     `json:"name" gorm:"type:varchar(100);not null"`
	Slug        string     `json:"slug" gorm:"type:varchar(100);not null"`
	Description string     `json:"description,omitempty" gorm:"type:text"`
	ParentID    *uuid.UUID `json:"parentId,omitempty" gorm:"type:uuid;index"`
	SortOrder   int        `json:"sortOrder" gorm:"not null;default:0"`
	IsActive    bool       `json:"isActive" gorm:"not null;default:true"`
	PostCount   int        `json:"postCount" gorm:"not null;default:0"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"not null;default:now()"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"not null;default:now()"`

	// Relationships
	Parent   *PostCategory  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []PostCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName specifies the table name for GORM
func (PostCategory) TableName() string {
	return "post_categories"
}

// Validate checks if the post category entity is valid
func (pc *PostCategory) Validate() error {
	if pc.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(pc.Name) > 100 {
		return fmt.Errorf("name cannot exceed 100 characters")
	}

	if pc.Slug == "" {
		return fmt.Errorf("slug is required")
	}

	if len(pc.Slug) > 100 {
		return fmt.Errorf("slug cannot exceed 100 characters")
	}

	if !pc.isValidSlug() {
		return fmt.Errorf("slug must be URL-safe (lowercase letters, numbers, hyphens only)")
	}

	if pc.SortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	// Prevent self-reference
	if pc.ParentID != nil && *pc.ParentID == pc.ID {
		return fmt.Errorf("category cannot be its own parent")
	}

	return nil
}

// isValidSlug checks if the slug is URL-safe
func (pc *PostCategory) isValidSlug() bool {
	if pc.Slug == "" {
		return false
	}

	for _, char := range pc.Slug {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	// Slug cannot start or end with hyphen
	return !strings.HasPrefix(pc.Slug, "-") && !strings.HasSuffix(pc.Slug, "-")
}

// GenerateSlug generates a URL-safe slug from the name
func (pc *PostCategory) GenerateSlug() {
	slug := strings.ToLower(pc.Name)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	var cleanSlug strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			cleanSlug.WriteRune(char)
		}
	}

	pc.Slug = cleanSlug.String()

	// Remove multiple consecutive hyphens
	for strings.Contains(pc.Slug, "--") {
		pc.Slug = strings.ReplaceAll(pc.Slug, "--", "-")
	}

	// Trim hyphens from start and end
	pc.Slug = strings.Trim(pc.Slug, "-")
}

// IsRoot checks if this category is a root category (no parent)
func (pc *PostCategory) IsRoot() bool {
	return pc.ParentID == nil
}

// GetDepth returns the depth level of this category in the hierarchy
func (pc *PostCategory) GetDepth() int {
	if pc.IsRoot() {
		return 0
	}
	// This would require recursive loading of parents
	// For now, return 1 if has parent, actual depth calculation would need repository
	return 1
}

// GetFullPath returns the full path of this category (parent/child/grandchild)
func (pc *PostCategory) GetFullPath() string {
	// This would require loading all parent categories
	// For now, just return the name
	return pc.Name
}

// IncrementPostCount increases the post count for this category
func (pc *PostCategory) IncrementPostCount() {
	pc.PostCount++
	pc.UpdatedAt = time.Now()
}

// DecrementPostCount decreases the post count for this category
func (pc *PostCategory) DecrementPostCount() {
	if pc.PostCount > 0 {
		pc.PostCount--
	}
	pc.UpdatedAt = time.Now()
}

// CanBeDeleted checks if this category can be deleted
func (pc *PostCategory) CanBeDeleted() bool {
	// Categories with posts or children cannot be deleted
	return pc.PostCount == 0 && len(pc.Children) == 0
}

// IsSystemCategory checks if this is a system-defined category
func (pc *PostCategory) IsSystemCategory() bool {
	systemCategories := []string{"uncategorized"}
	return contains(systemCategories, strings.ToLower(pc.Slug))
}

// BeforeCreate GORM hook called before creating a record
func (pc *PostCategory) BeforeCreate() error {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}

	// Generate slug if not provided
	if pc.Slug == "" {
		pc.GenerateSlug()
	}

	pc.CreatedAt = time.Now()
	pc.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate GORM hook called before updating a record
func (pc *PostCategory) BeforeUpdate() error {
	pc.UpdatedAt = time.Now()
	return nil
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
