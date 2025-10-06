package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

// PostTypeUsecase handles business logic for post type operations
// Manages extensible post type definitions and field schemas
type PostTypeUsecase struct {
	postTypeRepo domain.PostTypeRepository
	postRepo     domain.PostRepository
}

// PostTypeUsecaseInput represents input for post type operations
type PostTypeUsecaseInput struct {
	Name             string                 `json:"name" validate:"required,min=1,max=100"`
	Description      string                 `json:"description" validate:"max=500"`
	Slug             string                 `json:"slug,omitempty" validate:"max=120"`
	IsActive         bool                   `json:"is_active"`
	FieldDefinitions map[string]interface{} `json:"field_definitions,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ExtensionType    string                 `json:"extension_type,omitempty" validate:"max=50"`
}

// PostTypeUsecaseOutput represents output from post type operations
type PostTypeUsecaseOutput struct {
	PostType *domain.PostType `json:"post_type"`
	Message  string           `json:"message,omitempty"`
}

// ListPostTypesInput represents input for listing post types
type ListPostTypesInput struct {
	IsActive      *bool   `json:"is_active,omitempty"`
	IsSystem      *bool   `json:"is_system,omitempty"`
	ExtensionType *string `json:"extension_type,omitempty"`
	Limit         int     `json:"limit" validate:"min=1,max=100"`
	Offset        int     `json:"offset" validate:"min=0"`
	SortBy        string  `json:"sort_by,omitempty"`
	SortOrder     string  `json:"sort_order,omitempty"`
}

// ListPostTypesOutput represents output from listing post types
type ListPostTypesOutput struct {
	PostTypes  []*domain.PostType `json:"post_types"`
	TotalCount int64              `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
}

// PostTypeStatsOutput represents post type statistics
type PostTypeStatsOutput struct {
	UsageStats      []*domain.PostTypeUsageStats      `json:"usage_stats,omitempty"`
	FieldUsageStats []*domain.PostTypeFieldUsageStats `json:"field_usage_stats,omitempty"`
}

// ValidateFieldDefinitionsInput represents input for field validation
type ValidateFieldDefinitionsInput struct {
	FieldDefinitions map[string]interface{} `json:"field_definitions" validate:"required"`
}

// ValidatePostDataInput represents input for post data validation
type ValidatePostDataInput struct {
	PostTypeID uuid.UUID              `json:"post_type_id" validate:"required"`
	PostData   map[string]interface{} `json:"post_data" validate:"required"`
}

// NewPostTypeUsecase creates a new PostTypeUsecase instance
func NewPostTypeUsecase(
	postTypeRepo domain.PostTypeRepository,
	postRepo domain.PostRepository,
) *PostTypeUsecase {
	return &PostTypeUsecase{
		postTypeRepo: postTypeRepo,
		postRepo:     postRepo,
	}
}

// CreatePostType creates a new post type with validation
func (uc *PostTypeUsecase) CreatePostType(ctx context.Context, input *PostTypeUsecaseInput) (*PostTypeUsecaseOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Check if name already exists
	exists, err := uc.postTypeRepo.ExistsByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check if post type name exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("post type with name '%s' already exists", input.Name)
	}

	// Create the post type entity
	postType := &domain.PostType{
		ID:                uuid.New(),
		Name:              input.Name,
		DisplayName:       input.Name, // Use name as display name initially
		Description:       input.Description,
		IsActive:          input.IsActive,
		RequiresApproval:  false, // Default values
		AllowsScheduling:  true,
		AllowsAttachments: true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Set display name if different from name
	if input.Name != "" {
		// Create a more readable display name
		postType.DisplayName = input.Name
	}

	// Set field definitions if provided
	if len(input.FieldDefinitions) > 0 {
		if err := postType.SetFieldDefinitions(input.FieldDefinitions); err != nil {
			return nil, fmt.Errorf("failed to set field definitions: %w", err)
		}
	}

	// Set metadata if provided (store as JSON in FieldDefinitions for now)
	if len(input.Metadata) > 0 {
		// For now, we'll merge metadata into field definitions
		// In a more complete implementation, PostType would have a Metadata field
		if len(input.FieldDefinitions) == 0 {
			input.FieldDefinitions = make(map[string]interface{})
		}
		input.FieldDefinitions["_metadata"] = input.Metadata
	}

	// Save the post type
	if err := uc.postTypeRepo.Create(ctx, postType); err != nil {
		return nil, fmt.Errorf("failed to create post type: %w", err)
	}

	return &PostTypeUsecaseOutput{
		PostType: postType,
		Message:  "Post type created successfully",
	}, nil
}

// GetPostType retrieves a post type by ID
func (uc *PostTypeUsecase) GetPostType(ctx context.Context, postTypeID uuid.UUID) (*PostTypeUsecaseOutput, error) {
	if postTypeID == uuid.Nil {
		return nil, errors.New("post type ID cannot be nil")
	}

	postType, err := uc.postTypeRepo.GetByID(ctx, postTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post type: %w", err)
	}
	if postType == nil {
		return nil, errors.New("post type not found")
	}

	return &PostTypeUsecaseOutput{
		PostType: postType,
		Message:  "Post type retrieved successfully",
	}, nil
}

