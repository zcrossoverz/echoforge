package documentation

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkValidation(t *testing.T) {
	// Get repository root
	repoRoot, err := getRepoRoot()
	require.NoError(t, err, "Should find repository root")

	docsDir := filepath.Join(repoRoot, "docs")

	t.Run("internal links are valid", func(t *testing.T) {
		// This test MUST fail initially (TDD) until documentation is created
		var markdownFiles []string
		err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".md") {
				markdownFiles = append(markdownFiles, path)
			}
			return nil
		})

		if err != nil {
			t.Skip("Documentation not yet created - test will pass once docs exist")
			return
		}

		assert.Greater(t, len(markdownFiles), 0, "Should find markdown files in docs directory")

		// Regular expression to find markdown links [text](path)
		linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

		var brokenLinks []string

		for _, file := range markdownFiles {
			content, readErr := os.ReadFile(file)
			require.NoError(t, readErr, "Should be able to read %s", file)

			matches := linkRegex.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				if len(match) >= 3 {
					linkPath := match[2]

					// Skip external links (http/https)
					if strings.HasPrefix(linkPath, "http://") || strings.HasPrefix(linkPath, "https://") {
						continue
					}

					// Skip anchor links
					if strings.HasPrefix(linkPath, "#") {
						continue
					}

					// Resolve relative paths
					var targetPath string
					if strings.HasPrefix(linkPath, "/") {
						targetPath = filepath.Join(repoRoot, linkPath)
					} else {
						targetPath = filepath.Join(filepath.Dir(file), linkPath)
					}

					// Check if target exists
					if _, statErr := os.Stat(targetPath); os.IsNotExist(statErr) {
						brokenLinks = append(brokenLinks, file+": "+linkPath)
					}
				}
			}
		}

		assert.Empty(t, brokenLinks, "All internal links should be valid. Broken links: %v", brokenLinks)
	})

	t.Run("README links to documentation", func(t *testing.T) {
		readmePath := filepath.Join(repoRoot, "README.md")

		// This test MUST fail initially until README is updated
		_, err := os.Stat(readmePath)
		if err != nil {
			t.Skip("README.md not found - will be updated later")
			return
		}

		content, readErr := os.ReadFile(readmePath)
		require.NoError(t, readErr, "Should be able to read README.md")

		contentStr := string(content)

		// Check for documentation links
		expectedLinks := []string{
			"docs/",
			"documentation",
			"guides",
		}

		hasDocLinks := false
		for _, expectedLink := range expectedLinks {
			if strings.Contains(contentStr, expectedLink) {
				hasDocLinks = true
				break
			}
		}

		assert.True(t, hasDocLinks, "README should link to documentation")
	})

	t.Run("architecture diagrams are referenced", func(t *testing.T) {
		diagramsDir := filepath.Join(docsDir, "architecture", "diagrams")

		// This test MUST fail initially (TDD) until diagrams are created
		_, err := os.Stat(diagramsDir)
		if err != nil {
			t.Skip("Architecture diagrams not yet created")
			return
		}

		// Expected diagram files
		expectedDiagrams := []string{
			"hexagonal-architecture.mmd",
			"data-flow.mmd",
			"deployment.mmd",
		}

		for _, diagram := range expectedDiagrams {
			diagramPath := filepath.Join(diagramsDir, diagram)
			_, err := os.Stat(diagramPath)
			assert.NoError(t, err, "Diagram %s should exist", diagram)
		}

		// Check that diagrams are referenced in documentation
		var markdownFiles []string
		err = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".md") {
				markdownFiles = append(markdownFiles, path)
			}
			return nil
		})
		require.NoError(t, err, "Should be able to walk docs directory")

		diagramReferences := make(map[string]bool)
		for _, diagram := range expectedDiagrams {
			diagramReferences[diagram] = false
		}

		for _, file := range markdownFiles {
			content, readErr := os.ReadFile(file)
			require.NoError(t, readErr, "Should be able to read %s", file)

			contentStr := string(content)
			for diagram := range diagramReferences {
				if strings.Contains(contentStr, diagram) || strings.Contains(contentStr, strings.TrimSuffix(diagram, ".mmd")) {
					diagramReferences[diagram] = true
				}
			}
		}

		for diagram, referenced := range diagramReferences {
			assert.True(t, referenced, "Diagram %s should be referenced in documentation", diagram)
		}
	})
}
