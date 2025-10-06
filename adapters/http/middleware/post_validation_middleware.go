package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// PostValidationMiddleware provides post-specific validation capabilities
type PostValidationMiddleware struct {
	baseValidator *ValidationMiddleware
	maxTitleLen   int
	maxContentLen int
	maxExcerptLen int
}

// NewPostValidationMiddleware creates a new post validation middleware
func NewPostValidationMiddleware() *PostValidationMiddleware {
	return &PostValidationMiddleware{
		baseValidator: NewValidationMiddleware(),
		maxTitleLen:   255,
		maxContentLen: 1000000, // 1MB text content
		maxExcerptLen: 500,
	}
}

// PostCreateValidation validates post creation requests
func (pvm *PostValidationMiddleware) PostCreateValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Title       string                 `json:"title" validate:"required,min=1,max=255"`
			Content     string                 `json:"content" validate:"required,min=1"`
			PostTypeID  uuid.UUID              `json:"postTypeId" validate:"required"`
			Status      string                 `json:"status,omitempty"`
			CategoryIDs []uuid.UUID            `json:"categoryIds,omitempty"`
			TagIDs      []uuid.UUID            `json:"tagIds,omitempty"`
			ScheduledAt *string                `json:"scheduledAt,omitempty"`
			Metadata    map[string]interface{} `json:"metadata,omitempty"`
		}

		// Parse and validate JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Title) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Title is required and cannot be empty",
			})
			c.Abort()
			return
		}

		if strings.TrimSpace(req.Content) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content is required and cannot be empty",
			})
			c.Abort()
			return
		}

		// Validate field lengths
		if len(req.Title) > pvm.maxTitleLen {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Title exceeds maximum length",
				"max_length": pvm.maxTitleLen,
			})
			c.Abort()
			return
		}

		if len(req.Content) > pvm.maxContentLen {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Content exceeds maximum length",
				"max_length": pvm.maxContentLen,
			})
			c.Abort()
			return
		}

		// Validate status if provided
		if req.Status != "" {
			validStatuses := []string{
				string(domain.PostStatusDraft),
				string(domain.PostStatusScheduled),
				string(domain.PostStatusPublished),
				string(domain.PostStatusPendingApproval),
			}

			isValidStatus := false
			for _, status := range validStatuses {
				if req.Status == status {
					isValidStatus = true
					break
				}
			}

			if !isValidStatus {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":          "Invalid status value",
					"valid_statuses": validStatuses,
				})
				c.Abort()
				return
			}
		}

		// Validate scheduled date if provided
		if req.ScheduledAt != nil && *req.ScheduledAt != "" {
			_, err := time.Parse("2006-01-02T15:04:05Z", *req.ScheduledAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid scheduledAt format. Use ISO 8601 format (2006-01-02T15:04:05Z)",
				})
				c.Abort()
				return
			}
		}

		// Validate category limits
		if len(req.CategoryIDs) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum 10 categories allowed per post",
			})
			c.Abort()
			return
		}

		// Validate tag limits
		if len(req.TagIDs) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum 20 tags allowed per post",
			})
			c.Abort()
			return
		}

		// Store validated data in context
		c.Set("validated_post_data", req)
		c.Next()
	}
}

// PostUpdateValidation validates post update requests
func (pvm *PostValidationMiddleware) PostUpdateValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate post ID parameter
		postIDParam := c.Param("id")
		if postIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Post ID is required",
			})
			c.Abort()
			return
		}

		postID, err := uuid.Parse(postIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid post ID format",
			})
			c.Abort()
			return
		}

		var req struct {
			Title       *string                 `json:"title,omitempty"`
			Content     *string                 `json:"content,omitempty"`
			Status      *string                 `json:"status,omitempty"`
			CategoryIDs *[]uuid.UUID            `json:"categoryIds,omitempty"`
			TagIDs      *[]uuid.UUID            `json:"tagIds,omitempty"`
			ScheduledAt *string                 `json:"scheduledAt,omitempty"`
			Metadata    *map[string]interface{} `json:"metadata,omitempty"`
		}

		// Parse JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate title if provided
		if req.Title != nil {
			if strings.TrimSpace(*req.Title) == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Title cannot be empty",
				})
				c.Abort()
				return
			}
			if len(*req.Title) > pvm.maxTitleLen {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "Title exceeds maximum length",
					"max_length": pvm.maxTitleLen,
				})
				c.Abort()
				return
			}
		}

		// Validate content if provided
		if req.Content != nil {
			if strings.TrimSpace(*req.Content) == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Content cannot be empty",
				})
				c.Abort()
				return
			}
			if len(*req.Content) > pvm.maxContentLen {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "Content exceeds maximum length",
					"max_length": pvm.maxContentLen,
				})
				c.Abort()
				return
			}
		}

		// Validate status if provided
		if req.Status != nil {
			validStatuses := []string{
				string(domain.PostStatusDraft),
				string(domain.PostStatusScheduled),
				string(domain.PostStatusPublished),
				string(domain.PostStatusArchived),
				string(domain.PostStatusPendingApproval),
			}

			isValidStatus := false
			for _, status := range validStatuses {
				if *req.Status == status {
					isValidStatus = true
					break
				}
			}

			if !isValidStatus {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":          "Invalid status value",
					"valid_statuses": validStatuses,
				})
				c.Abort()
				return
			}
		}

		// Validate scheduled date if provided
		if req.ScheduledAt != nil && *req.ScheduledAt != "" {
			_, err := time.Parse("2006-01-02T15:04:05Z", *req.ScheduledAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid scheduledAt format. Use ISO 8601 format (2006-01-02T15:04:05Z)",
				})
				c.Abort()
				return
			}
		}

		// Validate category limits if provided
		if req.CategoryIDs != nil && len(*req.CategoryIDs) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum 10 categories allowed per post",
			})
			c.Abort()
			return
		}

		// Validate tag limits if provided
		if req.TagIDs != nil && len(*req.TagIDs) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum 20 tags allowed per post",
			})
			c.Abort()
			return
		}

		// Store validated data in context
		c.Set("validated_post_id", postID)
		c.Set("validated_update_data", req)
		c.Next()
	}
}

