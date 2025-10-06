package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// PostHandler handles HTTP requests for post management
type PostHandler struct {
	postUsecase *usecase.PostUsecase
}

// NewPostHandler creates a new post handler
func NewPostHandler(postUsecase *usecase.PostUsecase) *PostHandler {
	return &PostHandler{
		postUsecase: postUsecase,
	}
}

// CreatePostRequest represents the request body for creating a post
type CreatePostRequest struct {
	Title       string                 `json:"title" binding:"required,max=255"`
	Content     string                 `json:"content" binding:"required"`
	PostTypeID  uuid.UUID              `json:"postTypeId" binding:"required"`
	Status      string                 `json:"status,omitempty"`
	ScheduledAt *string                `json:"scheduledAt,omitempty"`
	CategoryIDs []uuid.UUID            `json:"categoryIds,omitempty"`
	TagIDs      []uuid.UUID            `json:"tagIds,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePostRequest represents the request body for updating a post
type UpdatePostRequest struct {
	Title       string                 `json:"title" binding:"required,max=255"`
	Content     string                 `json:"content" binding:"required"`
	PostTypeID  uuid.UUID              `json:"postTypeId" binding:"required"`
	Status      string                 `json:"status,omitempty"`
	ScheduledAt *string                `json:"scheduledAt,omitempty"`
	CategoryIDs []uuid.UUID            `json:"categoryIds,omitempty"`
	TagIDs      []uuid.UUID            `json:"tagIds,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PostResponse represents the response structure for a post
type PostResponse struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	AuthorID    uuid.UUID              `json:"authorId"`
	Author      *AuthorResponse        `json:"author,omitempty"`
	PostType    *PostTypeResponse      `json:"postType,omitempty"`
	Status      string                 `json:"status"`
	ScheduledAt *string                `json:"scheduledAt"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
	PublishedAt *string                `json:"publishedAt"`
	ViewCount   int64                  `json:"viewCount"`
	IsApproved  bool                   `json:"isApproved"`
	Categories  []CategoryResponse     `json:"categories,omitempty"`
	Tags        []TagResponse          `json:"tags,omitempty"`
	Attachments []AttachmentResponse   `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PostSummaryResponse represents a summarized post for list responses
type PostSummaryResponse struct {
	ID            uuid.UUID         `json:"id"`
	Title         string            `json:"title"`
	Content       string            `json:"content"` // Truncated
	AuthorID      uuid.UUID         `json:"authorId"`
	Author        *AuthorResponse   `json:"author,omitempty"`
	PostType      *PostTypeResponse `json:"postType,omitempty"`
	Status        string            `json:"status"`
	CreatedAt     string            `json:"createdAt"`
	PublishedAt   *string           `json:"publishedAt"`
	ViewCount     int64             `json:"viewCount"`
	CategoryCount int               `json:"categoryCount"`
	TagCount      int               `json:"tagCount"`
}

// AuthorResponse represents an author in responses
type AuthorResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// PostTypeResponse represents a post type in responses
type PostTypeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
}

// CategoryResponse represents a category in responses
type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// TagResponse represents a tag in responses
type TagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// AttachmentResponse represents an attachment in responses
type AttachmentResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	FileURL  string    `json:"fileUrl"`
	FileSize int64     `json:"fileSize"`
	MimeType string    `json:"mimeType"`
}

// ListPostsResponse represents the response for listing posts
type ListPostsResponse struct {
	Posts      []PostSummaryResponse  `json:"posts"`
	Pagination PaginationResponse     `json:"pagination"`
	Filters    map[string]interface{} `json:"filters"`
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// CreatePost handles POST /api/v1/posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get author ID from JWT context (assuming middleware sets this)
	authorIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	authorID, err := uuid.Parse(authorIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid author ID",
		})
		return
	}

	// Create usecase input
	input := &usecase.PostUsecaseInput{
		Title:       req.Title,
		Content:     req.Content,
		AuthorID:    authorID,
		PostTypeID:  req.PostTypeID,
		CategoryIDs: req.CategoryIDs,
		TagIDs:      req.TagIDs,
		Metadata:    req.Metadata,
	}

	// Set status
	if req.Status != "" {
		switch strings.ToLower(req.Status) {
		case "draft":
			input.Status = domain.PostStatusDraft
		case "scheduled":
			input.Status = domain.PostStatusScheduled
		case "published":
			input.Status = domain.PostStatusPublished
		default:
			input.Status = domain.PostStatusDraft
		}
	} else {
		input.Status = domain.PostStatusDraft
	}

	// Parse scheduled time if provided
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		scheduledAt, err := parseISO8601(*req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid scheduledAt format",
				"details": "Expected ISO 8601 format",
			})
			return
		}
		input.ScheduledAt = &scheduledAt
	}

	// Create post using use case
	output, err := h.postUsecase.CreatePost(c.Request.Context(), input)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "post type not found"):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "Invalid post type",
			})
		case strings.Contains(err.Error(), "unauthorized"):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create post",
			})
		}
		return
	}

	// Convert to response
	response := h.convertUsecaseOutputToResponse(output)

	c.JSON(http.StatusCreated, response)
}

