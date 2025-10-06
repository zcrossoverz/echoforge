package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zcrossoverz/echoforge/internal/domain"
	"github.com/zcrossoverz/echoforge/internal/usecase"
)

// SearchHandler handles HTTP requests for search operations
type SearchHandler struct {
	searchUsecase *usecase.PostSearchUsecase
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchUsecase *usecase.PostSearchUsecase) *SearchHandler {
	return &SearchHandler{
		searchUsecase: searchUsecase,
	}
}

// SearchResponse represents the response structure for search operations
type SearchResponse struct {
	Results    []SearchPostSummaryResponse     `json:"results"`
	Facets     map[string][]FacetValueResponse `json:"facets,omitempty"`
	Pagination PaginationResponse              `json:"pagination"`
	Query      SearchQueryResponse             `json:"query"`
	SearchTime string                          `json:"search_time,omitempty"`
	MaxScore   float64                         `json:"max_score,omitempty"`
}

// SearchPostSummaryResponse represents a summarized post in search results (extends PostSummaryResponse)
type SearchPostSummaryResponse struct {
	PostSummaryResponse
	Score      float64             `json:"score,omitempty"`
	Highlights map[string][]string `json:"highlights,omitempty"`
	Snippet    string              `json:"snippet,omitempty"`
}

// FacetValueResponse represents a facet value with count
type FacetValueResponse struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// SearchQueryResponse represents the executed search query
type SearchQueryResponse struct {
	Query   string                 `json:"q"`
	Filters map[string]interface{} `json:"filters"`
}

// GlobalSearch handles GET /api/v1/search
func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	// Parse required query parameter
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query 'q' is required",
		})
		return
	}

	// Parse search type
	searchType := c.DefaultQuery("type", "posts")
	if searchType != "posts" && searchType != "categories" && searchType != "tags" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid search type. Must be 'posts', 'categories', or 'tags'",
		})
		return
	}

	// For now, only support posts search
	if searchType != "posts" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Search for categories and tags not yet implemented",
		})
		return
	}

	// Parse pagination
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	// Parse optional filters
	var postTypeID *uuid.UUID
	if postTypeIDStr := c.Query("postTypeId"); postTypeIDStr != "" {
		if parsed, err := uuid.Parse(postTypeIDStr); err == nil {
			postTypeID = &parsed
		}
	}

	var categoryID *uuid.UUID
	if categoryIDStr := c.Query("categoryId"); categoryIDStr != "" {
		if parsed, err := uuid.Parse(categoryIDStr); err == nil {
			categoryID = &parsed
		}
	}

	var tagIDs []uuid.UUID
	if tagIDStr := c.Query("tagId"); tagIDStr != "" {
		if parsed, err := uuid.Parse(tagIDStr); err == nil {
			tagIDs = append(tagIDs, parsed)
		}
	}

	// Parse date filters
	var dateFrom *time.Time
	if dateFromStr := c.Query("dateFrom"); dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = &parsed
		}
	}

	var dateTo *time.Time
	if dateToStr := c.Query("dateTo"); dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			// Set to end of day
			endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &endOfDay
		}
	}

	// Create search input
	searchInput := &usecase.SearchPostsInput{
		Query:             query,
		PostTypeID:        postTypeID,
		CategoryID:        categoryID,
		TagIDs:            tagIDs,
		PublishedAfter:    dateFrom,
		PublishedBefore:   dateTo,
		Limit:             limit,
		Offset:            offset,
		SortBy:            c.DefaultQuery("sort", "relevance"),
		SortOrder:         c.DefaultQuery("order", "desc"),
		IncludeHighlights: c.Query("highlights") == "true",
		IncludeSnippets:   c.Query("snippets") == "true",
		Facets:            []string{"postTypes", "categories", "tags"},
	}

	// Execute search
	output, err := h.searchUsecase.SearchPosts(c.Request.Context(), searchInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed",
		})
		return
	}

	// Convert to response
	response := h.convertSearchOutputToResponse(output, query, page, limit)

	c.JSON(http.StatusOK, response)
}

// GetSuggestions handles GET /api/v1/search/suggestions
func (h *SearchHandler) GetSuggestions(c *gin.Context) {
	partial := strings.TrimSpace(c.Query("q"))
	if partial == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	maxSuggestions := 10
	if maxStr := c.Query("limit"); maxStr != "" {
		if parsed, err := strconv.Atoi(maxStr); err == nil && parsed > 0 && parsed <= 50 {
			maxSuggestions = parsed
		}
	}

	suggestions, err := h.searchUsecase.GetSearchSuggestions(c.Request.Context(), partial, maxSuggestions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get search suggestions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
	})
}

// GetTrending handles GET /api/v1/search/trending
func (h *SearchHandler) GetTrending(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	timeframe := c.DefaultQuery("timeframe", "week") // day, week, month

	input := &usecase.TrendingPostsInput{
		Timeframe: timeframe,
		Limit:     limit,
	}

	output, err := h.searchUsecase.GetTrendingPosts(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get trending posts",
		})
		return
	}

	// Convert to response format
	response := h.convertSearchOutputToTrendingResponse(output)

	c.JSON(http.StatusOK, response)
}

