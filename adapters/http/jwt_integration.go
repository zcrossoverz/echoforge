package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zcrossoverz/echoforge/adapters/http/handlers"
	"github.com/zcrossoverz/echoforge/adapters/http/middleware"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// JWTIntegrationConfig holds configuration for JWT integration with post routes
type JWTIntegrationConfig struct {
	Logger                   *zap.Logger
	AuthUseCase              *usecase.UserAuthenticationUseCase
	PostHandler              *handlers.PostHandler
	PostTypeHandler          *handlers.PostTypeHandler
	CategoryHandler          *handlers.CategoryHandler
	TagHandler               *handlers.TagHandler
	SearchHandler            *handlers.SearchHandler
	AttachmentHandler        *handlers.AttachmentHandler
	BulkHandler              *handlers.BulkHandler
	PostValidationMiddleware *middleware.PostValidationMiddleware
	FileSecurityMiddleware   *middleware.FileSecurityMiddleware
	PostRateLimiter          *middleware.PostRateLimiter
}

// SetupAuthenticatedPostRoutes sets up all post routes with proper JWT authentication
func SetupAuthenticatedPostRoutes(group *gin.RouterGroup, config *JWTIntegrationConfig) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(config.AuthUseCase)

	// Create post router config with auth middleware
	postRouterConfig := &PostRouterConfig{
		Logger:                   config.Logger,
		PostHandler:              config.PostHandler,
		PostTypeHandler:          config.PostTypeHandler,
		CategoryHandler:          config.CategoryHandler,
		TagHandler:               config.TagHandler,
		SearchHandler:            config.SearchHandler,
		AttachmentHandler:        config.AttachmentHandler,
		BulkHandler:              config.BulkHandler,
		AuthMiddleware:           authMiddleware.RequireAuth(),
		PostValidationMiddleware: config.PostValidationMiddleware,
		FileSecurityMiddleware:   config.FileSecurityMiddleware,
		PostRateLimiter:          config.PostRateLimiter,
	}

	// Setup post routes with authentication
	SetupPostRoutes(group, postRouterConfig)

	config.Logger.Info("Authenticated post routes configured successfully",
		zap.String("base_path", group.BasePath()),
		zap.Bool("auth_enabled", true),
	)
}

// SetupMixedPostRoutes sets up post routes with mixed authentication (some public, some protected)
func SetupMixedPostRoutes(group *gin.RouterGroup, config *JWTIntegrationConfig) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(config.AuthUseCase)

	// Initialize middleware
	postValidation := config.PostValidationMiddleware
	fileSecurity := config.FileSecurityMiddleware
	rateLimiter := config.PostRateLimiter

	// Public post endpoints (no authentication required)
	publicPosts := group.Group("/posts")
	{
		publicPosts.GET("",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostListValidation(),
			config.PostHandler.ListPosts,
		)

		publicPosts.GET("/:id",
			rateLimiter.GeneralPostRateLimit(),
			config.PostHandler.GetPost,
		)
	}

	// Protected post endpoints (authentication required)
	protectedPosts := group.Group("/posts")
	protectedPosts.Use(authMiddleware.RequireAuth())
	{
		protectedPosts.POST("",
			rateLimiter.PostCreationRateLimit(),
			postValidation.PostCreateValidation(),
			withUserContext(config.PostHandler.CreatePost),
		)

		protectedPosts.PUT("/:id",
			rateLimiter.GeneralPostRateLimit(),
			postValidation.PostUpdateValidation(),
			withOwnershipCheck(config.PostHandler.UpdatePost),
		)

		protectedPosts.DELETE("/:id",
			rateLimiter.GeneralPostRateLimit(),
			withOwnershipCheck(config.PostHandler.DeletePost),
		)

		protectedPosts.PATCH("/:id/status",
			rateLimiter.GeneralPostRateLimit(),
			withOwnershipCheck(config.PostHandler.UpdatePost),
		)

		protectedPosts.PATCH("/:id/schedule",
			rateLimiter.GeneralPostRateLimit(),
			withOwnershipCheck(config.PostHandler.UpdatePost),
		)
	}

	// Semi-protected routes (optional authentication for enhanced features)
	semiProtectedPosts := group.Group("/posts")
	semiProtectedPosts.Use(authMiddleware.OptionalAuth())
	{
		// These endpoints work without auth but provide enhanced features when authenticated
		semiProtectedPosts.GET("/my-posts",
			rateLimiter.GeneralPostRateLimit(),
			requireAuthForEndpoint(), // This specific endpoint requires auth
			postValidation.PostListValidation(),
			withUserContext(config.PostHandler.ListPosts),
		)
	}

	// File upload routes (always require authentication)
	protectedAttachments := group.Group("/attachments")
	protectedAttachments.Use(authMiddleware.RequireAuth())
	{
		protectedAttachments.POST("",
			rateLimiter.FileUploadRateLimit(),
			fileSecurity.ValidateFileUpload(),
			fileSecurity.ValidateAttachmentMetadata(),
			fileSecurity.SetSecurityHeaders(),
			withUserContext(config.AttachmentHandler.UploadAttachment),
		)

		protectedAttachments.PUT("/:id",
			rateLimiter.GeneralPostRateLimit(),
			withOwnershipCheck(config.AttachmentHandler.UpdateAttachment),
		)

		protectedAttachments.DELETE("/:id",
			rateLimiter.GeneralPostRateLimit(),
			withOwnershipCheck(config.AttachmentHandler.DeleteAttachment),
		)
	}

	// Public attachment downloads (no auth required)
	publicAttachments := group.Group("/attachments")
	{
		publicAttachments.GET("/post/:postId",
			rateLimiter.GeneralPostRateLimit(),
			config.AttachmentHandler.ListPostAttachments,
		)

		publicAttachments.GET("/:id/download",
			rateLimiter.GeneralPostRateLimit(),
			fileSecurity.ValidateFileDownload(),
			fileSecurity.SetSecurityHeaders(),
			config.AttachmentHandler.DownloadAttachment,
		)
	}

	// Bulk operations (always require authentication and admin role)
	adminBulk := group.Group("/bulk")
	adminBulk.Use(authMiddleware.RequireAuth())
	adminBulk.Use(requireAdminRole()) // Additional admin role check
	{
		adminBulk.PUT("/posts",
			rateLimiter.BulkOperationRateLimit(),
			postValidation.BulkOperationValidation(),
			config.BulkHandler.BulkUpdatePosts,
		)

		adminBulk.DELETE("/posts",
			rateLimiter.BulkOperationRateLimit(),
			postValidation.BulkOperationValidation(),
			config.BulkHandler.BulkDeletePosts,
		)
	}

	// Search routes (public with optional auth for personalization)
	searchGroup := group.Group("/search")
	searchGroup.Use(authMiddleware.OptionalAuth())
	{
		searchGroup.GET("/posts",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GlobalSearch,
		)

		searchGroup.GET("/suggestions",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetSuggestions,
		)

		searchGroup.GET("/trending",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetTrending,
		)

		searchGroup.GET("/popular",
			rateLimiter.SearchRateLimit(),
			config.SearchHandler.GetPopular,
		)
	}

	config.Logger.Info("Mixed authentication post routes configured successfully",
		zap.String("base_path", group.BasePath()),
		zap.Bool("auth_mixed", true),
	)
}

