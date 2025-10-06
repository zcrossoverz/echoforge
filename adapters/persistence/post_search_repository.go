package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// GormPostSearchRepository implements domain.PostSearchRepository using GORM
type GormPostSearchRepository struct {
	db *gorm.DB
}

// NewGormPostSearchRepository creates a new GORM-based post search repository
func NewGormPostSearchRepository(db *gorm.DB) domain.PostSearchRepository {
	return &GormPostSearchRepository{
		db: db,
	}
}

// SearchFullText performs full-text search on posts
func (r *GormPostSearchRepository) SearchFullText(ctx context.Context, query string, options domain.FullTextSearchOptions) (*domain.SearchResult, error) {
	if query == "" {
		return &domain.SearchResult{Posts: []*domain.PostSearchMatch{}}, nil
	}

	startTime := time.Now()
	dbQuery := r.db.WithContext(ctx).Model(&domain.Post{})

	// Apply text search
	dbQuery = dbQuery.Where("title ILIKE ? OR content ILIKE ?", "%"+query+"%", "%"+query+"%")

	// Apply filters
	if options.Status != nil {
		dbQuery = dbQuery.Where("status = ?", *options.Status)
	}
	if options.AuthorID != nil {
		dbQuery = dbQuery.Where("author_id = ?", *options.AuthorID)
	}
	if options.PostTypeID != nil {
		dbQuery = dbQuery.Where("post_type_id = ?", *options.PostTypeID)
	}

	// Get total count
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply sorting and pagination
	dbQuery = dbQuery.Order("created_at DESC")
	if options.Limit > 0 {
		dbQuery = dbQuery.Limit(options.Limit)
	}
	if options.Offset > 0 {
		dbQuery = dbQuery.Offset(options.Offset)
	}

	// Execute search
	var posts []*domain.Post
	if err := dbQuery.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	// Convert to search matches
	matches := make([]*domain.PostSearchMatch, len(posts))
	for i, post := range posts {
		matches[i] = &domain.PostSearchMatch{
			Post:  post,
			Score: 1.0,
		}
	}

	return &domain.SearchResult{
		Posts:      matches,
		TotalCount: total,
		SearchTime: time.Since(startTime),
	}, nil
}

// SearchByKeywords performs keyword-based search
func (r *GormPostSearchRepository) SearchByKeywords(ctx context.Context, keywords []string, options domain.KeywordSearchOptions) (*domain.SearchResult, error) {
	// Simplified implementation - delegate to full-text search
	query := ""
	if len(keywords) > 0 {
		query = keywords[0] // Use first keyword
	}

	ftOptions := domain.FullTextSearchOptions{
		Limit:      options.Limit,
		Offset:     options.Offset,
		Status:     options.Status,
		AuthorID:   options.AuthorID,
		PostTypeID: options.PostTypeID,
		CategoryID: options.CategoryID,
		SortBy:     options.SortBy,
		SortOrder:  options.SortOrder,
	}

	return r.SearchFullText(ctx, query, ftOptions)
}

// SearchSimilar finds posts similar to a given post
func (r *GormPostSearchRepository) SearchSimilar(ctx context.Context, postID uuid.UUID, options domain.SimilarSearchOptions) (*domain.SearchResult, error) {
	// Get source post
	var sourcePost domain.Post
	if err := r.db.WithContext(ctx).Where("id = ?", postID).First(&sourcePost).Error; err != nil {
		return nil, fmt.Errorf("failed to get source post: %w", err)
	}

	// Find similar posts (simplified - same post type)
	dbQuery := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("post_type_id = ? AND id != ?", sourcePost.PostTypeID, postID)

	if options.Status != nil {
		dbQuery = dbQuery.Where("status = ?", *options.Status)
	}

	var total int64
	dbQuery.Count(&total)

	if options.Limit > 0 {
		dbQuery = dbQuery.Limit(options.Limit)
	}

	var posts []*domain.Post
	if err := dbQuery.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to find similar posts: %w", err)
	}

	matches := make([]*domain.PostSearchMatch, len(posts))
	for i, post := range posts {
		matches[i] = &domain.PostSearchMatch{
			Post:  post,
			Score: 0.8, // Default similarity score
		}
	}

	return &domain.SearchResult{
		Posts:      matches,
		TotalCount: total,
	}, nil
}

// Implement all required interface methods with basic implementations

func (r *GormPostSearchRepository) SearchWithFacets(ctx context.Context, query string, facets []string, options domain.FacetedSearchOptions) (*domain.FacetedSearchResult, error) {
	searchResult, err := r.SearchFullText(ctx, query, options.FullTextSearchOptions)
	if err != nil {
		return nil, err
	}
	return &domain.FacetedSearchResult{
		SearchResult: *searchResult,
		Facets:       make(map[string][]domain.FacetValue),
	}, nil
}

func (r *GormPostSearchRepository) GetFacetValues(ctx context.Context, facetName string, options domain.FacetValuesOptions) ([]domain.FacetValue, error) {
	return []domain.FacetValue{}, nil
}

