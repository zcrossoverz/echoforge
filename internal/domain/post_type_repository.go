package domain

import (
	"context"

	"github.com/google/uuid"
)

// PostTypeRepository defines the interface for post type data access operations
// Manages extensible post type definitions with field schemas and metadata
type PostTypeRepository interface {
	// Core CRUD Operations
	Create(ctx context.Context, postType *PostType) error
	GetByID(ctx context.Context, id uuid.UUID) (*PostType, error)
	Update(ctx context.Context, postType *PostType) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and Querying
	List(ctx context.Context, options ListPostTypesOptions) ([]*PostType, error)
	Count(ctx context.Context) (int64, error)

	// Name and Slug Operations
	GetByName(ctx context.Context, name string) (*PostType, error)
	GetBySlug(ctx context.Context, slug string) (*PostType, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// System Type Operations
	GetSystemTypes(ctx context.Context) ([]*PostType, error)
	GetCustomTypes(ctx context.Context) ([]*PostType, error)

	// Field Definition Operations
	GetFieldDefinitions(ctx context.Context, postTypeID uuid.UUID) (map[string]interface{}, error)
	UpdateFieldDefinitions(ctx context.Context, postTypeID uuid.UUID, fieldDefinitions map[string]interface{}) error

	// Usage Statistics
	GetUsageStats(ctx context.Context, postTypeID uuid.UUID) (*PostTypeUsageStats, error)
	GetAllUsageStats(ctx context.Context) ([]*PostTypeUsageStats, error)

	// Validation
	ValidateFieldDefinitions(ctx context.Context, fieldDefinitions map[string]interface{}) error
	ValidatePostData(ctx context.Context, postTypeID uuid.UUID, postData map[string]interface{}) error

	// Metadata Operations
	GetMetadata(ctx context.Context, postTypeID uuid.UUID) (map[string]interface{}, error)
	UpdateMetadata(ctx context.Context, postTypeID uuid.UUID, metadata map[string]interface{}) error

	// Bulk Operations
	BulkUpdateStatus(ctx context.Context, postTypeIDs []uuid.UUID, isActive bool) error
	BulkDelete(ctx context.Context, postTypeIDs []uuid.UUID) error

	// Search and Filtering
	Search(ctx context.Context, query string, options SearchPostTypesOptions) ([]*PostType, error)

	// Extension-specific Operations
	GetByExtensionType(ctx context.Context, extensionType string) ([]*PostType, error)
	GetCompatibleTypes(ctx context.Context, requiredFields []string) ([]*PostType, error)
}

// ListPostTypesOptions provides filtering and sorting options for listing post types
type ListPostTypesOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Filtering
	IsActive      *bool
	IsSystem      *bool
	ExtensionType *string

	// Sorting
	SortBy    string // "name", "slug", "created_at", "updated_at", "post_count"
	SortOrder string // "asc", "desc"

	// Inclusion
	IncludeUsageStats bool
	IncludeMetadata   bool
}

// SearchPostTypesOptions provides options for searching post types
type SearchPostTypesOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Search scope
	SearchInName        bool
	SearchInDescription bool
	SearchInFields      bool

	// Filtering
	IsActive      *bool
	IsSystem      *bool
	ExtensionType *string

	// Sorting
	SortBy    string // "relevance", "name", "created_at", "post_count"
	SortOrder string // "asc", "desc"
}