// GetPostTypeBySlug retrieves a post type by slug
func (uc *PostTypeUsecase) GetPostTypeBySlug(ctx context.Context, slug string) (*PostTypeUsecaseOutput, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	postType, err := uc.postTypeRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get post type by slug: %w", err)
	}
	if postType == nil {
		return nil, errors.New("post type not found")
	}

	return &PostTypeUsecaseOutput{
		PostType: postType,
		Message:  "Post type retrieved successfully",
	}, nil
}

// UpdatePostType updates an existing post type
func (uc *PostTypeUsecase) UpdatePostType(ctx context.Context, postTypeID uuid.UUID, input *PostTypeUsecaseInput) (*PostTypeUsecaseOutput, error) {
	if postTypeID == uuid.Nil {
		return nil, errors.New("post type ID cannot be nil")
	}
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Get existing post type
	existingPostType, err := uc.postTypeRepo.GetByID(ctx, postTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing post type: %w", err)
	}
	if existingPostType == nil {
		return nil, errors.New("post type not found")
	}

	// Check if this is a system post type
	if existingPostType.IsSystemType() {
		return nil, fmt.Errorf("cannot modify system post type")
	}

	// Check if new name conflicts with existing names (excluding current)
	if input.Name != existingPostType.Name {
		exists, err := uc.postTypeRepo.ExistsByName(ctx, input.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check if post type name exists: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("post type with name '%s' already exists", input.Name)
		}
	}

	// Update post type fields
	existingPostType.Name = input.Name
	existingPostType.DisplayName = input.Name // Update display name too
	existingPostType.Description = input.Description
	existingPostType.IsActive = input.IsActive

	// Update field definitions if provided
	if len(input.FieldDefinitions) > 0 {
		if err := existingPostType.SetFieldDefinitions(input.FieldDefinitions); err != nil {
			return nil, fmt.Errorf("failed to update field definitions: %w", err)
		}
	}

	// Update metadata if provided (merge into field definitions)
	if len(input.Metadata) > 0 {
		fieldDefs, _ := existingPostType.GetFieldDefinitions()
		if fieldDefs == nil {
			fieldDefs = make(map[string]interface{})
		}
		fieldDefs["_metadata"] = input.Metadata
		if err := existingPostType.SetFieldDefinitions(fieldDefs); err != nil {
			return nil, fmt.Errorf("failed to update metadata: %w", err)
		}
	}

	// Save the updated post type
	if err := uc.postTypeRepo.Update(ctx, existingPostType); err != nil {
		return nil, fmt.Errorf("failed to update post type: %w", err)
	}

	return &PostTypeUsecaseOutput{
		PostType: existingPostType,
		Message:  "Post type updated successfully",
	}, nil
}

// DeletePostType deletes a post type
func (uc *PostTypeUsecase) DeletePostType(ctx context.Context, postTypeID uuid.UUID) error {
	if postTypeID == uuid.Nil {
		return errors.New("post type ID cannot be nil")
	}

	// Get the post type to verify it exists and can be deleted
	postType, err := uc.postTypeRepo.GetByID(ctx, postTypeID)
	if err != nil {
		return fmt.Errorf("failed to get post type: %w", err)
	}
	if postType == nil {
		return errors.New("post type not found")
	}

	// Check if this is a system post type
	if postType.IsSystemType() {
		return fmt.Errorf("cannot delete system post type")
	}

	// Check if there are posts using this type
	postCount, err := uc.postRepo.CountByType(ctx, postTypeID)
	if err != nil {
		return fmt.Errorf("failed to count posts for type: %w", err)
	}
	if postCount > 0 {
		return fmt.Errorf("cannot delete post type: %d posts are using this type", postCount)
	}

	// Delete the post type
	if err := uc.postTypeRepo.Delete(ctx, postTypeID); err != nil {
		return fmt.Errorf("failed to delete post type: %w", err)
	}

	return nil
}

// ListPostTypes lists post types with filtering and pagination
func (uc *PostTypeUsecase) ListPostTypes(ctx context.Context, input *ListPostTypesInput) (*ListPostTypesOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	// Build list options
	options := domain.ListPostTypesOptions{
		Limit:             input.Limit,
		Offset:            input.Offset,
		IsActive:          input.IsActive,
		IsSystem:          input.IsSystem,
		ExtensionType:     input.ExtensionType,
		SortBy:            input.SortBy,
		SortOrder:         input.SortOrder,
		IncludeUsageStats: true,
		IncludeMetadata:   true,
	}

	// Get post types
	postTypes, err := uc.postTypeRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list post types: %w", err)
	}

	// Get total count
	totalCount, err := uc.postTypeRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count post types: %w", err)
	}

	hasMore := input.Offset+input.Limit < int(totalCount)

	return &ListPostTypesOutput{
		PostTypes:  postTypes,
		TotalCount: totalCount,
		Limit:      input.Limit,
		Offset:     input.Offset,
		HasMore:    hasMore,
	}, nil
}

