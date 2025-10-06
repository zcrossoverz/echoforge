# Manga Site Extension Guide

## Overview

This guide walks you through creating a manga reading platform using Echoforge's multi-tenant architecture. The manga site will support multiple series, chapters, user bookmarks, ratings, and comments while maintaining proper tenant isolation.

## Prerequisites

- Go 1.25+ installed
- PostgreSQL 16+ database
- Basic understanding of Echoforge's architecture
- Familiarity with GORM and Gin frameworks

## Setup Process

### Step 1: Clone and Configure Base

```bash
# Clone the Echoforge repository
git clone https://github.com/zcrossoverz/echoforge.git
cd echoforge

# Install dependencies
go mod download

# Create your manga site database
createdb manga_reader_db
```

### Step 2: Create Site Configuration

Create a configuration file for your manga site:

```yaml
# config/manga-site.yaml
site:
  id: "manga-001"
  name: "Manga Reader Platform"
  description: "Multi-series manga reading platform with user engagement"
  type: "manga"

database:
  dsn: "postgres://username:password@localhost/manga_reader_db?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 10

features:
  comments: true
  ratings: true
  bookmarks: true
  notifications: true
  social_sharing: false

manga_specific:
  max_chapter_size_mb: 50
  supported_formats: ["jpg", "png", "webp"]
  reading_modes: ["single", "double", "webtoon"]
  default_reading_direction: "ltr"

authentication:
  jwt_secret: "${JWT_SECRET}"
  session_duration: "24h"
  rate_limit:
    requests_per_minute: 60
    burst: 10

performance:
  cache_duration: "1h"
  max_concurrent_uploads: 5
  image_optimization: true
```

### Step 3: Extend Domain Models

Create manga-specific domain entities in `internal/domain/manga/`:

```go
// internal/domain/manga/series.go
package manga

import (
	"time"
	"github.com/google/uuid"
)

type Series struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID      string    `gorm:"not null;index" json:"site_id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Artist      string    `json:"artist"`
	Status      SeriesStatus `gorm:"default:'ongoing'" json:"status"`
	CoverImage  string    `json:"cover_image"`
	Tags        []Tag     `gorm:"many2many:series_tags" json:"tags"`
	Chapters    []Chapter `json:"chapters"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SeriesStatus string

const (
	StatusOngoing   SeriesStatus = "ongoing"
	StatusCompleted SeriesStatus = "completed"
	StatusHiatus    SeriesStatus = "hiatus"
	StatusCancelled SeriesStatus = "cancelled"
)

type Chapter struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID   string    `gorm:"not null;index" json:"site_id"`
	SeriesID uuid.UUID `gorm:"not null" json:"series_id"`
	Number   float64   `gorm:"not null" json:"number"`
	Title    string    `json:"title"`
	Pages    []Page    `json:"pages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Page struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChapterID uuid.UUID `gorm:"not null" json:"chapter_id"`
	Number    int       `gorm:"not null" json:"number"`
	ImageURL  string    `gorm:"not null" json:"image_url"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
}

type Tag struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SiteID string    `gorm:"not null;index" json:"site_id"`
	Name   string    `gorm:"not null" json:"name"`
	Color  string    `json:"color"`
}
```

### Step 4: Multi-Tenant Repository Implementation

```go
// internal/adapters/persistence/manga_repository.go
package persistence

import (
	"github.com/zcrossoverz/echoforge/internal/domain/manga"
	"gorm.io/gorm"
)

type MangaRepository struct {
	db     *gorm.DB
	siteID string
}

func NewMangaRepository(db *gorm.DB, siteID string) *MangaRepository {
	return &MangaRepository{
		db:     db,
		siteID: siteID,
	}
}

func (r *MangaRepository) CreateSeries(series *manga.Series) error {
	series.SiteID = r.siteID
	return r.db.Create(series).Error
}

func (r *MangaRepository) GetSeriesByID(id string) (*manga.Series, error) {
	var series manga.Series
	err := r.db.Where("id = ? AND site_id = ?", id, r.siteID).
		Preload("Tags").
		Preload("Chapters").
		First(&series).Error
	return &series, err
}

func (r *MangaRepository) ListSeries(limit, offset int) ([]manga.Series, error) {
	var series []manga.Series
	err := r.db.Where("site_id = ?", r.siteID).
		Preload("Tags").
		Limit(limit).
		Offset(offset).
		Find(&series).Error
	return series, err
}

func (r *MangaRepository) CreateChapter(chapter *manga.Chapter) error {
	chapter.SiteID = r.siteID
	return r.db.Create(chapter).Error
}

func (r *MangaRepository) GetChapterWithPages(id string) (*manga.Chapter, error) {
	var chapter manga.Chapter
	err := r.db.Where("id = ? AND site_id = ?", id, r.siteID).
		Preload("Pages", func(db *gorm.DB) *gorm.DB {
			return db.Order("number ASC")
		}).
		First(&chapter).Error
	return &chapter, err
}
```

