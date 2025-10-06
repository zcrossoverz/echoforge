package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// CategoryHandler handles HTTP requests for category management
type CategoryHandler struct {
	categoryRepo domain.PostTypeCategoryRepository
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryRepo domain.PostTypeCategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
	}
}

// CategoryDetailResponse represents the response structure for a category
type CategoryDetailResponse struct {
	ID          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	Slug        string                   `json:"slug"`
	Description string                   `json:"description,omitempty"`
	ParentID    *uuid.UUID               `json:"parentId,omitempty"`
	SortOrder   int                      `json:"sortOrder"`
	IsActive    bool                     `json:"isActive"`
	PostCount   int                      `json:"postCount"`
	CreatedAt   string                   `json:"createdAt"`
	UpdatedAt   string                   `json:"updatedAt"`
	Children    []CategoryDetailResponse `json:"children,omitempty"`
}

// ListCategoriesResponse represents the response for listing categories
type ListCategoriesResponse struct {
	Categories []CategoryDetailResponse `json:"categories"`
}

// GetCategory handles GET /api/v1/categories/{id}
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Get category using repository
	category, err := h.categoryRepo.GetByID(c.Request.Context(), categoryID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve category",
			})
		}
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
		return
	}

	// Convert to response
	response := h.convertCategoryToResponse(category)

	c.JSON(http.StatusOK, response)
}

// ListCategories handles GET /api/v1/categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Check if we should show only root categories (hierarchical view)
	parentID := c.Query("parentId")
	includeEmpty := c.DefaultQuery("includeEmpty", "true") == "true"

	var categories []*domain.PostCategory
	var err error

	if parentID == "" {
		// Get root categories (hierarchical view)
		categories, err = h.categoryRepo.GetRootCategories(c.Request.Context())
	} else if parentID != "null" {
		// Get children of specific parent
		if parentUUID, parseErr := uuid.Parse(parentID); parseErr == nil {
			categories, err = h.categoryRepo.GetChildren(c.Request.Context(), parentUUID)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid parent ID",
			})
			return
		}
	} else {
		// Get all categories with pagination
		options := domain.ListCategoriesOptions{
			Limit:        limit,
			Offset:       (page - 1) * limit,
			IncludeEmpty: includeEmpty,
		}

		// Apply active filter (IsActive filtering not currently supported in options)
		// if isActive := c.Query("isActive"); isActive != "" {
		//     active := strings.ToLower(isActive) == "true"
		//     // Note: IsActive filtering would need to be added to ListCategoriesOptions
		// }

		categories, err = h.categoryRepo.List(c.Request.Context(), options)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve categories",
		})
		return
	}

	// Convert to responses
	responses := make([]CategoryDetailResponse, len(categories))
	for i, category := range categories {
		responses[i] = h.convertCategoryToResponse(category)

		// Load children for hierarchical view if it's a root category
		if parentID == "" && category.ParentID == nil {
			children, err := h.categoryRepo.GetChildren(c.Request.Context(), category.ID)
			if err == nil && len(children) > 0 {
				childResponses := make([]CategoryDetailResponse, len(children))
				for j, child := range children {
					childResponses[j] = h.convertCategoryToResponse(child)
				}
				responses[i].Children = childResponses
			}
		}
	}

	response := ListCategoriesResponse{
		Categories: responses,
	}

	c.JSON(http.StatusOK, response)
}

// CreateCategory handles POST /api/v1/categories (Admin only - placeholder)
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Category creation requires administrative privileges",
	})
}

// UpdateCategory handles PUT /api/v1/categories/{id} (Admin only - placeholder)
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Category modification requires administrative privileges",
	})
}

// DeleteCategory handles DELETE /api/v1/categories/{id} (Admin only - placeholder)
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Category deletion requires administrative privileges",
	})
}

// Helper methods

// convertCategoryToResponse converts a domain category to response
func (h *CategoryHandler) convertCategoryToResponse(category *domain.PostCategory) CategoryDetailResponse {
	response := CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ParentID:    category.ParentID,
		SortOrder:   category.SortOrder,
		IsActive:    category.IsActive,
		PostCount:   category.PostCount,
		CreatedAt:   category.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   category.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	return response
}
