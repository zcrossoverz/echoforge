# Blog Site Extension Guide

## Overview

This guide demonstrates how to extend Echoforge to create a multi-author blogging platform with features like post management, categories, tags, comments, and SEO optimization, all while maintaining multi-tenant isolation.

## Prerequisites

- Go 1.25+ installed
- PostgreSQL 16+ database
- Understanding of Echoforge's hexagonal architecture
- Familiarity with GORM and Gin frameworks

## Setup Process

### Step 1: Clone and Configure Base

```bash
# Clone the Echoforge repository
git clone https://github.com/zcrossoverz/echoforge.git
cd echoforge

# Install dependencies
go mod download

# Create your blog database
createdb blog_platform_db
```

### Step 2: Create Site Configuration

Create a blog-specific configuration file:

```yaml
# config/blog-site.yaml
site:
  id: "blog-001"
  name: "Personal Blog Platform"
  description: "Multi-author blogging platform with rich content management"
  type: "blog"

database:
  dsn: "postgres://username:password@localhost/blog_platform_db?sslmode=disable"
  max_open_conns: 30
  max_idle_conns: 15

features:
  comments: true
  social_sharing: true
  newsletter: true
  analytics: true
  seo_optimization: true
  content_scheduling: true

blog_specific:
  max_post_size_mb: 10
  allowed_file_types: ["jpg", "png", "gif", "pdf", "doc", "docx"]
  content_formats: ["markdown", "html", "plain_text"]
  moderation_enabled: true
  auto_excerpt_length: 150

authentication:
  jwt_secret: "${JWT_SECRET}"
  session_duration: "7d"
  rate_limit:
    requests_per_minute: 100
    burst: 20

seo:
  meta_description_length: 160
  slug_max_length: 100
  auto_generate_sitemap: true
  canonical_urls: true

performance:
  cache_duration: "2h"
  enable_compression: true
  max_concurrent_requests: 50
```

### Step 3: Extend Domain Models

Create blog-specific domain entities:

```go
// internal/domain/blog/post.go
package blog

import (
	"time"
	"github.com/google/uuid"
)

type Post struct {
	ID              uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID          string       `gorm:"not null;index" json:"site_id"`
	AuthorID        uuid.UUID    `gorm:"not null;index" json:"author_id"`
	Title           string       `gorm:"not null" json:"title"`
	Slug            string       `gorm:"not null;unique_index" json:"slug"`
	Excerpt         string       `json:"excerpt"`
	Content         string       `gorm:"type:text" json:"content"`
	ContentFormat   ContentFormat `gorm:"default:'markdown'" json:"content_format"`
	Status          PostStatus   `gorm:"default:'draft'" json:"status"`
	FeaturedImage   string       `json:"featured_image"`
	MetaTitle       string       `json:"meta_title"`
	MetaDescription string       `json:"meta_description"`
	PublishedAt     *time.Time   `json:"published_at"`
	ScheduledAt     *time.Time   `json:"scheduled_at"`
	ViewCount       int64        `gorm:"default:0" json:"view_count"`
	Categories      []Category   `gorm:"many2many:post_categories" json:"categories"`
	Tags            []Tag        `gorm:"many2many:post_tags" json:"tags"`
	Comments        []Comment    `json:"comments"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type ContentFormat string

const (
	FormatMarkdown  ContentFormat = "markdown"
	FormatHTML      ContentFormat = "html"
	FormatPlainText ContentFormat = "plain_text"
)

