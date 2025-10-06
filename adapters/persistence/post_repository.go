package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// GormPostRepository implements domain.PostRepository using GORM
type GormPostRepository struct {
	db *gorm.DB
}

// NewGormPostRepository creates a new GORM-based post repository
func NewGormPostRepository(db *gorm.DB) domain.PostRepository {
	return &GormPostRepository{
		db: db,
	}
}

// Create creates a new post in the database
func (r *GormPostRepository) Create(ctx context.Context, post *domain.Post) error {
	if post == nil {
		return fmt.Errorf("post cannot be nil")
	}

	// Use transaction for consistency
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a post by its ID
func (r *GormPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}

	var post domain.Post
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&post).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}

	return &post, nil
}

// Update updates an existing post
func (r *GormPostRepository) Update(ctx context.Context, post *domain.Post) error {
	if post == nil {
		return fmt.Errorf("post cannot be nil")
	}
	if post.ID == uuid.Nil {
		return fmt.Errorf("post ID cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the post
		result := tx.Save(post)
		if result.Error != nil {
			return fmt.Errorf("failed to update post: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("post not found")
		}
		return nil
	})
}

// Delete deletes a post by ID (soft delete by setting status to archived)
func (r *GormPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("id cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Soft delete by updating status to archived
		result := tx.Model(&domain.Post{}).
			Where("id = ?", id).
			Update("status", domain.PostStatusArchived)

		if result.Error != nil {
			return fmt.Errorf("failed to delete post: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("post not found")
		}
		return nil
	})
}

// List retrieves posts with filtering and pagination
func (r *GormPostRepository) List(ctx context.Context, options domain.ListPostsOptions) ([]*domain.Post, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{})

	// Apply filters
	query = r.applyListFilters(query, options)

	// Apply sorting
	query = r.applySorting(query, options.SortBy, options.SortOrder)

	// Apply pagination
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}

	var posts []*domain.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return posts, nil
}

// Count counts posts with filtering
func (r *GormPostRepository) Count(ctx context.Context, options domain.CountPostsOptions) (int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Post{})

	// Apply filters
	if options.Status != nil {
		query = query.Where("status = ?", *options.Status)
	}
	if options.AuthorID != nil {
		query = query.Where("author_id = ?", *options.AuthorID)
	}
	if options.PostTypeID != nil {
		query = query.Where("post_type_id = ?", *options.PostTypeID)
	}
	if options.CategoryID != nil {
		query = query.Joins("JOIN post_category_assignments pca ON posts.id = pca.post_id").
			Where("pca.category_id = ?", *options.CategoryID)
	}
	if options.TagID != nil {
		query = query.Joins("JOIN post_tag_assignments pta ON posts.id = pta.post_id").
			Where("pta.tag_id = ?", *options.TagID)
	}

	// Date filters
	if options.PublishedAfter != nil {
		query = query.Where("published_at > ?", *options.PublishedAfter)
	}
	if options.PublishedBefore != nil {
		query = query.Where("published_at < ?", *options.PublishedBefore)
	}
	if options.CreatedAfter != nil {
		query = query.Where("created_at > ?", *options.CreatedAfter)
	}
	if options.CreatedBefore != nil {
		query = query.Where("created_at < ?", *options.CreatedBefore)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}

	return count, nil
}

// GetByStatus retrieves posts by status with pagination
func (r *GormPostRepository) GetByStatus(ctx context.Context, status domain.PostStatus, limit, offset int) ([]*domain.Post, error) {
	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by status: %w", err)
	}

	return posts, nil
}

// GetPublished retrieves published posts with pagination
func (r *GormPostRepository) GetPublished(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	return r.GetByStatus(ctx, domain.PostStatusPublished, limit, offset)
}

// GetScheduled retrieves scheduled posts that should be published
func (r *GormPostRepository) GetScheduled(ctx context.Context, before time.Time) ([]*domain.Post, error) {
	var posts []*domain.Post
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", domain.PostStatusScheduled, before).
		Order("scheduled_at ASC").
		Find(&posts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled posts: %w", err)
	}

	return posts, nil
}

