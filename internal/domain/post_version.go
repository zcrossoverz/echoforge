package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChangeType represents the type of change made to a post
type ChangeType string

const (
	ChangeTypeCreate       ChangeType = "create"
	ChangeTypeUpdate       ChangeType = "update"
	ChangeTypeStatusChange ChangeType = "status_change"
	ChangeTypePublish      ChangeType = "publish"
	ChangeTypeUnpublish    ChangeType = "unpublish"
	ChangeTypeArchive      ChangeType = "archive"
	ChangeTypeRestore      ChangeType = "restore"
	ChangeTypeDelete       ChangeType = "delete"
)

// PostVersion represents a version/revision of a post
// Provides complete audit trail and rollback capabilities
type PostVersion struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"post_id" validate:"required"`
	VersionNumber int        `gorm:"not null;index" json:"version_number" validate:"required,min=1"`
	Title         string     `gorm:"size:255;not null" json:"title" validate:"required,max=255"`
	Content       string     `gorm:"type:text" json:"content"`
	Excerpt       string     `gorm:"size:500" json:"excerpt" validate:"max=500"`
	Status        PostStatus `gorm:"size:20;not null;index" json:"status" validate:"required"`
	AuthorID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id" validate:"required"`
	EditorID      *uuid.UUID `gorm:"type:uuid;index" json:"editor_id,omitempty"`
	ChangeType    ChangeType `gorm:"size:20;not null;index" json:"change_type" validate:"required"`
	ChangeReason  string     `gorm:"size:500" json:"change_reason" validate:"max=500"`
	ChangeSummary string     `gorm:"size:255" json:"change_summary" validate:"max=255"`
	IsCurrent     bool       `gorm:"default:false;index" json:"is_current"`
	Metadata      string     `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Post   *Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Author *User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Editor *User `gorm:"foreignKey:EditorID" json:"editor,omitempty"`
}

// BeforeCreate GORM hook - validate before creating
func (pv *PostVersion) BeforeCreate(tx *gorm.DB) error {
	return pv.Validate()
}

// BeforeUpdate GORM hook - validate before updating
func (pv *PostVersion) BeforeUpdate(tx *gorm.DB) error {
	return pv.Validate()
}

// Validate performs comprehensive validation on the PostVersion
func (pv *PostVersion) Validate() error {
	if err := pv.validateRequired(); err != nil {
		return err
	}
	if err := pv.validateVersionNumber(); err != nil {
		return err
	}
	if err := pv.validateChangeType(); err != nil {
		return err
	}
	if err := pv.validateStatus(); err != nil {
		return err
	}
	return nil
}

// validateRequired validates required fields
func (pv *PostVersion) validateRequired() error {
	if pv.PostID == uuid.Nil {
		return errors.New("post_id is required")
	}
	if strings.TrimSpace(pv.Title) == "" {
		return errors.New("title is required")
	}
	if pv.AuthorID == uuid.Nil {
		return errors.New("author_id is required")
	}
	return nil
}

// validateVersionNumber validates the version number
func (pv *PostVersion) validateVersionNumber() error {
	if pv.VersionNumber < 1 {
		return errors.New("version_number must be greater than 0")
	}
	if pv.VersionNumber > 10000 {
		return errors.New("version_number cannot exceed 10000")
	}
	return nil
}

// validateChangeType validates the change type
func (pv *PostVersion) validateChangeType() error {
	validTypes := []ChangeType{
		ChangeTypeCreate, ChangeTypeUpdate, ChangeTypeStatusChange,
		ChangeTypePublish, ChangeTypeUnpublish, ChangeTypeArchive,
		ChangeTypeRestore, ChangeTypeDelete,
	}

	for _, validType := range validTypes {
		if pv.ChangeType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid change_type: %s", pv.ChangeType)
}

// validateStatus validates the post status
func (pv *PostVersion) validateStatus() error {
	validStatuses := []PostStatus{
		PostStatusDraft, PostStatusPendingApproval, PostStatusPublished,
		PostStatusScheduled, PostStatusArchived,
	}

	for _, validStatus := range validStatuses {
		if pv.Status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s", pv.Status)
}

// IsEditorial checks if this is an editorial change (different editor than author)
func (pv *PostVersion) IsEditorial() bool {
	return pv.EditorID != nil && *pv.EditorID != pv.AuthorID
}

// GetChangeDescription returns a human-readable description of the change
func (pv *PostVersion) GetChangeDescription() string {
	switch pv.ChangeType {
	case ChangeTypeCreate:
		return "Post created"
	case ChangeTypeUpdate:
		return "Content updated"
	case ChangeTypeStatusChange:
		return fmt.Sprintf("Status changed to %s", pv.Status)
	case ChangeTypePublish:
		return "Post published"
	case ChangeTypeUnpublish:
		return "Post unpublished"
	case ChangeTypeArchive:
		return "Post archived"
	case ChangeTypeRestore:
		return "Post restored"
	case ChangeTypeDelete:
		return "Post deleted"
	default:
		return string(pv.ChangeType)
	}
}

// GetChangeSummaryOrDefault returns change summary or default description
func (pv *PostVersion) GetChangeSummaryOrDefault() string {
	if pv.ChangeSummary != "" {
		return pv.ChangeSummary
	}
	return pv.GetChangeDescription()
}

// SetAsCurrent marks this version as the current version
func (pv *PostVersion) SetAsCurrent() {
	pv.IsCurrent = true
}

// UnsetAsCurrent marks this version as not current
func (pv *PostVersion) UnsetAsCurrent() {
	pv.IsCurrent = false
}

// HasChanges compares this version with another to detect changes
func (pv *PostVersion) HasChanges(other *PostVersion) bool {
	if other == nil {
		return true
	}

	return pv.Title != other.Title ||
		pv.Content != other.Content ||
		pv.Excerpt != other.Excerpt ||
		pv.Status != other.Status
}

// GetContentPreview returns a preview of the content (first 200 characters)
func (pv *PostVersion) GetContentPreview() string {
	if len(pv.Content) <= 200 {
		return pv.Content
	}

	preview := pv.Content[:200]
	// Try to break at word boundary
	if lastSpace := strings.LastIndex(preview, " "); lastSpace > 150 {
		preview = preview[:lastSpace]
	}

	return preview + "..."
}

// GetVersionInfo returns version information as a map
func (pv *PostVersion) GetVersionInfo() map[string]interface{} {
	info := map[string]interface{}{
		"version_number": pv.VersionNumber,
		"change_type":    pv.ChangeType,
		"change_summary": pv.GetChangeSummaryOrDefault(),
		"is_current":     pv.IsCurrent,
		"is_editorial":   pv.IsEditorial(),
		"created_at":     pv.CreatedAt,
	}

	if pv.EditorID != nil {
		info["editor_id"] = *pv.EditorID
	}

	return info
}

// SetMetadata sets JSON metadata for the version
func (pv *PostVersion) SetMetadata(metadata map[string]interface{}) error {
	if len(metadata) == 0 {
		pv.Metadata = ""
		return nil
	}

	// Convert map to JSON string (simplified for demo)
	var jsonParts []string
	for key, value := range metadata {
		jsonParts = append(jsonParts, fmt.Sprintf(`"%s":"%v"`, key, value))
	}
	pv.Metadata = "{" + strings.Join(jsonParts, ",") + "}"

	return nil
}

// GetMetadata parses and returns the JSON metadata as a map
func (pv *PostVersion) GetMetadata() map[string]interface{} {
	// Simplified JSON parsing for demo - in production use json.Unmarshal
	metadata := make(map[string]interface{})
	if pv.Metadata == "" {
		return metadata
	}

	// This is a simplified parser - in real implementation use proper JSON parsing
	metadata["raw"] = pv.Metadata
	return metadata
}

// CreateFromPost creates a new version from a post
func (pv *PostVersion) CreateFromPost(post *Post, changeType ChangeType, editorID *uuid.UUID, reason, summary string) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}

	pv.PostID = post.ID
	pv.Title = post.Title
	pv.Content = post.Content
	pv.Excerpt = pv.generateExcerpt(post.Content)
	pv.Status = post.Status
	pv.AuthorID = post.AuthorID
	pv.EditorID = editorID
	pv.ChangeType = changeType
	pv.ChangeReason = strings.TrimSpace(reason)
	pv.ChangeSummary = strings.TrimSpace(summary)
	pv.IsCurrent = true
	pv.CreatedAt = time.Now()

	return pv.Validate()
}

// NewPostVersion creates a new PostVersion with validation
func NewPostVersion(postID, authorID uuid.UUID, versionNumber int, title, content string, status PostStatus, changeType ChangeType) (*PostVersion, error) {
	version := &PostVersion{
		ID:            uuid.New(),
		PostID:        postID,
		VersionNumber: versionNumber,
		Title:         strings.TrimSpace(title),
		Content:       content,
		Status:        status,
		AuthorID:      authorID,
		ChangeType:    changeType,
		IsCurrent:     false,
		CreatedAt:     time.Now(),
	}

	if err := version.Validate(); err != nil {
		return nil, err
	}

	return version, nil
}

// generateExcerpt creates an excerpt from content
func (pv *PostVersion) generateExcerpt(content string) string {
	if len(content) <= 500 {
		return content
	}

	excerpt := content[:500]
	// Try to break at word boundary
	if lastSpace := strings.LastIndex(excerpt, " "); lastSpace > 400 {
		excerpt = excerpt[:lastSpace]
	}

	return excerpt + "..."
}

// NewPostVersionFromPost creates a new version from an existing Post
func NewPostVersionFromPost(post *Post, versionNumber int, changeType ChangeType, editorID *uuid.UUID, reason, summary string) (*PostVersion, error) {
	if post == nil {
		return nil, errors.New("post cannot be nil")
	}

	version := &PostVersion{
		ID:            uuid.New(),
		PostID:        post.ID,
		VersionNumber: versionNumber,
		Title:         post.Title,
		Content:       post.Content,
		Excerpt:       "",
		Status:        post.Status,
		AuthorID:      post.AuthorID,
		EditorID:      editorID,
		ChangeType:    changeType,
		ChangeReason:  strings.TrimSpace(reason),
		ChangeSummary: strings.TrimSpace(summary),
		IsCurrent:     true,
		CreatedAt:     time.Now(),
	}

	// Generate excerpt from content
	version.Excerpt = version.generateExcerpt(post.Content)

	if err := version.Validate(); err != nil {
		return nil, err
	}

	return version, nil
}

// TableName returns the database table name for GORM
func (PostVersion) TableName() string {
	return "post_versions"
}