### Step 5: API Endpoints

```go
// adapters/http/manga_handler.go
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zcrossoverz/echoforge/internal/domain/manga"
)

type MangaHandler struct {
	repo manga.Repository
}

func NewMangaHandler(repo manga.Repository) *MangaHandler {
	return &MangaHandler{repo: repo}
}

func (h *MangaHandler) RegisterRoutes(r *gin.RouterGroup) {
	manga := r.Group("/manga")
	{
		manga.GET("/series", h.ListSeries)
		manga.POST("/series", h.CreateSeries) 
		manga.GET("/series/:id", h.GetSeries)
		manga.POST("/series/:id/chapters", h.CreateChapter)
		manga.GET("/chapters/:id", h.GetChapter)
	}
}

func (h *MangaHandler) ListSeries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	series, err := h.repo.ListSeries(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"series": series,
		"page":   page,
		"limit":  limit,
	})
}

func (h *MangaHandler) CreateSeries(c *gin.Context) {
	var series manga.Series
	if err := c.ShouldBindJSON(&series); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateSeries(&series); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, series)
}

func (h *MangaHandler) GetSeries(c *gin.Context) {
	id := c.Param("id")
	
	series, err := h.repo.GetSeriesByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	c.JSON(http.StatusOK, series)
}

func (h *MangaHandler) CreateChapter(c *gin.Context) {
	seriesID := c.Param("id")
	
	var chapter manga.Chapter
	if err := c.ShouldBindJSON(&chapter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that series exists and belongs to this site
	if _, err := h.repo.GetSeriesByID(seriesID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "series not found"})
		return
	}

	if err := h.repo.CreateChapter(&chapter); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chapter)
}

func (h *MangaHandler) GetChapter(c *gin.Context) {
	id := c.Param("id")
	
	chapter, err := h.repo.GetChapterWithPages(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	c.JSON(http.StatusOK, chapter)
}
```

### Step 6: Database Migrations

Create migration files for manga-specific tables:

```sql
-- migrations/007_create_manga_tables.up.sql
CREATE TABLE IF NOT EXISTS series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    artist VARCHAR(255),
    status VARCHAR(50) DEFAULT 'ongoing',
    cover_image VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_series_site_id ON series(site_id);
CREATE INDEX idx_series_status ON series(status);

CREATE TABLE IF NOT EXISTS chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    number DECIMAL(5,2) NOT NULL,
    title VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_chapters_site_id ON chapters(site_id);
CREATE INDEX idx_chapters_series_id ON chapters(series_id);
CREATE UNIQUE INDEX idx_chapters_series_number ON chapters(series_id, number);

CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    width INTEGER,
    height INTEGER
);

CREATE INDEX idx_pages_chapter_id ON pages(chapter_id);
CREATE UNIQUE INDEX idx_pages_chapter_number ON pages(chapter_id, number);

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7)
);

CREATE INDEX idx_tags_site_id ON tags(site_id);
CREATE UNIQUE INDEX idx_tags_site_name ON tags(site_id, name);

CREATE TABLE IF NOT EXISTS series_tags (
    series_id UUID REFERENCES series(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (series_id, tag_id)
);
```

### Step 7: Testing Your Manga Site

```bash
# Run database migrations
migrate -path migrations -database "postgres://username:password@localhost/manga_reader_db?sslmode=disable" up

# Start your manga site
go run cmd/server/main.go --config config/manga-site.yaml

# Test the API endpoints
curl -X POST http://localhost:8080/api/v1/manga/series \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Sample Manga",
    "description": "A test manga series",
    "author": "Test Author",
    "status": "ongoing"
  }'

# List series
curl http://localhost:8080/api/v1/manga/series

# Create a chapter
curl -X POST http://localhost:8080/api/v1/manga/series/{series-id}/chapters \
  -H "Content-Type: application/json" \
  -d '{
    "number": 1,
    "title": "Chapter 1: The Beginning",
    "pages": [
      {"number": 1, "image_url": "/images/ch1-p1.jpg"},
      {"number": 2, "image_url": "/images/ch1-p2.jpg"}
    ]
  }'
```

## Multi-Tenant Isolation Best Practices

### 1. Always Use site_id in Queries

Every database operation must include the `site_id` filter:

```go
// CORRECT: Always filter by site_id
err := r.db.Where("id = ? AND site_id = ?", id, r.siteID).First(&series).Error

// WRONG: Never query without site_id filtering
err := r.db.Where("id = ?", id).First(&series).Error
```

### 2. Validate Cross-Site References

When creating relationships, ensure all entities belong to the same site:

```go
func (r *MangaRepository) CreateChapter(chapter *manga.Chapter) error {
    // Verify the series exists and belongs to this site
    var series manga.Series
    if err := r.db.Where("id = ? AND site_id = ?", chapter.SeriesID, r.siteID).First(&series).Error; err != nil {
        return errors.New("series not found or doesn't belong to this site")
    }
    
    chapter.SiteID = r.siteID
    return r.db.Create(chapter).Error
}
```

### 3. Configuration Validation

Validate that your site configuration is properly isolated:

```go
func validateConfig(config *MangaConfig) error {
    if config.Site.ID == "" {
        return errors.New("site_id is required")
    }
    
    if config.Database.DSN == "" {
        return errors.New("database DSN is required")
    }
    
    // Ensure site_id follows naming conventions
    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(config.Site.ID) {
        return errors.New("site_id must contain only lowercase letters, numbers, and hyphens")
    }
    
    return nil
}
```

## Performance Optimization

### 1. Database Indexing

Ensure proper indexes for manga-specific queries:

```sql
-- Indexes for common query patterns
CREATE INDEX idx_series_site_updated ON series(site_id, updated_at DESC);
CREATE INDEX idx_chapters_series_number ON chapters(series_id, number);
CREATE INDEX idx_series_tags_composite ON series_tags(series_id, tag_id);
```

### 2. Image Handling

Implement efficient image storage and delivery:

```go
type ImageService struct {
    basePath string
    maxSize  int64
}

func (s *ImageService) UploadPage(siteID string, file io.Reader) (string, error) {
    // Create site-specific directory
    sitePath := filepath.Join(s.basePath, siteID, "pages")
    
    // Generate unique filename
    filename := fmt.Sprintf("%d-%s.jpg", time.Now().Unix(), uuid.New().String()[:8])
    
    // Optimize and save image
    return s.optimizeAndSave(sitePath, filename, file)
}
```

### 3. Caching Strategy

Implement Redis caching for frequently accessed data:

```go
func (r *MangaRepository) GetSeriesWithCache(id string) (*manga.Series, error) {
    cacheKey := fmt.Sprintf("series:%s:%s", r.siteID, id)
    
    // Try cache first
    if cached, err := r.cache.Get(cacheKey); err == nil {
        var series manga.Series
        json.Unmarshal(cached, &series)
        return &series, nil
    }
    
    // Fallback to database
    series, err := r.GetSeriesByID(id)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    if data, err := json.Marshal(series); err == nil {
        r.cache.Set(cacheKey, data, time.Hour)
    }
    
    return series, nil
}
```

## Security Considerations

### 1. Content Validation

Validate uploaded content to prevent security issues:

```go
func validateImageUpload(file multipart.File) error {
    // Check file size
    if size := getFileSize(file); size > 50*1024*1024 { // 50MB
        return errors.New("file too large")
    }
    
    // Validate image format
    if !isValidImageFormat(file) {
        return errors.New("invalid image format")
    }
    
    // Scan for malware (in production)
    return nil
}
```

### 2. Access Control

Implement proper authorization for manga operations:

```go
func (h *MangaHandler) requireAuth(c *gin.Context) {
    token := c.GetHeader("Authorization")
    if !h.authService.ValidateToken(token) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        c.Abort()
        return
    }
}
```

## Deployment

Deploy your manga site using Docker:

```dockerfile
# Dockerfile.manga
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o manga-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/manga-server .
COPY --from=builder /app/config/manga-site.yaml ./config/

CMD ["./manga-server", "--config", "config/manga-site.yaml"]
```

```bash
# Build and run
docker build -f Dockerfile.manga -t manga-site:latest .
docker run -p 8080:8080 -e JWT_SECRET=your-secret manga-site:latest
```

## Troubleshooting

### Common Issues

1. **Database Connection Fails**
   - Verify PostgreSQL is running
   - Check connection string format
   - Ensure database exists

2. **Image Upload Issues**
   - Check file permissions on upload directory
   - Verify image format support
   - Check file size limits

3. **Multi-Tenant Data Leakage**
   - Always verify `site_id` filtering in queries
   - Test with multiple site configurations
   - Audit database queries in logs

### Performance Issues

1. **Slow Series Listing**
   - Add indexes on commonly queried fields
   - Implement pagination
   - Use database query optimization

2. **Image Loading Slow**
   - Implement CDN for static assets
   - Use image compression
   - Add proper caching headers

## Next Steps

1. **Add User Engagement Features**
   - Comments and ratings system
   - User reading progress tracking
   - Bookmarks and favorites

2. **Enhanced Reading Experience**
   - Progressive image loading
   - Offline reading support
   - Multiple reading modes

3. **Content Management**
   - Bulk chapter upload
   - Series metadata management
   - Tag management system

4. **Analytics and Monitoring**
   - Reading statistics
   - Popular series tracking
   - Performance monitoring

This manga site extension demonstrates Echoforge's flexibility while maintaining proper multi-tenant isolation and security practices. The modular approach allows for easy customization and scaling as your manga platform grows.