// GetPendingApproval retrieves posts pending approval
func (r *GormPostRepository) GetPendingApproval(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	return r.GetByStatus(ctx, domain.PostStatusPendingApproval, limit, offset)
}

// GetByAuthor retrieves posts by author with pagination
func (r *GormPostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Post, error) {
	if authorID == uuid.Nil {
		return nil, fmt.Errorf("author ID cannot be nil")
	}

	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Where("author_id = ?", authorID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by author: %w", err)
	}

	return posts, nil
}

// CountByAuthor counts posts by author
func (r *GormPostRepository) CountByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error) {
	if authorID == uuid.Nil {
		return 0, fmt.Errorf("author ID cannot be nil")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("author_id = ?", authorID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count posts by author: %w", err)
	}

	return count, nil
}

// GetByType retrieves posts by post type with pagination
func (r *GormPostRepository) GetByType(ctx context.Context, postTypeID uuid.UUID, limit, offset int) ([]*domain.Post, error) {
	if postTypeID == uuid.Nil {
		return nil, fmt.Errorf("post type ID cannot be nil")
	}

	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Where("post_type_id = ?", postTypeID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by type: %w", err)
	}

	return posts, nil
}

// CountByType counts posts by post type
func (r *GormPostRepository) CountByType(ctx context.Context, postTypeID uuid.UUID) (int64, error) {
	if postTypeID == uuid.Nil {
		return 0, fmt.Errorf("post type ID cannot be nil")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", postTypeID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count posts by type: %w", err)
	}

	return count, nil
}

// GetByCategory retrieves posts by category with pagination
func (r *GormPostRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*domain.Post, error) {
	if categoryID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be nil")
	}

	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Joins("JOIN post_category_assignments pca ON posts.id = pca.post_id").
		Where("pca.category_id = ?", categoryID).
		Order("posts.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by category: %w", err)
	}

	return posts, nil
}

// GetByCategorySlug retrieves posts by category slug with pagination
func (r *GormPostRepository) GetByCategorySlug(ctx context.Context, categorySlug string, limit, offset int) ([]*domain.Post, error) {
	if categorySlug == "" {
		return nil, fmt.Errorf("category slug cannot be empty")
	}

	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Joins("JOIN post_category_assignments pca ON posts.id = pca.post_id").
		Joins("JOIN post_categories pc ON pca.category_id = pc.id").
		Where("pc.slug = ?", categorySlug).
		Order("posts.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by category slug: %w", err)
	}

	return posts, nil
}

// GetByTag retrieves posts by tag with pagination
func (r *GormPostRepository) GetByTag(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]*domain.Post, error) {
	if tagID == uuid.Nil {
		return nil, fmt.Errorf("tag ID cannot be nil")
	}

	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Joins("JOIN post_tag_assignments pta ON posts.id = pta.post_id").
		Where("pta.tag_id = ?", tagID).
		Order("posts.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by tag: %w", err)
	}

	return posts, nil
}

// GetByTagSlug retrieves posts by tag slug with pagination
func (r *GormPostRepository) GetByTagSlug(ctx context.Context, tagSlug string, limit, offset int) ([]*domain.Post, error) {
	if tagSlug == "" {
		return nil, fmt.Errorf("tag slug cannot be empty")
	}

	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Joins("JOIN post_tag_assignments pta ON posts.id = pta.post_id").
		Joins("JOIN post_tags pt ON pta.tag_id = pt.id").
		Where("pt.slug = ?", tagSlug).
		Order("posts.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by tag slug: %w", err)
	}

	return posts, nil
}

