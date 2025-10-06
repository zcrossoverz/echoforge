package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// GormPostTypeRepository implements domain.PostTypeRepository using GORM
type GormPostTypeRepository struct {
	db *gorm.DB
}

// NewGormPostTypeRepository creates a new GORM-based post type repository
func NewGormPostTypeRepository(db *gorm.DB) domain.PostTypeRepository {
	return &GormPostTypeRepository{
		db: db,
	}
}

// Create creates a new post type in the database
func (r *GormPostTypeRepository) Create(ctx context.Context, postType *domain.PostType) error {
	if postType == nil {
		return fmt.Errorf("post type cannot be nil")
	}

	// Use transaction for consistency
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(postType).Error; err != nil {
			return fmt.Errorf("failed to create post type: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a post type by its ID
func (r *GormPostTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PostType, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}

	var postType domain.PostType
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&postType).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post type by id: %w", err)
	}

	return &postType, nil
}

// GetByName retrieves a post type by its name
func (r *GormPostTypeRepository) GetByName(ctx context.Context, name string) (*domain.PostType, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	var postType domain.PostType
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&postType).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post type by name: %w", err)
	}

	return &postType, nil
}

// GetBySlug retrieves a post type by its slug
func (r *GormPostTypeRepository) GetBySlug(ctx context.Context, slug string) (*domain.PostType, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	var postType domain.PostType
	err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&postType).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post type by slug: %w", err)
	}

	return &postType, nil
}

// Update updates an existing post type
func (r *GormPostTypeRepository) Update(ctx context.Context, postType *domain.PostType) error {
	if postType == nil {
		return fmt.Errorf("post type cannot be nil")
	}
	if postType.ID == uuid.Nil {
		return fmt.Errorf("post type ID cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the post type
		result := tx.Save(postType)
		if result.Error != nil {
			return fmt.Errorf("failed to update post type: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("post type not found")
		}
		return nil
	})
}

// Delete deletes a post type by ID (soft delete by setting status to inactive)
func (r *GormPostTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("id cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Soft delete by setting is_active to false
		result := tx.Model(&domain.PostType{}).
			Where("id = ?", id).
			Update("is_active", false)

		if result.Error != nil {
			return fmt.Errorf("failed to delete post type: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("post type not found")
		}
		return nil
	})
}

// List retrieves all post types with filtering and pagination
func (r *GormPostTypeRepository) List(ctx context.Context, options domain.ListPostTypesOptions) ([]*domain.PostType, error) {
	query := r.db.WithContext(ctx).Model(&domain.PostType{})

	// Apply filters
	if options.IsActive != nil {
		query = query.Where("is_active = ?", *options.IsActive)
	}
	if options.IsSystem != nil {
		query = query.Where("is_system = ?", *options.IsSystem)
	}

	// Apply sorting
	if options.SortBy != "" {
		sortOrder := "ASC"
		if options.SortOrder != "" {
			sortOrder = options.SortOrder
		}

		// Validate sort fields
		validSortFields := map[string]bool{
			"name":       true,
			"slug":       true,
			"created_at": true,
			"updated_at": true,
		}

		if validSortFields[options.SortBy] {
			query = query.Order(fmt.Sprintf("%s %s", options.SortBy, sortOrder))
		} else {
			query = query.Order("name ASC")
		}
	} else {
		query = query.Order("name ASC")
	}

	// Apply pagination
	if options.Limit > 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset > 0 {
		query = query.Offset(options.Offset)
	}

	var postTypes []*domain.PostType
	if err := query.Find(&postTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to list post types: %w", err)
	}

	return postTypes, nil
}

// Count counts post types with filtering
func (r *GormPostTypeRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.PostType{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count post types: %w", err)
	}

	return count, nil
}

// GetSystemTypes retrieves all system post types
func (r *GormPostTypeRepository) GetSystemTypes(ctx context.Context) ([]*domain.PostType, error) {
	var postTypes []*domain.PostType
	err := r.db.WithContext(ctx).
		Where("name IN (?)", []string{"blog", "manga", "news"}).
		Find(&postTypes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get system post types: %w", err)
	}

	return postTypes, nil
}

// GetCustomTypes retrieves all custom (non-system) post types
func (r *GormPostTypeRepository) GetCustomTypes(ctx context.Context) ([]*domain.PostType, error) {
	var postTypes []*domain.PostType
	err := r.db.WithContext(ctx).
		Where("name NOT IN (?)", []string{"blog", "manga", "news"}).
		Find(&postTypes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get custom post types: %w", err)
	}

	return postTypes, nil
}

// GetFieldDefinitions returns the field definitions for a post type
func (r *GormPostTypeRepository) GetFieldDefinitions(ctx context.Context, postTypeID uuid.UUID) (map[string]interface{}, error) {
	if postTypeID == uuid.Nil {
		return nil, fmt.Errorf("post type ID cannot be nil")
	}

	postType, err := r.GetByID(ctx, postTypeID)
	if err != nil {
		return nil, err
	}
	if postType == nil {
		return nil, fmt.Errorf("post type not found")
	}

	return postType.GetFieldDefinitions()
}

// UpdateFieldDefinitions updates the field definitions for a post type
func (r *GormPostTypeRepository) UpdateFieldDefinitions(ctx context.Context, postTypeID uuid.UUID, fieldDefinitions map[string]interface{}) error {
	if postTypeID == uuid.Nil {
		return fmt.Errorf("post type ID cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the post type
		var postType domain.PostType
		if err := tx.Where("id = ?", postTypeID).First(&postType).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("post type not found")
			}
			return fmt.Errorf("failed to get post type: %w", err)
		}

		// Set field definitions
		if err := postType.SetFieldDefinitions(fieldDefinitions); err != nil {
			return fmt.Errorf("failed to set field definitions: %w", err)
		}

		// Save
		if err := tx.Save(&postType).Error; err != nil {
			return fmt.Errorf("failed to update field definitions: %w", err)
		}

		return nil
	})
}