type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusPublished PostStatus = "published"
	StatusScheduled PostStatus = "scheduled"
	StatusArchived  PostStatus = "archived"
)

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID      string    `gorm:"not null;index" json:"site_id"`
	Name        string    `gorm:"not null" json:"name"`
	Slug        string    `gorm:"not null;unique_index" json:"slug"`
	Description string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Children    []Category `gorm:"foreignkey:ParentID" json:"children"`
	PostCount   int64     `gorm:"default:0" json:"post_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID    string    `gorm:"not null;index" json:"site_id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"not null;unique_index" json:"slug"`
	Color     string    `json:"color"`
	PostCount int64     `gorm:"default:0" json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"not null;index" json:"post_id"`
	AuthorID  *uuid.UUID `json:"author_id"`
	AuthorName string    `json:"author_name"`
	AuthorEmail string   `json:"author_email"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Status    CommentStatus `gorm:"default:'pending'" json:"status"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Children  []Comment `gorm:"foreignkey:ParentID" json:"children"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentStatus string

const (
	CommentPending  CommentStatus = "pending"
	CommentApproved CommentStatus = "approved"
	CommentSpam     CommentStatus = "spam"
	CommentDeleted  CommentStatus = "deleted"
)

type Author struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID      string    `gorm:"not null;index" json:"site_id"`
	Name        string    `gorm:"not null" json:"name"`
	Email       string    `gorm:"not null;unique_index" json:"email"`
	Bio         string    `json:"bio"`
	Avatar      string    `json:"avatar"`
	SocialLinks map[string]string `gorm:"type:jsonb" json:"social_links"`
	Role        AuthorRole `gorm:"default:'author'" json:"role"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	PostCount   int64     `gorm:"default:0" json:"post_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AuthorRole string

const (
	RoleAdmin      AuthorRole = "admin"
	RoleEditor     AuthorRole = "editor"
	RoleAuthor     AuthorRole = "author"
	RoleContributor AuthorRole = "contributor"
)
```

### Step 4: Multi-Tenant Repository Implementation

```go
// internal/adapters/persistence/blog_repository.go
package persistence

import (
	"fmt"
	"time"
	"github.com/zcrossoverz/echoforge/internal/domain/blog"
	"gorm.io/gorm"
)

type BlogRepository struct {
	db     *gorm.DB
	siteID string
}

func NewBlogRepository(db *gorm.DB, siteID string) *BlogRepository {
	return &BlogRepository{
		db:     db,
		siteID: siteID,
	}
}

// Post operations
func (r *BlogRepository) CreatePost(post *blog.Post) error {
	post.SiteID = r.siteID
	
	// Generate slug if not provided
	if post.Slug == "" {
		post.Slug = r.generateSlug(post.Title)
	}
	
	// Auto-generate excerpt if not provided
	if post.Excerpt == "" {
		post.Excerpt = r.generateExcerpt(post.Content, 150)
	}
	
	return r.db.Create(post).Error
}

func (r *BlogRepository) GetPostBySlug(slug string) (*blog.Post, error) {
	var post blog.Post
	err := r.db.Where("slug = ? AND site_id = ?", slug, r.siteID).
		Preload("Categories").
		Preload("Tags").
		Preload("Comments", "status = ?", blog.CommentApproved).
		First(&post).Error
	return &post, err
}

func (r *BlogRepository) ListPosts(status blog.PostStatus, limit, offset int) ([]blog.Post, error) {
	var posts []blog.Post
	query := r.db.Where("site_id = ?", r.siteID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Preload("Categories").
		Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	
	return posts, err
}

func (r *BlogRepository) UpdatePost(post *blog.Post) error {
	// Ensure the post belongs to this site
	existingPost, err := r.GetPostByID(post.ID.String())
	if err != nil {
		return err
	}
	
	if existingPost.SiteID != r.siteID {
		return fmt.Errorf("post does not belong to site %s", r.siteID)
	}
	
	post.SiteID = r.siteID
	return r.db.Save(post).Error
}

func (r *BlogRepository) DeletePost(id string) error {
	result := r.db.Where("id = ? AND site_id = ?", id, r.siteID).Delete(&blog.Post{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// Category operations
func (r *BlogRepository) CreateCategory(category *blog.Category) error {
	category.SiteID = r.siteID
	
	if category.Slug == "" {
		category.Slug = r.generateSlug(category.Name)
	}
	
	return r.db.Create(category).Error
}

func (r *BlogRepository) ListCategories() ([]blog.Category, error) {
	var categories []blog.Category
	err := r.db.Where("site_id = ?", r.siteID).
		Order("name ASC").
		Find(&categories).Error
	return categories, err
}

// Tag operations
func (r *BlogRepository) CreateTag(tag *blog.Tag) error {
	tag.SiteID = r.siteID
	
	if tag.Slug == "" {
		tag.Slug = r.generateSlug(tag.Name)
	}
	
	return r.db.Create(tag).Error
}

func (r *BlogRepository) ListTags() ([]blog.Tag, error) {
	var tags []blog.Tag
	err := r.db.Where("site_id = ?", r.siteID).
		Order("name ASC").
		Find(&tags).Error
	return tags, err
}

// Comment operations
func (r *BlogRepository) CreateComment(comment *blog.Comment) error {
	// Verify the post exists and belongs to this site
	var post blog.Post
	if err := r.db.Where("id = ? AND site_id = ?", comment.PostID, r.siteID).First(&post).Error; err != nil {
		return fmt.Errorf("post not found or doesn't belong to this site")
	}
	
	return r.db.Create(comment).Error
}

func (r *BlogRepository) GetCommentsForPost(postID string, status blog.CommentStatus) ([]blog.Comment, error) {
	var comments []blog.Comment
	query := r.db.Where("post_id = ?", postID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// Verify post belongs to this site
	var post blog.Post
	if err := r.db.Where("id = ? AND site_id = ?", postID, r.siteID).First(&post).Error; err != nil {
		return nil, fmt.Errorf("post not found or doesn't belong to this site")
	}
	
	err := query.Order("created_at ASC").Find(&comments).Error
	return comments, err
}

// Author operations
func (r *BlogRepository) CreateAuthor(author *blog.Author) error {
	author.SiteID = r.siteID
	return r.db.Create(author).Error
}

func (r *BlogRepository) GetAuthorByEmail(email string) (*blog.Author, error) {
	var author blog.Author
	err := r.db.Where("email = ? AND site_id = ?", email, r.siteID).First(&author).Error
	return &author, err
}

// Utility functions
func (r *BlogRepository) generateSlug(text string) string {
	// Implement slug generation logic
	// This is a simplified version - use a proper slug library in production
	slug := strings.ToLower(text)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Ensure uniqueness by checking database
	originalSlug := slug
	counter := 1
	for {
		var count int64
		r.db.Model(&blog.Post{}).Where("slug = ? AND site_id = ?", slug, r.siteID).Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}
	
	return slug
}

func (r *BlogRepository) generateExcerpt(content string, length int) string {
	// Strip HTML/Markdown and create excerpt
	text := stripMarkdown(content)
	if len(text) <= length {
		return text
	}
	
	// Find the last space before the length limit
	excerpt := text[:length]
	if lastSpace := strings.LastIndex(excerpt, " "); lastSpace > 0 {
		excerpt = excerpt[:lastSpace]
	}
	
	return excerpt + "..."
}
```

