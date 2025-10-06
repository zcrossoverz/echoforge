package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// TagHandler handles HTTP requests for tag management
type TagHandler struct {
	postRepo domain.PostRepository // Using PostRepository until dedicated TagRepository is implemented
}

// NewTagHandler creates a new tag handler
func NewTagHandler(postRepo domain.PostRepository) *TagHandler {
	return &TagHandler{
		postRepo: postRepo,
	}
}

// TagDetailResponse represents detailed tag information
type TagDetailResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	UsageCount  int       `json:"usage_count"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// ListTagsResponse represents the response for listing tags
type ListTagsResponse struct {
	Tags []TagDetailResponse `json:"tags"`
}

// CreateTagRequest represents the request payload for creating a tag
type CreateTagRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	Color       string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// ListTags handles GET /api/v1/tags
func (h *TagHandler) ListTags(c *gin.Context) {
	// Parse query parameters
	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	search := strings.TrimSpace(c.Query("search"))

	// TODO: This is a placeholder implementation
	// When TagRepository is implemented, replace this with proper tag listing
	// For now, return empty list with proper structure
	response := ListTagsResponse{
		Tags: []TagDetailResponse{},
	}

	// Suppress unused variable warnings until TagRepository is implemented
	_ = limit
	_ = offset
	_ = search

	// Placeholder: In a real implementation, this would:
	// 1. Call tagRepo.List(ctx, ListTagsOptions{Limit: limit, Offset: offset, Search: search})
	// 2. Convert domain.PostTag entities to TagDetailResponse
	// 3. Handle sorting by usage count, name, or creation date

	c.JSON(http.StatusOK, response)
}

// GetTag handles GET /api/v1/tags/{id}
func (h *TagHandler) GetTag(c *gin.Context) {
	idParam := c.Param("id")
	tagID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tag ID",
		})
		return
	}

	// TODO: This is a placeholder implementation
	// When TagRepository is implemented, replace this with:
	// tag, err := h.tagRepo.FindByID(c.Request.Context(), tagID)
	_ = tagID // Suppress unused variable warning

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Tag not found - TagRepository not yet implemented",
	})
}

// CreateTag handles POST /api/v1/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Set default color if not provided
	if req.Color == "" {
		req.Color = "#6B7280" // Default gray color
	}

	// TODO: This is a placeholder implementation
	// When TagRepository is implemented, replace this with:
	// 1. Create new PostTag domain entity
	// 2. Call h.tagRepo.Create(c.Request.Context(), tag)
	// 3. Return created tag as TagDetailResponse

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Tag creation not yet implemented - requires TagRepository",
	})
}

// UpdateTag handles PUT /api/v1/tags/{id}
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idParam := c.Param("id")
	tagID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tag ID",
		})
		return
	}

	var req CreateTagRequest // Reuse same structure for updates
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// TODO: This is a placeholder implementation
	// When TagRepository is implemented, replace this with:
	// 1. Find existing tag by ID
	// 2. Update tag properties
	// 3. Call h.tagRepo.Update(c.Request.Context(), tag)
	// 4. Return updated tag as TagDetailResponse
	_ = tagID // Suppress unused variable warning

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Tag update not yet implemented - requires TagRepository",
	})
}

// DeleteTag handles DELETE /api/v1/tags/{id}
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idParam := c.Param("id")
	tagID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tag ID",
		})
		return
	}

	// TODO: This is a placeholder implementation
	// When TagRepository is implemented, replace this with:
	// 1. Check if tag is system tag (cannot be deleted)
	// 2. Check if tag is in use (might want to prevent deletion)
	// 3. Call h.tagRepo.Delete(c.Request.Context(), tagID)
	_ = tagID // Suppress unused variable warning

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Tag deletion not yet implemented - requires TagRepository",
	})
}

// Helper methods

// convertTagToResponse converts a domain PostTag to TagDetailResponse
func (h *TagHandler) convertTagToResponse(tag *domain.PostTag) TagDetailResponse {
	return TagDetailResponse{
		ID:          tag.ID,
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		UsageCount:  tag.UsageCount,
		IsSystem:    tag.IsSystem,
		CreatedAt:   tag.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   tag.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// convertTagToSimpleResponse converts a domain PostTag to simple TagResponse for use in other handlers
func (h *TagHandler) convertTagToSimpleResponse(tag *domain.PostTag) TagResponse {
	return TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
		Slug: tag.Slug,
	}
}
