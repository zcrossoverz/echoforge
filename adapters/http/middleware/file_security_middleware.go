package middleware

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// FileSecurityMiddleware provides file upload security validation
type FileSecurityMiddleware struct {
	maxFileSize       int64
	maxFiles          int
	allowedExtensions map[string]bool
	allowedMimeTypes  map[string]bool
	blockedKeywords   []string
}

// NewFileSecurityMiddleware creates a new file security middleware
func NewFileSecurityMiddleware() *FileSecurityMiddleware {
	return &FileSecurityMiddleware{
		maxFileSize: 100 * 1024 * 1024, // 100MB
		maxFiles:    10,                // Maximum 10 files per upload
		allowedExtensions: map[string]bool{
			// Images
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
			".bmp": true, ".svg": true, ".ico": true,
			// Documents
			".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
			".ppt": true, ".pptx": true, ".txt": true, ".rtf": true, ".odt": true,
			// Archives
			".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
			// Media
			".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true,
			".mp3": true, ".wav": true, ".ogg": true, ".flac": true,
			// Code/Text
			".json": true, ".xml": true, ".csv": true, ".yaml": true, ".yml": true,
			".md": true, ".log": true,
		},
		allowedMimeTypes: map[string]bool{
			// Images
			"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
			"image/bmp": true, "image/svg+xml": true, "image/x-icon": true,
			// Documents
			"application/pdf":    true,
			"application/msword": true,
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
			"application/vnd.ms-excel": true,
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
			"application/vnd.ms-powerpoint":                                             true,
			"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
			"text/plain": true, "application/rtf": true,
			"application/vnd.oasis.opendocument.text": true,
			// Archives
			"application/zip": true, "application/x-rar-compressed": true,
			"application/x-7z-compressed": true, "application/x-tar": true,
			"application/gzip": true,
			// Media
			"video/mp4": true, "video/x-msvideo": true, "video/quicktime": true,
			"video/x-ms-wmv": true, "video/x-flv": true,
			"audio/mpeg": true, "audio/wav": true, "audio/ogg": true, "audio/flac": true,
			// Code/Text
			"application/json": true, "application/xml": true, "text/xml": true,
			"text/csv": true, "application/x-yaml": true, "text/yaml": true,
			"text/markdown": true, "text/x-log": true,
		},
		blockedKeywords: []string{
			// Script extensions
			".exe", ".bat", ".cmd", ".com", ".scr", ".vbs", ".js", ".jar",
			".php", ".asp", ".aspx", ".jsp", ".py", ".rb", ".pl", ".sh",
			// System files
			".dll", ".sys", ".ini", ".cfg", ".conf",
			// Potentially dangerous
			".htaccess", ".htpasswd",
		},
	}
}

// ValidateFileUpload validates file upload requests
func (fsm *FileSecurityMiddleware) ValidateFileUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is a multipart form
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content-Type must be multipart/form-data for file uploads",
			})
			c.Abort()
			return
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to parse multipart form",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Get uploaded files
		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No files provided",
			})
			c.Abort()
			return
		}

		// Check file count limit
		if len(files) > fsm.maxFiles {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Too many files. Maximum %d files allowed", fsm.maxFiles),
			})
			c.Abort()
			return
		}

		// Validate each file
		var validationErrors []string
		var totalSize int64

		for i, fileHeader := range files {
			// Validate individual file
			if errors := fsm.validateSingleFile(fileHeader, i+1); len(errors) > 0 {
				validationErrors = append(validationErrors, errors...)
			}
			totalSize += fileHeader.Size
		}

		// Check total size limit
		if totalSize > fsm.maxFileSize*int64(len(files)) {
			validationErrors = append(validationErrors, fmt.Sprintf("Total file size exceeds limit of %d MB", (fsm.maxFileSize*int64(len(files)))/(1024*1024)))
		}

		// Return validation errors if any
		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "File validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated files in context
		c.Set("validated_files", files)
		c.Set("validated_form", form)
		c.Next()
	}
}

