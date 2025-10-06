package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostStatus represents the status of a post
type PostStatus string

const (
	PostStatusDraft           PostStatus = "draft"
	PostStatusScheduled       PostStatus = "scheduled"
	PostStatusPublished       PostStatus = "published"
	PostStatusArchived        PostStatus = "archived"
	PostStatusPendingApproval PostStatus = "pending_approval"
)

// ValidPostStatuses contains all valid post status values
var ValidPostStatuses = []PostStatus{
	PostStatusDraft,
	PostStatusScheduled,
	PostStatusPublished,
	PostStatusArchived,
	PostStatusPendingApproval,
}

// Post represents the core post entity
type Post struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string     `json:"title" gorm:"type:varchar(255);not null"`
	Content     string     `json:"content" gorm:"type:text;not null"`
	AuthorID    uuid.UUID  `json:"authorId" gorm:"type:uuid;not null;index"`
	PostTypeID  uuid.UUID  `json:"postTypeId" gorm:"type:uuid;not null;index"`
	Status      PostStatus `json:"status" gorm:"type:post_status;not null;default:'draft';index"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty" gorm:"index"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"not null;default:now()"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"not null;default:now()"`
	PublishedAt *time.Time `json:"publishedAt,omitempty" gorm:"index"`
	ViewCount   int        `json:"viewCount" gorm:"not null;default:0"`
	IsApproved  bool       `json:"isApproved" gorm:"not null;default:false"`
	ApprovedBy  *uuid.UUID `json:"approvedBy,omitempty" gorm:"type:uuid"`
	ApprovedAt  *time.Time `json:"approvedAt,omitempty"`

	// Relationships (loaded separately to avoid circular imports)
	Author      *User            `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
	PostType    *PostType        `json:"postType,omitempty" gorm:"foreignKey:PostTypeID"`
	Categories  []PostCategory   `json:"categories,omitempty" gorm:"many2many:post_category_assignments;"`
	Tags        []PostTag        `json:"tags,omitempty" gorm:"many2many:post_tag_assignments;"`
	Attachments []PostAttachment `json:"attachments,omitempty" gorm:"foreignKey:PostID"`
	Metadata    []PostMetadata   `json:"metadata,omitempty" gorm:"foreignKey:PostID"`
}

// TableName specifies the table name for GORM
func (Post) TableName() string {
	return "posts"
}

// Validate checks if the post entity is valid
func (p *Post) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(p.Title) > 255 {
		return fmt.Errorf("title cannot exceed 255 characters")
	}

	if p.Content == "" {
		return fmt.Errorf("content is required")
	}

	if len(p.Content) > 1024*1024 { // 1MB limit
		return fmt.Errorf("content cannot exceed 1MB")
	}

	if p.AuthorID == uuid.Nil {
		return fmt.Errorf("author ID is required")
	}

	if p.PostTypeID == uuid.Nil {
		return fmt.Errorf("post type ID is required")
	}

	if !p.isValidStatus() {
		return fmt.Errorf("invalid status: %s", p.Status)
	}

	// Validate scheduled post requirements
	if p.Status == PostStatusScheduled {
		if p.ScheduledAt == nil {
			return fmt.Errorf("scheduled posts must have a scheduled time")
		}
		if p.ScheduledAt.Before(time.Now()) {
			return fmt.Errorf("scheduled time must be in the future")
		}
		// Minimum 5 minutes from now (configurable)
		minScheduleTime := time.Now().Add(5 * time.Minute)
		if p.ScheduledAt.Before(minScheduleTime) {
			return fmt.Errorf("scheduled time must be at least 5 minutes from now")
		}
	}

	return nil
}

// isValidStatus checks if the post status is valid
func (p *Post) isValidStatus() bool {
	for _, validStatus := range ValidPostStatuses {
		if p.Status == validStatus {
			return true
		}
	}
	return false
}

// TransitionTo attempts to transition the post to a new status
func (p *Post) TransitionTo(newStatus PostStatus) error {
	if !p.isValidTransition(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", p.Status, newStatus)
	}

	oldStatus := p.Status
	p.Status = newStatus
	p.UpdatedAt = time.Now()

	// Set published timestamp when transitioning to published
	if newStatus == PostStatusPublished && oldStatus != PostStatusPublished {
		now := time.Now()
		p.PublishedAt = &now
	}

	return nil
}

// isValidTransition checks if a status transition is allowed
func (p *Post) isValidTransition(newStatus PostStatus) bool {
	validTransitions := map[PostStatus][]PostStatus{
		PostStatusDraft:           {PostStatusPublished, PostStatusScheduled, PostStatusArchived, PostStatusPendingApproval},
		PostStatusScheduled:       {PostStatusPublished, PostStatusDraft, PostStatusArchived},
		PostStatusPublished:       {PostStatusArchived},
		PostStatusArchived:        {PostStatusDraft},
		PostStatusPendingApproval: {PostStatusPublished, PostStatusDraft, PostStatusArchived},
	}

	allowedStatuses, exists := validTransitions[p.Status]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedStatuses {
		if allowedStatus == newStatus {
			return true
		}
	}

	return false
}

// Approve approves the post by setting approval fields
func (p *Post) Approve(approverID uuid.UUID) error {
	if approverID == uuid.Nil {
		return fmt.Errorf("approver ID is required")
	}

	p.IsApproved = true
	p.ApprovedBy = &approverID
	now := time.Now()
	p.ApprovedAt = &now
	p.UpdatedAt = time.Now()

	return nil
}

// IncrementViewCount increases the view count by 1
func (p *Post) IncrementViewCount() {
	p.ViewCount++
	// Note: UpdatedAt is not changed for view count increments
}

// IsPublished checks if the post is currently published
func (p *Post) IsPublished() bool {
	return p.Status == PostStatusPublished
}

// IsScheduled checks if the post is scheduled for future publication
func (p *Post) IsScheduled() bool {
	return p.Status == PostStatusScheduled
}

// IsDraft checks if the post is in draft status
func (p *Post) IsDraft() bool {
	return p.Status == PostStatusDraft
}

// IsArchived checks if the post is archived
func (p *Post) IsArchived() bool {
	return p.Status == PostStatusArchived
}

// RequiresApproval checks if the post is pending approval
func (p *Post) RequiresApproval() bool {
	return p.Status == PostStatusPendingApproval
}

// CanBePublished checks if the post can be published
func (p *Post) CanBePublished() bool {
	// Can only publish from draft, scheduled, or pending approval
	return p.Status == PostStatusDraft ||
		p.Status == PostStatusScheduled ||
		p.Status == PostStatusPendingApproval
}

// GetTruncatedContent returns truncated content for listings
func (p *Post) GetTruncatedContent(maxLength int) string {
	if len(p.Content) <= maxLength {
		return p.Content
	}
	return p.Content[:maxLength] + "..."
}

// BeforeCreate GORM hook called before creating a record
func (p *Post) BeforeCreate() error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate GORM hook called before updating a record
func (p *Post) BeforeUpdate() error {
	p.UpdatedAt = time.Now()
	return nil
}