### Step 5: API Endpoints

```go
// adapters/http/blog_handler.go
package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcrossoverz/echoforge/internal/domain/blog"
)

type BlogHandler struct {
	repo blog.Repository
}

func NewBlogHandler(repo blog.Repository) *BlogHandler {
	return &BlogHandler{repo: repo}
}

func (h *BlogHandler) RegisterRoutes(r *gin.RouterGroup) {
	blog := r.Group("/blog")
	{
		// Post routes
		blog.GET("/posts", h.ListPosts)
		blog.POST("/posts", h.CreatePost)
		blog.GET("/posts/:slug", h.GetPost)
		blog.PUT("/posts/:id", h.UpdatePost)
		blog.DELETE("/posts/:id", h.DeletePost)
		
		// Category routes
		blog.GET("/categories", h.ListCategories)
		blog.POST("/categories", h.CreateCategory)
		
		// Tag routes
		blog.GET("/tags", h.ListTags)
		blog.POST("/tags", h.CreateTag)
		
		// Comment routes
		blog.GET("/posts/:slug/comments", h.GetComments)
		blog.POST("/posts/:slug/comments", h.CreateComment)
		
		// Author routes
		blog.GET("/authors", h.ListAuthors)
		blog.POST("/authors", h.CreateAuthor)
	}
}

func (h *BlogHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	
	status := blog.PostStatus(c.DefaultQuery("status", "published"))
	
	posts, err := h.repo.ListPosts(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"page":  page,
		"limit": limit,
	})
}

func (h *BlogHandler) CreatePost(c *gin.Context) {
	var post blog.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Set published_at if status is published
	if post.Status == blog.StatusPublished && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}
	
	if err := h.repo.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, post)
}

func (h *BlogHandler) GetPost(c *gin.Context) {
	slug := c.Param("slug")
	
	post, err := h.repo.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	
	// Increment view count
	h.repo.IncrementViewCount(post.ID.String())
	
	c.JSON(http.StatusOK, post)
}

func (h *BlogHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	
	var post blog.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parse ID
	postID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}
	post.ID = postID
	
	if err := h.repo.UpdatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, post)
}

func (h *BlogHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.repo.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

func (h *BlogHandler) CreateComment(c *gin.Context) {
	slug := c.Param("slug")
	
	// Get post by slug to get ID
	post, err := h.repo.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	
	var comment blog.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	comment.PostID = post.ID
	
	if err := h.repo.CreateComment(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, comment)
}

func (h *BlogHandler) GetComments(c *gin.Context) {
	slug := c.Param("slug")
	
	// Get post by slug to get ID
	post, err := h.repo.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	
	comments, err := h.repo.GetCommentsForPost(post.ID.String(), blog.CommentApproved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
```

