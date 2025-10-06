package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zcrossoverz/echoforge/internal/domain"
)

// PostSearchUsecase handles business logic for post search and filtering operations
// Provides advanced search capabilities with full-text search, faceted search, and recommendations
type PostSearchUsecase struct {
	searchRepo   domain.PostSearchRepository
	postRepo     domain.PostRepository
	postTypeRepo domain.PostTypeRepository
}

// SearchPostsInput represents input for searching posts
type SearchPostsInput struct {
	Query             string             `json:"query" validate:"max=500"`
	SearchFields      []string           `json:"search_fields,omitempty"`
	Status            *domain.PostStatus `json:"status,omitempty"`
	AuthorID          *uuid.UUID         `json:"author_id,omitempty"`
	PostTypeID        *uuid.UUID         `json:"post_type_id,omitempty"`
	CategoryID        *uuid.UUID         `json:"category_id,omitempty"`
	TagIDs            []uuid.UUID        `json:"tag_ids,omitempty"`
	PublishedAfter    *time.Time         `json:"published_after,omitempty"`
	PublishedBefore   *time.Time         `json:"published_before,omitempty"`
	Limit             int                `json:"limit" validate:"min=1,max=100"`
	Offset            int                `json:"offset" validate:"min=0"`
	SortBy            string             `json:"sort_by,omitempty"`
	SortOrder         string             `json:"sort_order,omitempty"`
	Operator          string             `json:"operator,omitempty"` // "AND", "OR"
	MinScore          float64            `json:"min_score,omitempty"`
	BoostFields       map[string]float64 `json:"boost_fields,omitempty"`
	IncludeHighlights bool               `json:"include_highlights,omitempty"`
	IncludeSnippets   bool               `json:"include_snippets,omitempty"`
	Facets            []string           `json:"facets,omitempty"`
}

// SearchPostsOutput represents output from searching posts
type SearchPostsOutput struct {
	Results     *domain.SearchResult           `json:"results"`
	Facets      map[string][]domain.FacetValue `json:"facets,omitempty"`
	SearchTime  time.Duration                  `json:"search_time"`
	TotalCount  int64                          `json:"total_count"`
	MaxScore    float64                        `json:"max_score,omitempty"`
	Suggestions []string                       `json:"suggestions,omitempty"`
}

// SimilarPostsInput represents input for finding similar posts
type SimilarPostsInput struct {
	PostID        uuid.UUID          `json:"post_id" validate:"required"`
	Algorithm     string             `json:"algorithm,omitempty"` // "content", "tags", "categories", "combined"
	MinSimilarity float64            `json:"min_similarity,omitempty"`
	Status        *domain.PostStatus `json:"status,omitempty"`
	PostTypeID    *uuid.UUID         `json:"post_type_id,omitempty"`
	Limit         int                `json:"limit" validate:"min=1,max=50"`
	BoostRecent   bool               `json:"boost_recent,omitempty"`
	BoostPopular  bool               `json:"boost_popular,omitempty"`
}

// RecommendationsInput represents input for getting recommendations
type RecommendationsInput struct {
	UserID        uuid.UUID          `json:"user_id" validate:"required"`
	Algorithm     string             `json:"algorithm,omitempty"` // "collaborative", "content", "hybrid"
	Strategies    []string           `json:"strategies,omitempty"`
	Status        *domain.PostStatus `json:"status,omitempty"`
	PostTypeID    *uuid.UUID         `json:"post_type_id,omitempty"`
	Limit         int                `json:"limit" validate:"min=1,max=50"`
	Diversify     bool               `json:"diversify,omitempty"`
	ExcludeViewed bool               `json:"exclude_viewed,omitempty"`
}

// TrendingPostsInput represents input for getting trending posts
type TrendingPostsInput struct {
	Timeframe  string             `json:"timeframe" validate:"required"` // "1h", "24h", "7d", "30d"
	Algorithm  string             `json:"algorithm,omitempty"`           // "views", "engagement", "velocity", "combined"
	Status     *domain.PostStatus `json:"status,omitempty"`
	PostTypeID *uuid.UUID         `json:"post_type_id,omitempty"`
	CategoryID *uuid.UUID         `json:"category_id,omitempty"`
	MinViews   int64              `json:"min_views,omitempty"`
	Limit      int                `json:"limit" validate:"min=1,max=100"`
}