// PostTypeUsageStats represents usage statistics for a post type
type PostTypeUsageStats struct {
	PostTypeID     uuid.UUID `json:"post_type_id"`
	PostTypeName   string    `json:"post_type_name"`
	PostTypeSlug   string    `json:"post_type_slug"`
	TotalPosts     int64     `json:"total_posts"`
	PublishedPosts int64     `json:"published_posts"`
	DraftPosts     int64     `json:"draft_posts"`
	PendingPosts   int64     `json:"pending_posts"`
	ArchivedPosts  int64     `json:"archived_posts"`
	TotalViews     int64     `json:"total_views"`
	AverageViews   float64   `json:"average_views"`
	UniqueAuthors  int64     `json:"unique_authors"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

// PostTypeFieldUsageStats represents usage statistics for post type fields
type PostTypeFieldUsageStats struct {
	PostTypeID   uuid.UUID              `json:"post_type_id"`
	FieldName    string                 `json:"field_name"`
	FieldType    string                 `json:"field_type"`
	UsageCount   int64                  `json:"usage_count"`
	UsagePercent float64                `json:"usage_percent"`
	ValueStats   map[string]interface{} `json:"value_stats,omitempty"`
}

// PostTypeCategoryRepository defines the interface for post category data access operations
// Manages hierarchical category structures with parent-child relationships
type PostTypeCategoryRepository interface {
	// Core CRUD Operations
	Create(ctx context.Context, category *PostCategory) error
	GetByID(ctx context.Context, id uuid.UUID) (*PostCategory, error)
	Update(ctx context.Context, category *PostCategory) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and Querying
	List(ctx context.Context, options ListCategoriesOptions) ([]*PostCategory, error)
	Count(ctx context.Context) (int64, error)

	// Name and Slug Operations
	GetByName(ctx context.Context, name string) (*PostCategory, error)
	GetBySlug(ctx context.Context, slug string) (*PostCategory, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// Hierarchical Operations
	GetRootCategories(ctx context.Context) ([]*PostCategory, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*PostCategory, error)
	GetParent(ctx context.Context, categoryID uuid.UUID) (*PostCategory, error)
	GetAncestors(ctx context.Context, categoryID uuid.UUID) ([]*PostCategory, error)
	GetDescendants(ctx context.Context, categoryID uuid.UUID) ([]*PostCategory, error)
	GetSiblings(ctx context.Context, categoryID uuid.UUID) ([]*PostCategory, error)

	// System Category Operations
	GetSystemCategories(ctx context.Context) ([]*PostCategory, error)
	GetCustomCategories(ctx context.Context) ([]*PostCategory, error)

	// Usage Statistics
	GetUsageStats(ctx context.Context, categoryID uuid.UUID) (*CategoryUsageStats, error)
	GetAllUsageStats(ctx context.Context) ([]*CategoryUsageStats, error)

	// Post Assignment Operations
	GetPostCategories(ctx context.Context, postID uuid.UUID) ([]*PostCategory, error)
	AssignPostToCategory(ctx context.Context, postID, categoryID uuid.UUID) error
	UnassignPostFromCategory(ctx context.Context, postID, categoryID uuid.UUID) error
	ReassignPostCategories(ctx context.Context, postID uuid.UUID, categoryIDs []uuid.UUID) error

	// Bulk Operations
	BulkDelete(ctx context.Context, categoryIDs []uuid.UUID) error
	BulkUpdateParent(ctx context.Context, categoryIDs []uuid.UUID, parentID *uuid.UUID) error

	// Search and Filtering
	Search(ctx context.Context, query string, options SearchCategoriesOptions) ([]*PostCategory, error)

	// Tree Operations
	GetCategoryTree(ctx context.Context) ([]*CategoryTreeNode, error)
	ValidateHierarchy(ctx context.Context, categoryID, parentID uuid.UUID) error
}

// ListCategoriesOptions provides filtering and sorting options for listing categories
type ListCategoriesOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Filtering
	ParentID *uuid.UUID
	IsSystem *bool

	// Hierarchical filtering
	RootOnly     bool
	IncludeEmpty bool // Include categories with no posts

	// Sorting
	SortBy    string // "name", "slug", "created_at", "post_count", "hierarchy"
	SortOrder string // "asc", "desc"

	// Inclusion
	IncludeUsageStats bool
	IncludeChildren   bool
	IncludeParent     bool
}

// SearchCategoriesOptions provides options for searching categories
type SearchCategoriesOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Search scope
	SearchInName        bool
	SearchInDescription bool

	// Filtering
	ParentID *uuid.UUID
	IsSystem *bool

	// Sorting
	SortBy    string // "relevance", "name", "post_count"
	SortOrder string // "asc", "desc"
}

// CategoryUsageStats represents usage statistics for a category
type CategoryUsageStats struct {
	CategoryID       uuid.UUID `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	CategorySlug     string    `json:"category_slug"`
	PostCount        int64     `json:"post_count"`
	DirectPostCount  int64     `json:"direct_post_count"`
	SubcategoryCount int64     `json:"subcategory_count"`
	TotalViews       int64     `json:"total_views"`
	AverageViews     float64   `json:"average_views"`
	UniqueAuthors    int64     `json:"unique_authors"`
	Level            int       `json:"level"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

// CategoryTreeNode represents a node in the category tree
type CategoryTreeNode struct {
	Category  *PostCategory       `json:"category"`
	Children  []*CategoryTreeNode `json:"children,omitempty"`
	Level     int                 `json:"level"`
	PostCount int64               `json:"post_count"`
}
