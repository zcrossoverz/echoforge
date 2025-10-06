package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// BulkHandler handles HTTP requests for bulk operations on posts
type BulkHandler struct {
	postRepo domain.PostRepository
}

// NewBulkHandler creates a new bulk operations handler
func NewBulkHandler(postRepo domain.PostRepository) *BulkHandler {
	return &BulkHandler{
		postRepo: postRepo,
	}
}

// BulkOperation represents supported bulk operations
type BulkOperation string

const (
	BulkOperationUpdateStatus     BulkOperation = "update_status"
	BulkOperationAddCategories    BulkOperation = "add_categories"
	BulkOperationRemoveCategories BulkOperation = "remove_categories"
	BulkOperationAddTags          BulkOperation = "add_tags"
	BulkOperationRemoveTags       BulkOperation = "remove_tags"
	BulkOperationArchive          BulkOperation = "archive"
)

// BulkOperationRequest represents the request payload for bulk operations
type BulkOperationRequest struct {
	PostIDs               []uuid.UUID       `json:"postIds" validate:"required,min=1,max=100"`
	Operation             BulkOperation     `json:"operation" validate:"required"`
	Data                  BulkOperationData `json:"data"`
	ApplyApprovalWorkflow bool              `json:"applyApprovalWorkflow"`
}

// BulkOperationData represents the data for specific bulk operations
type BulkOperationData struct {
	Status      *string     `json:"status,omitempty"`
	CategoryIDs []uuid.UUID `json:"categoryIds,omitempty"`
	TagIDs      []uuid.UUID `json:"tagIds,omitempty"`
}

// BulkOperationResponse represents the response from bulk operations
type BulkOperationResponse struct {
	ProcessedCount    int                   `json:"processedCount"`
	FailedCount       int                   `json:"failedCount"`
	Results           []BulkOperationResult `json:"results"`
	ApprovalRequired  bool                  `json:"approvalRequired"`
	ApprovalRequestID *uuid.UUID            `json:"approvalRequestId,omitempty"`
}