// PopularPostsInput represents input for getting popular posts
type PopularPostsInput struct {
	Timeframe  string             `json:"timeframe" validate:"required"` // "all", "1y", "30d", "7d", "24h"
	MetricType string             `json:"metric_type,omitempty"`         // "views", "engagement", "shares", "combined"
	Status     *domain.PostStatus `json:"status,omitempty"`
	PostTypeID *uuid.UUID         `json:"post_type_id,omitempty"`
	CategoryID *uuid.UUID         `json:"category_id,omitempty"`
	MinViews   int64              `json:"min_views,omitempty"`
	Limit      int                `json:"limit" validate:"min=1,max=100"`
}

// SearchAnalyticsOutput represents search analytics data
type SearchAnalyticsOutput struct {
	Analytics *domain.SearchAnalytics `json:"analytics"`
}

// NewPostSearchUsecase creates a new PostSearchUsecase instance
func NewPostSearchUsecase(
	searchRepo domain.PostSearchRepository,
	postRepo domain.PostRepository,
	postTypeRepo domain.PostTypeRepository,
) *PostSearchUsecase {
	return &PostSearchUsecase{
		searchRepo:   searchRepo,
		postRepo:     postRepo,
		postTypeRepo: postTypeRepo,
	}
}

// SearchPosts performs full-text search with advanced filtering
func (uc *PostSearchUsecase) SearchPosts(ctx context.Context, input *SearchPostsInput) (*SearchPostsOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	startTime := time.Now()

	// If no query provided, use regular listing
	if input.Query == "" {
		return uc.listPostsAsSearch(ctx, input, startTime)
	}

	// Prepare search options
	searchOptions := domain.FullTextSearchOptions{
		Limit:             input.Limit,
		Offset:            input.Offset,
		Operator:          input.Operator,
		MinScore:          input.MinScore,
		BoostFields:       input.BoostFields,
		SearchFields:      input.SearchFields,
		Status:            input.Status,
		AuthorID:          input.AuthorID,
		PostTypeID:        input.PostTypeID,
		CategoryID:        input.CategoryID,
		TagIDs:            input.TagIDs,
		PublishedAfter:    input.PublishedAfter,
		PublishedBefore:   input.PublishedBefore,
		IncludeHighlights: input.IncludeHighlights,
		IncludeSnippets:   input.IncludeSnippets,
		SortBy:            input.SortBy,
		SortOrder:         input.SortOrder,
	}

	// Set defaults for search options
	if searchOptions.Operator == "" {
		searchOptions.Operator = "AND"
	}
	if searchOptions.SortBy == "" {
		searchOptions.SortBy = "relevance"
	}
	if searchOptions.SortOrder == "" {
		searchOptions.SortOrder = "desc"
	}

	// Perform the search
	var results *domain.SearchResult
	var facets map[string][]domain.FacetValue
	var err error

	if len(input.Facets) > 0 {
		// Use faceted search
		facetedOptions := domain.FacetedSearchOptions{
			FullTextSearchOptions: searchOptions,
			FacetFields:           input.Facets,
			FacetLimits:           make(map[string]int),
			FacetMinCounts:        make(map[string]int),
			FacetSorts:            make(map[string]string),
		}

		// Set facet defaults
		for _, facet := range input.Facets {
			facetedOptions.FacetLimits[facet] = 10
			facetedOptions.FacetMinCounts[facet] = 1
			facetedOptions.FacetSorts[facet] = "count"
		}

		facetedResult, err := uc.searchRepo.SearchWithFacets(ctx, input.Query, input.Facets, facetedOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to perform faceted search: %w", err)
		}

		results = &facetedResult.SearchResult
		facets = facetedResult.Facets
	} else {
		// Use regular full-text search
		results, err = uc.searchRepo.SearchFullText(ctx, input.Query, searchOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to perform full-text search: %w", err)
		}
	}

	searchTime := time.Since(startTime)

	return &SearchPostsOutput{
		Results:     results,
		Facets:      facets,
		SearchTime:  searchTime,
		TotalCount:  results.TotalCount,
		MaxScore:    results.MaxScore,
		Suggestions: results.Suggestions,
	}, nil
}