// withUserContext injects authenticated user information into the request context
func withUserContext(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// The auth middleware already sets user info, just pass through
		// Additional user context processing can be added here if needed

		user, exists := middleware.GetCurrentUser(c)
		if exists {
			// Log user action for audit purposes
			c.Set("audit_user_id", user.ID)
			c.Set("audit_user_email", user.Email)
		}

		handler(c)
	}
}

// withOwnershipCheck adds ownership verification for resource modification
func withOwnershipCheck(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current user
		userID, exists := middleware.GetCurrentUserID(c)
		if !exists {
			c.JSON(401, gin.H{
				"error": "Authentication required for ownership verification",
			})
			c.Abort()
			return
		}

		// Store user ID for ownership verification in handlers
		c.Set("requesting_user_id", userID)

		// The actual ownership check should be performed in the handler
		// as it requires database access to verify the resource owner
		handler(c)
	}
}

// requireAuthForEndpoint forces authentication for endpoints that should require it
func requireAuthForEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		_, exists := middleware.GetCurrentUser(c)
		if !exists {
			c.JSON(401, gin.H{
				"error": "Authentication required for this endpoint",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// requireAdminRole checks if the authenticated user has admin privileges
func requireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current user
		user, exists := middleware.GetCurrentUser(c)
		if !exists {
			c.JSON(401, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		// Check if user has admin role (placeholder - actual role checking depends on user domain)
		// This is a simplified check - in production, you'd check against user roles/permissions
		isAdmin := checkUserAdminRole(user)
		if !isAdmin {
			c.JSON(403, gin.H{
				"error": "Admin privileges required",
				"code":  "INSUFFICIENT_PRIVILEGES",
			})
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

// checkUserAdminRole checks if a user has admin privileges
func checkUserAdminRole(user *usecase.UserDTO) bool {
	// Placeholder implementation
	// In a real application, this would check the user's roles/permissions
	// For now, return false as we don't have role system implemented
	return false
}

// SetupAuthRouteDocumentation provides documentation for authenticated routes
func SetupAuthRouteDocumentation(group *gin.RouterGroup) {
	// Documentation endpoint for authenticated routes
	group.GET("/auth-info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"authentication": gin.H{
				"type":   "Bearer JWT",
				"header": "Authorization: Bearer <token>",
				"endpoints": gin.H{
					"login":    "POST /api/v1/auth/login",
					"register": "POST /api/v1/auth/register",
					"logout":   "POST /api/v1/auth/logout",
					"profile":  "GET /api/v1/auth/profile",
				},
			},
			"protected_routes": []string{
				"POST /api/v1/posts",
				"PUT /api/v1/posts/:id",
				"DELETE /api/v1/posts/:id",
				"PATCH /api/v1/posts/:id/status",
				"PATCH /api/v1/posts/:id/schedule",
				"POST /api/v1/attachments",
				"PUT /api/v1/attachments/:id",
				"DELETE /api/v1/attachments/:id",
				"PUT /api/v1/bulk/posts",
				"DELETE /api/v1/bulk/posts",
			},
			"public_routes": []string{
				"GET /api/v1/posts",
				"GET /api/v1/posts/:id",
				"GET /api/v1/post-types",
				"GET /api/v1/categories",
				"GET /api/v1/tags",
				"GET /api/v1/search/*",
				"GET /api/v1/attachments/post/:postId",
				"GET /api/v1/attachments/:id/download",
			},
			"optional_auth_routes": []string{
				"GET /api/v1/search/* (enhanced features when authenticated)",
			},
		})
	})
}

// GetAuthenticationStatus returns current authentication status
func GetAuthenticationStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := middleware.GetCurrentUser(c)
		if !exists {
			c.JSON(200, gin.H{
				"authenticated": false,
				"user":          nil,
			})
			return
		}

		c.JSON(200, gin.H{
			"authenticated": true,
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"created_at": user.CreatedAt,
			},
		})
	}
}