// ExistsByName checks if a post type exists by name
func (r *GormPostTypeRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("name cannot be empty")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.PostType{}).
		Where("name = ?", name).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check post type existence by name: %w", err)
	}

	return count > 0, nil
}

// ExistsBySlug checks if a post type exists by slug
func (r *GormPostTypeRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	if slug == "" {
		return false, fmt.Errorf("slug cannot be empty")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.PostType{}).
		Where("slug = ?", slug).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check post type existence by slug: %w", err)
	}

	return count > 0, nil
}

// GetAllUsageStats gets usage statistics for all post types
func (r *GormPostTypeRepository) GetAllUsageStats(ctx context.Context) ([]*domain.PostTypeUsageStats, error) {
	var postTypes []*domain.PostType
	if err := r.db.WithContext(ctx).Find(&postTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to get post types: %w", err)
	}

	var allStats []*domain.PostTypeUsageStats
	for _, postType := range postTypes {
		stats, err := r.GetUsageStats(ctx, postType.ID)
		if err != nil {
			return nil, err
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// ValidatePostData validates post data against a post type's field definitions
func (r *GormPostTypeRepository) ValidatePostData(ctx context.Context, postTypeID uuid.UUID, postData map[string]interface{}) error {
	if postTypeID == uuid.Nil {
		return fmt.Errorf("post type ID cannot be nil")
	}

	postType, err := r.GetByID(ctx, postTypeID)
	if err != nil {
		return err
	}
	if postType == nil {
		return fmt.Errorf("post type not found")
	}

	return postType.ValidatePostMetadata(postData)
}

// GetMetadata gets metadata for a post type (currently same as field definitions)
func (r *GormPostTypeRepository) GetMetadata(ctx context.Context, postTypeID uuid.UUID) (map[string]interface{}, error) {
	return r.GetFieldDefinitions(ctx, postTypeID)
}

// UpdateMetadata updates metadata for a post type (currently same as field definitions)
func (r *GormPostTypeRepository) UpdateMetadata(ctx context.Context, postTypeID uuid.UUID, metadata map[string]interface{}) error {
	return r.UpdateFieldDefinitions(ctx, postTypeID, metadata)
}

// BulkUpdateStatus updates status for multiple post types
func (r *GormPostTypeRepository) BulkUpdateStatus(ctx context.Context, postTypeIDs []uuid.UUID, isActive bool) error {
	if len(postTypeIDs) == 0 {
		return fmt.Errorf("post type IDs cannot be empty")
	}

	result := r.db.WithContext(ctx).
		Model(&domain.PostType{}).
		Where("id IN ?", postTypeIDs).
		Update("is_active", isActive)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk update status: %w", result.Error)
	}

	return nil
}

// GetByExtensionType gets post types by extension type (using name pattern)
func (r *GormPostTypeRepository) GetByExtensionType(ctx context.Context, extensionType string) ([]*domain.PostType, error) {
	if extensionType == "" {
		return nil, fmt.Errorf("extension type cannot be empty")
	}

	var postTypes []*domain.PostType
	err := r.db.WithContext(ctx).
		Where("name LIKE ?", extensionType+"%").
		Find(&postTypes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get post types by extension type: %w", err)
	}

	return postTypes, nil
}

// GetCompatibleTypes gets post types that have all required fields
func (r *GormPostTypeRepository) GetCompatibleTypes(ctx context.Context, requiredFields []string) ([]*domain.PostType, error) {
	if len(requiredFields) == 0 {
		return r.List(ctx, domain.ListPostTypesOptions{})
	}

	var postTypes []*domain.PostType
	if err := r.db.WithContext(ctx).Find(&postTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to get post types: %w", err)
	}

	var compatibleTypes []*domain.PostType
	for _, postType := range postTypes {
		fieldDefs, err := postType.GetFieldDefinitions()
		if err != nil {
			continue // Skip if can't parse field definitions
		}

		isCompatible := true
		for _, required := range requiredFields {
			if _, exists := fieldDefs[required]; !exists {
				isCompatible = false
				break
			}
		}

		if isCompatible {
			compatibleTypes = append(compatibleTypes, postType)
		}
	}

	return compatibleTypes, nil
}

// BulkDelete deletes multiple post types (soft delete by deactivating)
func (r *GormPostTypeRepository) BulkDelete(ctx context.Context, postTypeIDs []uuid.UUID) error {
	if len(postTypeIDs) == 0 {
		return fmt.Errorf("post type IDs cannot be empty")
	}

	// Don't allow deletion of system types
	var systemTypes []*domain.PostType
	err := r.db.WithContext(ctx).
		Where("id IN ? AND name IN ?", postTypeIDs, []string{"blog", "manga", "news"}).
		Find(&systemTypes).Error

	if err != nil {
		return fmt.Errorf("failed to check for system types: %w", err)
	}

	if len(systemTypes) > 0 {
		return fmt.Errorf("cannot delete system post types")
	}

	// Soft delete by setting is_active to false
	result := r.db.WithContext(ctx).
		Model(&domain.PostType{}).
		Where("id IN ?", postTypeIDs).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete post types: %w", result.Error)
	}

	return nil
}

// GetUsageStats gets detailed usage statistics for a post type
func (r *GormPostTypeRepository) GetUsageStats(ctx context.Context, id uuid.UUID) (*domain.PostTypeUsageStats, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("id cannot be nil")
	}

	// Get post type info
	postType, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if postType == nil {
		return nil, fmt.Errorf("post type not found")
	}

	stats := &domain.PostTypeUsageStats{
		PostTypeID:   id,
		PostTypeName: postType.Name,
		PostTypeSlug: postType.Name, // Using name as slug for now
		CreatedAt:    postType.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    postType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Get total posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", id).
		Count(&stats.TotalPosts)

	// Get published posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", id, domain.PostStatusPublished).
		Count(&stats.PublishedPosts)

	// Get draft posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", id, domain.PostStatusDraft).
		Count(&stats.DraftPosts)

	// Get pending posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", id, domain.PostStatusPendingApproval).
		Count(&stats.PendingPosts)

	// Get archived posts
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ? AND status = ?", id, domain.PostStatusArchived).
		Count(&stats.ArchivedPosts)

	// Get unique authors count
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", id).
		Distinct("author_id").
		Count(&stats.UniqueAuthors)

	// Get total views
	r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("post_type_id = ?", id).
		Select("COALESCE(SUM(view_count), 0)").
		Row().
		Scan(&stats.TotalViews)

	// Calculate average views
	if stats.TotalPosts > 0 {
		stats.AverageViews = float64(stats.TotalViews) / float64(stats.TotalPosts)
	}

	return stats, nil
}

// Search performs text search on post type names and descriptions
func (r *GormPostTypeRepository) Search(ctx context.Context, query string, options domain.SearchPostTypesOptions) ([]*domain.PostType, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	dbQuery := r.db.WithContext(ctx).Model(&domain.PostType{})

	// Basic text search
	searchCondition := "name ILIKE ? OR description ILIKE ?"
	searchValue := "%" + query + "%"
	dbQuery = dbQuery.Where(searchCondition, searchValue, searchValue)

	// Apply filters
	if options.IsActive != nil {
		dbQuery = dbQuery.Where("is_active = ?", *options.IsActive)
	}
	if options.IsSystem != nil {
		systemTypes := []string{"blog", "manga", "news"}
		if *options.IsSystem {
			dbQuery = dbQuery.Where("name IN ?", systemTypes)
		} else {
			dbQuery = dbQuery.Where("name NOT IN ?", systemTypes)
		}
	}

	// Apply sorting
	if options.SortBy != "" {
		sortOrder := "ASC"
		if options.SortOrder != "" {
			sortOrder = options.SortOrder
		}

		// Validate sort fields
		validSortFields := map[string]bool{
			"name":       true,
			"created_at": true,
			"updated_at": true,
		}

		if validSortFields[options.SortBy] {
			dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", options.SortBy, sortOrder))
		} else {
			dbQuery = dbQuery.Order("name ASC")
		}
	} else {
		dbQuery = dbQuery.Order("name ASC")
	}

	// Apply pagination
	if options.Limit > 0 {
		dbQuery = dbQuery.Limit(options.Limit)
	}
	if options.Offset > 0 {
		dbQuery = dbQuery.Offset(options.Offset)
	}

	var postTypes []*domain.PostType
	if err := dbQuery.Find(&postTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to search post types: %w", err)
	}

	return postTypes, nil
}

// ValidateFieldDefinitions validates field definitions format
func (r *GormPostTypeRepository) ValidateFieldDefinitions(ctx context.Context, fieldDefs map[string]interface{}) error {
	// This could be enhanced with JSON schema validation
	// For now, just check if it's valid JSON-like structure
	if fieldDefs == nil {
		return nil // Empty definitions are allowed
	}

	// Basic validation - ensure all field definitions have required properties
	for fieldName, fieldDef := range fieldDefs {
		if fieldName == "" {
			return fmt.Errorf("field name cannot be empty")
		}

		// Field definition should be a map
		defMap, ok := fieldDef.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field definition for '%s' must be an object", fieldName)
		}

		// Check for required properties
		if _, hasType := defMap["type"]; !hasType {
			return fmt.Errorf("field definition for '%s' must have a 'type' property", fieldName)
		}

		// Validate field type
		fieldType, ok := defMap["type"].(string)
		if !ok {
			return fmt.Errorf("field type for '%s' must be a string", fieldName)
		}

		validTypes := map[string]bool{
			"string":      true,
			"text":        true,
			"number":      true,
			"boolean":     true,
			"date":        true,
			"datetime":    true,
			"url":         true,
			"email":       true,
			"select":      true,
			"multiselect": true,
			"file":        true,
			"image":       true,
		}

		if !validTypes[fieldType] {
			return fmt.Errorf("invalid field type '%s' for field '%s'", fieldType, fieldName)
		}

		// Additional validation for select fields
		if fieldType == "select" || fieldType == "multiselect" {
			if _, hasOptions := defMap["options"]; !hasOptions {
				return fmt.Errorf("select field '%s' must have 'options' property", fieldName)
			}
		}
	}

	return nil
}