// SearchSimilarPosts finds posts similar to a given post
func (uc *PostSearchUsecase) SearchSimilarPosts(ctx context.Context, input *SimilarPostsInput) (*SearchPostsOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	if input.PostID == uuid.Nil {
		return nil, errors.New("post ID cannot be nil")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.Limit > 50 {
		input.Limit = 50
	}
	if input.Algorithm == "" {
		input.Algorithm = "combined"
	}
	if input.MinSimilarity == 0 {
		input.MinSimilarity = 0.3 // 30% similarity threshold
	}

	startTime := time.Now()

	// Prepare similarity options
	options := domain.SimilarSearchOptions{
		Limit:           input.Limit,
		Offset:          0,
		Algorithm:       input.Algorithm,
		MinSimilarity:   input.MinSimilarity,
		BoostRecent:     input.BoostRecent,
		BoostPopular:    input.BoostPopular,
		ExcludeOriginal: true,
		Status:          input.Status,
		PostTypeID:      input.PostTypeID,
		IncludeScore:    true,
		IncludeReason:   true,
	}

	// Perform similarity search
	results, err := uc.searchRepo.SearchSimilar(ctx, input.PostID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar posts: %w", err)
	}

	searchTime := time.Since(startTime)

	return &SearchPostsOutput{
		Results:    results,
		SearchTime: searchTime,
		TotalCount: results.TotalCount,
		MaxScore:   results.MaxScore,
	}, nil
}

// GetRecommendations gets personalized post recommendations for a user
func (uc *PostSearchUsecase) GetRecommendations(ctx context.Context, input *RecommendationsInput) (*SearchPostsOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	if input.UserID == uuid.Nil {
		return nil, errors.New("user ID cannot be nil")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 50 {
		input.Limit = 50
	}
	if input.Algorithm == "" {
		input.Algorithm = "hybrid"
	}

	startTime := time.Now()

	// Prepare recommendation options
	options := domain.RecommendationOptions{
		Limit:          input.Limit,
		Offset:         0,
		Algorithm:      input.Algorithm,
		Strategies:     input.Strategies,
		Diversify:      input.Diversify,
		UserHistory:    true,
		UserCategories: true,
		UserTags:       true,
		Status:         input.Status,
		PostTypeID:     input.PostTypeID,
		ExcludeViewed:  input.ExcludeViewed,
		IncludeScore:   true,
		IncludeReason:  true,
	}

	// Get recommendations
	posts, err := uc.searchRepo.GetRecommendations(ctx, input.UserID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Convert to search result format
	postMatches := make([]*domain.PostSearchMatch, len(posts))
	for i, post := range posts {
		postMatches[i] = &domain.PostSearchMatch{
			Post:   post,
			Score:  0.8, // Default recommendation score
			Reason: "personalized recommendation",
		}
	}

	results := &domain.SearchResult{
		Posts:      postMatches,
		TotalCount: int64(len(posts)),
		SearchTime: time.Since(startTime),
		MaxScore:   1.0,
	}

	return &SearchPostsOutput{
		Results:    results,
		SearchTime: time.Since(startTime),
		TotalCount: results.TotalCount,
		MaxScore:   results.MaxScore,
	}, nil
}

// GetTrendingPosts gets trending posts based on various metrics
func (uc *PostSearchUsecase) GetTrendingPosts(ctx context.Context, input *TrendingPostsInput) (*SearchPostsOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	if input.Timeframe == "" {
		return nil, errors.New("timeframe is required")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.Algorithm == "" {
		input.Algorithm = "combined"
	}

	startTime := time.Now()

	// Prepare trending options
	options := domain.TrendingOptions{
		Limit:          input.Limit,
		Offset:         0,
		Timeframe:      input.Timeframe,
		Algorithm:      input.Algorithm,
		MinViews:       input.MinViews,
		Status:         input.Status,
		PostTypeID:     input.PostTypeID,
		CategoryID:     input.CategoryID,
		IncludeScore:   true,
		IncludeMetrics: true,
	}

	// Get trending posts
	posts, err := uc.searchRepo.GetTrending(ctx, input.Timeframe, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending posts: %w", err)
	}

	// Convert to search result format
	postMatches := make([]*domain.PostSearchMatch, len(posts))
	for i, post := range posts {
		postMatches[i] = &domain.PostSearchMatch{
			Post:   post,
			Score:  0.9, // High score for trending posts
			Reason: fmt.Sprintf("trending in %s", input.Timeframe),
		}
	}

	results := &domain.SearchResult{
		Posts:      postMatches,
		TotalCount: int64(len(posts)),
		SearchTime: time.Since(startTime),
		MaxScore:   1.0,
	}

	return &SearchPostsOutput{
		Results:    results,
		SearchTime: time.Since(startTime),
		TotalCount: results.TotalCount,
		MaxScore:   results.MaxScore,
	}, nil
}

// GetPopularPosts gets popular posts based on various metrics
func (uc *PostSearchUsecase) GetPopularPosts(ctx context.Context, input *PopularPostsInput) (*SearchPostsOutput, error) {
	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	if input.Timeframe == "" {
		return nil, errors.New("timeframe is required")
	}

	// Set defaults
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.MetricType == "" {
		input.MetricType = "views"
	}

	startTime := time.Now()

	// Prepare popular options
	options := domain.PopularOptions{
		Limit:          input.Limit,
		Offset:         0,
		Timeframe:      input.Timeframe,
		MetricType:     input.MetricType,
		MinViews:       input.MinViews,
		Status:         input.Status,
		PostTypeID:     input.PostTypeID,
		CategoryID:     input.CategoryID,
		IncludeMetrics: true,
	}

	// Get popular posts
	posts, err := uc.searchRepo.GetPopular(ctx, input.Timeframe, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular posts: %w", err)
	}

	// Convert to search result format
	postMatches := make([]*domain.PostSearchMatch, len(posts))
	for i, post := range posts {
		postMatches[i] = &domain.PostSearchMatch{
			Post:   post,
			Score:  0.85, // High score for popular posts
			Reason: fmt.Sprintf("popular in %s", input.Timeframe),
		}
	}

	results := &domain.SearchResult{
		Posts:      postMatches,
		TotalCount: int64(len(posts)),
		SearchTime: time.Since(startTime),
		MaxScore:   1.0,
	}

	return &SearchPostsOutput{
		Results:    results,
		SearchTime: time.Since(startTime),
		TotalCount: results.TotalCount,
		MaxScore:   results.MaxScore,
	}, nil
}

// GetSearchAnalytics gets search analytics for a given timeframe
func (uc *PostSearchUsecase) GetSearchAnalytics(ctx context.Context, timeframe string) (*SearchAnalyticsOutput, error) {
	if timeframe == "" {
		timeframe = "30d" // Default to 30 days
	}

	analytics, err := uc.searchRepo.GetSearchAnalytics(ctx, timeframe)
	if err != nil {
		return nil, fmt.Errorf("failed to get search analytics: %w", err)
	}

	return &SearchAnalyticsOutput{
		Analytics: analytics,
	}, nil
}

// GetSearchSuggestions gets auto-completion suggestions for a partial query
func (uc *PostSearchUsecase) GetSearchSuggestions(ctx context.Context, partial string, maxSuggestions int) ([]string, error) {
	if partial == "" {
		return []string{}, nil
	}

	if maxSuggestions <= 0 {
		maxSuggestions = 10
	}

	suggestions, err := uc.searchRepo.GetSearchSuggestions(ctx, partial, maxSuggestions)
	if err != nil {
		return nil, fmt.Errorf("failed to get search suggestions: %w", err)
	}

	return suggestions, nil
}

// listPostsAsSearch converts regular listing to search result format
func (uc *PostSearchUsecase) listPostsAsSearch(ctx context.Context, input *SearchPostsInput, startTime time.Time) (*SearchPostsOutput, error) {
	// Use regular post listing when no query is provided
	listOptions := domain.ListPostsOptions{
		Limit:             input.Limit,
		Offset:            input.Offset,
		Status:            input.Status,
		AuthorID:          input.AuthorID,
		PostTypeID:        input.PostTypeID,
		CategoryID:        input.CategoryID,
		PublishedAfter:    input.PublishedAfter,
		PublishedBefore:   input.PublishedBefore,
		SortBy:            input.SortBy,
		SortOrder:         input.SortOrder,
		IncludeCategories: true,
		IncludeTags:       true,
		IncludeAuthor:     true,
		IncludePostType:   true,
	}

	posts, err := uc.postRepo.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	// Get total count
	countOptions := domain.CountPostsOptions{
		Status:          input.Status,
		AuthorID:        input.AuthorID,
		PostTypeID:      input.PostTypeID,
		CategoryID:      input.CategoryID,
		PublishedAfter:  input.PublishedAfter,
		PublishedBefore: input.PublishedBefore,
	}

	totalCount, err := uc.postRepo.Count(ctx, countOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	// Convert to search result format
	postMatches := make([]*domain.PostSearchMatch, len(posts))
	for i, post := range posts {
		postMatches[i] = &domain.PostSearchMatch{
			Post:  post,
			Score: 1.0, // Default score for listing
		}
	}

	results := &domain.SearchResult{
		Posts:      postMatches,
		TotalCount: totalCount,
		SearchTime: time.Since(startTime),
		MaxScore:   1.0,
	}

	return &SearchPostsOutput{
		Results:    results,
		SearchTime: time.Since(startTime),
		TotalCount: totalCount,
		MaxScore:   1.0,
	}, nil
}