// Search performs basic text search in posts
func (r *GormPostRepository) Search(ctx context.Context, query string, options domain.SearchPostsOptions) ([]*domain.Post, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	dbQuery := r.db.WithContext(ctx).Model(&domain.Post{})

	// Basic text search (can be enhanced with full-text search later)
	searchCondition := "title ILIKE ? OR content ILIKE ?"
	searchValue := "%" + query + "%"
	dbQuery = dbQuery.Where(searchCondition, searchValue, searchValue)

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

	// Apply sorting
	dbQuery = r.applySorting(dbQuery, options.SortBy, options.SortOrder)

	// Apply pagination
	if options.Limit > 0 {
		dbQuery = dbQuery.Limit(options.Limit)
	}
	if options.Offset > 0 {
		dbQuery = dbQuery.Offset(options.Offset)
	}

	var posts []*domain.Post
	if err := dbQuery.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	return posts, nil
}

// GetByDateRange retrieves posts within a date range
func (r *GormPostRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*domain.Post, error) {
	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts by date range: %w", err)
	}

	return posts, nil
}

// IncrementViewCount increments the view count for a post
func (r *GormPostRepository) IncrementViewCount(ctx context.Context, postID uuid.UUID) error {
	if postID == uuid.Nil {
		return fmt.Errorf("post ID cannot be nil")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		Update("view_count", gorm.Expr("view_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to increment view count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// GetMostViewed retrieves most viewed posts
func (r *GormPostRepository) GetMostViewed(ctx context.Context, limit int, since time.Time) ([]*domain.Post, error) {
	var posts []*domain.Post
	query := r.db.WithContext(ctx).
		Where("created_at >= ?", since).
		Order("view_count DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get most viewed posts: %w", err)
	}

	return posts, nil
}

// Approve approves a post
func (r *GormPostRepository) Approve(ctx context.Context, postID, approverID uuid.UUID) error {
	if postID == uuid.Nil {
		return fmt.Errorf("post ID cannot be nil")
	}
	if approverID == uuid.Nil {
		return fmt.Errorf("approver ID cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&domain.Post{}).
			Where("id = ?", postID).
			Updates(map[string]interface{}{
				"is_approved": true,
				"approved_by": approverID,
				"approved_at": now,
				"updated_at":  now,
			})

		if result.Error != nil {
			return fmt.Errorf("failed to approve post: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("post not found")
		}

		return nil
	})
}

// Reject rejects a post (not implemented in basic version)
func (r *GormPostRepository) Reject(ctx context.Context, postID, approverID uuid.UUID, reason string) error {
	// For now, we'll just set status back to draft
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&domain.Post{}).
			Where("id = ?", postID).
			Updates(map[string]interface{}{
				"status":     domain.PostStatusDraft,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return fmt.Errorf("failed to reject post: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("post not found")
		}

		return nil
	})
}

// BulkUpdateStatus updates status for multiple posts
func (r *GormPostRepository) BulkUpdateStatus(ctx context.Context, postIDs []uuid.UUID, status domain.PostStatus) error {
	if len(postIDs) == 0 {
		return fmt.Errorf("post IDs cannot be empty")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id IN ?", postIDs).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to bulk update status: %w", result.Error)
	}

	return nil
}

// BulkDelete deletes multiple posts (soft delete)
func (r *GormPostRepository) BulkDelete(ctx context.Context, postIDs []uuid.UUID) error {
	if len(postIDs) == 0 {
		return fmt.Errorf("post IDs cannot be empty")
	}

	return r.BulkUpdateStatus(ctx, postIDs, domain.PostStatusArchived)
}

// AddCategory adds a category to a post
func (r *GormPostRepository) AddCategory(ctx context.Context, postID, categoryID uuid.UUID) error {
	if postID == uuid.Nil || categoryID == uuid.Nil {
		return fmt.Errorf("post ID and category ID cannot be nil")
	}

	// Use raw SQL to insert into junction table
	err := r.db.WithContext(ctx).Exec(
		"INSERT INTO post_category_assignments (post_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		postID, categoryID,
	).Error

	if err != nil {
		return fmt.Errorf("failed to add category to post: %w", err)
	}

	return nil
}

// RemoveCategory removes a category from a post
func (r *GormPostRepository) RemoveCategory(ctx context.Context, postID, categoryID uuid.UUID) error {
	if postID == uuid.Nil || categoryID == uuid.Nil {
		return fmt.Errorf("post ID and category ID cannot be nil")
	}

	err := r.db.WithContext(ctx).Exec(
		"DELETE FROM post_category_assignments WHERE post_id = ? AND category_id = ?",
		postID, categoryID,
	).Error

	if err != nil {
		return fmt.Errorf("failed to remove category from post: %w", err)
	}

	return nil
}

// AddTag adds a tag to a post
func (r *GormPostRepository) AddTag(ctx context.Context, postID, tagID uuid.UUID) error {
	if postID == uuid.Nil || tagID == uuid.Nil {
		return fmt.Errorf("post ID and tag ID cannot be nil")
	}

	// Use raw SQL to insert into junction table
	err := r.db.WithContext(ctx).Exec(
		"INSERT INTO post_tag_assignments (post_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		postID, tagID,
	).Error

	if err != nil {
		return fmt.Errorf("failed to add tag to post: %w", err)
	}

	return nil
}

// RemoveTag removes a tag from a post
func (r *GormPostRepository) RemoveTag(ctx context.Context, postID, tagID uuid.UUID) error {
	if postID == uuid.Nil || tagID == uuid.Nil {
		return fmt.Errorf("post ID and tag ID cannot be nil")
	}

	err := r.db.WithContext(ctx).Exec(
		"DELETE FROM post_tag_assignments WHERE post_id = ? AND tag_id = ?",
		postID, tagID,
	).Error

	if err != nil {
		return fmt.Errorf("failed to remove tag from post: %w", err)
	}

	return nil
}

// GetStatsByAuthor gets statistics for an author
func (r *GormPostRepository) GetStatsByAuthor(ctx context.Context, authorID uuid.UUID) (*domain.AuthorStats, error) {
	if authorID == uuid.Nil {
		return nil, fmt.Errorf("author ID cannot be nil")
	}

	stats := &domain.AuthorStats{
		AuthorID: authorID,
	}

	// Get total posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("author_id = ?", authorID).
		Count(&stats.TotalPosts)

	// Get published posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("author_id = ? AND status = ?", authorID, domain.PostStatusPublished).
		Count(&stats.PublishedPosts)

	// Get draft posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("author_id = ? AND status = ?", authorID, domain.PostStatusDraft).
		Count(&stats.DraftPosts)

	// Get pending posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("author_id = ? AND status = ?", authorID, domain.PostStatusPendingApproval).
		Count(&stats.PendingPosts)

	// Get total views
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("author_id = ?", authorID).
		Select("COALESCE(SUM(view_count), 0)").
		Row().
		Scan(&stats.TotalViews)

	// Calculate average views
	if stats.TotalPosts > 0 {
		stats.AverageViews = float64(stats.TotalViews) / float64(stats.TotalPosts)
	}

	// Get first and last post dates
	var firstPost, lastPost domain.Post
	if err := r.db.WithContext(ctx).Where("author_id = ?", authorID).Order("created_at ASC").First(&firstPost).Error; err == nil {
		stats.FirstPostAt = &firstPost.CreatedAt
	}
	if err := r.db.WithContext(ctx).Where("author_id = ?", authorID).Order("created_at DESC").First(&lastPost).Error; err == nil {
		stats.LastPostAt = &lastPost.CreatedAt
	}

	return stats, nil
}

// GetStatsByType gets statistics for a post type
func (r *GormPostRepository) GetStatsByType(ctx context.Context, postTypeID uuid.UUID) (*domain.TypeStats, error) {
	if postTypeID == uuid.Nil {
		return nil, fmt.Errorf("post type ID cannot be nil")
	}

	stats := &domain.TypeStats{
		PostTypeID: postTypeID,
	}

	// Get total posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", postTypeID).
		Count(&stats.TotalPosts)

	// Get published posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", postTypeID, domain.PostStatusPublished).
		Count(&stats.PublishedPosts)

	// Get draft posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", postTypeID, domain.PostStatusDraft).
		Count(&stats.DraftPosts)

	// Get pending posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", postTypeID, domain.PostStatusPendingApproval).
		Count(&stats.PendingPosts)

	// Get total views
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", postTypeID).
		Select("COALESCE(SUM(view_count), 0)").
		Row().
		Scan(&stats.TotalViews)

	// Calculate average views
	if stats.TotalPosts > 0 {
		stats.AverageViews = float64(stats.TotalViews) / float64(stats.TotalPosts)
	}

	// Get unique authors count
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", postTypeID).
		Distinct("author_id").
		Count(&stats.UniqueAuthors)

	// Get first and last post dates
	var firstPost, lastPost domain.Post
	if err := r.db.WithContext(ctx).Where("post_type_id = ?", postTypeID).Order("created_at ASC").First(&firstPost).Error; err == nil {
		stats.FirstPostAt = &firstPost.CreatedAt
	}
	if err := r.db.WithContext(ctx).Where("post_type_id = ?", postTypeID).Order("created_at DESC").First(&lastPost).Error; err == nil {
		stats.LastPostAt = &lastPost.CreatedAt
	}

	return stats, nil
}

// GetActivityStats gets overall activity statistics
func (r *GormPostRepository) GetActivityStats(ctx context.Context, since time.Time) (*domain.ActivityStats, error) {
	stats := &domain.ActivityStats{
		Period:    fmt.Sprintf("since %s", since.Format("2006-01-02")),
		CreatedAt: time.Now(),
	}

	// Get total posts since date
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ?", since).
		Count(&stats.TotalPosts)

	// Get published posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ? AND status = ?", since, domain.PostStatusPublished).
		Count(&stats.PublishedPosts)

	// Get draft posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ? AND status = ?", since, domain.PostStatusDraft).
		Count(&stats.DraftPosts)

	// Get pending posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ? AND status = ?", since, domain.PostStatusPendingApproval).
		Count(&stats.PendingPosts)

	// Get archived posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ? AND status = ?", since, domain.PostStatusArchived).
		Count(&stats.DeletedPosts)

	// Get total views
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ?", since).
		Select("COALESCE(SUM(view_count), 0)").
		Row().
		Scan(&stats.TotalViews)

	// Get unique authors count
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("created_at >= ?", since).
		Distinct("author_id").
		Count(&stats.UniqueAuthors)

	// TODO: Get popular tags and top categories (requires joins)
	stats.PopularTags = []string{}
	stats.TopCategories = []string{}

	return stats, nil
}

// Helper methods

// applyListFilters applies filtering conditions to a query
func (r *GormPostRepository) applyListFilters(query *gorm.DB, options domain.ListPostsOptions) *gorm.DB {
	if options.Status != nil {
		query = query.Where("status = ?", *options.Status)
	}
	if options.AuthorID != nil {
		query = query.Where("author_id = ?", *options.AuthorID)
	}
	if options.PostTypeID != nil {
		query = query.Where("post_type_id = ?", *options.PostTypeID)
	}
	if options.CategoryID != nil {
		query = query.Joins("JOIN post_category_assignments pca ON posts.id = pca.post_id").
			Where("pca.category_id = ?", *options.CategoryID)
	}
	if options.TagID != nil {
		query = query.Joins("JOIN post_tag_assignments pta ON posts.id = pta.post_id").
			Where("pta.tag_id = ?", *options.TagID)
	}

	// Date filters
	if options.PublishedAfter != nil {
		query = query.Where("published_at > ?", *options.PublishedAfter)
	}
	if options.PublishedBefore != nil {
		query = query.Where("published_at < ?", *options.PublishedBefore)
	}
	if options.CreatedAfter != nil {
		query = query.Where("created_at > ?", *options.CreatedAfter)
	}
	if options.CreatedBefore != nil {
		query = query.Where("created_at < ?", *options.CreatedBefore)
	}

	return query
}

// applySorting applies sorting to a query
func (r *GormPostRepository) applySorting(query *gorm.DB, sortBy, sortOrder string) *gorm.DB {
	// Default sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sort fields
	validSortFields := map[string]bool{
		"created_at":   true,
		"updated_at":   true,
		"published_at": true,
		"title":        true,
		"view_count":   true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
}