// GetSystemPostTypes retrieves all system post types
func (uc *PostTypeUsecase) GetSystemPostTypes(ctx context.Context) (*ListPostTypesOutput, error) {
	postTypes, err := uc.postTypeRepo.GetSystemTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system post types: %w", err)
	}

	return &ListPostTypesOutput{
		PostTypes:  postTypes,
		TotalCount: int64(len(postTypes)),
		Limit:      len(postTypes),
		Offset:     0,
		HasMore:    false,
	}, nil
}

// GetCustomPostTypes retrieves all custom post types
func (uc *PostTypeUsecase) GetCustomPostTypes(ctx context.Context) (*ListPostTypesOutput, error) {
	postTypes, err := uc.postTypeRepo.GetCustomTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom post types: %w", err)
	}

	return &ListPostTypesOutput{
		PostTypes:  postTypes,
		TotalCount: int64(len(postTypes)),
		Limit:      len(postTypes),
		Offset:     0,
		HasMore:    false,
	}, nil
}

// GetPostTypeStats gets usage statistics for a post type
func (uc *PostTypeUsecase) GetPostTypeStats(ctx context.Context, postTypeID uuid.UUID) (*PostTypeStatsOutput, error) {
	if postTypeID == uuid.Nil {
		return nil, errors.New("post type ID cannot be nil")
	}

	stats, err := uc.postTypeRepo.GetUsageStats(ctx, postTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post type stats: %w", err)
	}

	return &PostTypeStatsOutput{
		UsageStats: []*domain.PostTypeUsageStats{stats},
	}, nil
}

// GetAllPostTypeStats gets usage statistics for all post types
func (uc *PostTypeUsecase) GetAllPostTypeStats(ctx context.Context) (*PostTypeStatsOutput, error) {
	stats, err := uc.postTypeRepo.GetAllUsageStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all post type stats: %w", err)
	}

	return &PostTypeStatsOutput{
		UsageStats: stats,
	}, nil
}

// ValidateFieldDefinitions validates field definitions for a post type
func (uc *PostTypeUsecase) ValidateFieldDefinitions(ctx context.Context, input *ValidateFieldDefinitionsInput) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}

	if len(input.FieldDefinitions) == 0 {
		return errors.New("field definitions cannot be empty")
	}

	return uc.postTypeRepo.ValidateFieldDefinitions(ctx, input.FieldDefinitions)
}

// ValidatePostData validates post data against a post type's field definitions
func (uc *PostTypeUsecase) ValidatePostData(ctx context.Context, input *ValidatePostDataInput) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}

	if input.PostTypeID == uuid.Nil {
		return errors.New("post type ID cannot be nil")
	}

	if len(input.PostData) == 0 {
		return errors.New("post data cannot be empty")
	}

	return uc.postTypeRepo.ValidatePostData(ctx, input.PostTypeID, input.PostData)
}

// GetCompatiblePostTypes gets post types that are compatible with required fields
func (uc *PostTypeUsecase) GetCompatiblePostTypes(ctx context.Context, requiredFields []string) (*ListPostTypesOutput, error) {
	if len(requiredFields) == 0 {
		return nil, errors.New("required fields cannot be empty")
	}

	postTypes, err := uc.postTypeRepo.GetCompatibleTypes(ctx, requiredFields)
	if err != nil {
		return nil, fmt.Errorf("failed to get compatible post types: %w", err)
	}

	return &ListPostTypesOutput{
		PostTypes:  postTypes,
		TotalCount: int64(len(postTypes)),
		Limit:      len(postTypes),
		Offset:     0,
		HasMore:    false,
	}, nil
}

// SearchPostTypes searches post types with advanced filtering
func (uc *PostTypeUsecase) SearchPostTypes(ctx context.Context, query string, input *ListPostTypesInput) (*ListPostTypesOutput, error) {
	if query == "" {
		return uc.ListPostTypes(ctx, input)
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	// Build search options
	options := domain.SearchPostTypesOptions{
		Limit:               input.Limit,
		Offset:              input.Offset,
		SearchInName:        true,
		SearchInDescription: true,
		SearchInFields:      true,
		IsActive:            input.IsActive,
		IsSystem:            input.IsSystem,
		ExtensionType:       input.ExtensionType,
		SortBy:              "relevance",
		SortOrder:           "desc",
	}

	// Search post types
	postTypes, err := uc.postTypeRepo.Search(ctx, query, options)
	if err != nil {
		return nil, fmt.Errorf("failed to search post types: %w", err)
	}

	return &ListPostTypesOutput{
		PostTypes:  postTypes,
		TotalCount: int64(len(postTypes)), // TODO: Get actual count from search
		Limit:      input.Limit,
		Offset:     input.Offset,
		HasMore:    len(postTypes) == input.Limit,
	}, nil
}
