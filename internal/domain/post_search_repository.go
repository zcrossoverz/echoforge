package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostSearchRepository defines the interface for advanced search and filtering operations
// Provides full-text search, faceted search, and complex filtering capabilities
type PostSearchRepository interface {
	// Full-Text Search
	SearchFullText(ctx context.Context, query string, options FullTextSearchOptions) (*SearchResult, error)
	SearchByKeywords(ctx context.Context, keywords []string, options KeywordSearchOptions) (*SearchResult, error)
	SearchSimilar(ctx context.Context, postID uuid.UUID, options SimilarSearchOptions) (*SearchResult, error)

	// Faceted Search
	SearchWithFacets(ctx context.Context, query string, facets []string, options FacetedSearchOptions) (*FacetedSearchResult, error)
	GetFacetValues(ctx context.Context, facetName string, options FacetValuesOptions) ([]FacetValue, error)

	// Advanced Filtering
	FilterByMultipleConditions(ctx context.Context, conditions []FilterCondition, options FilterOptions) ([]*Post, error)
	FilterByDateRange(ctx context.Context, field string, startDate, endDate time.Time, options FilterOptions) ([]*Post, error)
	FilterByNumericRange(ctx context.Context, field string, min, max float64, options FilterOptions) ([]*Post, error)

	// Geographic Search (if location metadata exists)
	SearchByLocation(ctx context.Context, lat, lng float64, radius float64, options LocationSearchOptions) ([]*Post, error)
	SearchByBounds(ctx context.Context, northLat, southLat, eastLng, westLng float64, options LocationSearchOptions) ([]*Post, error)

	// Trending and Popular Content
	GetTrending(ctx context.Context, timeframe string, options TrendingOptions) ([]*Post, error)
	GetPopular(ctx context.Context, timeframe string, options PopularOptions) ([]*Post, error)
	GetRecentlyViewed(ctx context.Context, userID uuid.UUID, options RecentViewsOptions) ([]*Post, error)

	// Recommendation Engine
	GetRecommendations(ctx context.Context, userID uuid.UUID, options RecommendationOptions) ([]*Post, error)
	GetRelatedPosts(ctx context.Context, postID uuid.UUID, options RelatedPostsOptions) ([]*Post, error)

	// Analytics and Insights
	GetSearchAnalytics(ctx context.Context, timeframe string) (*SearchAnalytics, error)
	GetPopularSearchTerms(ctx context.Context, limit int, timeframe string) ([]SearchTerm, error)

	// Auto-completion and Suggestions
	GetSearchSuggestions(ctx context.Context, partial string, maxSuggestions int) ([]string, error)
	GetPopularTags(ctx context.Context, limit int, timeframe string) ([]TagSuggestion, error)

	// Export and Bulk Operations
	ExportSearchResults(ctx context.Context, query string, format string, options ExportOptions) ([]byte, error)
	BulkIndex(ctx context.Context, posts []*Post) error
	BulkDelete(ctx context.Context, postIDs []uuid.UUID) error

	// Search Index Management
	ReindexPost(ctx context.Context, postID uuid.UUID) error
	ReindexAll(ctx context.Context) error
	GetIndexStats(ctx context.Context) (*IndexStats, error)
}

// FullTextSearchOptions provides options for full-text search operations
type FullTextSearchOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Search behavior
	Operator      string             // "AND", "OR"
	MinScore      float64            // Minimum relevance score
	BoostFields   map[string]float64 // Field boost multipliers
	SearchFields  []string           // Fields to search in
	ExcludeFields []string           // Fields to exclude from search

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID
	TagIDs     []uuid.UUID

	// Date filtering
	PublishedAfter  *time.Time
	PublishedBefore *time.Time

	// Result options
	IncludeHighlights bool
	IncludeSnippets   bool
	SnippetLength     int
	HighlightTags     []string // HTML tags for highlighting

	// Sorting
	SortBy    string // "relevance", "date", "popularity", "title"
	SortOrder string // "asc", "desc"
}

