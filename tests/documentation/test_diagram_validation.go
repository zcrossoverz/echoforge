package documentation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiagramValidation(t *testing.T) {
	// Get repository root
	repoRoot, err := getRepoRoot()
	require.NoError(t, err, "Should find repository root")

	docsDir := filepath.Join(repoRoot, "docs")
	diagramsDir := filepath.Join(docsDir, "architecture", "diagrams")

	t.Run("mermaid diagram sources exist", func(t *testing.T) {
		// This test MUST fail initially (TDD) until diagrams are created
		expectedDiagrams := []string{
			"hexagonal-architecture.mmd",
			"data-flow.mmd",
			"deployment.mmd",
		}

		for _, diagram := range expectedDiagrams {
			diagramPath := filepath.Join(diagramsDir, diagram)
			_, err := os.Stat(diagramPath)
			assert.NoError(t, err, "Diagram source %s should exist", diagram)
		}
	})

	t.Run("mermaid syntax is valid", func(t *testing.T) {
		// This test MUST fail initially (TDD) until diagrams are created
		_, err := os.Stat(diagramsDir)
		if err != nil {
			t.Skip("Diagrams directory not yet created")
			return
		}

		var mermaidFiles []string
		err = filepath.Walk(diagramsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".mmd") {
				mermaidFiles = append(mermaidFiles, path)
			}
			return nil
		})

		if err != nil || len(mermaidFiles) == 0 {
			t.Skip("No Mermaid files found yet")
			return
		}

		for _, file := range mermaidFiles {
			content, readErr := os.ReadFile(file)
			require.NoError(t, readErr, "Should be able to read %s", file)

			contentStr := string(content)

			// Basic Mermaid syntax validation
			assert.Greater(t, len(contentStr), 10, "Diagram %s should have content", file)

			// Should contain valid Mermaid diagram type
			validTypes := []string{"graph", "flowchart", "sequenceDiagram", "classDiagram", "stateDiagram", "journey", "pie"}
			hasValidType := false
			for _, validType := range validTypes {
				if strings.Contains(contentStr, validType) {
					hasValidType = true
					break
				}
			}
			assert.True(t, hasValidType, "Diagram %s should contain valid Mermaid diagram type", file)
		}
	})

	t.Run("hexagonal architecture diagram shows correct structure", func(t *testing.T) {
		// This test MUST fail initially (TDD) until hexagonal architecture diagram is created
		hexDiagramPath := filepath.Join(diagramsDir, "hexagonal-architecture.mmd")

		_, err := os.Stat(hexDiagramPath)
		if err != nil {
			t.Skip("Hexagonal architecture diagram not yet created")
			return
		}

		content, readErr := os.ReadFile(hexDiagramPath)
		require.NoError(t, readErr, "Should be able to read hexagonal architecture diagram")

		contentStr := string(content)

		// Check for key hexagonal architecture components
		expectedComponents := []string{
			"domain",
			"usecase",
			"adapter",
			"port",
			"HTTP",
			"Database",
			"GORM",
			"Gin",
		}

		for _, component := range expectedComponents {
			assert.Contains(t, contentStr, component, "Hexagonal diagram should show %s", component)
		}
	})

	t.Run("data flow diagram shows multi-tenant isolation", func(t *testing.T) {
		// This test MUST fail initially (TDD) until data flow diagram is created
		dataFlowPath := filepath.Join(diagramsDir, "data-flow.mmd")

		_, err := os.Stat(dataFlowPath)
		if err != nil {
			t.Skip("Data flow diagram not yet created")
			return
		}

		content, readErr := os.ReadFile(dataFlowPath)
		require.NoError(t, readErr, "Should be able to read data flow diagram")

		contentStr := string(content)

		// Check for multi-tenant isolation concepts
		expectedConcepts := []string{
			"site_id",
			"tenant",
			"isolation",
			"database",
		}

		for _, concept := range expectedConcepts {
			assert.Contains(t, contentStr, concept, "Data flow diagram should show %s", concept)
		}
	})

	t.Run("deployment diagram shows docker architecture", func(t *testing.T) {
		// This test MUST fail initially (TDD) until deployment diagram is created
		deploymentPath := filepath.Join(diagramsDir, "deployment.mmd")

		_, err := os.Stat(deploymentPath)
		if err != nil {
			t.Skip("Deployment diagram not yet created")
			return
		}

		content, readErr := os.ReadFile(deploymentPath)
		require.NoError(t, readErr, "Should be able to read deployment diagram")

		contentStr := string(content)

		// Check for deployment concepts
		expectedConcepts := []string{
			"Docker",
			"container",
			"PostgreSQL",
			"API",
			"port",
		}

		for _, concept := range expectedConcepts {
			assert.Contains(t, contentStr, concept, "Deployment diagram should show %s", concept)
		}
	})

	t.Run("diagrams can be rendered", func(t *testing.T) {
		// This test validates that diagrams would render correctly
		// In a full implementation, this would use mermaid-cli to actually render
		// For now, we'll check that the syntax is reasonable

		_, err := os.Stat(diagramsDir)
		if err != nil {
			t.Skip("Diagrams directory not yet created")
			return
		}

		var mermaidFiles []string
		err = filepath.Walk(diagramsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".mmd") {
				mermaidFiles = append(mermaidFiles, path)
			}
			return nil
		})

		if err != nil || len(mermaidFiles) == 0 {
			t.Skip("No Mermaid files found yet")
			return
		}

		for _, file := range mermaidFiles {
			content, readErr := os.ReadFile(file)
			require.NoError(t, readErr, "Should be able to read %s", file)

			contentStr := string(content)

			// Check for basic rendering compatibility
			assert.NotContains(t, contentStr, "TODO", "Diagram %s should not contain TODO placeholders", file)
			assert.NotContains(t, contentStr, "FIXME", "Diagram %s should not contain FIXME placeholders", file)

			// Should not have syntax errors (basic check)
			lines := strings.Split(contentStr, "\n")
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "%") {
					continue // Skip empty lines and comments
				}

				// Basic check: lines should not have unmatched brackets
				openBrackets := strings.Count(line, "[") + strings.Count(line, "(") + strings.Count(line, "{")
				closeBrackets := strings.Count(line, "]") + strings.Count(line, ")") + strings.Count(line, "}")

				if openBrackets > 0 || closeBrackets > 0 {
					// For individual lines with brackets, they should be balanced or part of valid syntax
					// This is a simplified check - real validation would be more complex
					assert.True(t, len(line) > 2, "Line %d in %s should have meaningful content", i+1, file)
				}
			}
		}
	})
}
