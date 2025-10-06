# Portfolio Site Extension Guide

## Overview

This guide demonstrates how to extend Echoforge to create a professional portfolio platform for creatives, developers, and professionals to showcase their work, skills, and achievements with multi-tenant capabilities.

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

# Create your portfolio database
createdb portfolio_platform_db
```

### Step 2: Create Site Configuration

Create a portfolio-specific configuration file:

```yaml
# config/portfolio-site.yaml
site:
  id: "portfolio-001"
  name: "Creative Portfolio Platform"
  description: "Professional portfolio platform for showcasing creative work"
  type: "portfolio"

database:
  dsn: "postgres://username:password@localhost/portfolio_platform_db?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 10

features:
  contact_form: true
  testimonials: true
  analytics: true
  seo_optimization: true
  social_integration: true
  resume_builder: true
  project_galleries: true

portfolio_specific:
  max_project_images: 20
  max_file_size_mb: 15
  allowed_file_types: ["jpg", "jpeg", "png", "gif", "pdf", "mp4", "webm", "sketch", "psd", "ai"]
  thumbnail_sizes: [150, 300, 600, 1200]
  enable_3d_viewer: true
  project_visibility: ["public", "private", "password_protected"]

authentication:
  jwt_secret: "${JWT_SECRET}"
  session_duration: "30d"
  rate_limit:
    requests_per_minute: 60
    burst: 15

seo:
  meta_description_length: 160
  enable_schema_markup: true
  auto_generate_sitemap: true
  open_graph_enabled: true

performance:
  cache_duration: "4h"
  enable_compression: true
  image_optimization: true
  lazy_loading: true
  max_concurrent_requests: 30
```

### Step 3: Extend Domain Models

Create portfolio-specific domain entities:

```go
// internal/domain/portfolio/project.go
package portfolio

import (
	"time"
	"github.com/google/uuid"
)