// KeywordSearchOptions provides options for keyword-based search
type KeywordSearchOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Keyword behavior
	ExactMatch     bool
	FuzzyTolerance int  // Edit distance for fuzzy matching
	MinWordLength  int  // Minimum word length to consider
	Stemming       bool // Enable word stemming
	Synonyms       bool // Enable synonym expansion

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID

	// Result options
	IncludeKeywordStats bool
	HighlightKeywords   bool

	// Sorting
	SortBy    string
	SortOrder string
}

// SimilarSearchOptions provides options for finding similar posts
type SimilarSearchOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Similarity algorithm
	Algorithm     string  // "content", "tags", "categories", "combined"
	MinSimilarity float64 // Minimum similarity score (0.0 - 1.0)
	BoostRecent   bool    // Boost more recent posts
	BoostPopular  bool    // Boost popular posts

	// Filtering
	ExcludeOriginal bool
	Status          *PostStatus
	AuthorID        *uuid.UUID
	PostTypeID      *uuid.UUID

	// Result options
	IncludeScore  bool
	IncludeReason bool // Reason for similarity
}

// FacetedSearchOptions provides options for faceted search
type FacetedSearchOptions struct {
	// Basic search options
	FullTextSearchOptions

	// Facet configuration
	FacetFields    []string          // Fields to generate facets for
	FacetLimits    map[string]int    // Maximum values per facet
	FacetMinCounts map[string]int    // Minimum count to include facet value
	FacetSorts     map[string]string // Sort order for facet values ("count", "value")

	// Selected facets (filters)
	SelectedFacets map[string][]string
}

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string        `json:"field"`
	Operator string        `json:"operator"` // "eq", "ne", "gt", "gte", "lt", "lte", "in", "nin", "contains", "starts_with"
	Value    interface{}   `json:"value"`
	Values   []interface{} `json:"values,omitempty"` // For "in" and "nin" operators
}

// FilterOptions provides options for filtering operations
type FilterOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Logical operators
	LogicalOperator string // "AND", "OR"

	// Sorting
	SortBy    string
	SortOrder string

	// Result options
	IncludeCategories bool
	IncludeTags       bool
	IncludeAuthor     bool
	IncludePostType   bool
}

// LocationSearchOptions provides options for location-based search
type LocationSearchOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Units
	DistanceUnit string // "km", "miles"

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID

	// Sorting
	SortBy    string // "distance", "relevance", "date"
	SortOrder string

	// Result options
	IncludeDistance bool
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Posts       []*PostSearchMatch `json:"posts"`
	TotalCount  int64              `json:"total_count"`
	SearchTime  time.Duration      `json:"search_time"`
	MaxScore    float64            `json:"max_score,omitempty"`
	Suggestions []string           `json:"suggestions,omitempty"`
}

// PostSearchMatch represents a post match in search results
type PostSearchMatch struct {
	Post       *Post               `json:"post"`
	Score      float64             `json:"score,omitempty"`
	Highlights map[string][]string `json:"highlights,omitempty"`
	Snippets   map[string]string   `json:"snippets,omitempty"`
	Reason     string              `json:"reason,omitempty"`
	Distance   float64             `json:"distance,omitempty"`
}

// FacetedSearchResult represents the result of a faceted search
type FacetedSearchResult struct {
	SearchResult
	Facets map[string][]FacetValue `json:"facets"`
}

// FacetValue represents a single facet value with count
type FacetValue struct {
	Value    string `json:"value"`
	Count    int64  `json:"count"`
	Selected bool   `json:"selected,omitempty"`
}

// FacetValuesOptions provides options for retrieving facet values
type FacetValuesOptions struct {
	// Filtering
	Query    string // Filter facet values by query
	MinCount int    // Minimum count to include
	Limit    int    // Maximum number of values

	// Sorting
	SortBy    string // "count", "value"
	SortOrder string // "asc", "desc"
}

// TrendingOptions provides options for retrieving trending content
type TrendingOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Timeframe
	Timeframe string // "1h", "24h", "7d", "30d"

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID

	// Algorithm
	Algorithm string // "views", "engagement", "velocity", "combined"
	MinViews  int64  // Minimum view count

	// Result options
	IncludeScore   bool
	IncludeMetrics bool
}

