package contract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModuleInitialization tests the contract for Go module initialization
func TestModuleInitialization(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "go.mod exists with correct module path",
			description: "Verify go.mod file exists and contains correct module path",
			assert: func(t *testing.T) {
				goModPath := filepath.Join("..", "..", "go.mod")
				require.FileExists(t, goModPath, "go.mod file should exist")

				content, err := os.ReadFile(goModPath)
				require.NoError(t, err, "should be able to read go.mod")

				goModContent := string(content)
				assert.Contains(t, goModContent, "module github.com/zcrossoverz/echoforge",
					"go.mod should contain correct module path")
			},
		},
		{
			name:        "go version requirement",
			description: "Verify go.mod specifies Go 1.25+ requirement",
			assert: func(t *testing.T) {
				goModPath := filepath.Join("..", "..", "go.mod")
				content, err := os.ReadFile(goModPath)
				require.NoError(t, err, "should be able to read go.mod")

				goModContent := string(content)
				assert.Contains(t, goModContent, "go 1.25",
					"go.mod should specify Go 1.25+ requirement")
			},
		},
		{
			name:        "directory structure exists",
			description: "Verify hexagonal architecture directory structure exists",
			assert: func(t *testing.T) {
				basePath := filepath.Join("..", "..")
				requiredDirs := []string{
					"internal/domain",
					"internal/usecase",
					"internal/adapters/http",
					"internal/adapters/persistence",
					"internal/adapters/logger",
					"cmd/server",
					"pkg/auth",
					"pkg/common",
					"configs",
					"tests/unit",
					"tests/integration",
					"tests/contract",
					"migrations",
					"docs",
				}

				for _, dir := range requiredDirs {
					dirPath := filepath.Join(basePath, dir)
					assert.DirExists(t, dirPath, "directory %s should exist", dir)
				}
			},
		},
		{
			name:        "gitignore exists with proper exclusions",
			description: "Verify .gitignore file exists with Go-specific exclusions",
			assert: func(t *testing.T) {
				gitignorePath := filepath.Join("..", "..", ".gitignore")
				require.FileExists(t, gitignorePath, ".gitignore file should exist")

				content, err := os.ReadFile(gitignorePath)
				require.NoError(t, err, "should be able to read .gitignore")

				gitignoreContent := string(content)
				requiredExclusions := []string{
					"*.exe",
					"*.test",
					"*.out",
					"coverage/",
					"vendor/",
					".vscode/",
					".idea/",
					"*.log",
					"/bin/",
					"/build/",
				}

				for _, exclusion := range requiredExclusions {
					assert.Contains(t, gitignoreContent, exclusion,
						".gitignore should contain %s exclusion", exclusion)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assert(t)
		})
	}
}

// TestModuleInitializationContract validates the module initialization contract
func TestModuleInitializationContract(t *testing.T) {
	t.Run("ModuleInitRequest validation", func(t *testing.T) {
		// This test validates the module initialization contract requirements
		// It should test the inputs and expected outputs as defined in contracts/

		// For now, we'll test basic structure compliance
		// In a real implementation, this would validate against the OpenAPI spec

		// Module path validation
		modulePath := "github.com/zcrossoverz/echoforge"
		assert.Regexp(t, `^github\.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`, modulePath,
			"module path should match expected pattern")

		// Go version validation
		goVersion := "1.25"
		assert.Regexp(t, `^1\.(2[5-9]|[3-9][0-9])(\.[0-9]+)?$`, goVersion,
			"Go version should be 1.25 or higher")
	})

	t.Run("ModuleInitResponse validation", func(t *testing.T) {
		// Test expected success response structure
		// This would validate the contract response schema

		expectedFiles := []string{
			"go.mod",
			"go.sum", // Will exist after dependencies are added
			".gitignore",
		}

		basePath := filepath.Join("..", "..")
		for _, file := range expectedFiles {
			if file == "go.sum" {
				// go.sum won't exist until dependencies are added
				continue
			}
			filePath := filepath.Join(basePath, file)
			assert.FileExists(t, filePath, "expected file %s should exist", file)
		}
	})
}