### Step 6: Database Migrations

```sql
-- migrations/008_create_blog_tables.up.sql
CREATE TABLE IF NOT EXISTS authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar VARCHAR(500),
    social_links JSONB,
    role VARCHAR(50) DEFAULT 'author',
    is_active BOOLEAN DEFAULT true,
    post_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_authors_site_id ON authors(site_id);
CREATE UNIQUE INDEX idx_authors_site_email ON authors(site_id, email);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    post_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_categories_site_id ON categories(site_id);
CREATE UNIQUE INDEX idx_categories_site_slug ON categories(site_id, slug);

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    color VARCHAR(7),
    post_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tags_site_id ON tags(site_id);
CREATE UNIQUE INDEX idx_tags_site_slug ON tags(site_id, slug);

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    excerpt TEXT,
    content TEXT,
    content_format VARCHAR(50) DEFAULT 'markdown',
    status VARCHAR(50) DEFAULT 'draft',
    featured_image VARCHAR(500),
    meta_title VARCHAR(255),
    meta_description VARCHAR(500),
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_posts_site_id ON posts(site_id);
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_published_at ON posts(published_at DESC);
CREATE UNIQUE INDEX idx_posts_site_slug ON posts(site_id, slug);

CREATE TABLE IF NOT EXISTS post_categories (
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);

CREATE TABLE IF NOT EXISTS post_tags (
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID REFERENCES authors(id) ON DELETE SET NULL,
    author_name VARCHAR(255),
    author_email VARCHAR(255),
    content TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_status ON comments(status);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
```

### Step 7: SEO and Performance Features

```go
// internal/domain/blog/seo.go
package blog

type SEOService struct {
	repo Repository
}

func NewSEOService(repo Repository) *SEOService {
	return &SEOService{repo: repo}
}

func (s *SEOService) GenerateSitemap(siteID string) (string, error) {
	posts, err := s.repo.ListPosts(StatusPublished, 1000, 0)
	if err != nil {
		return "", err
	}
	
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	
	for _, post := range posts {
		sitemap += fmt.Sprintf(`
	<url>
		<loc>https://yourdomain.com/blog/%s</loc>
		<lastmod>%s</lastmod>
		<changefreq>monthly</changefreq>
		<priority>0.8</priority>
	</url>`, post.Slug, post.UpdatedAt.Format("2006-01-02"))
	}
	
	sitemap += `
</urlset>`
	
	return sitemap, nil
}

func (s *SEOService) OptimizePost(post *Post) {
	// Auto-generate meta title if not provided
	if post.MetaTitle == "" {
		post.MetaTitle = post.Title
		if len(post.MetaTitle) > 60 {
			post.MetaTitle = post.MetaTitle[:57] + "..."
		}
	}
	
	// Auto-generate meta description if not provided
	if post.MetaDescription == "" {
		post.MetaDescription = post.Excerpt
		if len(post.MetaDescription) > 160 {
			post.MetaDescription = post.MetaDescription[:157] + "..."
		}
	}
}
```

### Step 8: Testing Your Blog Site

```bash
# Run database migrations
migrate -path migrations -database "postgres://username:password@localhost/blog_platform_db?sslmode=disable" up

# Start your blog site
go run cmd/server/main.go --config config/blog-site.yaml

# Create an author
curl -X POST http://localhost:8080/api/v1/blog/authors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "bio": "Tech blogger and software developer",
    "role": "author"
  }'

# Create a blog post
curl -X POST http://localhost:8080/api/v1/blog/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Getting Started with Go",
    "content": "# Introduction\n\nGo is a powerful programming language...",
    "content_format": "markdown",
    "status": "published",
    "author_id": "author-uuid-here"
  }'

# List published posts
curl http://localhost:8080/api/v1/blog/posts

# Get a specific post
curl http://localhost:8080/api/v1/blog/posts/getting-started-with-go
```