// PostListValidation validates post listing query parameters
func (pvm *PostValidationMiddleware) PostListValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate pagination parameters
		page := 1
		if pageStr := c.Query("page"); pageStr != "" {
			if parsedPage, err := strconv.Atoi(pageStr); err != nil || parsedPage < 1 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid page parameter. Must be a positive integer",
				})
				c.Abort()
				return
			} else {
				page = parsedPage
			}
		}

		limit := 20
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err != nil || parsedLimit < 1 || parsedLimit > 100 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid limit parameter. Must be between 1 and 100",
				})
				c.Abort()
				return
			} else {
				limit = parsedLimit
			}
		}

		// Validate status filter
		status := c.Query("status")
		if status != "" {
			validStatuses := []string{
				string(domain.PostStatusDraft),
				string(domain.PostStatusScheduled),
				string(domain.PostStatusPublished),
				string(domain.PostStatusArchived),
				string(domain.PostStatusPendingApproval),
			}

			isValidStatus := false
			for _, validStatus := range validStatuses {
				if status == validStatus {
					isValidStatus = true
					break
				}
			}

			if !isValidStatus {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":          "Invalid status filter",
					"valid_statuses": validStatuses,
				})
				c.Abort()
				return
			}
		}

		// Validate postTypeId filter
		postTypeID := c.Query("postTypeId")
		if postTypeID != "" {
			if _, err := uuid.Parse(postTypeID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid postTypeId format",
				})
				c.Abort()
				return
			}
		}

		// Validate categoryId filter
		categoryID := c.Query("categoryId")
		if categoryID != "" {
			if _, err := uuid.Parse(categoryID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid categoryId format",
				})
				c.Abort()
				return
			}
		}

		// Validate tagId filter
		tagID := c.Query("tagId")
		if tagID != "" {
			if _, err := uuid.Parse(tagID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid tagId format",
				})
				c.Abort()
				return
			}
		}

		// Validate sort parameter
		sort := c.Query("sort")
		if sort != "" {
			validSorts := []string{"created_at", "updated_at", "published_at", "title", "view_count"}
			isValidSort := false
			for _, validSort := range validSorts {
				if sort == validSort {
					isValidSort = true
					break
				}
			}

			if !isValidSort {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":       "Invalid sort parameter",
					"valid_sorts": validSorts,
				})
				c.Abort()
				return
			}
		}

		// Validate order parameter
		order := c.Query("order")
		if order != "" && order != "asc" && order != "desc" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid order parameter. Must be 'asc' or 'desc'",
			})
			c.Abort()
			return
		}

		// Store validated parameters in context
		c.Set("validated_page", page)
		c.Set("validated_limit", limit)
		c.Next()
	}
}

// BulkOperationValidation validates bulk operation requests
func (pvm *PostValidationMiddleware) BulkOperationValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PostIDs   []uuid.UUID `json:"postIds" validate:"required,min=1,max=100"`
			Operation string      `json:"operation" validate:"required"`
			Data      struct {
				Status      *string     `json:"status,omitempty"`
				CategoryIDs []uuid.UUID `json:"categoryIds,omitempty"`
				TagIDs      []uuid.UUID `json:"tagIds,omitempty"`
			} `json:"data"`
		}

		// Parse JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate post IDs
		if len(req.PostIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "At least one post ID is required",
			})
			c.Abort()
			return
		}

		if len(req.PostIDs) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum 100 posts allowed per bulk operation",
			})
			c.Abort()
			return
		}

		// Validate operation
		validOperations := []string{
			"update_status", "add_categories", "remove_categories",
			"add_tags", "remove_tags", "archive",
		}

		isValidOperation := false
		for _, op := range validOperations {
			if req.Operation == op {
				isValidOperation = true
				break
			}
		}

		if !isValidOperation {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":            "Invalid bulk operation",
				"valid_operations": validOperations,
			})
			c.Abort()
			return
		}

		// Validate operation-specific data
		switch req.Operation {
		case "update_status":
			if req.Data.Status == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Status is required for update_status operation",
				})
				c.Abort()
				return
			}
		case "add_categories", "remove_categories":
			if len(req.Data.CategoryIDs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "At least one category ID is required for category operations",
				})
				c.Abort()
				return
			}
		case "add_tags", "remove_tags":
			if len(req.Data.TagIDs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "At least one tag ID is required for tag operations",
				})
				c.Abort()
				return
			}
		}

		c.Set("validated_bulk_data", req)
		c.Next()
	}
}
