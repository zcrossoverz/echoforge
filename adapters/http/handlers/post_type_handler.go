package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// PostTypeHandler handles HTTP requests for post type management
type PostTypeHandler struct {
	postTypeUsecase *usecase.PostTypeUsecase
}

// NewPostTypeHandler creates a new post type handler
func NewPostTypeHandler(postTypeUsecase *usecase.PostTypeUsecase) *PostTypeHandler {
	return &PostTypeHandler{
		postTypeUsecase: postTypeUsecase,
	}
}

// PostTypeDetailResponse represents detailed response for post types
type PostTypeDetailResponse struct {
	ID                uuid.UUID               `json:"id"`
	Name              string                  `json:"name"`
	Slug              string                  `json:"slug"`
	Description       *string                 `json:"description"`
	Icon              *string                 `json:"icon"`
	Color             *string                 `json:"color"`
	SortOrder         int                     `json:"sort_order"`
	IsActive          bool                    `json:"is_active"`
	RequiresApproval  bool                    `json:"requires_approval"`
	AllowsScheduling  bool                    `json:"allows_scheduling"`
	AllowsAttachments bool                    `json:"allows_attachments"`
	PostCount         int                     `json:"post_count"`
	IsDefault         bool                    `json:"is_default"`
	FieldDefinitions  []PostTypeFieldResponse `json:"field_definitions,omitempty"`
	CreatedAt         string                  `json:"created_at"`
	UpdatedAt         string                  `json:"updated_at"`
}

// PostTypeFieldResponse represents field definition in response
type PostTypeFieldResponse struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	Default     interface{}            `json:"default,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ListPostTypesResponse represents the response for listing post types
type ListPostTypesResponse struct {
	PostTypes []PostTypeResponse `json:"postTypes"`
}

// GetPostType handles GET /api/v1/post-types/{id}
func (h *PostTypeHandler) GetPostType(c *gin.Context) {
	idParam := c.Param("id")
	postTypeID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid post type ID",
		})
		return
	}

	// Get post type using use case
	output, err := h.postTypeUsecase.GetPostType(c.Request.Context(), postTypeID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post type not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve post type",
			})
		}
		return
	}

	// Convert to response
	response := h.convertPostTypeToResponse(output.PostType)

	c.JSON(http.StatusOK, response)
}

// ListPostTypes handles GET /api/v1/post-types
func (h *PostTypeHandler) ListPostTypes(c *gin.Context) {
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
	input := &usecase.ListPostTypesInput{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Apply filters
	if isActive := c.Query("isActive"); isActive != "" {
		active := strings.ToLower(isActive) == "true"
		input.IsActive = &active
	}

	if isSystem := c.Query("isSystem"); isSystem != "" {
		system := strings.ToLower(isSystem) == "true"
		input.IsSystem = &system
	}

	// Apply sorting
	if sortBy := c.Query("sortBy"); sortBy != "" {
		input.SortBy = sortBy
	}
	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		input.SortOrder = sortOrder
	}

	// Get post types using use case
	output, err := h.postTypeUsecase.ListPostTypes(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve post types",
		})
		return
	}

	// Convert to responses
	responses := make([]PostTypeResponse, len(output.PostTypes))
	for i, postType := range output.PostTypes {
		responses[i] = h.convertPostTypeToSimpleResponse(postType)
	}

	response := ListPostTypesResponse{
		PostTypes: responses,
	}

	c.JSON(http.StatusOK, response)
}

// CreatePostType handles POST /api/v1/post-types (Admin only - placeholder)
func (h *PostTypeHandler) CreatePostType(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Post type creation requires administrative privileges",
	})
}

// UpdatePostType handles PUT /api/v1/post-types/{id} (Admin only - placeholder)
func (h *PostTypeHandler) UpdatePostType(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Post type modification requires administrative privileges",
	})
}

// DeletePostType handles DELETE /api/v1/post-types/{id} (Admin only - placeholder)
func (h *PostTypeHandler) DeletePostType(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Post type deletion requires administrative privileges",
	})
}

// Helper methods

// convertPostTypeToResponse converts a domain post type to response
func (h *PostTypeHandler) convertPostTypeToResponse(postType *domain.PostType) PostTypeDetailResponse {
	var description *string
	if postType.Description != "" {
		description = &postType.Description
	}

	response := PostTypeDetailResponse{
		ID:                postType.ID,
		Name:              postType.Name,
		Slug:              postType.Name, // Use name as slug for now
		Description:       description,
		Icon:              nil, // Not available in domain
		Color:             nil, // Not available in domain
		SortOrder:         0,   // Not available in domain
		IsActive:          postType.IsActive,
		RequiresApproval:  postType.RequiresApproval,
		AllowsScheduling:  postType.AllowsScheduling,
		AllowsAttachments: postType.AllowsAttachments,
		PostCount:         0,     // Would need to be calculated
		IsDefault:         false, // Not available in domain
		CreatedAt:         postType.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         postType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Convert JSON field definitions to structured response
	if len(postType.FieldDefinitions) > 0 {
		var fieldDefs []PostTypeFieldResponse
		var rawFields []map[string]interface{}
		if err := json.Unmarshal(postType.FieldDefinitions, &rawFields); err == nil {
			for _, field := range rawFields {
				fieldResp := PostTypeFieldResponse{}
				if name, ok := field["name"].(string); ok {
					fieldResp.Name = name
				}
				if fieldType, ok := field["type"].(string); ok {
					fieldResp.Type = fieldType
				}
				if required, ok := field["required"].(bool); ok {
					fieldResp.Required = required
				}
				if defaultVal, ok := field["default"]; ok {
					fieldResp.Default = defaultVal
				}
				if validation, ok := field["validation"].(map[string]interface{}); ok {
					fieldResp.Validation = validation
				}
				if description, ok := field["description"].(string); ok {
					fieldResp.Description = description
				}
				fieldDefs = append(fieldDefs, fieldResp)
			}
			response.FieldDefinitions = fieldDefs
		}
	}

	return response
}

// convertPostTypeToSimpleResponse converts a domain post type to simple response for lists
func (h *PostTypeHandler) convertPostTypeToSimpleResponse(postType *domain.PostType) PostTypeResponse {
	return PostTypeResponse{
		ID:          postType.ID,
		Name:        postType.Name,
		DisplayName: postType.DisplayName,
	}
}
