package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zcrossoverz/echoforge/adapters/http/handlers"
	"github.com/zcrossoverz/echoforge/adapters/http/middleware"
)

// PostRouterConfig holds configuration for post-related routes
type PostRouterConfig struct {
	Logger                   *zap.Logger
	PostHandler              *handlers.PostHandler
	PostTypeHandler          *handlers.PostTypeHandler
	CategoryHandler          *handlers.CategoryHandler
	TagHandler               *handlers.TagHandler
	SearchHandler            *handlers.SearchHandler
	AttachmentHandler        *handlers.AttachmentHandler
	BulkHandler              *handlers.BulkHandler
	AuthMiddleware           gin.HandlerFunc
	PostValidationMiddleware *middleware.PostValidationMiddleware
	FileSecurityMiddleware   *middleware.FileSecurityMiddleware
	PostRateLimiter          *middleware.PostRateLimiter
}

// SetupPostRoutes configures all post-related routes with proper middleware
func SetupPostRoutes(group *gin.RouterGroup, config *PostRouterConfig) {
	// Initialize middleware
	postValidation := config.PostValidationMiddleware
	fileSecurity := config.FileSecurityMiddleware
	rateLimiter := config.PostRateLimiter

	// Post management routes
	posts := group.Group("/posts")
	{
		// Public post endpoints (read-only)
		posts.GET("",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			config.PostHandler.ListPosts,
		)

		posts.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.PostHandler.GetPost,
		)

		// Protected post endpoints (require authentication)
		posts.POST("",
			config.AuthMiddleware,
			rateLimiter.PostCreationRateLimit(),
			postValidation.PostCreateValidation(),
			config.PostHandler.CreatePost,
		)

		posts.PUT("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostUpdateValidation(),
			config.PostHandler.UpdatePost,
		)

		posts.DELETE("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.PostHandler.DeletePost,
		)

		// Post status management - use UpdatePost with specific validation
		posts.PATCH("/:id/status",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.PostHandler.UpdatePost, // UpdatePost handles status changes
		)

		// Post scheduling - use UpdatePost with scheduled date
		posts.PATCH("/:id/schedule",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.PostHandler.UpdatePost, // UpdatePost handles scheduling
		)
	}

	// Post type management routes
	postTypes := group.Group("/post-types")
	{
		// Public post type endpoints
		postTypes.GET("",
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.ListPostTypes,
		)

		postTypes.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.GetPostType,
		)

		// Admin-only endpoints (placeholder for future admin middleware)
		postTypes.POST("",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.CreatePostType,
		)

		postTypes.PUT("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.UpdatePostType,
		)

		postTypes.DELETE("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.DeletePostType,
		)
	}

	// Category management routes
	categories := group.Group("/categories")
	{
		// Public category endpoints
		categories.GET("",
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.ListCategories,
		)

		categories.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.GetCategory,
		)

		// Get posts by category - placeholder for future implementation
		categories.GET("/:id/posts",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			func(c *gin.Context) {
				c.JSON(501, gin.H{"error": "Category posts endpoint not implemented yet"})
			},
		)

		// Protected category endpoints
		categories.POST("",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.CreateCategory,
		)

		categories.PUT("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.UpdateCategory,
		)

		categories.DELETE("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.DeleteCategory,
		)
	}

	// Tag management routes
	tags := group.Group("/tags")
	{
		// Public tag endpoints
		tags.GET("",
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.ListTags,
		)

		tags.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.GetTag,
		)

		// Get posts by tag - placeholder for future implementation
		tags.GET("/:id/posts",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			func(c *gin.Context) {
				c.JSON(501, gin.H{"error": "Tag posts endpoint not implemented yet"})
			},
		)

		// Protected tag endpoints
		tags.POST("",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.CreateTag,
		)

		tags.PUT("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.UpdateTag,
		)

		tags.DELETE("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.DeleteTag,
		)
	}

	// Search and filtering routes
	search := group.Group("/search")
	{
		// Global search
		search.GET("/posts",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GlobalSearch,
		)

		// Search suggestions
		search.GET("/suggestions",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetSuggestions,
		)

		// Trending posts
		search.GET("/trending",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetTrending,
		)

		// Popular posts
		search.GET("/popular",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetPopular,
		)
	}

	// File attachment routes
	attachments := group.Group("/attachments")
	{
		// Upload files (protected)
		attachments.POST("",
			config.AuthMiddleware,
			rateLimiter.FileUploadRateLimit(),
			fileSecurity.ValidateFileUpload(),
			fileSecurity.ValidateAttachmentMetadata(),
			fileSecurity.SetSecurityHeaders(),
			config.AttachmentHandler.UploadAttachment,
		)

		// List post attachments
		attachments.GET("/post/:postId",
			rateLimiter.GeneralPostRateLimit(),
			config.AttachmentHandler.ListPostAttachments,
		)

		// Download attachment
		attachments.GET("/:id/download",
			rateLimiter.GeneralPostRateLimit(),
			fileSecurity.ValidateFileDownload(),
			fileSecurity.SetSecurityHeaders(),
			config.AttachmentHandler.DownloadAttachment,
		)

		// Get attachment metadata - placeholder for future implementation
		attachments.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			func(c *gin.Context) {
				c.JSON(501, gin.H{"error": "Get attachment metadata endpoint not implemented yet"})
			},
		)

		// Update attachment metadata (protected)
		attachments.PUT("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.AttachmentHandler.UpdateAttachment,
		)

		// Delete attachment (protected)
		attachments.DELETE("/:id",
			config.AuthMiddleware,
			rateLimiter.GeneralPostRateLimit(),
			config.AttachmentHandler.DeleteAttachment,
		)
	}

	// Bulk operations routes (protected, heavily rate limited)
	bulk := group.Group("/bulk")
	{
		bulk.Use(config.AuthMiddleware)
		bulk.Use(rateLimiter.BulkOperationRateLimit())
		bulk.Use(postValidation.BulkOperationValidation())

		// Bulk update posts
		bulk.PUT("/posts", config.BulkHandler.BulkUpdatePosts)

		// Bulk delete posts
		bulk.DELETE("/posts", config.BulkHandler.BulkDeletePosts)
	}

	// Rate limit status endpoint (for debugging)
	if gin.Mode() != gin.ReleaseMode {
		group.GET("/rate-limit-status", rateLimiter.GetRateLimitStatus())
	}

	config.Logger.Info("Post routes configured successfully",
		zap.String("base_path", group.BasePath()),
		zap.Int("route_groups", 7), // posts, post-types, categories, tags, search, attachments, bulk
	)
}