// BulkOperationResult represents the result for a single post in bulk operation
type BulkOperationResult struct {
	PostID  uuid.UUID `json:"postId"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
}

// BulkUpdatePosts handles POST /api/v1/posts/bulk
func (h *BulkHandler) BulkUpdatePosts(c *gin.Context) {
	var req BulkOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	// Validate request
	if len(req.PostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one post ID is required",
		})
		return
	}

	if len(req.PostIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 100 posts allowed per bulk operation",
		})
		return
	}

	// Validate operation-specific data
	if err := h.validateOperationData(req.Operation, req.Data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: Extract user ID from JWT for permission checking
	// userID := getUserIDFromContext(c)

	// Process bulk operation
	response, err := h.processBulkOperation(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process bulk operation",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BulkDeletePosts handles DELETE /api/v1/posts/bulk
func (h *BulkHandler) BulkDeletePosts(c *gin.Context) {
	var req struct {
		PostIDs               []uuid.UUID `json:"postIds" validate:"required,min=1,max=100"`
		ApplyApprovalWorkflow bool        `json:"applyApprovalWorkflow"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	if len(req.PostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one post ID is required",
		})
		return
	}

	if len(req.PostIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 100 posts allowed per bulk operation",
		})
		return
	}

	// Convert to bulk operation request
	bulkReq := BulkOperationRequest{
		PostIDs:               req.PostIDs,
		Operation:             BulkOperationArchive,
		ApplyApprovalWorkflow: req.ApplyApprovalWorkflow,
	}

	// Process bulk operation
	response, err := h.processBulkOperation(c, bulkReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process bulk delete operation",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

// processBulkOperation processes the bulk operation based on the operation type
func (h *BulkHandler) processBulkOperation(c *gin.Context, req BulkOperationRequest) (*BulkOperationResponse, error) {
	results := make([]BulkOperationResult, 0, len(req.PostIDs))
	processedCount := 0
	failedCount := 0

	switch req.Operation {
	case BulkOperationUpdateStatus:
		if req.Data.Status == nil {
			for _, postID := range req.PostIDs {
				results = append(results, BulkOperationResult{
					PostID:  postID,
					Success: false,
					Error:   "Status is required for update_status operation",
				})
				failedCount++
			}
		} else {
			status := domain.PostStatus(*req.Data.Status)
			err := h.postRepo.BulkUpdateStatus(c.Request.Context(), req.PostIDs, status)
			if err != nil {
				// If bulk operation fails, mark all as failed
				for _, postID := range req.PostIDs {
					results = append(results, BulkOperationResult{
						PostID:  postID,
						Success: false,
						Error:   "Failed to update status",
					})
					failedCount++
				}
			} else {
				// If bulk operation succeeds, mark all as successful
				for _, postID := range req.PostIDs {
					results = append(results, BulkOperationResult{
						PostID:  postID,
						Success: true,
					})
					processedCount++
				}
			}
		}

	case BulkOperationArchive:
		err := h.postRepo.BulkDelete(c.Request.Context(), req.PostIDs)
		if err != nil {
			for _, postID := range req.PostIDs {
				results = append(results, BulkOperationResult{
					PostID:  postID,
					Success: false,
					Error:   "Failed to archive post",
				})
				failedCount++
			}
		} else {
			for _, postID := range req.PostIDs {
				results = append(results, BulkOperationResult{
					PostID:  postID,
					Success: true,
				})
				processedCount++
			}
		}

	case BulkOperationAddCategories, BulkOperationRemoveCategories:
		// TODO: Implement category operations when PostCategoryRepository is available
		for _, postID := range req.PostIDs {
			results = append(results, BulkOperationResult{
				PostID:  postID,
				Success: false,
				Error:   "Category operations not yet implemented",
			})
			failedCount++
		}

	case BulkOperationAddTags, BulkOperationRemoveTags:
		// TODO: Implement tag operations when PostTagRepository is available
		for _, postID := range req.PostIDs {
			results = append(results, BulkOperationResult{
				PostID:  postID,
				Success: false,
				Error:   "Tag operations not yet implemented",
			})
			failedCount++
		}

	default:
		for _, postID := range req.PostIDs {
			results = append(results, BulkOperationResult{
				PostID:  postID,
				Success: false,
				Error:   "Unsupported bulk operation",
			})
			failedCount++
		}
	}

	// TODO: Handle approval workflow
	approvalRequired := false
	var approvalRequestID *uuid.UUID
	if req.ApplyApprovalWorkflow {
		// In a real implementation, this would:
		// 1. Check if the current user needs approval for these operations
		// 2. Create an approval request if needed
		// 3. Queue the operations for approval
		approvalRequired = false // Placeholder
	}

	return &BulkOperationResponse{
		ProcessedCount:    processedCount,
		FailedCount:       failedCount,
		Results:           results,
		ApprovalRequired:  approvalRequired,
		ApprovalRequestID: approvalRequestID,
	}, nil
}

// validateOperationData validates the data for specific bulk operations
func (h *BulkHandler) validateOperationData(operation BulkOperation, data BulkOperationData) error {
	switch operation {
	case BulkOperationUpdateStatus:
		if data.Status == nil {
			return &ValidationError{Message: "Status is required for update_status operation"}
		}
		// Validate status value
		status := domain.PostStatus(*data.Status)
		validStatuses := []domain.PostStatus{
			domain.PostStatusDraft,
			domain.PostStatusScheduled,
			domain.PostStatusPublished,
			domain.PostStatusArchived,
			domain.PostStatusPendingApproval,
		}
		isValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			return &ValidationError{Message: "Invalid status value"}
		}

	case BulkOperationAddCategories, BulkOperationRemoveCategories:
		if len(data.CategoryIDs) == 0 {
			return &ValidationError{Message: "At least one category ID is required for category operations"}
		}

	case BulkOperationAddTags, BulkOperationRemoveTags:
		if len(data.TagIDs) == 0 {
			return &ValidationError{Message: "At least one tag ID is required for tag operations"}
		}

	case BulkOperationArchive:
		// No additional validation needed for archive operation

	default:
		return &ValidationError{Message: "Unsupported bulk operation"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
