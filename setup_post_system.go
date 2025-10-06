package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zcrossoverz/echoforge/internal/domain"
)

// seedDefaultData seeds the database with default post system data
func seedDefaultData(db *gorm.DB) error {
	fmt.Println("Seeding default post system data...")

	// Create default post types with field definitions
	defaultPostTypes := []struct {
		PostType         domain.PostType
		FieldDefinitions map[string]interface{}
	}{
		{
			PostType: domain.PostType{
				ID:                uuid.New(),
				Name:              "blog_post",
				DisplayName:       "Blog Post",
				Description:       "Standard blog post with rich content support",
				IsActive:          true,
				RequiresApproval:  false,
				AllowsScheduling:  true,
				AllowsAttachments: true,
			},
			FieldDefinitions: map[string]interface{}{
				"title":    map[string]interface{}{"type": "string", "required": true, "max_length": 255},
				"content":  map[string]interface{}{"type": "text", "required": true},
				"excerpt":  map[string]interface{}{"type": "string", "max_length": 500},
				"featured": map[string]interface{}{"type": "boolean", "default": false},
			},
		},
		{
			PostType: domain.PostType{
				ID:                uuid.New(),
				Name:              "news_article",
				DisplayName:       "News Article",
				Description:       "News article with publication date and source",
				IsActive:          true,
				RequiresApproval:  true,
				AllowsScheduling:  true,
				AllowsAttachments: true,
			},
			FieldDefinitions: map[string]interface{}{
				"title":    map[string]interface{}{"type": "string", "required": true, "max_length": 255},
				"content":  map[string]interface{}{"type": "text", "required": true},
				"source":   map[string]interface{}{"type": "string", "max_length": 100},
				"urgent":   map[string]interface{}{"type": "boolean", "default": false},
				"location": map[string]interface{}{"type": "string", "max_length": 100},
			},
		},
		{
			PostType: domain.PostType{
				ID:                uuid.New(),
				Name:              "page",
				DisplayName:       "Static Page",
				Description:       "Static page content",
				IsActive:          true,
				RequiresApproval:  false,
				AllowsScheduling:  false,
				AllowsAttachments: true,
			},
			FieldDefinitions: map[string]interface{}{
				"title":     map[string]interface{}{"type": "string", "required": true, "max_length": 255},
				"content":   map[string]interface{}{"type": "text", "required": true},
				"show_nav":  map[string]interface{}{"type": "boolean", "default": true},
				"nav_order": map[string]interface{}{"type": "integer", "default": 0},
			},
		},
	}

	for _, postTypeData := range defaultPostTypes {
		var existingPostType domain.PostType
		if err := db.Where("name = ?", postTypeData.PostType.Name).First(&existingPostType).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Set field definitions using the domain method
				if err := postTypeData.PostType.SetFieldDefinitions(postTypeData.FieldDefinitions); err != nil {
					return fmt.Errorf("failed to set field definitions for %s: %v", postTypeData.PostType.Name, err)
				}

				if err := db.Create(&postTypeData.PostType).Error; err != nil {
					return fmt.Errorf("failed to create post type %s: %v", postTypeData.PostType.Name, err)
				}
				fmt.Printf("✓ Created post type: %s\n", postTypeData.PostType.DisplayName)
			} else {
				return fmt.Errorf("failed to check post type %s: %v", postTypeData.PostType.Name, err)
			}
		} else {
			fmt.Printf("→ Post type already exists: %s\n", postTypeData.PostType.DisplayName)
		}
	}

	// Seed default categories
	defaultCategories := []domain.PostCategory{
		{
			ID:          uuid.New(),
			Name:        "Technology",
			Slug:        "technology",
			Description: "Technology and software development posts",
			SortOrder:   1,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "Business",
			Slug:        "business",
			Description: "Business and entrepreneurship content",
			SortOrder:   2,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "Lifestyle",
			Slug:        "lifestyle",
			Description: "Lifestyle and personal development",
			SortOrder:   3,
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			Name:        "News",
			Slug:        "news",
			Description: "News and current events",
			SortOrder:   4,
			IsActive:    true,
		},
	}

	for _, category := range defaultCategories {
		var existingCategory domain.PostCategory
		if err := db.Where("slug = ?", category.Slug).First(&existingCategory).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					return fmt.Errorf("failed to create category %s: %v", category.Name, err)
				}
				fmt.Printf("✓ Created category: %s\n", category.Name)
			} else {
				return fmt.Errorf("failed to check category %s: %v", category.Name, err)
			}
		} else {
			fmt.Printf("→ Category already exists: %s\n", category.Name)
		}
	}

	// Seed default tags
	defaultTags := []domain.PostTag{
		{
			ID:          uuid.New(),
			Name:        "golang",
			Slug:        "golang",
			Color:       "#00ADD8",
			Description: "Go programming language",
			Metadata:    "{}",
		},
		{
			ID:          uuid.New(),
			Name:        "web-development",
			Slug:        "web-development",
			Color:       "#007ACC",
			Description: "Web development and frameworks",
			Metadata:    "{}",
		},
		{
			ID:          uuid.New(),
			Name:        "tutorial",
			Slug:        "tutorial",
			Color:       "#28A745",
			Description: "Educational tutorials and guides",
			Metadata:    "{}",
		},
		{
			ID:          uuid.New(),
			Name:        "api",
			Slug:        "api",
			Color:       "#FFC107",
			Description: "API development and design",
			Metadata:    "{}",
		},
		{
			ID:          uuid.New(),
			Name:        "database",
			Slug:        "database",
			Color:       "#DC3545",
			Description: "Database design and optimization",
			Metadata:    "{}",
		},
	}

	for _, tag := range defaultTags {
		var existingTag domain.PostTag
		if err := db.Where("slug = ?", tag.Slug).First(&existingTag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tag).Error; err != nil {
					return fmt.Errorf("failed to create tag %s: %v", tag.Name, err)
				}
				fmt.Printf("✓ Created tag: %s\n", tag.Name)
			} else {
				return fmt.Errorf("failed to check tag %s: %v", tag.Name, err)
			}
		} else {
			fmt.Printf("→ Tag already exists: %s\n", tag.Name)
		}
	}

	fmt.Println("✓ Default post system data seeding completed!")
	return nil
}

