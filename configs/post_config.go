package configs

import (
	"fmt"
	"time"
)

// PostConfig holds configuration for the post system
type PostConfig struct {
	// Content limits
	MaxContentSize      int64 `mapstructure:"max_content_size" yaml:"max_content_size"`
	MaxTitleLength      int   `mapstructure:"max_title_length" yaml:"max_title_length"`
	MaxAttachmentSize   int64 `mapstructure:"max_attachment_size" yaml:"max_attachment_size"`
	MaxAttachmentsCount int   `mapstructure:"max_attachments_count" yaml:"max_attachments_count"`

	// Version management
	MaxVersionsPerPost int `mapstructure:"max_versions_per_post" yaml:"max_versions_per_post"`

	// Scheduling
	MinScheduleDelay   time.Duration `mapstructure:"min_schedule_delay" yaml:"min_schedule_delay"`
	MaxScheduleAdvance time.Duration `mapstructure:"max_schedule_advance" yaml:"max_schedule_advance"`

	// File upload settings
	AllowedMimeTypes []string `mapstructure:"allowed_mime_types" yaml:"allowed_mime_types"`
	UploadPath       string   `mapstructure:"upload_path" yaml:"upload_path"`

	// Search and pagination
	DefaultPageSize  int `mapstructure:"default_page_size" yaml:"default_page_size"`
	MaxPageSize      int `mapstructure:"max_page_size" yaml:"max_page_size"`
	SearchMaxResults int `mapstructure:"search_max_results" yaml:"search_max_results"`

	// Performance settings
	ViewCountUpdateBatch int           `mapstructure:"view_count_batch" yaml:"view_count_batch"`
	CacheExpiration      time.Duration `mapstructure:"cache_expiration" yaml:"cache_expiration"`

	// Approval workflow
	RequireApprovalByDefault bool     `mapstructure:"require_approval_by_default" yaml:"require_approval_by_default"`
	AutoApproveAuthors       []string `mapstructure:"auto_approve_authors" yaml:"auto_approve_authors"`

	// Rate limiting (requests per hour)
	CreatePostRateLimit    int `mapstructure:"create_post_rate_limit" yaml:"create_post_rate_limit"`
	UpdatePostRateLimit    int `mapstructure:"update_post_rate_limit" yaml:"update_post_rate_limit"`
	UploadRateLimit        int `mapstructure:"upload_rate_limit" yaml:"upload_rate_limit"`
	BulkOperationRateLimit int `mapstructure:"bulk_operation_rate_limit" yaml:"bulk_operation_rate_limit"`
}

// DefaultPostConfig returns the default configuration for the post system
func DefaultPostConfig() *PostConfig {
	return &PostConfig{
		// Content limits
		MaxContentSize:      1024 * 1024, // 1MB
		MaxTitleLength:      255,
		MaxAttachmentSize:   100 * 1024 * 1024, // 100MB
		MaxAttachmentsCount: 10,

		// Version management
		MaxVersionsPerPost: 5,

		// Scheduling
		MinScheduleDelay:   time.Minute * 5,      // 5 minutes minimum
		MaxScheduleAdvance: time.Hour * 24 * 365, // 1 year maximum

		// File upload settings
		AllowedMimeTypes: []string{
			"image/jpeg", "image/png", "image/gif", "image/webp",
			"application/pdf", "text/plain", "text/markdown",
			"video/mp4", "video/webm", "audio/mpeg", "audio/wav",
		},
		UploadPath: "./uploads/posts",

		// Search and pagination
		DefaultPageSize:  20,
		MaxPageSize:      100,
		SearchMaxResults: 1000,

		// Performance settings
		ViewCountUpdateBatch: 10,
		CacheExpiration:      time.Hour,

		// Approval workflow
		RequireApprovalByDefault: false,
		AutoApproveAuthors:       []string{},

		// Rate limiting (per hour)
		CreatePostRateLimit:    100,
		UpdatePostRateLimit:    200,
		UploadRateLimit:        10,
		BulkOperationRateLimit: 5,
	}
}

// Validate checks if the post configuration is valid
func (c *PostConfig) Validate() error {
	if c.MaxContentSize <= 0 {
		return fmt.Errorf("max_content_size must be positive")
	}

	if c.MaxTitleLength <= 0 || c.MaxTitleLength > 500 {
		return fmt.Errorf("max_title_length must be between 1 and 500")
	}

	if c.MaxAttachmentSize <= 0 {
		return fmt.Errorf("max_attachment_size must be positive")
	}

	if c.MaxAttachmentsCount <= 0 {
		return fmt.Errorf("max_attachments_count must be positive")
	}

	if c.MaxVersionsPerPost < 1 || c.MaxVersionsPerPost > 50 {
		return fmt.Errorf("max_versions_per_post must be between 1 and 50")
	}

	if c.MinScheduleDelay < 0 {
		return fmt.Errorf("min_schedule_delay cannot be negative")
	}

	if c.MaxScheduleAdvance <= c.MinScheduleDelay {
		return fmt.Errorf("max_schedule_advance must be greater than min_schedule_delay")
	}

	if c.DefaultPageSize <= 0 || c.DefaultPageSize > c.MaxPageSize {
		return fmt.Errorf("default_page_size must be positive and not exceed max_page_size")
	}

	if c.MaxPageSize <= 0 || c.MaxPageSize > 1000 {
		return fmt.Errorf("max_page_size must be between 1 and 1000")
	}

	if len(c.UploadPath) == 0 {
		return fmt.Errorf("upload_path cannot be empty")
	}

	return nil
}