func (r *GormPostSearchRepository) FilterByMultipleConditions(ctx context.Context, conditions []domain.FilterCondition, options domain.FilterOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{})
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) FilterByDateRange(ctx context.Context, field string, startDate, endDate time.Time, options domain.FilterOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), startDate, endDate)
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) FilterByNumericRange(ctx context.Context, field string, min, max float64, options domain.FilterOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), min, max)
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) SearchByLocation(ctx context.Context, lat, lng float64, radius float64, options domain.LocationSearchOptions) ([]*domain.Post, error) {
	return []*domain.Post{}, nil // Not implemented
}

func (r *GormPostSearchRepository) SearchByBounds(ctx context.Context, northLat, southLat, eastLng, westLng float64, options domain.LocationSearchOptions) ([]*domain.Post, error) {
	return []*domain.Post{}, nil // Not implemented
}

func (r *GormPostSearchRepository) GetTrending(ctx context.Context, timeframe string, options domain.TrendingOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("status = ?", domain.PostStatusPublished).
		Order("view_count DESC")

	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}

	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) GetPopular(ctx context.Context, timeframe string, options domain.PopularOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("status = ?", domain.PostStatusPublished).
		Order("view_count DESC")

	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}

	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) GetRecentlyViewed(ctx context.Context, userID uuid.UUID, options domain.RecentViewsOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("status = ?", domain.PostStatusPublished).
		Order("created_at DESC")

	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}

	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) GetRecommendations(ctx context.Context, userID uuid.UUID, options domain.RecommendationOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("status = ?", domain.PostStatusPublished).
		Order("view_count DESC, created_at DESC")

	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}

	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) GetRelatedPosts(ctx context.Context, postID uuid.UUID, options domain.RelatedPostsOptions) ([]*domain.Post, error) {
	var sourcePost domain.Post
	if err := r.db.WithContext(ctx).Where("id = ?", postID).First(&sourcePost).Error; err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("post_type_id = ? AND id != ?", sourcePost.PostTypeID, postID)

	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}

	var posts []*domain.Post
	return posts, query.Find(&posts).Error
}

func (r *GormPostSearchRepository) GetSearchAnalytics(ctx context.Context, timeframe string) (*domain.SearchAnalytics, error) {
	return &domain.SearchAnalytics{
		Period:           timeframe,
		TotalSearches:    0,
		UniqueSearchers:  0,
		AverageResults:   0,
		ZeroResultRate:   0,
		ClickThroughRate: 0,
		TopQueries:       []domain.SearchTerm{},
		TopCategories:    []string{},
		SearchTrends:     make(map[string]int64),
	}, nil
}

func (r *GormPostSearchRepository) GetPopularSearchTerms(ctx context.Context, limit int, timeframe string) ([]domain.SearchTerm, error) {
	return []domain.SearchTerm{}, nil
}

func (r *GormPostSearchRepository) GetSearchSuggestions(ctx context.Context, partial string, maxSuggestions int) ([]string, error) {
	if partial == "" {
		return []string{}, nil
	}

	var titles []string
	rows, err := r.db.WithContext(ctx).
		Raw("SELECT title FROM posts WHERE status = ? AND title ILIKE ? LIMIT ?",
			domain.PostStatusPublished, partial+"%", maxSuggestions).
		Rows()

	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err == nil {
			titles = append(titles, title)
		}
	}

	return titles, nil
}

func (r *GormPostSearchRepository) GetPopularTags(ctx context.Context, limit int, timeframe string) ([]domain.TagSuggestion, error) {
	return []domain.TagSuggestion{}, nil
}

func (r *GormPostSearchRepository) ExportSearchResults(ctx context.Context, query string, format string, options domain.ExportOptions) ([]byte, error) {
	searchResult, err := r.SearchFullText(ctx, query, domain.FullTextSearchOptions{})
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.Marshal(searchResult)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (r *GormPostSearchRepository) BulkIndex(ctx context.Context, posts []*domain.Post) error {
	return nil // Not implemented for GORM
}

func (r *GormPostSearchRepository) BulkDelete(ctx context.Context, postIDs []uuid.UUID) error {
	if len(postIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id IN ?", postIDs).
		Update("status", domain.PostStatusArchived).Error
}

func (r *GormPostSearchRepository) ReindexPost(ctx context.Context, postID uuid.UUID) error {
	return nil // Not implemented for GORM
}

func (r *GormPostSearchRepository) ReindexAll(ctx context.Context) error {
	return nil // Not implemented for GORM
}

func (r *GormPostSearchRepository) GetIndexStats(ctx context.Context) (*domain.IndexStats, error) {
	var totalDocs int64
	r.db.Model(&domain.Post{}).Count(&totalDocs)

	var publishedDocs int64
	r.db.Model(&domain.Post{}).Where("status = ?", domain.PostStatusPublished).Count(&publishedDocs)

	return &domain.IndexStats{
		TotalDocuments:   totalDocs,
		IndexedDocuments: publishedDocs,
		PendingDocuments: totalDocs - publishedDocs,
		IndexSize:        0,
		LastIndexed:      time.Now(),
		IndexHealth:      "green",
		FieldStats:       make(map[string]int64),
		SearchStats:      make(map[string]interface{}),
	}, nil
}