// SetupPublicPostRoutes configures only public (non-authenticated) post routes
func SetupPublicPostRoutes(group *gin.RouterGroup, config *PostRouterConfig) {
	rateLimiter := config.PostRateLimiter
	postValidation := config.PostValidationMiddleware

	// Public read-only post endpoints
	posts := group.Group("/posts")
	{
		posts.GET("",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			config.PostHandler.ListPosts,
		)

		posts.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.PostHandler.GetPost,
		)
	}

	// Public post type endpoints
	postTypes := group.Group("/post-types")
	{
		postTypes.GET("",
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.ListPostTypes,
		)

		postTypes.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.PostTypeHandler.GetPostType,
		)
	}

	// Public category endpoints
	categories := group.Group("/categories")
	{
		categories.GET("",
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.ListCategories,
		)

		categories.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.CategoryHandler.GetCategory,
		)

		categories.GET("/:id/posts",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			func(c *gin.Context) {
				c.JSON(501, gin.H{"error": "Category posts endpoint not implemented yet"})
			},
		)
	}

	// Public tag endpoints
	tags := group.Group("/tags")
	{
		tags.GET("",
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.ListTags,
		)

		tags.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.TagHandler.GetTag,
		)

		tags.GET("/:id/posts",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			func(c *gin.Context) {
				c.JSON(501, gin.H{"error": "Tag posts endpoint not implemented yet"})
			},
		)
	}

	// Public search endpoints
	search := group.Group("/search")
	{
		search.GET("/posts",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GlobalSearch,
		)

		search.GET("/suggestions",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetSuggestions,
		)

		search.GET("/trending",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetTrending,
		)

		search.GET("/popular",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetPopular,
		)
	}

	config.Logger.Info("Public post routes configured successfully",
		zap.String("base_path", group.BasePath()),
		zap.Int("route_groups", 4), // posts, post-types, categories, tags, search
	)
}