## Advanced Features

### Content Scheduling

```go
func (r *BlogRepository) PublishScheduledPosts() error {
	now := time.Now()
	
	var posts []blog.Post
	err := r.db.Where("status = ? AND scheduled_at <= ? AND site_id = ?", 
		blog.StatusScheduled, now, r.siteID).Find(&posts).Error
	if err != nil {
		return err
	}
	
	for _, post := range posts {
		post.Status = blog.StatusPublished
		post.PublishedAt = &now
		r.db.Save(&post)
	}
	
	return nil
}
```

### Newsletter Integration

```go
type NewsletterService struct {
	emailService EmailService
	repo         Repository
}

func (n *NewsletterService) SendNewPostNotification(post *blog.Post) error {
	subscribers, err := n.getSubscribers()
	if err != nil {
		return err
	}
	
	for _, subscriber := range subscribers {
		email := &Email{
			To:      subscriber.Email,
			Subject: fmt.Sprintf("New Post: %s", post.Title),
			Body:    n.generateEmailTemplate(post),
		}
		
		n.emailService.Send(email)
	}
	
	return nil
}
```

## Performance Optimization

### Caching Strategy

```go
type CachedBlogRepository struct {
	*BlogRepository
	cache Cache
}

func (r *CachedBlogRepository) GetPostBySlug(slug string) (*blog.Post, error) {
	cacheKey := fmt.Sprintf("post:%s:%s", r.siteID, slug)
	
	// Try cache first
	if cached, err := r.cache.Get(cacheKey); err == nil {
		var post blog.Post
		json.Unmarshal(cached, &post)
		return &post, nil
	}
	
	// Fallback to database
	post, err := r.BlogRepository.GetPostBySlug(slug)
	if err != nil {
		return nil, err
	}
	
	// Cache for future requests
	if data, err := json.Marshal(post); err == nil {
		r.cache.Set(cacheKey, data, 2*time.Hour)
	}
	
	return post, nil
}
```

### Database Optimization

```sql
-- Additional indexes for performance
CREATE INDEX idx_posts_site_status_published ON posts(site_id, status, published_at DESC) WHERE status = 'published';
CREATE INDEX idx_comments_post_status ON comments(post_id, status) WHERE status = 'approved';
CREATE INDEX idx_posts_author_published ON posts(author_id, published_at DESC) WHERE status = 'published';
```

## Security Considerations

### Content Sanitization

```go
import "github.com/microcosm-cc/bluemonday"

func sanitizeContent(content string, format blog.ContentFormat) string {
	switch format {
	case blog.FormatHTML:
		p := bluemonday.UGCPolicy()
		return p.Sanitize(content)
	case blog.FormatMarkdown:
		// Convert markdown to HTML then sanitize
		html := convertMarkdownToHTML(content)
		p := bluemonday.UGCPolicy()
		return p.Sanitize(html)
	default:
		return html.EscapeString(content)
	}
}
```

### Anti-Spam Measures

```go
func (r *BlogRepository) CreateComment(comment *blog.Comment) error {
	// Basic spam detection
	if r.isSpam(comment) {
		comment.Status = blog.CommentSpam
	}
	
	return r.db.Create(comment).Error
}

func (r *BlogRepository) isSpam(comment *blog.Comment) bool {
	// Simple spam detection rules
	spamKeywords := []string{"viagra", "casino", "loan", "bitcoin"}
	
	content := strings.ToLower(comment.Content)
	for _, keyword := range spamKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	
	// Check for excessive links
	linkCount := strings.Count(content, "http")
	return linkCount > 3
}
```

## Deployment with Docker

```dockerfile
# Dockerfile.blog
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o blog-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/blog-server .
COPY --from=builder /app/config/blog-site.yaml ./config/

CMD ["./blog-server", "--config", "config/blog-site.yaml"]
```

## Next Steps

1. **Rich Text Editor Integration**
2. **Media Management System**
3. **Advanced SEO Features**
4. **Analytics Dashboard**
5. **Email Newsletter System**
6. **Social Media Integration**
7. **Multi-language Support**
8. **Theme Customization**

This blog platform demonstrates Echoforge's ability to create content-rich applications while maintaining proper architecture and multi-tenant isolation.