// GetPost handles GET /api/v1/posts/{id}
func (h *PostHandler) GetPost(c *gin.Context) {
	idParam := c.Param("id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid post ID",
		})
		return
	}

	// Get post using use case (no user ID needed for Get)
	output, err := h.postUsecase.GetPost(c.Request.Context(), postID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
		case strings.Contains(err.Error(), "access denied"):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve post",
			})
		}
		return
	}

	// Convert to response
	response := h.convertUsecaseOutputToResponse(output)

	c.JSON(http.StatusOK, response)
}

// UpdatePost handles PUT /api/v1/posts/{id}
func (h *PostHandler) UpdatePost(c *gin.Context) {
	idParam := c.Param("id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid post ID",
		})
		return
	}

	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get author ID from JWT context
	authorIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	authorID, err := uuid.Parse(authorIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid author ID",
		})
		return
	}

	// Create usecase input
	input := &usecase.PostUsecaseInput{
		Title:       req.Title,
		Content:     req.Content,
		AuthorID:    authorID,
		PostTypeID:  req.PostTypeID,
		CategoryIDs: req.CategoryIDs,
		TagIDs:      req.TagIDs,
		Metadata:    req.Metadata,
	}

	// Set status
	if req.Status != "" {
		switch strings.ToLower(req.Status) {
		case "draft":
			input.Status = domain.PostStatusDraft
		case "scheduled":
			input.Status = domain.PostStatusScheduled
		case "published":
			input.Status = domain.PostStatusPublished
		default:
			input.Status = domain.PostStatusDraft
		}
	}

	// Parse scheduled time if provided
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		scheduledAt, err := parseISO8601(*req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid scheduledAt format",
				"details": "Expected ISO 8601 format",
			})
			return
		}
		input.ScheduledAt = &scheduledAt
	}

	// Update post using use case
	output, err := h.postUsecase.UpdatePost(c.Request.Context(), postID, input)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
		case strings.Contains(err.Error(), "not authorized"):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not post author or insufficient permissions",
			})
		case strings.Contains(err.Error(), "concurrent modification"):
			c.JSON(http.StatusConflict, gin.H{
				"error": "Concurrent modification conflict",
			})
		case strings.Contains(err.Error(), "post type not found"):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "Invalid post type",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update post",
			})
		}
		return
	}

	// Convert to response
	response := h.convertUsecaseOutputToResponse(output)

	c.JSON(http.StatusOK, response)
}

