package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

// PostUsecase handles business logic for post operations
// Coordinates between domain entities and repository interfaces
type PostUsecase struct {
	postRepo       domain.PostRepository
	postTypeRepo   domain.PostTypeRepository
	categoryRepo   domain.PostTypeCategoryRepository
	tagRepo        domain.PostRepository // Will be extracted to TagRepository later
	attachmentRepo domain.PostRepository // Will be extracted to AttachmentRepository later
	versionRepo    domain.PostRepository // Will be extracted to VersionRepository later
	metadataRepo   domain.PostRepository // Will be extracted to MetadataRepository later
}

// PostUsecaseInput represents input for post operations
type PostUsecaseInput struct {
	Title       string                 `json:"title" validate:"required,min=1,max=255"`
	Content     string                 `json:"content" validate:"required"`
	AuthorID    uuid.UUID              `json:"author_id" validate:"required"`
	PostTypeID  uuid.UUID              `json:"post_type_id" validate:"required"`
	Status      domain.PostStatus      `json:"status,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	CategoryIDs []uuid.UUID            `json:"category_ids,omitempty"`
	TagIDs      []uuid.UUID            `json:"tag_ids,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PostUsecaseOutput represents output from post operations
type PostUsecaseOutput struct {
	Post       *domain.Post           `json:"post"`
	Categories []*domain.PostCategory `json:"categories,omitempty"`
	Tags       []*domain.PostTag      `json:"tags,omitempty"`
	PostType   *domain.PostType       `json:"post_type,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

// ListPostsInput represents input for listing posts
type ListPostsInput struct {
	Status     *domain.PostStatus `json:"status,omitempty"`
	AuthorID   *uuid.UUID         `json:"author_id,omitempty"`
	PostTypeID *uuid.UUID         `json:"post_type_id,omitempty"`
	CategoryID *uuid.UUID         `json:"category_id,omitempty"`
	TagID      *uuid.UUID         `json:"tag_id,omitempty"`
	Limit      int                `json:"limit" validate:"min=1,max=100"`
	Offset     int                `json:"offset" validate:"min=0"`
	SortBy     string             `json:"sort_by,omitempty"`
	SortOrder  string             `json:"sort_order,omitempty"`
}

// ListPostsOutput represents output from listing posts
type ListPostsOutput struct {
	Posts      []*domain.Post `json:"posts"`
	TotalCount int64          `json:"total_count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	HasMore    bool           `json:"has_more"`
}

// PostStatsOutput represents post statistics
type PostStatsOutput struct {
	AuthorStats   *domain.AuthorStats   `json:"author_stats,omitempty"`
	TypeStats     *domain.TypeStats     `json:"type_stats,omitempty"`
	ActivityStats *domain.ActivityStats `json:"activity_stats,omitempty"`
}

// NewPostUsecase creates a new PostUsecase instance
func NewPostUsecase(
	postRepo domain.PostRepository,
	postTypeRepo domain.PostTypeRepository,
	categoryRepo domain.PostTypeCategoryRepository,
) *PostUsecase {
	return &PostUsecase{
		postRepo:     postRepo,
		postTypeRepo: postTypeRepo,
		categoryRepo: categoryRepo,
		// TODO: Inject specific repositories when they're implemented
		tagRepo:        postRepo,
		attachmentRepo: postRepo,
		versionRepo:    postRepo,
		metadataRepo:   postRepo,
	}
}

// CreatePost creates a new post with validation and business logic
func (uc *PostUsecase) CreatePost(ctx context.Context, input *PostUsecaseInput) (*PostUsecaseOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Validate post type exists
	postType, err := uc.postTypeRepo.GetByID(ctx, input.PostTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post type: %w", err)
	}
	if postType == nil {
		return nil, errors.New("post type not found")
	}

	// Create the post entity
	post := &domain.Post{
		ID:         uuid.New(),
		Title:      input.Title,
		Content:    input.Content,
		AuthorID:   input.AuthorID,
		PostTypeID: input.PostTypeID,
		Status:     domain.PostStatusDraft, // Default status
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Set status if provided
	if input.Status != "" {
		if err := post.TransitionTo(input.Status); err != nil {
			return nil, fmt.Errorf("failed to set post status: %w", err)
		}
	}

	// Set scheduled time if provided
	if input.ScheduledAt != nil {
		post.ScheduledAt = input.ScheduledAt
		if err := post.TransitionTo(domain.PostStatusScheduled); err != nil {
			return nil, fmt.Errorf("failed to schedule post: %w", err)
		}
	}

	// Validate the post
	if err := post.Validate(); err != nil {
		return nil, fmt.Errorf("post validation failed: %w", err)
	}

	// Save the post
	if err := uc.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Handle categories if provided
	var categories []*domain.PostCategory
	if len(input.CategoryIDs) > 0 {
		for _, categoryID := range input.CategoryIDs {
			if err := uc.postRepo.AddCategory(ctx, post.ID, categoryID); err != nil {
				return nil, fmt.Errorf("failed to add category %s: %w", categoryID, err)
			}
		}
		// Fetch category details for response
		for _, categoryID := range input.CategoryIDs {
			category, err := uc.categoryRepo.GetByID(ctx, categoryID)
			if err == nil && category != nil {
				categories = append(categories, category)
			}
		}
	}

	// Handle tags if provided
	var tags []*domain.PostTag
	if len(input.TagIDs) > 0 {
		for _, tagID := range input.TagIDs {
			if err := uc.postRepo.AddTag(ctx, post.ID, tagID); err != nil {
				return nil, fmt.Errorf("failed to add tag %s: %w", tagID, err)
			}
		}
		// TODO: Fetch tag details when TagRepository is implemented
	}

	// Create initial version
	if err := uc.createInitialVersion(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create initial version: %w", err)
	}

	return &PostUsecaseOutput{
		Post:       post,
		Categories: categories,
		Tags:       tags,
		PostType:   postType,
		Message:    "Post created successfully",
	}, nil
}

// GetPost retrieves a post by ID
func (uc *PostUsecase) GetPost(ctx context.Context, postID uuid.UUID) (*PostUsecaseOutput, error) {
	if postID == uuid.Nil {
		return nil, errors.New("post ID cannot be nil")
	}

	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Increment view count
	if err := uc.postRepo.IncrementViewCount(ctx, postID); err != nil {
		// Log error but don't fail the request
		// In production, this should use proper logging
		fmt.Printf("Warning: failed to increment view count for post %s: %v\n", postID, err)
	}

	// Get post type
	var postType *domain.PostType
	if post.PostTypeID != uuid.Nil {
		postType, _ = uc.postTypeRepo.GetByID(ctx, post.PostTypeID)
	}

	return &PostUsecaseOutput{
		Post:     post,
		PostType: postType,
		Message:  "Post retrieved successfully",
	}, nil
}

// UpdatePost updates an existing post
func (uc *PostUsecase) UpdatePost(ctx context.Context, postID uuid.UUID, input *PostUsecaseInput) (*PostUsecaseOutput, error) {
	if postID == uuid.Nil {
		return nil, errors.New("post ID cannot be nil")
	}
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Get existing post
	existingPost, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing post: %w", err)
	}
	if existingPost == nil {
		return nil, errors.New("post not found")
	}

	// Check if user can modify this post
	if existingPost.AuthorID != input.AuthorID {
		return nil, errors.New("user not authorized to modify this post")
	}

	// Create version before updating
	if err := uc.createVersionBeforeUpdate(ctx, existingPost, input.AuthorID); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Update post fields
	existingPost.Title = input.Title
	existingPost.Content = input.Content
	existingPost.UpdatedAt = time.Now()

	// Update status if provided
	if input.Status != "" && input.Status != existingPost.Status {
		if err := existingPost.TransitionTo(input.Status); err != nil {
			return nil, fmt.Errorf("failed to update post status: %w", err)
		}
	}

	// Update scheduled time if provided
	if input.ScheduledAt != nil {
		existingPost.ScheduledAt = input.ScheduledAt
		if err := existingPost.TransitionTo(domain.PostStatusScheduled); err != nil {
			return nil, fmt.Errorf("failed to update scheduled time: %w", err)
		}
	}

	// Save the updated post
	if err := uc.postRepo.Update(ctx, existingPost); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Get post type
	var postType *domain.PostType
	if existingPost.PostTypeID != uuid.Nil {
		postType, _ = uc.postTypeRepo.GetByID(ctx, existingPost.PostTypeID)
	}

	return &PostUsecaseOutput{
		Post:     existingPost,
		PostType: postType,
		Message:  "Post updated successfully",
	}, nil
}

// DeletePost deletes a post
func (uc *PostUsecase) DeletePost(ctx context.Context, postID, userID uuid.UUID) error {
	if postID == uuid.Nil {
		return errors.New("post ID cannot be nil")
	}
	if userID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}

	// Get the post to verify ownership
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return errors.New("post not found")
	}

	// Check if user can delete this post
	if post.AuthorID != userID {
		return errors.New("user not authorized to delete this post")
	}

	// Create version before deletion
	if err := uc.createVersionBeforeDelete(ctx, post, userID); err != nil {
		return fmt.Errorf("failed to create deletion version: %w", err)
	}

	// Soft delete by setting status to archived
	if err := post.TransitionTo(domain.PostStatusArchived); err != nil {
		return fmt.Errorf("failed to archive post: %w", err)
	}

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post status: %w", err)
	}

	return nil
}