type Project struct {
	ID             uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID         string        `gorm:"not null;index" json:"site_id"`
	OwnerID        uuid.UUID     `gorm:"not null;index" json:"owner_id"`
	Title          string        `gorm:"not null" json:"title"`
	Slug           string        `gorm:"not null;unique_index" json:"slug"`
	Description    string        `json:"description"`
	LongDescription string       `gorm:"type:text" json:"long_description"`
	Category       string        `gorm:"not null;index" json:"category"`
	Status         ProjectStatus `gorm:"default:'draft'" json:"status"`
	FeaturedImage  string        `json:"featured_image"`
	ThumbnailImage string        `json:"thumbnail_image"`
	Technologies   []string      `gorm:"type:text[]" json:"technologies"`
	Skills         []string      `gorm:"type:text[]" json:"skills"`
	ClientName     string        `json:"client_name"`
	ProjectURL     string        `json:"project_url"`
	GitHubURL      string        `json:"github_url"`
	DemoURL        string        `json:"demo_url"`
	StartDate      *time.Time    `json:"start_date"`
	EndDate        *time.Time    `json:"end_date"`
	Duration       string        `json:"duration"`
	Priority       int           `gorm:"default:0" json:"priority"`
	ViewCount      int64         `gorm:"default:0" json:"view_count"`
	IsFeatured     bool          `gorm:"default:false" json:"is_featured"`
	Images         []ProjectImage `json:"images"`
	Videos         []ProjectVideo `json:"videos"`
	Files          []ProjectFile  `json:"files"`
	Tags           []Tag         `gorm:"many2many:project_tags" json:"tags"`
	MetaTitle      string        `json:"meta_title"`
	MetaDescription string       `json:"meta_description"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type ProjectStatus string

const (
	StatusDraft     ProjectStatus = "draft"
	StatusPublished ProjectStatus = "published"
	StatusArchived  ProjectStatus = "archived"
	StatusPrivate   ProjectStatus = "private"
)

type ProjectImage struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"not null;index" json:"project_id"`
	FileName    string    `gorm:"not null" json:"file_name"`
	OriginalURL string    `gorm:"not null" json:"original_url"`
	ThumbnailURL string   `json:"thumbnail_url"`
	MediumURL   string    `json:"medium_url"`
	LargeURL    string    `json:"large_url"`
	Caption     string    `json:"caption"`
	AltText     string    `json:"alt_text"`
	FileSize    int64     `json:"file_size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectVideo struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"not null;index" json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	VideoURL    string    `gorm:"not null" json:"video_url"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Duration    int       `json:"duration"` // in seconds
	FileSize    int64     `json:"file_size"`
	VideoType   string    `json:"video_type"` // mp4, webm, youtube, vimeo
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectFile struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"not null;index" json:"project_id"`
	FileName    string    `gorm:"not null" json:"file_name"`
	FileURL     string    `gorm:"not null" json:"file_url"`
	FileType    string    `gorm:"not null" json:"file_type"`
	FileSize    int64     `json:"file_size"`
	Description string    `json:"description"`
	IsDownloadable bool   `gorm:"default:false" json:"is_downloadable"`
	DownloadCount int64   `gorm:"default:0" json:"download_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID    string    `gorm:"not null;index" json:"site_id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"not null;unique_index" json:"slug"`
	Color     string    `json:"color"`
	Category  string    `json:"category"` // skill, technology, tool, etc.
	UsageCount int64    `gorm:"default:0" json:"usage_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Profile struct {
	ID              uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID          string            `gorm:"not null;index" json:"site_id"`
	UserID          uuid.UUID         `gorm:"not null;unique_index" json:"user_id"`
	DisplayName     string            `gorm:"not null" json:"display_name"`
	JobTitle        string            `json:"job_title"`
	Bio             string            `gorm:"type:text" json:"bio"`
	Avatar          string            `json:"avatar"`
	CoverImage      string            `json:"cover_image"`
	Location        string            `json:"location"`
	Website         string            `json:"website"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	ResumeURL       string            `json:"resume_url"`
	SocialLinks     map[string]string `gorm:"type:jsonb" json:"social_links"`
	Skills          []Skill           `json:"skills"`
	Experiences     []Experience      `json:"experiences"`
	Education       []Education       `json:"education"`
	Awards          []Award           `json:"awards"`
	Services        []Service         `json:"services"`
	Testimonials    []Testimonial     `json:"testimonials"`
	ContactSettings ContactSettings   `gorm:"type:jsonb" json:"contact_settings"`
	SEOSettings     SEOSettings       `gorm:"type:jsonb" json:"seo_settings"`
	IsPublic        bool              `gorm:"default:true" json:"is_public"`
	ViewCount       int64             `gorm:"default:0" json:"view_count"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type Skill struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID  uuid.UUID  `gorm:"not null;index" json:"profile_id"`
	Name       string     `gorm:"not null" json:"name"`
	Level      SkillLevel `json:"level"`
	Category   string     `json:"category"`
	YearsExp   int        `json:"years_experience"`
	IsFeatured bool       `gorm:"default:false" json:"is_featured"`
	SortOrder  int        `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time  `json:"created_at"`
}

type SkillLevel string

const (
	LevelBeginner     SkillLevel = "beginner"
	LevelIntermediate SkillLevel = "intermediate"
	LevelAdvanced     SkillLevel = "advanced"
	LevelExpert       SkillLevel = "expert"
)

type Experience struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID   uuid.UUID `gorm:"not null;index" json:"profile_id"`
	Company     string    `gorm:"not null" json:"company"`
	Position    string    `gorm:"not null" json:"position"`
	Description string    `gorm:"type:text" json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	IsCurrent   bool      `gorm:"default:false" json:"is_current"`
	Location    string    `json:"location"`
	CompanyURL  string    `json:"company_url"`
	Achievements []string `gorm:"type:text[]" json:"achievements"`
	Technologies []string `gorm:"type:text[]" json:"technologies"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Education struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID   uuid.UUID `gorm:"not null;index" json:"profile_id"`
	Institution string    `gorm:"not null" json:"institution"`
	Degree      string    `gorm:"not null" json:"degree"`
	FieldOfStudy string   `json:"field_of_study"`
	Grade       string    `json:"grade"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Location    string    `json:"location"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Award struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID   uuid.UUID `gorm:"not null;index" json:"profile_id"`
	Title       string    `gorm:"not null" json:"title"`
	Organization string   `json:"organization"`
	Description string    `json:"description"`
	AwardDate   time.Time `json:"award_date"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"image_url"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Service struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID   uuid.UUID `gorm:"not null;index" json:"profile_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       string    `json:"price"`
	Duration    string    `json:"duration"`
	Features    []string  `gorm:"type:text[]" json:"features"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Testimonial struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID   uuid.UUID `gorm:"not null;index" json:"profile_id"`
	ClientName  string    `gorm:"not null" json:"client_name"`
	ClientTitle string    `json:"client_title"`
	ClientCompany string  `json:"client_company"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Rating      int       `json:"rating"` // 1-5 stars
	ClientPhoto string    `json:"client_photo"`
	ProjectID   *uuid.UUID `json:"project_id"`
	IsApproved  bool      `gorm:"default:false" json:"is_approved"`
	IsFeatured  bool      `gorm:"default:false" json:"is_featured"`
	CreatedAt   time.Time `json:"created_at"`
}

type ContactSettings struct {
	EmailEnabled    bool   `json:"email_enabled"`
	FormEnabled     bool   `json:"form_enabled"`
	PhoneEnabled    bool   `json:"phone_enabled"`
	LinkedInEnabled bool   `json:"linkedin_enabled"`
	CalendlyURL     string `json:"calendly_url"`
	Timezone        string `json:"timezone"`
}

type SEOSettings struct {
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	Keywords        []string `json:"keywords"`
	OGImage         string   `json:"og_image"`
	TwitterHandle   string   `json:"twitter_handle"`
}
```

### Step 4: Multi-Tenant Repository Implementation

```go
// internal/adapters/persistence/portfolio_repository.go
package persistence

import (
	"fmt"
	"strings"
	"time"
	"github.com/zcrossoverz/echoforge/internal/domain/portfolio"
	"gorm.io/gorm"
)

type PortfolioRepository struct {
	db     *gorm.DB
	siteID string
}

func NewPortfolioRepository(db *gorm.DB, siteID string) *PortfolioRepository {
	return &PortfolioRepository{
		db:     db,
		siteID: siteID,
	}
}

// Project operations
func (r *PortfolioRepository) CreateProject(project *portfolio.Project) error {
	project.SiteID = r.siteID
	
	// Generate slug if not provided
	if project.Slug == "" {
		project.Slug = r.generateSlug(project.Title)
	}
	
	// Auto-generate meta fields if not provided
	if project.MetaTitle == "" {
		project.MetaTitle = project.Title
	}
	
	if project.MetaDescription == "" && project.Description != "" {
		project.MetaDescription = r.truncateText(project.Description, 160)
	}
	
	return r.db.Create(project).Error
}

func (r *PortfolioRepository) GetProjectBySlug(slug string) (*portfolio.Project, error) {
	var project portfolio.Project
	err := r.db.Where("slug = ? AND site_id = ?", slug, r.siteID).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Videos", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Files").
		Preload("Tags").
		First(&project).Error
	return &project, err
}

func (r *PortfolioRepository) ListProjects(category string, status portfolio.ProjectStatus, featured bool, limit, offset int) ([]portfolio.Project, error) {
	var projects []portfolio.Project
	query := r.db.Where("site_id = ?", r.siteID)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if featured {
		query = query.Where("is_featured = ?", true)
	}
	
	err := query.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Where("sort_order = 0").Limit(1) // Only featured image for list
	}).
		Preload("Tags").
		Order("priority DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error
	
	return projects, err
}

func (r *PortfolioRepository) UpdateProject(project *portfolio.Project) error {
	// Ensure the project belongs to this site
	existingProject, err := r.GetProjectByID(project.ID.String())
	if err != nil {
		return err
	}
	
	if existingProject.SiteID != r.siteID {
		return fmt.Errorf("project does not belong to site %s", r.siteID)
	}
	
	project.SiteID = r.siteID
	return r.db.Save(project).Error
}

func (r *PortfolioRepository) DeleteProject(id string) error {
	result := r.db.Where("id = ? AND site_id = ?", id, r.siteID).Delete(&portfolio.Project{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *PortfolioRepository) IncrementViewCount(projectID string) error {
	return r.db.Model(&portfolio.Project{}).
		Where("id = ? AND site_id = ?", projectID, r.siteID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// Profile operations
func (r *PortfolioRepository) CreateProfile(profile *portfolio.Profile) error {
	profile.SiteID = r.siteID
	return r.db.Create(profile).Error
}

func (r *PortfolioRepository) GetProfileByUserID(userID string) (*portfolio.Profile, error) {
	var profile portfolio.Profile
	err := r.db.Where("user_id = ? AND site_id = ?", userID, r.siteID).
		Preload("Skills", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Experiences", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Education", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Awards", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Services", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("sort_order ASC")
		}).
		Preload("Testimonials", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_approved = ?", true).Order("is_featured DESC, created_at DESC")
		}).
		First(&profile).Error
	return &profile, err
}

func (r *PortfolioRepository) UpdateProfile(profile *portfolio.Profile) error {
	// Ensure the profile belongs to this site
	existingProfile, err := r.GetProfileByUserID(profile.UserID.String())
	if err != nil {
		return err
	}
	
	if existingProfile.SiteID != r.siteID {
		return fmt.Errorf("profile does not belong to site %s", r.siteID)
	}
	
	profile.SiteID = r.siteID
	return r.db.Save(profile).Error
}

// Media operations
func (r *PortfolioRepository) AddProjectImage(image *portfolio.ProjectImage) error {
	// Verify project belongs to this site
	var project portfolio.Project
	if err := r.db.Where("id = ? AND site_id = ?", image.ProjectID, r.siteID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found or doesn't belong to this site")
	}
	
	return r.db.Create(image).Error
}

func (r *PortfolioRepository) AddProjectVideo(video *portfolio.ProjectVideo) error {
	// Verify project belongs to this site
	var project portfolio.Project
	if err := r.db.Where("id = ? AND site_id = ?", video.ProjectID, r.siteID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found or doesn't belong to this site")
	}
	
	return r.db.Create(video).Error
}

func (r *PortfolioRepository) AddProjectFile(file *portfolio.ProjectFile) error {
	// Verify project belongs to this site
	var project portfolio.Project
	if err := r.db.Where("id = ? AND site_id = ?", file.ProjectID, r.siteID).First(&project).Error; err != nil {
		return fmt.Errorf("project not found or doesn't belong to this site")
	}
	
	return r.db.Create(file).Error
}

// Tag operations
func (r *PortfolioRepository) CreateTag(tag *portfolio.Tag) error {
	tag.SiteID = r.siteID
	
	if tag.Slug == "" {
		tag.Slug = r.generateSlug(tag.Name)
	}
	
	return r.db.Create(tag).Error
}

func (r *PortfolioRepository) ListTags(category string) ([]portfolio.Tag, error) {
	var tags []portfolio.Tag
	query := r.db.Where("site_id = ?", r.siteID)
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	err := query.Order("usage_count DESC, name ASC").Find(&tags).Error
	return tags, err
}

// Search and filtering
func (r *PortfolioRepository) SearchProjects(searchTerm string, filters map[string]interface{}, limit, offset int) ([]portfolio.Project, error) {
	var projects []portfolio.Project
	query := r.db.Where("site_id = ?", r.siteID)
	
	if searchTerm != "" {
		searchPattern := "%" + strings.ToLower(searchTerm) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR ? = ANY(LOWER(technologies::text)::text[])", 
			searchPattern, searchPattern, strings.ToLower(searchTerm))
	}
	
	// Apply filters
	for key, value := range filters {
		switch key {
		case "category":
			query = query.Where("category = ?", value)
		case "technology":
			query = query.Where("? = ANY(technologies)", value)
		case "client":
			query = query.Where("LOWER(client_name) LIKE ?", "%"+strings.ToLower(value.(string))+"%")
		case "year":
			query = query.Where("EXTRACT(YEAR FROM created_at) = ?", value)
		}
	}
	
	err := query.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Where("sort_order = 0").Limit(1)
	}).
		Preload("Tags").
		Order("priority DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error
	
	return projects, err
}

// Analytics
func (r *PortfolioRepository) GetProjectStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total projects
	var totalProjects int64
	r.db.Model(&portfolio.Project{}).Where("site_id = ?", r.siteID).Count(&totalProjects)
	stats["total_projects"] = totalProjects
	
	// Projects by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.Model(&portfolio.Project{}).
		Where("site_id = ?", r.siteID).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts)
	stats["by_status"] = statusCounts
	
	// Total views
	var totalViews int64
	r.db.Model(&portfolio.Project{}).
		Where("site_id = ?", r.siteID).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&totalViews)
	stats["total_views"] = totalViews
	
	// Most viewed projects
	var topProjects []portfolio.Project
	r.db.Where("site_id = ?", r.siteID).
		Order("view_count DESC").
		Limit(5).
		Find(&topProjects)
	stats["top_viewed"] = topProjects
	
	return stats, nil
}

// Utility functions
func (r *PortfolioRepository) generateSlug(text string) string {
	// Implement slug generation logic
	slug := strings.ToLower(text)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	
	// Ensure uniqueness
	originalSlug := slug
	counter := 1
	for {
		var count int64
		r.db.Model(&portfolio.Project{}).Where("slug = ? AND site_id = ?", slug, r.siteID).Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}
	
	return slug
}

func (r *PortfolioRepository) truncateText(text string, length int) string {
	if len(text) <= length {
		return text
	}
	
	truncated := text[:length]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	
	return truncated + "..."
}
```

### Step 5: API Endpoints

```go
// adapters/http/portfolio_handler.go
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zcrossoverz/echoforge/internal/domain/portfolio"
)

type PortfolioHandler struct {
	repo portfolio.Repository
}

func NewPortfolioHandler(repo portfolio.Repository) *PortfolioHandler {
	return &PortfolioHandler{repo: repo}
}

func (h *PortfolioHandler) RegisterRoutes(r *gin.RouterGroup) {
	portfolio := r.Group("/portfolio")
	{
		// Project routes
		portfolio.GET("/projects", h.ListProjects)
		portfolio.POST("/projects", h.CreateProject)
		portfolio.GET("/projects/search", h.SearchProjects)
		portfolio.GET("/projects/stats", h.GetProjectStats)
		portfolio.GET("/projects/:slug", h.GetProject)
		portfolio.PUT("/projects/:id", h.UpdateProject)
		portfolio.DELETE("/projects/:id", h.DeleteProject)
		
		// Media routes
		portfolio.POST("/projects/:id/images", h.AddProjectImage)
		portfolio.POST("/projects/:id/videos", h.AddProjectVideo)
		portfolio.POST("/projects/:id/files", h.AddProjectFile)
		
		// Profile routes
		portfolio.GET("/profile", h.GetProfile)
		portfolio.PUT("/profile", h.UpdateProfile)
		portfolio.POST("/profile", h.CreateProfile)
		
		// Tag routes
		portfolio.GET("/tags", h.ListTags)
		portfolio.POST("/tags", h.CreateTag)
		
		// Analytics routes
		portfolio.GET("/analytics", h.GetAnalytics)
	}
}

func (h *PortfolioHandler) ListProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	offset := (page - 1) * limit
	
	category := c.Query("category")
	status := portfolio.ProjectStatus(c.DefaultQuery("status", "published"))
	featured, _ := strconv.ParseBool(c.Query("featured"))
	
	projects, err := h.repo.ListProjects(category, status, featured, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"page":     page,
		"limit":    limit,
	})
}

func (h *PortfolioHandler) CreateProject(c *gin.Context) {
	var project portfolio.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.repo.CreateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, project)
}

func (h *PortfolioHandler) GetProject(c *gin.Context) {
	slug := c.Param("slug")
	
	project, err := h.repo.GetProjectBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	
	// Increment view count
	h.repo.IncrementViewCount(project.ID.String())
	
	c.JSON(http.StatusOK, project)
}

func (h *PortfolioHandler) SearchProjects(c *gin.Context) {
	searchTerm := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	offset := (page - 1) * limit
	
	filters := make(map[string]interface{})
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if technology := c.Query("technology"); technology != "" {
		filters["technology"] = technology
	}
	if client := c.Query("client"); client != "" {
		filters["client"] = client
	}
	if year := c.Query("year"); year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			filters["year"] = yearInt
		}
	}
	
	projects, err := h.repo.SearchProjects(searchTerm, filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"projects":    projects,
		"search_term": searchTerm,
		"filters":     filters,
		"page":        page,
		"limit":       limit,
	})
}

func (h *PortfolioHandler) GetProfile(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		// Get from JWT token
		userID = c.GetString("user_id")
	}
	
	profile, err := h.repo.GetProfileByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	
	c.JSON(http.StatusOK, profile)
}

func (h *PortfolioHandler) UpdateProfile(c *gin.Context) {
	var profile portfolio.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.repo.UpdateProfile(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, profile)
}

func (h *PortfolioHandler) AddProjectImage(c *gin.Context) {
	projectID := c.Param("id")
	
	var image portfolio.ProjectImage
	if err := c.ShouldBindJSON(&image); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parse project ID
	id, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	image.ProjectID = id
	
	if err := h.repo.AddProjectImage(&image); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, image)
}

func (h *PortfolioHandler) GetProjectStats(c *gin.Context) {
	stats, err := h.repo.GetProjectStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
```

### Step 6: Database Migrations

```sql
-- migrations/009_create_portfolio_tables.up.sql
CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    job_title VARCHAR(255),
    bio TEXT,
    avatar VARCHAR(500),
    cover_image VARCHAR(500),
    location VARCHAR(255),
    website VARCHAR(500),
    email VARCHAR(255),
    phone VARCHAR(50),
    resume_url VARCHAR(500),
    social_links JSONB,
    contact_settings JSONB,
    seo_settings JSONB,
    is_public BOOLEAN DEFAULT true,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_profiles_site_id ON profiles(site_id);
CREATE UNIQUE INDEX idx_profiles_site_user ON profiles(site_id, user_id);

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    color VARCHAR(7),
    category VARCHAR(100),
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tags_site_id ON tags(site_id);
CREATE UNIQUE INDEX idx_tags_site_slug ON tags(site_id, slug);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    owner_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    long_description TEXT,
    category VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    featured_image VARCHAR(500),
    thumbnail_image VARCHAR(500),
    technologies TEXT[],
    skills TEXT[],
    client_name VARCHAR(255),
    project_url VARCHAR(500),
    github_url VARCHAR(500),
    demo_url VARCHAR(500),
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    duration VARCHAR(100),
    priority INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    is_featured BOOLEAN DEFAULT false,
    meta_title VARCHAR(255),
    meta_description VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_projects_site_id ON projects(site_id);
CREATE INDEX idx_projects_owner_id ON projects(owner_id);
CREATE INDEX idx_projects_category ON projects(category);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_featured ON projects(is_featured);
CREATE UNIQUE INDEX idx_projects_site_slug ON projects(site_id, slug);

CREATE TABLE IF NOT EXISTS project_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    original_url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    medium_url VARCHAR(500),
    large_url VARCHAR(500),
    caption TEXT,
    alt_text VARCHAR(255),
    file_size BIGINT,
    width INTEGER,
    height INTEGER,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_project_images_project_id ON project_images(project_id);
CREATE INDEX idx_project_images_sort ON project_images(project_id, sort_order);

CREATE TABLE IF NOT EXISTS project_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255),
    description TEXT,
    video_url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    duration INTEGER,
    file_size BIGINT,
    video_type VARCHAR(50),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_project_videos_project_id ON project_videos(project_id);

CREATE TABLE IF NOT EXISTS project_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    file_size BIGINT,
    description TEXT,
    is_downloadable BOOLEAN DEFAULT false,
    download_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_project_files_project_id ON project_files(project_id);

CREATE TABLE IF NOT EXISTS project_tags (
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, tag_id)
);

CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    level VARCHAR(50),
    category VARCHAR(100),
    years_experience INTEGER,
    is_featured BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_skills_profile_id ON skills(profile_id);

CREATE TABLE IF NOT EXISTS experiences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    company VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    is_current BOOLEAN DEFAULT false,
    location VARCHAR(255),
    company_url VARCHAR(500),
    achievements TEXT[],
    technologies TEXT[],
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_experiences_profile_id ON experiences(profile_id);

CREATE TABLE IF NOT EXISTS education (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    institution VARCHAR(255) NOT NULL,
    degree VARCHAR(255) NOT NULL,
    field_of_study VARCHAR(255),
    grade VARCHAR(50),
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    location VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_education_profile_id ON education(profile_id);

CREATE TABLE IF NOT EXISTS awards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    organization VARCHAR(255),
    description TEXT,
    award_date TIMESTAMP WITH TIME ZONE,
    url VARCHAR(500),
    image_url VARCHAR(500),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_awards_profile_id ON awards(profile_id);

CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price VARCHAR(100),
    duration VARCHAR(100),
    features TEXT[],
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_services_profile_id ON services(profile_id);

CREATE TABLE IF NOT EXISTS testimonials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    client_name VARCHAR(255) NOT NULL,
    client_title VARCHAR(255),
    client_company VARCHAR(255),
    content TEXT NOT NULL,
    rating INTEGER,
    client_photo VARCHAR(500),
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    is_approved BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_testimonials_profile_id ON testimonials(profile_id);
CREATE INDEX idx_testimonials_approved ON testimonials(is_approved);
```

### Step 7: Testing Your Portfolio Site

```bash
# Run database migrations
migrate -path migrations -database "postgres://username:password@localhost/portfolio_platform_db?sslmode=disable" up

# Start your portfolio site
go run cmd/server/main.go --config config/portfolio-site.yaml

# Create a profile
curl -X POST http://localhost:8080/api/v1/portfolio/profile \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Jane Developer",
    "job_title": "Full Stack Developer & UI/UX Designer",
    "bio": "Passionate developer with 5+ years of experience creating beautiful, functional web applications.",
    "location": "San Francisco, CA",
    "website": "https://janedeveloper.com",
    "email": "jane@example.com"
  }'

# Create a project
curl -X POST http://localhost:8080/api/v1/portfolio/projects \
  -H "Content-Type: application/json" \
  -d '{
    "title": "E-commerce Platform",
    "description": "Modern e-commerce platform built with React and Node.js",
    "category": "web-development",
    "technologies": ["React", "Node.js", "PostgreSQL", "Docker"],
    "status": "published",
    "project_url": "https://example-ecommerce.com",
    "github_url": "https://github.com/jane/ecommerce-platform"
  }'

# List projects
curl http://localhost:8080/api/v1/portfolio/projects

# Search projects
curl "http://localhost:8080/api/v1/portfolio/projects/search?q=ecommerce&category=web-development"
```

## Advanced Features

### Image Processing and Optimization

```go
// pkg/media/image_processor.go
package media

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	
	"github.com/nfnt/resize"
)

type ImageProcessor struct {
	config ImageProcessorConfig
}

type ImageProcessorConfig struct {
	ThumbnailSizes []int
	Quality        int
	OutputFormat   string
}

func NewImageProcessor(config ImageProcessorConfig) *ImageProcessor {
	return &ImageProcessor{config: config}
}

func (p *ImageProcessor) ProcessProjectImage(inputPath, outputDir string) (*ProcessedImage, error) {
	// Open original image
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	
	processed := &ProcessedImage{
		OriginalFormat: format,
		OriginalWidth:  img.Bounds().Dx(),
		OriginalHeight: img.Bounds().Dy(),
		Variants:       make(map[string]string),
	}
	
	// Generate thumbnails
	for _, size := range p.config.ThumbnailSizes {
		resized := resize.Thumbnail(uint(size), uint(size), img, resize.Lanczos3)
		
		filename := fmt.Sprintf("thumb_%dx%d.jpg", size, size)
		outputPath := filepath.Join(outputDir, filename)
		
		if err := p.saveImage(resized, outputPath, "jpeg"); err != nil {
			return nil, err
		}
		
		processed.Variants[fmt.Sprintf("thumb_%d", size)] = outputPath
	}
	
	return processed, nil
}

type ProcessedImage struct {
	OriginalFormat string
	OriginalWidth  int
	OriginalHeight int
	Variants       map[string]string
}

func (p *ImageProcessor) saveImage(img image.Image, path, format string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	switch format {
	case "jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: p.config.Quality})
	case "png":
		return png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
```

### SEO Enhancement

```go
// internal/domain/portfolio/seo.go
package portfolio

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SEOService struct {
	repo Repository
}

func NewSEOService(repo Repository) *SEOService {
	return &SEOService{repo: repo}
}

func (s *SEOService) GenerateStructuredData(project *Project) (string, error) {
	schema := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "CreativeWork",
		"name":     project.Title,
		"description": project.Description,
		"url":      fmt.Sprintf("https://yourdomain.com/projects/%s", project.Slug),
		"image":    project.FeaturedImage,
		"creator": map[string]interface{}{
			"@type": "Person",
			"name":  "Portfolio Owner Name",
		},
		"dateCreated": project.CreatedAt.Format(time.RFC3339),
		"keywords":    strings.Join(project.Technologies, ", "),
	}
	
	if project.ProjectURL != "" {
		schema["sameAs"] = project.ProjectURL
	}
	
	jsonLD, err := json.MarshalIndent(schema, "", "  ")
	return string(jsonLD), err
}

func (s *SEOService) GenerateOpenGraphTags(project *Project) map[string]string {
	return map[string]string{
		"og:title":       project.Title,
		"og:description": project.Description,
		"og:image":       project.FeaturedImage,
		"og:url":         fmt.Sprintf("https://yourdomain.com/projects/%s", project.Slug),
		"og:type":        "website",
		"og:site_name":   "Portfolio",
	}
}

func (s *SEOService) GenerateSitemap(siteID string) (string, error) {
	projects, err := s.repo.ListProjects("", StatusPublished, false, 1000, 0)
	if err != nil {
		return "", err
	}
	
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	
	for _, project := range projects {
		sitemap += fmt.Sprintf(`
	<url>
		<loc>https://yourdomain.com/projects/%s</loc>
		<lastmod>%s</lastmod>
		<changefreq>monthly</changefreq>
		<priority>0.8</priority>
	</url>`, project.Slug, project.UpdatedAt.Format("2006-01-02"))
	}
	
	sitemap += `
</urlset>`
	
	return sitemap, nil
}
```

## Performance Optimization

### Caching Strategy

```go
type CachedPortfolioRepository struct {
	*PortfolioRepository
	cache Cache
}

func (r *CachedPortfolioRepository) GetProjectBySlug(slug string) (*portfolio.Project, error) {
	cacheKey := fmt.Sprintf("project:%s:%s", r.siteID, slug)
	
	// Try cache first
	if cached, err := r.cache.Get(cacheKey); err == nil {
		var project portfolio.Project
		json.Unmarshal(cached, &project)
		return &project, nil
	}
	
	// Fallback to database
	project, err := r.PortfolioRepository.GetProjectBySlug(slug)
	if err != nil {
		return nil, err
	}
	
	// Cache for future requests
	if data, err := json.Marshal(project); err == nil {
		r.cache.Set(cacheKey, data, 4*time.Hour)
	}
	
	return project, nil
}
```

## Deployment

```dockerfile
# Dockerfile.portfolio
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o portfolio-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates imagemagick
WORKDIR /root/

COPY --from=builder /app/portfolio-server .
COPY --from=builder /app/config/portfolio-site.yaml ./config/

CMD ["./portfolio-server", "--config", "config/portfolio-site.yaml"]
```

This portfolio platform demonstrates Echoforge's flexibility in creating media-rich, professional showcase applications with proper multi-tenant architecture and advanced features like SEO optimization, analytics, and media processing.