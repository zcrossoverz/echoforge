package documentation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuideValidation(t *testing.T) {
	// Get repository root
	repoRoot, err := getRepoRoot()
	require.NoError(t, err, "Should find repository root")

	docsDir := filepath.Join(repoRoot, "docs")

	t.Run("manga site guide exists and is valid", func(t *testing.T) {
		guidePath := filepath.Join(docsDir, "guides", "site-extension", "manga-site-setup.md")

		// This test MUST fail initially (TDD)
		_, err := os.Stat(guidePath)
		assert.NoError(t, err, "Manga site guide should exist")

		if err == nil {
			content, readErr := os.ReadFile(guidePath)
			require.NoError(t, readErr, "Should be able to read manga guide")

			// Validate content requirements
			contentStr := string(content)
			assert.Contains(t, contentStr, "manga", "Guide should mention manga")
			assert.Contains(t, contentStr, "site_id", "Guide should explain multi-tenant isolation")
			assert.Contains(t, contentStr, "configuration", "Guide should include configuration")
			assert.Greater(t, len(contentStr), 1000, "Guide should be comprehensive (>1000 chars)")
		}
	})

	t.Run("blog site guide exists and is valid", func(t *testing.T) {
		guidePath := filepath.Join(docsDir, "guides", "site-extension", "blog-site-setup.md")

		// This test MUST fail initially (TDD)
		_, err := os.Stat(guidePath)
		assert.NoError(t, err, "Blog site guide should exist")

		if err == nil {
			content, readErr := os.ReadFile(guidePath)
			require.NoError(t, readErr, "Should be able to read blog guide")

			// Validate content requirements
			contentStr := string(content)
			assert.Contains(t, contentStr, "blog", "Guide should mention blog")
			assert.Contains(t, contentStr, "site_id", "Guide should explain multi-tenant isolation")
			assert.Contains(t, contentStr, "configuration", "Guide should include configuration")
			assert.Greater(t, len(contentStr), 1000, "Guide should be comprehensive (>1000 chars)")
		}
	})

	t.Run("portfolio site guide exists and is valid", func(t *testing.T) {
		guidePath := filepath.Join(docsDir, "guides", "site-extension", "portfolio-site-setup.md")

		// This test MUST fail initially (TDD)
		_, err := os.Stat(guidePath)
		assert.NoError(t, err, "Portfolio site guide should exist")

		if err == nil {
			content, readErr := os.ReadFile(guidePath)
			require.NoError(t, readErr, "Should be able to read portfolio guide")

			// Validate content requirements
			contentStr := string(content)
			assert.Contains(t, contentStr, "portfolio", "Guide should mention portfolio")
			assert.Contains(t, contentStr, "site_id", "Guide should explain multi-tenant isolation")
			assert.Contains(t, contentStr, "configuration", "Guide should include configuration")
			assert.Greater(t, len(contentStr), 1000, "Guide should be comprehensive (>1000 chars)")
		}
	})

	t.Run("customization patterns documentation exists", func(t *testing.T) {
		customDir := filepath.Join(docsDir, "guides", "customization")

		// This test MUST fail initially (TDD)
		_, err := os.Stat(customDir)
		assert.NoError(t, err, "Customization directory should exist")

		// Check for expected customization files
		expectedFiles := []string{
			"authentication-patterns.md",
			"ui-customization.md",
			"data-model-extensions.md",
		}

		for _, expectedFile := range expectedFiles {
			filePath := filepath.Join(customDir, expectedFile)
			_, err := os.Stat(filePath)
			assert.NoError(t, err, "Customization file %s should exist", expectedFile)
		}
	})
}

func getRepoRoot() (string, error) {
	// Find the repository root by looking for the main go.mod with the correct module name
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Check if this is the main project go.mod by looking for specific content
			content, readErr := os.ReadFile(goModPath)
			if readErr == nil && (strings.Contains(string(content), "github.com/zcrossoverz/echoforge") ||
				strings.Contains(string(content), "cmd/server/main.go") ||
				len(strings.Split(string(content), "\n")) > 10) { // Main go.mod is larger
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