// ListPosts lists posts with filtering and pagination
func (uc *PostUsecase) ListPosts(ctx context.Context, input *ListPostsInput) (*ListPostsOutput, error) {
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
	options := domain.ListPostsOptions{
		Limit:             input.Limit,
		Offset:            input.Offset,
		Status:            input.Status,
		AuthorID:          input.AuthorID,
		PostTypeID:        input.PostTypeID,
		CategoryID:        input.CategoryID,
		TagID:             input.TagID,
		SortBy:            input.SortBy,
		SortOrder:         input.SortOrder,
		IncludeCategories: true,
		IncludeTags:       true,
		IncludeAuthor:     true,
		IncludePostType:   true,
	}

	// Get posts
	posts, err := uc.postRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	// Get total count
	countOptions := domain.CountPostsOptions{
		Status:     input.Status,
		AuthorID:   input.AuthorID,
		PostTypeID: input.PostTypeID,
		CategoryID: input.CategoryID,
		TagID:      input.TagID,
	}

	totalCount, err := uc.postRepo.Count(ctx, countOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	hasMore := input.Offset+input.Limit < int(totalCount)

	return &ListPostsOutput{
		Posts:      posts,
		TotalCount: totalCount,
		Limit:      input.Limit,
		Offset:     input.Offset,
		HasMore:    hasMore,
	}, nil
}

// PublishPost publishes a post
func (uc *PostUsecase) PublishPost(ctx context.Context, postID, userID uuid.UUID) (*PostUsecaseOutput, error) {
	if postID == uuid.Nil {
		return nil, errors.New("post ID cannot be nil")
	}

	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Check authorization
	if post.AuthorID != userID {
		return nil, errors.New("user not authorized to publish this post")
	}

	// Create version before publishing
	if err := uc.createVersionBeforePublish(ctx, post, userID); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Publish the post
	if err := post.TransitionTo(domain.PostStatusPublished); err != nil {
		return nil, fmt.Errorf("failed to publish post: %w", err)
	}

	// Update in repository
	if err := uc.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return &PostUsecaseOutput{
		Post:    post,
		Message: "Post published successfully",
	}, nil
}

// ApprovePost approves a post (for approval workflow)
func (uc *PostUsecase) ApprovePost(ctx context.Context, postID, approverID uuid.UUID) (*PostUsecaseOutput, error) {
	if postID == uuid.Nil {
		return nil, errors.New("post ID cannot be nil")
	}
	if approverID == uuid.Nil {
		return nil, errors.New("approver ID cannot be nil")
	}

	// Use repository's approve method
	if err := uc.postRepo.Approve(ctx, postID, approverID); err != nil {
		return nil, fmt.Errorf("failed to approve post: %w", err)
	}

	// Get updated post
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved post: %w", err)
	}

	return &PostUsecaseOutput{
		Post:    post,
		Message: "Post approved successfully",
	}, nil
}