// GetPostRouteInfo returns information about post-related routes
func GetPostRouteInfo() []RouteInfo {
	return []RouteInfo{
		// Post routes
		{Method: "GET", Path: "/api/v1/posts", Protected: false, Description: "List posts with filtering and pagination"},
		{Method: "GET", Path: "/api/v1/posts/:id", Protected: false, Description: "Get specific post by ID"},
		{Method: "POST", Path: "/api/v1/posts", Protected: true, Description: "Create new post"},
		{Method: "PUT", Path: "/api/v1/posts/:id", Protected: true, Description: "Update existing post"},
		{Method: "DELETE", Path: "/api/v1/posts/:id", Protected: true, Description: "Delete post"},
		{Method: "PATCH", Path: "/api/v1/posts/:id/status", Protected: true, Description: "Update post status"},
		{Method: "PATCH", Path: "/api/v1/posts/:id/schedule", Protected: true, Description: "Schedule post publication"},

		// Post type routes
		{Method: "GET", Path: "/api/v1/post-types", Protected: false, Description: "List all post types"},
		{Method: "GET", Path: "/api/v1/post-types/:id", Protected: false, Description: "Get specific post type"},
		{Method: "POST", Path: "/api/v1/post-types", Protected: true, Description: "Create new post type"},
		{Method: "PUT", Path: "/api/v1/post-types/:id", Protected: true, Description: "Update post type"},
		{Method: "DELETE", Path: "/api/v1/post-types/:id", Protected: true, Description: "Delete post type"},

		// Category routes
		{Method: "GET", Path: "/api/v1/categories", Protected: false, Description: "List categories with hierarchy"},
		{Method: "GET", Path: "/api/v1/categories/:id", Protected: false, Description: "Get specific category"},
		{Method: "GET", Path: "/api/v1/categories/:id/posts", Protected: false, Description: "Get posts in category"},
		{Method: "POST", Path: "/api/v1/categories", Protected: true, Description: "Create new category"},
		{Method: "PUT", Path: "/api/v1/categories/:id", Protected: true, Description: "Update category"},
		{Method: "DELETE", Path: "/api/v1/categories/:id", Protected: true, Description: "Delete category"},

		// Tag routes
		{Method: "GET", Path: "/api/v1/tags", Protected: false, Description: "List all tags"},
		{Method: "GET", Path: "/api/v1/tags/:id", Protected: false, Description: "Get specific tag"},
		{Method: "GET", Path: "/api/v1/tags/:id/posts", Protected: false, Description: "Get posts with tag"},
		{Method: "POST", Path: "/api/v1/tags", Protected: true, Description: "Create new tag"},
		{Method: "PUT", Path: "/api/v1/tags/:id", Protected: true, Description: "Update tag"},
		{Method: "DELETE", Path: "/api/v1/tags/:id", Protected: true, Description: "Delete tag"},

		// Search routes
		{Method: "GET", Path: "/api/v1/search/posts", Protected: false, Description: "Global post search"},
		{Method: "GET", Path: "/api/v1/search/suggestions", Protected: false, Description: "Search suggestions"},
		{Method: "GET", Path: "/api/v1/search/trending", Protected: false, Description: "Trending posts"},
		{Method: "GET", Path: "/api/v1/search/popular", Protected: false, Description: "Popular posts"},

		// Attachment routes
		{Method: "POST", Path: "/api/v1/attachments", Protected: true, Description: "Upload file attachment"},
		{Method: "GET", Path: "/api/v1/attachments/post/:postId", Protected: false, Description: "List post attachments"},
		{Method: "GET", Path: "/api/v1/attachments/:id/download", Protected: false, Description: "Download attachment"},
		{Method: "GET", Path: "/api/v1/attachments/:id", Protected: false, Description: "Get attachment metadata"},
		{Method: "PUT", Path: "/api/v1/attachments/:id", Protected: true, Description: "Update attachment metadata"},
		{Method: "DELETE", Path: "/api/v1/attachments/:id", Protected: true, Description: "Delete attachment"},

		// Bulk operation routes
		{Method: "PUT", Path: "/api/v1/bulk/posts", Protected: true, Description: "Bulk update posts"},
		{Method: "DELETE", Path: "/api/v1/bulk/posts", Protected: true, Description: "Bulk delete posts"},
	}
}

// RouteInfo represents information about a route
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Protected   bool   `json:"protected"`
	Description string `json:"description"`
}