// GetPopular handles GET /api/v1/search/popular
func (h *SearchHandler) GetPopular(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	var postTypeID *uuid.UUID
	if postTypeIDStr := c.Query("postTypeId"); postTypeIDStr != "" {
		if parsed, err := uuid.Parse(postTypeIDStr); err == nil {
			postTypeID = &parsed
		}
	}

	input := &usecase.PopularPostsInput{
		PostTypeID: postTypeID,
		Limit:      limit,
	}

	output, err := h.searchUsecase.GetPopularPosts(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get popular posts",
		})
		return
	}

	// Convert to response format
	response := h.convertSearchOutputToPopularResponse(output)

	c.JSON(http.StatusOK, response)
}

// Helper methods

// convertSearchOutputToResponse converts search use case output to HTTP response
func (h *SearchHandler) convertSearchOutputToResponse(output *usecase.SearchPostsOutput, query string, page, limit int) *SearchResponse {
	// Convert results
	results := make([]SearchPostSummaryResponse, 0)
	if output.Results != nil && output.Results.Posts != nil {
		results = make([]SearchPostSummaryResponse, len(output.Results.Posts))
		for i, postMatch := range output.Results.Posts {
			results[i] = h.convertPostMatchToSearchResponse(postMatch)
		}
	}

	// Convert facets
	facets := make(map[string][]FacetValueResponse)
	for facetName, facetValues := range output.Facets {
		facetResponses := make([]FacetValueResponse, len(facetValues))
		for i, fv := range facetValues {
			facetResponses[i] = FacetValueResponse{
				Value: fv.Value,
				Count: fv.Count,
			}
		}
		facets[facetName] = facetResponses
	}

	// Calculate total pages
	totalPages := int(output.TotalCount) / limit
	if int(output.TotalCount)%limit > 0 {
		totalPages++
	}

	return &SearchResponse{
		Results: results,
		Facets:  facets,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      output.TotalCount,
			TotalPages: totalPages,
		},
		Query: SearchQueryResponse{
			Query:   query,
			Filters: make(map[string]interface{}),
		},
		SearchTime: output.SearchTime.String(),
		MaxScore:   output.MaxScore,
	}
}

// convertPostMatchToSearchResponse converts a PostSearchMatch to SearchPostSummaryResponse
func (h *SearchHandler) convertPostMatchToSearchResponse(postMatch *domain.PostSearchMatch) SearchPostSummaryResponse {
	// Convert the core post data
	post := postMatch.Post
	var publishedAt *string
	if post.PublishedAt != nil {
		formatted := post.PublishedAt.Format("2006-01-02T15:04:05Z")
		publishedAt = &formatted
	}

	// Create base post summary (truncate content for excerpt)
	content := post.Content
	if len(content) > 200 {
		content = content[:200] + "..."
	}

	baseSummary := PostSummaryResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       content,
		AuthorID:      post.AuthorID,
		Status:        string(post.Status),
		CreatedAt:     post.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ViewCount:     int64(post.ViewCount),
		CategoryCount: 0, // TODO: Get from relationships when available
		TagCount:      0, // TODO: Get from relationships when available
		PublishedAt:   publishedAt,
	}

	// Create search-specific response
	searchResponse := SearchPostSummaryResponse{
		PostSummaryResponse: baseSummary,
		Score:               postMatch.Score,
		Highlights:          postMatch.Highlights,
	}

	// Add snippet from PostSearchMatch
	if len(postMatch.Snippets) > 0 {
		// Take the first snippet available
		for _, snippet := range postMatch.Snippets {
			searchResponse.Snippet = snippet
			break
		}
	}

	return searchResponse
}

// convertSearchOutputToTrendingResponse converts trending posts output to response
func (h *SearchHandler) convertSearchOutputToTrendingResponse(output *usecase.SearchPostsOutput) gin.H {
	results := make([]SearchPostSummaryResponse, 0)
	if output.Results != nil && output.Results.Posts != nil {
		results = make([]SearchPostSummaryResponse, len(output.Results.Posts))
		for i, postMatch := range output.Results.Posts {
			results[i] = h.convertPostMatchToSearchResponse(postMatch)
		}
	}

	return gin.H{
		"trending": results,
		"total":    output.TotalCount,
	}
}

// convertSearchOutputToPopularResponse converts popular posts output to response
func (h *SearchHandler) convertSearchOutputToPopularResponse(output *usecase.SearchPostsOutput) gin.H {
	results := make([]SearchPostSummaryResponse, 0)
	if output.Results != nil && output.Results.Posts != nil {
		results = make([]SearchPostSummaryResponse, len(output.Results.Posts))
		for i, postMatch := range output.Results.Posts {
			results[i] = h.convertPostMatchToSearchResponse(postMatch)
		}
	}

	return gin.H{
		"popular": results,
		"total":   output.TotalCount,
	}
}