// DeletePost handles DELETE /api/v1/posts/{id}
func (h *PostHandler) DeletePost(c *gin.Context) {
	idParam := c.Param("id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid post ID",
		})
		return
	}

	// Get author ID from JWT context
	authorIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	authorID, err := uuid.Parse(authorIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid author ID",
		})
		return
	}

	// Delete post using use case
	err = h.postUsecase.DeletePost(c.Request.Context(), postID, authorID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
		case strings.Contains(err.Error(), "access denied"):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Not post author or insufficient permissions",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete post",
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListPosts handles GET /api/v1/posts
func (h *PostHandler) ListPosts(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build list options
	input := &usecase.ListPostsInput{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Apply filters
	if status := c.Query("status"); status != "" {
		switch strings.ToLower(status) {
		case "published":
			s := domain.PostStatusPublished
			input.Status = &s
		case "draft":
			s := domain.PostStatusDraft
			input.Status = &s
		case "archived":
			s := domain.PostStatusArchived
			input.Status = &s
		}
	}

	if postTypeID := c.Query("postTypeId"); postTypeID != "" {
		if id, err := uuid.Parse(postTypeID); err == nil {
			input.PostTypeID = &id
		}
	}

	if authorID := c.Query("authorId"); authorID != "" {
		if id, err := uuid.Parse(authorID); err == nil {
			input.AuthorID = &id
		}
	}

	if categoryID := c.Query("categoryId"); categoryID != "" {
		if id, err := uuid.Parse(categoryID); err == nil {
			input.CategoryID = &id
		}
	}

	if tagID := c.Query("tagId"); tagID != "" {
		if id, err := uuid.Parse(tagID); err == nil {
			input.TagID = &id
		}
	}

	// Apply sorting
	if sortBy := c.Query("sortBy"); sortBy != "" {
		input.SortBy = sortBy
	}
	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		input.SortOrder = sortOrder
	}

	// Get posts using use case
	output, err := h.postUsecase.ListPosts(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve posts",
		})
		return
	}

	// Convert to summary responses
	summaries := make([]PostSummaryResponse, len(output.Posts))
	for i, post := range output.Posts {
		summaries[i] = h.convertPostToSummary(post)
	}

	// Calculate pagination
	totalPages := int((output.TotalCount + int64(limit) - 1) / int64(limit))

	response := ListPostsResponse{
		Posts: summaries,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      output.TotalCount,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Filters: map[string]interface{}{},
	}

	// Add applied filters to response
	if input.Status != nil {
		response.Filters["status"] = string(*input.Status)
	}
	if input.PostTypeID != nil {
		response.Filters["postTypeId"] = input.PostTypeID.String()
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

// convertUsecaseOutputToResponse converts a usecase output to API response
func (h *PostHandler) convertUsecaseOutputToResponse(output *usecase.PostUsecaseOutput) PostResponse {
	post := output.Post
	response := PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		Content:    post.Content,
		AuthorID:   post.AuthorID,
		Status:     string(post.Status),
		CreatedAt:  post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  post.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ViewCount:  int64(post.ViewCount),
		IsApproved: post.IsApproved,
	}

	// Convert metadata from []interface{} to map[string]interface{}
	if len(post.Metadata) > 0 {
		metadata := make(map[string]interface{})
		// For now, convert slice to simple map with indexed keys
		// In a real implementation, this would depend on the actual metadata structure
		for i, v := range post.Metadata {
			metadata[fmt.Sprintf("item_%d", i)] = v
		}
		response.Metadata = metadata
	}

	// Handle optional fields
	if post.ScheduledAt != nil {
		scheduledAt := post.ScheduledAt.Format("2006-01-02T15:04:05Z")
		response.ScheduledAt = &scheduledAt
	}

	if post.PublishedAt != nil {
		publishedAt := post.PublishedAt.Format("2006-01-02T15:04:05Z")
		response.PublishedAt = &publishedAt
	}

	// Add post type from output
	if output.PostType != nil {
		response.PostType = &PostTypeResponse{
			ID:          output.PostType.ID,
			Name:        output.PostType.Name,
			DisplayName: output.PostType.DisplayName,
		}
	}

	// Add categories from output
	if len(output.Categories) > 0 {
		categories := make([]CategoryResponse, len(output.Categories))
		for i, cat := range output.Categories {
			categories[i] = CategoryResponse{
				ID:   cat.ID,
				Name: cat.Name,
				Slug: cat.Slug,
			}
		}
		response.Categories = categories
	}

	// Add tags from output
	if len(output.Tags) > 0 {
		tags := make([]TagResponse, len(output.Tags))
		for i, tag := range output.Tags {
			tags[i] = TagResponse{
				ID:   tag.ID,
				Name: tag.Name,
				Slug: tag.Slug,
			}
		}
		response.Tags = tags
	}

	return response
}

// convertPostToSummary converts a domain post to summary response
func (h *PostHandler) convertPostToSummary(post *domain.Post) PostSummaryResponse {
	content := post.Content
	if len(content) > 200 {
		content = content[:200] + "..."
	}

	response := PostSummaryResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       content,
		AuthorID:      post.AuthorID,
		Status:        string(post.Status),
		CreatedAt:     post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ViewCount:     int64(post.ViewCount),
		CategoryCount: 0, // TODO: Get from relationships
		TagCount:      0, // TODO: Get from relationships
	}

	if post.PublishedAt != nil {
		publishedAt := post.PublishedAt.Format("2006-01-02T15:04:05Z")
		response.PublishedAt = &publishedAt
	}

	// TODO: Add author and post type from relationships

	return response
}

// parseISO8601 parses an ISO 8601 date string
func parseISO8601(dateStr string) (time.Time, error) {
	// Try multiple ISO 8601 formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid ISO 8601 format")
}