// GetAuthorStats gets statistics for an author
func (uc *PostUsecase) GetAuthorStats(ctx context.Context, authorID uuid.UUID) (*PostStatsOutput, error) {
	if authorID == uuid.Nil {
		return nil, errors.New("author ID cannot be nil")
	}

	stats, err := uc.postRepo.GetStatsByAuthor(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get author stats: %w", err)
	}

	return &PostStatsOutput{
		AuthorStats: stats,
	}, nil
}

// GetTypeStats gets statistics for a post type
func (uc *PostUsecase) GetTypeStats(ctx context.Context, postTypeID uuid.UUID) (*PostStatsOutput, error) {
	if postTypeID == uuid.Nil {
		return nil, errors.New("post type ID cannot be nil")
	}

	stats, err := uc.postRepo.GetStatsByType(ctx, postTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}

	return &PostStatsOutput{
		TypeStats: stats,
	}, nil
}

// GetActivityStats gets overall activity statistics
func (uc *PostUsecase) GetActivityStats(ctx context.Context, since time.Time) (*PostStatsOutput, error) {
	stats, err := uc.postRepo.GetActivityStats(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity stats: %w", err)
	}

	return &PostStatsOutput{
		ActivityStats: stats,
	}, nil
}

// Helper methods for version management

func (uc *PostUsecase) createInitialVersion(ctx context.Context, post *domain.Post) error {
	// TODO: Implement version creation when PostVersion repository is available
	// This is a placeholder for the initial version creation
	return nil
}

func (uc *PostUsecase) createVersionBeforeUpdate(ctx context.Context, post *domain.Post, editorID uuid.UUID) error {
	// TODO: Implement version creation before update
	return nil
}

func (uc *PostUsecase) createVersionBeforeDelete(ctx context.Context, post *domain.Post, editorID uuid.UUID) error {
	// TODO: Implement version creation before deletion
	return nil
}

func (uc *PostUsecase) createVersionBeforePublish(ctx context.Context, post *domain.Post, editorID uuid.UUID) error {
	// TODO: Implement version creation before publishing
	return nil
}