// PopularOptions provides options for retrieving popular content
type PopularOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Timeframe
	Timeframe string // "all", "1y", "30d", "7d", "24h"

	// Filtering
	Status     *PostStatus
	AuthorID   *uuid.UUID
	PostTypeID *uuid.UUID
	CategoryID *uuid.UUID

	// Metrics
	MetricType string // "views", "engagement", "shares", "combined"
	MinViews   int64

	// Result options
	IncludeMetrics bool
}

// RecentViewsOptions provides options for retrieving recently viewed posts
type RecentViewsOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Timeframe
	MaxAge time.Duration // Maximum age of view records

	// Filtering
	Status     *PostStatus
	PostTypeID *uuid.UUID

	// Deduplication
	Deduplicate bool // Remove duplicate posts
}

// RecommendationOptions provides options for getting recommendations
type RecommendationOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Algorithm
	Algorithm  string   // "collaborative", "content", "hybrid"
	Strategies []string // Multiple recommendation strategies
	Diversify  bool     // Ensure diverse recommendations

	// User context
	UserHistory    bool // Use user's post history
	UserCategories bool // Use user's preferred categories
	UserTags       bool // Use user's preferred tags

	// Filtering
	Status        *PostStatus
	PostTypeID    *uuid.UUID
	ExcludeViewed bool // Exclude previously viewed posts

	// Result options
	IncludeScore  bool
	IncludeReason bool
}

// RelatedPostsOptions provides options for finding related posts
type RelatedPostsOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Relationship types
	BySameAuthor   bool
	BySameCategory bool
	BySameTags     bool
	ByContent      bool
	ByMetadata     bool

	// Filtering
	Status          *PostStatus
	PostTypeID      *uuid.UUID
	ExcludeOriginal bool

	// Scoring
	WeightAuthor   float64
	WeightCategory float64
	WeightTags     float64
	WeightContent  float64

	// Result options
	IncludeScore  bool
	IncludeReason bool
}

// SearchAnalytics represents search analytics data
type SearchAnalytics struct {
	Period           string           `json:"period"`
	TotalSearches    int64            `json:"total_searches"`
	UniqueSearchers  int64            `json:"unique_searchers"`
	AverageResults   float64          `json:"average_results"`
	ZeroResultRate   float64          `json:"zero_result_rate"`
	ClickThroughRate float64          `json:"click_through_rate"`
	TopQueries       []SearchTerm     `json:"top_queries"`
	TopCategories    []string         `json:"top_categories"`
	SearchTrends     map[string]int64 `json:"search_trends"`
}

// SearchTerm represents a search term with statistics
type SearchTerm struct {
	Term         string  `json:"term"`
	Count        int64   `json:"count"`
	ResultCount  int64   `json:"result_count"`
	ClickCount   int64   `json:"click_count"`
	ClickRate    float64 `json:"click_rate"`
	LastSearched string  `json:"last_searched"`
}

// TagSuggestion represents a tag suggestion with usage statistics
type TagSuggestion struct {
	Tag       *PostTag `json:"tag"`
	Usage     int64    `json:"usage"`
	Trend     string   `json:"trend"` // "rising", "stable", "declining"
	Relevance float64  `json:"relevance"`
}

// ExportOptions provides options for exporting search results
type ExportOptions struct {
	Format   string   `json:"format"`   // "json", "csv", "xml"
	Fields   []string `json:"fields"`   // Fields to include in export
	Compress bool     `json:"compress"` // Whether to compress the output
}

// IndexStats represents search index statistics
type IndexStats struct {
	TotalDocuments   int64                  `json:"total_documents"`
	IndexedDocuments int64                  `json:"indexed_documents"`
	PendingDocuments int64                  `json:"pending_documents"`
	IndexSize        int64                  `json:"index_size"` // Size in bytes
	LastIndexed      time.Time              `json:"last_indexed"`
	IndexHealth      string                 `json:"index_health"` // "green", "yellow", "red"
	FieldStats       map[string]int64       `json:"field_stats"`
	SearchStats      map[string]interface{} `json:"search_stats"`
}
