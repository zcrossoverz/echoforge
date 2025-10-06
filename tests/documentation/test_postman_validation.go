package documentation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PostmanCollection represents the basic structure of a Postman collection
type PostmanCollection struct {
	Info struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	} `json:"info"`
	Item []PostmanItem `json:"item"`
}

type PostmanItem struct {
	Name    string          `json:"name"`
	Item    []PostmanItem   `json:"item,omitempty"`
	Request *PostmanRequest `json:"request,omitempty"`
}

type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header"`
	URL    PostmanURL      `json:"url"`
	Body   *PostmanBody    `json:"body,omitempty"`
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanURL struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

func TestPostmanValidation(t *testing.T) {
	// Get repository root
	repoRoot, err := getRepoRoot()
	require.NoError(t, err, "Should find repository root")

	docsDir := filepath.Join(repoRoot, "docs")
	postmanDir := filepath.Join(docsDir, "postman")

	t.Run("postman collection exists and is valid json", func(t *testing.T) {
		// This test MUST fail initially (TDD) until Postman collection is created
		collectionPath := filepath.Join(postmanDir, "echoforge-api.json")

		_, err := os.Stat(collectionPath)
		if err != nil {
			assert.Fail(t, "Postman collection should exist at %s", collectionPath)
			return
		}

		content, readErr := os.ReadFile(collectionPath)
		require.NoError(t, readErr, "Should be able to read Postman collection")

		var collection PostmanCollection
		jsonErr := json.Unmarshal(content, &collection)
		assert.NoError(t, jsonErr, "Collection should be valid JSON")

		if jsonErr == nil {
			// Validate collection structure
			assert.NotEmpty(t, collection.Info.Name, "Collection should have a name")
			assert.NotEmpty(t, collection.Info.Version, "Collection should have a version")
			assert.NotEmpty(t, collection.Info.Description, "Collection should have a description")
			assert.Greater(t, len(collection.Item), 0, "Collection should have items")
		}
	})

	t.Run("collection includes authentication endpoints", func(t *testing.T) {
		// This test MUST fail initially (TDD) until collection includes auth
		collectionPath := filepath.Join(postmanDir, "echoforge-api.json")

		_, err := os.Stat(collectionPath)
		if err != nil {
			t.Skip("Postman collection not yet created")
			return
		}

		content, readErr := os.ReadFile(collectionPath)
		require.NoError(t, readErr, "Should be able to read Postman collection")

		var collection PostmanCollection
		jsonErr := json.Unmarshal(content, &collection)
		require.NoError(t, jsonErr, "Collection should be valid JSON")

		// Look for authentication endpoints
		authEndpoints := []string{"register", "login", "auth"}
		hasAuthEndpoints := false

		for _, item := range collection.Item {
			if containsAuthEndpoint(item, authEndpoints) {
				hasAuthEndpoints = true
				break
			}
		}

		assert.True(t, hasAuthEndpoints, "Collection should include authentication endpoints")
	})

	t.Run("collection includes all major API endpoints", func(t *testing.T) {
		// This test MUST fail initially (TDD) until collection is comprehensive
		collectionPath := filepath.Join(postmanDir, "echoforge-api.json")

		_, err := os.Stat(collectionPath)
		if err != nil {
			t.Skip("Postman collection not yet created")
			return
		}

		content, readErr := os.ReadFile(collectionPath)
		require.NoError(t, readErr, "Should be able to read Postman collection")

		var collection PostmanCollection
		jsonErr := json.Unmarshal(content, &collection)
		require.NoError(t, jsonErr, "Collection should be valid JSON")

		// Expected endpoint categories
		expectedCategories := []string{
			"auth", "user", "post", "health",
		}

		foundCategories := make(map[string]bool)
		for _, category := range expectedCategories {
			foundCategories[category] = false
		}

		for _, item := range collection.Item {
			checkItemForCategories(item, foundCategories)
		}

		for category, found := range foundCategories {
			assert.True(t, found, "Collection should include %s endpoints", category)
		}
	})

	t.Run("environment configurations exist", func(t *testing.T) {
		// This test MUST fail initially (TDD) until environments are created
		envDir := filepath.Join(postmanDir, "environments")

		_, err := os.Stat(envDir)
		if err != nil {
			assert.Fail(t, "Environment directory should exist at %s", envDir)
			return
		}

		expectedEnvs := []string{
			"dev.json",
			"staging.json",
			"prod.json",
		}

		for _, env := range expectedEnvs {
			envPath := filepath.Join(envDir, env)
			_, err := os.Stat(envPath)
			assert.NoError(t, err, "Environment file %s should exist", env)

			if err == nil {
				content, readErr := os.ReadFile(envPath)
				require.NoError(t, readErr, "Should be able to read %s", env)

				var envConfig map[string]interface{}
				jsonErr := json.Unmarshal(content, &envConfig)
				assert.NoError(t, jsonErr, "Environment %s should be valid JSON", env)

				if jsonErr == nil {
					// Check for required environment variables
					assert.Contains(t, envConfig, "name", "Environment should have name")
					assert.Contains(t, envConfig, "values", "Environment should have values")
				}
			}
		}
	})

	t.Run("collection uses environment variables", func(t *testing.T) {
		// This test MUST fail initially (TDD) until collection uses variables
		collectionPath := filepath.Join(postmanDir, "echoforge-api.json")

		_, err := os.Stat(collectionPath)
		if err != nil {
			t.Skip("Postman collection not yet created")
			return
		}

		content, readErr := os.ReadFile(collectionPath)
		require.NoError(t, readErr, "Should be able to read Postman collection")

		var collection PostmanCollection
		jsonErr := json.Unmarshal(content, &collection)
		require.NoError(t, jsonErr, "Collection should be valid JSON")

		// Check that requests use environment variables
		usesVariables := false
		for _, item := range collection.Item {
			if checkItemUsesVariables(item) {
				usesVariables = true
				break
			}
		}

		assert.True(t, usesVariables, "Collection should use environment variables like {{base_url}}")
	})
}

func containsAuthEndpoint(item PostmanItem, authEndpoints []string) bool {
	// Check item name
	for _, auth := range authEndpoints {
		if contains(item.Name, auth) {
			return true
		}
	}

	// Check sub-items
	for _, subItem := range item.Item {
		if containsAuthEndpoint(subItem, authEndpoints) {
			return true
		}
	}

	// Check request URL if present
	if item.Request != nil {
		for _, auth := range authEndpoints {
			if contains(item.Request.URL.Raw, auth) {
				return true
			}
		}
	}

	return false
}

func checkItemForCategories(item PostmanItem, categories map[string]bool) {
	// Check item name
	for category := range categories {
		if contains(item.Name, category) {
			categories[category] = true
		}
	}

	// Check sub-items
	for _, subItem := range item.Item {
		checkItemForCategories(subItem, categories)
	}

	// Check request URL if present
	if item.Request != nil {
		for category := range categories {
			if contains(item.Request.URL.Raw, category) {
				categories[category] = true
			}
		}
	}
}

func checkItemUsesVariables(item PostmanItem) bool {
	// Check request URL for variables
	if item.Request != nil {
		if contains(item.Request.URL.Raw, "{{") && contains(item.Request.URL.Raw, "}}") {
			return true
		}
	}

	// Check sub-items
	for _, subItem := range item.Item {
		if checkItemUsesVariables(subItem) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