func main() {
	// Database connection string
	// In production, this should come from configuration
	dsn := "host=localhost user=postgres password=admin dbname=bloggo port=5432 sslmode=disable TimeZone=UTC"

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to database successfully!")

	// Create custom ENUM types first (if they don't exist)
	fmt.Println("Creating custom ENUM types...")

	// Check if post_status enum exists
	var count int64
	db.Raw("SELECT COUNT(*) FROM pg_type WHERE typname = 'post_status'").Scan(&count)
	if count == 0 {
		if err := db.Exec("CREATE TYPE post_status AS ENUM ('draft', 'scheduled', 'published', 'archived', 'pending_approval')").Error; err != nil {
			log.Printf("Failed to create post_status ENUM: %v", err)
		} else {
			fmt.Println("✓ Created post_status ENUM")
		}
	} else {
		fmt.Println("→ post_status ENUM already exists")
	}

	// Check if metadata_type enum exists
	db.Raw("SELECT COUNT(*) FROM pg_type WHERE typname = 'metadata_type'").Scan(&count)
	if count == 0 {
		if err := db.Exec("CREATE TYPE metadata_type AS ENUM ('string', 'integer', 'boolean', 'json', 'date')").Error; err != nil {
			log.Printf("Failed to create metadata_type ENUM: %v", err)
		} else {
			fmt.Println("✓ Created metadata_type ENUM")
		}
	} else {
		fmt.Println("→ metadata_type ENUM already exists")
	}

	fmt.Println("✓ ENUM types ready!")

	// Auto-migrate all post-related tables in correct order
	fmt.Println("Running auto-migration for post system...")

	// First migrate the basic tables (one at a time to avoid relationship conflicts)
	err = db.AutoMigrate(&domain.PostType{})
	if err != nil {
		log.Fatal("Failed to migrate PostType:", err)
	}
	fmt.Println("✓ PostType table migrated!")

	err = db.AutoMigrate(&domain.PostCategory{})
	if err != nil {
		log.Fatal("Failed to migrate PostCategory:", err)
	}
	fmt.Println("✓ PostCategory table migrated!")

	err = db.AutoMigrate(&domain.PostTag{})
	if err != nil {
		log.Fatal("Failed to migrate PostTag:", err)
	}
	fmt.Println("✓ PostTag table migrated!")

	// Then migrate Post table (this may create some junction tables)
	err = db.AutoMigrate(&domain.Post{})
	if err != nil {
		log.Fatal("Failed to migrate Post:", err)
	}
	fmt.Println("✓ Post table migrated!")

	// Finally migrate PostAttachment
	err = db.AutoMigrate(&domain.PostAttachment{})
	if err != nil {
		log.Fatal("Failed to migrate PostAttachment:", err)
	}
	fmt.Println("✓ PostAttachment table migrated!")

	fmt.Println("✓ Auto-migration completed!")

	// Seed default data
	if err := seedDefaultData(db); err != nil {
		log.Fatal("Failed to seed default data:", err)
	}

	fmt.Println("\n🎉 Post system setup completed successfully!")
	fmt.Println("\nDefault data created:")
	fmt.Println("- 3 post types: Blog Post, News Article, Static Page")
	fmt.Println("- 4 categories: Technology, Business, Lifestyle, News")
	fmt.Println("- 5 tags: golang, web-development, tutorial, api, database")
	fmt.Println("\nYou can now start creating posts!")
}
