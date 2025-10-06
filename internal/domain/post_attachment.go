package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttachmentType represents the type of attachment
type AttachmentType string

const (
	AttachmentTypeImage    AttachmentType = "image"
	AttachmentTypeVideo    AttachmentType = "video"
	AttachmentTypeAudio    AttachmentType = "audio"
	AttachmentTypeDocument AttachmentType = "document"
	AttachmentTypeArchive  AttachmentType = "archive"
	AttachmentTypeOther    AttachmentType = "other"
)

// PostAttachment represents a file attachment for posts
// Handles various file types with security and metadata tracking
type PostAttachment struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"post_id" validate:"required"`
	FileName      string         `gorm:"size:255;not null" json:"file_name" validate:"required,max=255"`
	OriginalName  string         `gorm:"size:255;not null" json:"original_name" validate:"required,max=255"`
	FileSize      int64          `gorm:"not null" json:"file_size" validate:"required,min=1"`
	ContentType   string         `gorm:"size:100;not null" json:"content_type" validate:"required,max=100"`
	Type          AttachmentType `gorm:"size:20;not null;index" json:"type" validate:"required"`
	FilePath      string         `gorm:"size:500;not null" json:"file_path" validate:"required,max=500"`
	URL           string         `gorm:"size:500" json:"url" validate:"max=500"`
	Alt           string         `gorm:"size:255" json:"alt" validate:"max=255"`
	Caption       string         `gorm:"size:500" json:"caption" validate:"max=500"`
	Width         int            `gorm:"default:0" json:"width"`
	Height        int            `gorm:"default:0" json:"height"`
	Duration      int            `gorm:"default:0" json:"duration"` // For video/audio in seconds
	IsPublic      bool           `gorm:"default:true" json:"is_public"`
	DownloadCount int            `gorm:"default:0" json:"download_count"`
	Metadata      string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Post *Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
}

// BeforeCreate GORM hook - validate before creating
func (pa *PostAttachment) BeforeCreate(tx *gorm.DB) error {
	if err := pa.DetermineType(); err != nil {
		return err
	}
	return pa.Validate()
}

// BeforeUpdate GORM hook - validate before updating
func (pa *PostAttachment) BeforeUpdate(tx *gorm.DB) error {
	if err := pa.DetermineType(); err != nil {
		return err
	}
	return pa.Validate()
}

// Validate performs comprehensive validation on the PostAttachment
func (pa *PostAttachment) Validate() error {
	if err := pa.validateRequired(); err != nil {
		return err
	}
	if err := pa.validateFileSize(); err != nil {
		return err
	}
	if err := pa.validateContentType(); err != nil {
		return err
	}
	if err := pa.validateDimensions(); err != nil {
		return err
	}
	return nil
}

// validateRequired validates required fields
func (pa *PostAttachment) validateRequired() error {
	if pa.PostID == uuid.Nil {
		return errors.New("post_id is required")
	}
	if strings.TrimSpace(pa.FileName) == "" {
		return errors.New("file_name is required")
	}
	if strings.TrimSpace(pa.OriginalName) == "" {
		return errors.New("original_name is required")
	}
	if strings.TrimSpace(pa.FilePath) == "" {
		return errors.New("file_path is required")
	}
	return nil
}

// validateFileSize validates file size constraints
func (pa *PostAttachment) validateFileSize() error {
	if pa.FileSize <= 0 {
		return errors.New("file_size must be greater than 0")
	}

	// Maximum file size: 100MB (configurable in production)
	maxSize := int64(100 * 1024 * 1024)
	if pa.FileSize > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", pa.FileSize, maxSize)
	}

	return nil
}

// validateContentType validates content type
func (pa *PostAttachment) validateContentType() error {
	if strings.TrimSpace(pa.ContentType) == "" {
		return errors.New("content_type is required")
	}

	// Basic MIME type validation
	if !strings.Contains(pa.ContentType, "/") {
		return errors.New("invalid content_type format")
	}

	return nil
}

// validateDimensions validates dimensions for media files
func (pa *PostAttachment) validateDimensions() error {
	if pa.Type == AttachmentTypeImage || pa.Type == AttachmentTypeVideo {
		if pa.Width < 0 || pa.Height < 0 {
			return errors.New("width and height cannot be negative")
		}
		if pa.Width > 10000 || pa.Height > 10000 {
			return errors.New("width and height cannot exceed 10000 pixels")
		}
	}

	if pa.Type == AttachmentTypeVideo || pa.Type == AttachmentTypeAudio {
		if pa.Duration < 0 {
			return errors.New("duration cannot be negative")
		}
		if pa.Duration > 86400 { // Max 24 hours
			return errors.New("duration cannot exceed 24 hours")
		}
	}

	return nil
}

// DetermineType determines attachment type based on content type
func (pa *PostAttachment) DetermineType() error {
	if pa.ContentType == "" {
		return errors.New("content_type is required to determine attachment type")
	}

	contentType := strings.ToLower(pa.ContentType)

	switch {
	case strings.HasPrefix(contentType, "image/"):
		pa.Type = AttachmentTypeImage
	case strings.HasPrefix(contentType, "video/"):
		pa.Type = AttachmentTypeVideo
	case strings.HasPrefix(contentType, "audio/"):
		pa.Type = AttachmentTypeAudio
	case contentType == "application/pdf" ||
		strings.HasPrefix(contentType, "text/") ||
		contentType == "application/msword" ||
		contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		pa.Type = AttachmentTypeDocument
	case contentType == "application/zip" ||
		contentType == "application/x-zip-compressed" ||
		contentType == "application/x-rar-compressed" ||
		contentType == "application/x-tar" ||
		contentType == "application/gzip":
		pa.Type = AttachmentTypeArchive
	default:
		pa.Type = AttachmentTypeOther
	}

	return nil
}

// GetFileExtension returns the file extension
func (pa *PostAttachment) GetFileExtension() string {
	return strings.ToLower(filepath.Ext(pa.FileName))
}

// IsImage checks if the attachment is an image
func (pa *PostAttachment) IsImage() bool {
	return pa.Type == AttachmentTypeImage
}

// IsVideo checks if the attachment is a video
func (pa *PostAttachment) IsVideo() bool {
	return pa.Type == AttachmentTypeVideo
}

// IsAudio checks if the attachment is an audio file
func (pa *PostAttachment) IsAudio() bool {
	return pa.Type == AttachmentTypeAudio
}

// IsDocument checks if the attachment is a document
func (pa *PostAttachment) IsDocument() bool {
	return pa.Type == AttachmentTypeDocument
}

// IncrementDownload increments the download count
func (pa *PostAttachment) IncrementDownload() {
	pa.DownloadCount++
}

// SetDimensions sets width and height for image/video attachments
func (pa *PostAttachment) SetDimensions(width, height int) error {
	if !pa.IsImage() && !pa.IsVideo() {
		return errors.New("dimensions can only be set for image or video attachments")
	}

	if width < 0 || height < 0 {
		return errors.New("dimensions cannot be negative")
	}

	pa.Width = width
	pa.Height = height
	return nil
}

// SetDuration sets duration for video/audio attachments
func (pa *PostAttachment) SetDuration(duration int) error {
	if !pa.IsVideo() && !pa.IsAudio() {
		return errors.New("duration can only be set for video or audio attachments")
	}

	if duration < 0 {
		return errors.New("duration cannot be negative")
	}

	pa.Duration = duration
	return nil
}

// GetDisplaySize returns human-readable file size
func (pa *PostAttachment) GetDisplaySize() string {
	size := float64(pa.FileSize)
	units := []string{"B", "KB", "MB", "GB"}

	for _, unit := range units {
		if size < 1024 {
			return fmt.Sprintf("%.1f %s", size, unit)
		}
		size /= 1024
	}

	return fmt.Sprintf("%.1f TB", size)
}

// GetDisplayDuration returns human-readable duration
func (pa *PostAttachment) GetDisplayDuration() string {
	if pa.Duration == 0 {
		return "0s"
	}

	hours := pa.Duration / 3600
	minutes := (pa.Duration % 3600) / 60
	seconds := pa.Duration % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// SetMetadata sets JSON metadata for the attachment
func (pa *PostAttachment) SetMetadata(metadata map[string]interface{}) error {
	if len(metadata) == 0 {
		pa.Metadata = ""
		return nil
	}

	// Convert map to JSON string (simplified for demo)
	var jsonParts []string
	for key, value := range metadata {
		jsonParts = append(jsonParts, fmt.Sprintf(`"%s":"%v"`, key, value))
	}
	pa.Metadata = "{" + strings.Join(jsonParts, ",") + "}"

	return nil
}

// GetMetadata parses and returns the JSON metadata as a map
func (pa *PostAttachment) GetMetadata() map[string]interface{} {
	// Simplified JSON parsing for demo - in production use json.Unmarshal
	metadata := make(map[string]interface{})
	if pa.Metadata == "" {
		return metadata
	}

	// This is a simplified parser - in real implementation use proper JSON parsing
	metadata["raw"] = pa.Metadata
	return metadata
}

// NewPostAttachment creates a new PostAttachment with validation
func NewPostAttachment(postID uuid.UUID, fileName, originalName, filePath, contentType string, fileSize int64) (*PostAttachment, error) {
	attachment := &PostAttachment{
		ID:            uuid.New(),
		PostID:        postID,
		FileName:      strings.TrimSpace(fileName),
		OriginalName:  strings.TrimSpace(originalName),
		FileSize:      fileSize,
		ContentType:   strings.TrimSpace(contentType),
		FilePath:      strings.TrimSpace(filePath),
		IsPublic:      true,
		DownloadCount: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := attachment.DetermineType(); err != nil {
		return nil, err
	}

	if err := attachment.Validate(); err != nil {
		return nil, err
	}

	return attachment, nil
}

// TableName returns the database table name for GORM
func (PostAttachment) TableName() string {
	return "post_attachments"
}
