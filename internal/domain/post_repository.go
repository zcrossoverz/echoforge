package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostRepository defines the interface for post data access operations
// Following Repository pattern with domain-driven design principles
type PostRepository interface {
	// Core CRUD Operations
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and Pagination
	List(ctx context.Context, options ListPostsOptions) ([]*Post, error)
	Count(ctx context.Context, options CountPostsOptions) (int64, error)

	// Status-based Operations
	GetByStatus(ctx context.Context, status PostStatus, limit, offset int) ([]*Post, error)
	GetPublished(ctx context.Context, limit, offset int) ([]*Post, error)
	GetScheduled(ctx context.Context, before time.Time) ([]*Post, error)
	GetPendingApproval(ctx context.Context, limit, offset int) ([]*Post, error)

	// Author-based Operations
	GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*Post, error)
	CountByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error)

	// Type-based Operations
	GetByType(ctx context.Context, postTypeID uuid.UUID, limit, offset int) ([]*Post, error)
	CountByType(ctx context.Context, postTypeID uuid.UUID) (int64, error)

	// Category-based Operations
	GetByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*Post, error)
	GetByCategorySlug(ctx context.Context, categorySlug string, limit, offset int) ([]*Post, error)

	// Tag-based Operations
	GetByTag(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]*Post, error)
	GetByTagSlug(ctx context.Context, tagSlug string, limit, offset int) ([]*Post, error)

	// Search and Filtering
	Search(ctx context.Context, query string, options SearchPostsOptions) ([]*Post, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*Post, error)

	// View Tracking
	IncrementViewCount(ctx context.Context, postID uuid.UUID) error
	GetMostViewed(ctx context.Context, limit int, since time.Time) ([]*Post, error)

	// Approval Workflow
	Approve(ctx context.Context, postID, approverID uuid.UUID) error
	Reject(ctx context.Context, postID, approverID uuid.UUID, reason string) error

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, postIDs []uuid.UUID, status PostStatus) error
	BulkDelete(ctx context.Context, postIDs []uuid.UUID) error

	// Relationship Management
	AddCategory(ctx context.Context, postID, categoryID uuid.UUID) error
	RemoveCategory(ctx context.Context, postID, categoryID uuid.UUID) error
	AddTag(ctx context.Context, postID, tagID uuid.UUID) error
	RemoveTag(ctx context.Context, postID, tagID uuid.UUID) error

	// Statistics and Analytics
	GetStatsByAuthor(ctx context.Context, authorID uuid.UUID) (*AuthorStats, error)
	GetStatsByType(ctx context.Context, postTypeID uuid.UUID) (*TypeStats, error)
	GetActivityStats(ctx context.Context, since time.Time) (*ActivityStats, error)
}

// ListPostsOptions provides filtering and sorting options for listing posts
type ListPostsOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID
	TagID      *uuid.UUID

	// Date Filtering
	PublishedAfter  *time.Time
	PublishedBefore *time.Time
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time

	// Sorting
	SortBy    string // "created_at", "updated_at", "published_at", "title", "view_count"
	SortOrder string // "asc", "desc"

	// Inclusion
	IncludeCategories bool
	IncludeTags       bool
	IncludeAuthor     bool
	IncludePostType   bool
}

// CountPostsOptions provides filtering options for counting posts
type CountPostsOptions struct {
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID
	TagID      *uuid.UUID

	// Date Filtering
	PublishedAfter  *time.Time
	PublishedBefore *time.Time
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
}

// SearchPostsOptions provides options for searching posts
type SearchPostsOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Search scope
	SearchInTitle   bool
	SearchInContent bool
	SearchInExcerpt bool

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID
	TagID      *uuid.UUID

	// Date Filtering
	PublishedAfter  *time.Time
	PublishedBefore *time.Time

	// Sorting
	SortBy    string // "relevance", "created_at", "updated_at", "published_at", "view_count"
	SortOrder string // "asc", "desc"

	// Inclusion
	IncludeCategories bool
	IncludeTags       bool
	IncludeAuthor     bool
	IncludePostType   bool
}

// AuthorStats represents statistics for an author
type AuthorStats struct {
	AuthorID       uuid.UUID  `json:"author_id"`
	TotalPosts     int64      `json:"total_posts"`
	PublishedPosts int64      `json:"published_posts"`
	DraftPosts     int64      `json:"draft_posts"`
	PendingPosts   int64      `json:"pending_posts"`
	TotalViews     int64      `json:"total_views"`
	AverageViews   float64    `json:"average_views"`
	FirstPostAt    *time.Time `json:"first_post_at,omitempty"`
	LastPostAt     *time.Time `json:"last_post_at,omitempty"`
}

// TypeStats represents statistics for a post type
type TypeStats struct {
	PostTypeID     uuid.UUID  `json:"post_type_id"`
	TotalPosts     int64      `json:"total_posts"`
	PublishedPosts int64      `json:"published_posts"`
	DraftPosts     int64      `json:"draft_posts"`
	PendingPosts   int64      `json:"pending_posts"`
	TotalViews     int64      `json:"total_views"`
	AverageViews   float64    `json:"average_views"`
	UniqueAuthors  int64      `json:"unique_authors"`
	FirstPostAt    *time.Time `json:"first_post_at,omitempty"`
	LastPostAt     *time.Time `json:"last_post_at,omitempty"`
}

// ActivityStats represents overall activity statistics
type ActivityStats struct {
	Period         string    `json:"period"` // "24h", "7d", "30d", etc.
	TotalPosts     int64     `json:"total_posts"`
	PublishedPosts int64     `json:"published_posts"`
	DraftPosts     int64     `json:"draft_posts"`
	PendingPosts   int64     `json:"pending_posts"`
	DeletedPosts   int64     `json:"deleted_posts"`
	TotalViews     int64     `json:"total_views"`
	UniqueAuthors  int64     `json:"unique_authors"`
	PopularTags    []string  `json:"popular_tags"`
	TopCategories  []string  `json:"top_categories"`
	CreatedAt      time.Time `json:"created_at"`
}