// validateSingleFile validates a single uploaded file
func (fsm *FileSecurityMiddleware) validateSingleFile(fileHeader *multipart.FileHeader, fileIndex int) []string {
	var errors []string
	filename := fileHeader.Filename

	// Check file size
	if fileHeader.Size > fsm.maxFileSize {
		errors = append(errors, fmt.Sprintf("File %d (%s): exceeds maximum size of %d MB",
			fileIndex, filename, fsm.maxFileSize/(1024*1024)))
	}

	// Check for empty files
	if fileHeader.Size == 0 {
		errors = append(errors, fmt.Sprintf("File %d (%s): empty files are not allowed", fileIndex, filename))
	}

	// Validate filename
	if filename == "" {
		errors = append(errors, fmt.Sprintf("File %d: filename is required", fileIndex))
		return errors
	}

	// Check for dangerous filename patterns
	if fsm.isDangerousFilename(filename) {
		errors = append(errors, fmt.Sprintf("File %d (%s): dangerous filename detected", fileIndex, filename))
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		errors = append(errors, fmt.Sprintf("File %d (%s): file extension is required", fileIndex, filename))
	} else if !fsm.allowedExtensions[ext] {
		errors = append(errors, fmt.Sprintf("File %d (%s): file extension '%s' is not allowed", fileIndex, filename, ext))
	}

	// Check for blocked keywords in filename
	lowerFilename := strings.ToLower(filename)
	for _, keyword := range fsm.blockedKeywords {
		if strings.Contains(lowerFilename, keyword) {
			errors = append(errors, fmt.Sprintf("File %d (%s): contains blocked keyword '%s'", fileIndex, filename, keyword))
		}
	}

	// Validate MIME type by opening the file
	file, err := fileHeader.Open()
	if err != nil {
		errors = append(errors, fmt.Sprintf("File %d (%s): failed to open file for validation", fileIndex, filename))
		return errors
	}
	defer file.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		errors = append(errors, fmt.Sprintf("File %d (%s): failed to read file content", fileIndex, filename))
		return errors
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer[:n])
	if !fsm.allowedMimeTypes[mimeType] {
		errors = append(errors, fmt.Sprintf("File %d (%s): MIME type '%s' is not allowed", fileIndex, filename, mimeType))
	}

	// Additional security checks
	if fsm.containsSuspiciousContent(buffer[:n]) {
		errors = append(errors, fmt.Sprintf("File %d (%s): contains suspicious content", fileIndex, filename))
	}

	return errors
}

// isDangerousFilename checks for dangerous filename patterns
func (fsm *FileSecurityMiddleware) isDangerousFilename(filename string) bool {
	// Check for path traversal attempts
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return true
	}

	// Check for reserved names (Windows)
	reservedNames := []string{
		"con", "prn", "aux", "nul",
		"com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9",
		"lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9",
	}

	baseFilename := strings.ToLower(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, reserved := range reservedNames {
		if baseFilename == reserved {
			return true
		}
	}

	// Check for control characters
	for _, char := range filename {
		if char < 32 && char != 9 && char != 10 && char != 13 { // Allow tab, LF, CR
			return true
		}
	}

	return false
}

// containsSuspiciousContent checks for suspicious content in file headers
func (fsm *FileSecurityMiddleware) containsSuspiciousContent(content []byte) bool {
	// Convert to string for pattern matching
	contentStr := strings.ToLower(string(content))

	// Check for script tags and suspicious patterns
	suspiciousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"<?php", "<%", "<jsp:", "#!/bin/", "#!/usr/bin/",
		"cmd.exe", "powershell", "bash",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(contentStr, pattern) {
			return true
		}
	}

	return false
}

// ValidateAttachmentMetadata validates attachment metadata in request body
func (fsm *FileSecurityMiddleware) ValidateAttachmentMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PostID      string                 `json:"postId"`
			Description string                 `json:"description,omitempty"`
			AltText     string                 `json:"altText,omitempty"`
			Metadata    map[string]interface{} `json:"metadata,omitempty"`
		}

		// For multipart forms, parse form values
		if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			req.PostID = c.PostForm("postId")
			req.Description = c.PostForm("description")
			req.AltText = c.PostForm("altText")
		} else {
			// For JSON requests
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid request format",
					"details": err.Error(),
				})
				c.Abort()
				return
			}
		}

		// Validate required fields
		if strings.TrimSpace(req.PostID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Post ID is required",
			})
			c.Abort()
			return
		}

		// Validate description length
		if len(req.Description) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Description cannot exceed 1000 characters",
			})
			c.Abort()
			return
		}

		// Validate alt text length
		if len(req.AltText) > 255 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Alt text cannot exceed 255 characters",
			})
			c.Abort()
			return
		}

		// Store validated metadata in context
		c.Set("validated_metadata", req)
		c.Next()
	}
}

// SetSecurityHeaders sets security-related headers for file operations
func (fsm *FileSecurityMiddleware) SetSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent embedding in frames
		c.Header("X-Frame-Options", "DENY")

		// XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy for file operations
		c.Header("Content-Security-Policy", "default-src 'none'; script-src 'none'; object-src 'none';")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// ValidateFileDownload validates file download requests
func (fsm *FileSecurityMiddleware) ValidateFileDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		attachmentID := c.Param("id")
		if attachmentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Attachment ID is required",
			})
			c.Abort()
			return
		}

		// Validate UUID format
		if len(attachmentID) != 36 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid attachment ID format",
			})
			c.Abort()
			return
		}

		// Check for path traversal in any query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if strings.Contains(value, "..") || strings.Contains(value, "/") || strings.Contains(value, "\\") {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("Invalid characters in parameter '%s'", key),
					})
					c.Abort()
					return
				}
			}
		}

		c.Set("validated_attachment_id", attachmentID)
		c.Next()
	}
}
