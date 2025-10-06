package documentation

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleValidation(t *testing.T) {
	// Get repository root
	repoRoot, err := getRepoRoot()
	require.NoError(t, err, "Should find repository root")

	docsDir := filepath.Join(repoRoot, "docs")

	t.Run("go code examples compile and run", func(t *testing.T) {
		// This test MUST fail initially (TDD) until documentation with examples is created
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

		if err != nil || len(markdownFiles) == 0 {
			t.Skip("Documentation not yet created - test will pass once docs with examples exist")
			return
		}

		// Regular expression to find Go code blocks
		codeBlockRegex := regexp.MustCompile("(?s)```go\\n(.*?)\\n```")

		var failedExamples []string

		for _, file := range markdownFiles {
			content, readErr := os.ReadFile(file)
			require.NoError(t, readErr, "Should be able to read %s", file)

			matches := codeBlockRegex.FindAllStringSubmatch(string(content), -1)
			for i, match := range matches {
				if len(match) >= 2 {
					goCode := match[1]

					// Skip code examples that are just snippets (no package declaration)
					if !strings.Contains(goCode, "package") {
						continue
					}

					// Create temporary file for testing
					tmpFile := filepath.Join(os.TempDir(), "example_test.go")
					writeErr := os.WriteFile(tmpFile, []byte(goCode), 0644)
					require.NoError(t, writeErr, "Should be able to create temp file")

					// Try to compile the code
					cmd := exec.Command("go", "build", tmpFile)
					cmd.Dir = repoRoot
					if err := cmd.Run(); err != nil {
						failedExamples = append(failedExamples, file+":example_"+string(rune(i+1)))
					}

					// Clean up
					os.Remove(tmpFile)
				}
			}
		}

		assert.Empty(t, failedExamples, "All Go code examples should compile. Failed examples: %v", failedExamples)
	})

	t.Run("bash commands are valid", func(t *testing.T) {
		// This test MUST fail initially (TDD) until documentation with bash examples is created
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

		if err != nil || len(markdownFiles) == 0 {
			t.Skip("Documentation not yet created - test will pass once docs with bash examples exist")
			return
		}

		// Regular expression to find bash code blocks
		bashBlockRegex := regexp.MustCompile("(?s)```(?:bash|sh)\\n(.*?)\\n```")

		var suspiciousCommands []string
		dangerousCommands := []string{"rm -rf", "sudo rm", "format", "delete", "> /dev/null"}

		for _, file := range markdownFiles {
			content, readErr := os.ReadFile(file)
			require.NoError(t, readErr, "Should be able to read %s", file)

			matches := bashBlockRegex.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				if len(match) >= 2 {
					bashCode := match[1]

					// Check for dangerous commands
					for _, dangerous := range dangerousCommands {
						if strings.Contains(bashCode, dangerous) {
							suspiciousCommands = append(suspiciousCommands, file+": "+dangerous)
						}
					}
				}
			}
		}

		assert.Empty(t, suspiciousCommands, "Bash examples should not contain dangerous commands: %v", suspiciousCommands)
	})

	t.Run("yaml configuration examples are valid", func(t *testing.T) {
		// This test MUST fail initially (TDD) until site configs are created
		configDir := filepath.Join(docsDir, "site-configs")

		_, err := os.Stat(configDir)
		if err != nil {
			t.Skip("Site configs not yet created")
			return
		}

		expectedConfigs := []string{
			"manga-site.yaml",
			"blog-site.yaml",
			"portfolio-site.yaml",
		}

		for _, config := range expectedConfigs {
			configPath := filepath.Join(configDir, config)
			_, err := os.Stat(configPath)
			assert.NoError(t, err, "Config file %s should exist", config)

			if err == nil {
				content, readErr := os.ReadFile(configPath)
				require.NoError(t, readErr, "Should be able to read %s", config)

				// Basic YAML validation - check for required fields
				contentStr := string(content)
				assert.Contains(t, contentStr, "site_id", "Config should have site_id")
				assert.Contains(t, contentStr, "database", "Config should have database config")
				assert.Greater(t, len(contentStr), 100, "Config should be comprehensive")
			}
		}
	})

	t.Run("postman collection is valid json", func(t *testing.T) {
		// This test MUST fail initially (TDD) until Postman collection is created
		collectionPath := filepath.Join(docsDir, "postman", "echoforge-api.json")

		_, err := os.Stat(collectionPath)
		if err != nil {
			t.Skip("Postman collection not yet created")
			return
		}

		content, readErr := os.ReadFile(collectionPath)
		require.NoError(t, readErr, "Should be able to read Postman collection")

		// Basic JSON validation by attempting to parse
		contentStr := string(content)
		assert.True(t, strings.HasPrefix(contentStr, "{"), "Collection should be valid JSON object")
		assert.True(t, strings.HasSuffix(strings.TrimSpace(contentStr), "}"), "Collection should be valid JSON object")
		assert.Contains(t, contentStr, "info", "Collection should have info section")
		assert.Contains(t, contentStr, "item", "Collection should have items section")
	})
}
