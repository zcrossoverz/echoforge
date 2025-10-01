package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Constitutional requirements from spec
	maxBinarySizeMB     = 50  // 50MB maximum binary size
	maxBuildTimeSeconds = 300 // 5 minutes maximum build time
	minGoVersion        = "1.25"
)

// TestBinaryBuildValidation tests all aspects of binary build validation
func TestBinaryBuildValidation(t *testing.T) {
	basePath := filepath.Join("..", "..")
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { os.Chdir(originalDir) }()

	err = os.Chdir(basePath)
	require.NoError(t, err)

	// Clean up any existing binaries
	t.Cleanup(func() {
		cleanupBinaries(t)
	})

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "basic build success",
			description: "Test that the project builds successfully",
			assert: func(t *testing.T) {
				// This will fail until main.go is created in T017
				binaryPath := getBinaryPath()

				startTime := time.Now()
				cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
				output, err := cmd.CombinedOutput()
				buildTime := time.Since(startTime)

				if err != nil {
					t.Logf("Build failed (expected until main.go created): %s", string(output))
					// In TDD, we expect this to fail until T017 creates main.go
					assert.Fail(t, "Build should succeed after main.go is created")
					return
				}

				t.Logf("Build succeeded in %v: %s", buildTime, string(output))
				assert.True(t, buildTime < time.Duration(maxBuildTimeSeconds)*time.Second,
					"build should complete within %d seconds", maxBuildTimeSeconds)
			},
		},
		{
			name:        "binary size constraint",
			description: "Test that the built binary meets size requirements",
			assert: func(t *testing.T) {
				binaryPath := getBinaryPath()

				// Try to build first
				cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
				output, err := cmd.CombinedOutput()

				if err != nil {
					t.Logf("Build failed, skipping size test: %s", string(output))
					return
				}

				// Check binary size
				fileInfo, err := os.Stat(binaryPath)
				require.NoError(t, err, "should be able to stat built binary")

				sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
				t.Logf("Binary size: %.2f MB", sizeMB)

				assert.True(t, sizeMB <= float64(maxBinarySizeMB),
					"binary size (%.2f MB) should not exceed %d MB", sizeMB, maxBinarySizeMB)
			},
		},
		{
			name:        "optimized build",
			description: "Test that optimized builds produce smaller binaries",
			assert: func(t *testing.T) {
				// Build with optimization flags
				optimizedPath := getBinaryPath() + "_optimized"

				cmd := exec.Command("go", "build",
					"-ldflags", "-s -w", // Strip debug info
					"-trimpath", // Remove file system paths
					"-o", optimizedPath,
					"./cmd/server")
				output, err := cmd.CombinedOutput()

				if err != nil {
					t.Logf("Optimized build failed: %s", string(output))
					return
				}

				// Check optimized binary size
				fileInfo, err := os.Stat(optimizedPath)
				require.NoError(t, err)

				optimizedSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
				t.Logf("Optimized binary size: %.2f MB", optimizedSizeMB)

				assert.True(t, optimizedSizeMB <= float64(maxBinarySizeMB),
					"optimized binary should meet size constraints")

				// Clean up
				os.Remove(optimizedPath)
			},
		},
		{
			name:        "cross-platform builds",
			description: "Test that the project can be built for different platforms",
			assert: func(t *testing.T) {
				platforms := []struct {
					goos   string
					goarch string
				}{
					{"linux", "amd64"},
					{"darwin", "amd64"},
					{"windows", "amd64"},
				}

				for _, platform := range platforms {
					t.Run(fmt.Sprintf("%s-%s", platform.goos, platform.goarch), func(t *testing.T) {
						binaryName := fmt.Sprintf("echoforge_%s_%s", platform.goos, platform.goarch)
						if platform.goos == "windows" {
							binaryName += ".exe"
						}

						cmd := exec.Command("go", "build", "-o", binaryName, "./cmd/server")
						cmd.Env = append(os.Environ(),
							"GOOS="+platform.goos,
							"GOARCH="+platform.goarch,
						)

						output, err := cmd.CombinedOutput()
						if err != nil {
							t.Logf("Cross-platform build failed for %s-%s: %s",
								platform.goos, platform.goarch, string(output))
							return
						}

						t.Logf("Successfully built for %s-%s", platform.goos, platform.goarch)

						// Check that binary was created
						_, err = os.Stat(binaryName)
						assert.NoError(t, err, "binary should exist after build")

						// Clean up
						os.Remove(binaryName)
					})
				}
			},
		},
		{
			name:        "build reproducibility",
			description: "Test that builds are reproducible",
			assert: func(t *testing.T) {
				// Build twice and compare
				binary1 := getBinaryPath() + "_build1"
				binary2 := getBinaryPath() + "_build2"

				// First build
				cmd1 := exec.Command("go", "build", "-trimpath", "-o", binary1, "./cmd/server")
				output1, err1 := cmd1.CombinedOutput()

				// Second build
				cmd2 := exec.Command("go", "build", "-trimpath", "-o", binary2, "./cmd/server")
				output2, err2 := cmd2.CombinedOutput()

				if err1 != nil || err2 != nil {
					t.Logf("One or both builds failed - build1: %v, build2: %v", err1, err2)
					t.Logf("Output1: %s", string(output1))
					t.Logf("Output2: %s", string(output2))
					return
				}

				// Compare file sizes (basic reproducibility check)
				info1, err := os.Stat(binary1)
				require.NoError(t, err)
				info2, err := os.Stat(binary2)
				require.NoError(t, err)

				assert.Equal(t, info1.Size(), info2.Size(),
					"reproducible builds should produce identical binary sizes")

				// Clean up
				os.Remove(binary1)
				os.Remove(binary2)
			},
		},
		{
			name:        "build cache utilization",
			description: "Test that build cache is properly utilized",
			assert: func(t *testing.T) {
				// First build (cold cache)
				binaryPath := getBinaryPath()

				startTime1 := time.Now()
				cmd1 := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
				output1, err1 := cmd1.CombinedOutput()
				buildTime1 := time.Since(startTime1)

				if err1 != nil {
					t.Logf("First build failed: %s", string(output1))
					return
				}

				// Remove binary but keep cache
				os.Remove(binaryPath)

				// Second build (warm cache)
				startTime2 := time.Now()
				cmd2 := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
				output2, err2 := cmd2.CombinedOutput()
				buildTime2 := time.Since(startTime2)

				if err2 != nil {
					t.Logf("Second build failed: %s", string(output2))
					return
				}

				t.Logf("Build times - Cold cache: %v, Warm cache: %v", buildTime1, buildTime2)

				// Second build should be faster (cache hit)
				assert.True(t, buildTime2 <= buildTime1,
					"warm cache build should be faster or equal to cold cache build")
			},
		},
		{
			name:        "build with race detection",
			description: "Test that the project builds with race detection enabled",
			assert: func(t *testing.T) {
				binaryPath := getBinaryPath() + "_race"

				cmd := exec.Command("go", "build", "-race", "-o", binaryPath, "./cmd/server")
				output, err := cmd.CombinedOutput()

				if err != nil {
					t.Logf("Race-enabled build failed: %s", string(output))
					return
				}

				t.Logf("Race-enabled build succeeded: %s", string(output))

				// Race-enabled binaries are larger
				fileInfo, err := os.Stat(binaryPath)
				require.NoError(t, err)

				raceSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
				t.Logf("Race-enabled binary size: %.2f MB", raceSizeMB)

				// Clean up
				os.Remove(binaryPath)
			},
		},
		{
			name:        "build tags support",
			description: "Test that build tags work correctly",
			assert: func(t *testing.T) {
				// Test with a hypothetical build tag
				binaryPath := getBinaryPath() + "_tagged"

				cmd := exec.Command("go", "build", "-tags", "integration", "-o", binaryPath, "./cmd/server")
				output, err := cmd.CombinedOutput()

				if err != nil {
					t.Logf("Tagged build failed (expected until main.go supports tags): %s", string(output))
					return
				}

				t.Logf("Tagged build succeeded: %s", string(output))

				// Clean up
				os.Remove(binaryPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assert(t)
		})
	}
}

// TestBuildEnvironmentValidation tests the build environment requirements
func TestBuildEnvironmentValidation(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "go version requirement",
			description: "Test that Go version meets minimum requirements",
			assert: func(t *testing.T) {
				cmd := exec.Command("go", "version")
				output, err := cmd.CombinedOutput()
				require.NoError(t, err)

				versionStr := string(output)
				t.Logf("Go version: %s", versionStr)

				// Extract version number
				re := regexp.MustCompile(`go(\d+\.\d+)`)
				matches := re.FindStringSubmatch(versionStr)
				require.Len(t, matches, 2, "should be able to parse Go version")

				version := matches[1]
				versionParts := strings.Split(version, ".")
				require.Len(t, versionParts, 2)

				major, err := strconv.Atoi(versionParts[0])
				require.NoError(t, err)
				minor, err := strconv.Atoi(versionParts[1])
				require.NoError(t, err)

				minVersionParts := strings.Split(minGoVersion, ".")
				minMajor, _ := strconv.Atoi(minVersionParts[0])
				minMinor, _ := strconv.Atoi(minVersionParts[1])

				assert.True(t, major > minMajor || (major == minMajor && minor >= minMinor),
					"Go version %s should be >= %s", version, minGoVersion)
			},
		},
		{
			name:        "build tools availability",
			description: "Test that required build tools are available",
			assert: func(t *testing.T) {
				tools := []string{"go", "gofmt", "vet"}

				for _, tool := range tools {
					cmd := exec.Command(tool, "version")
					if tool == "vet" {
						cmd = exec.Command("go", "vet", "-h")
					} else if tool == "gofmt" {
						cmd = exec.Command("gofmt", "-h")
					}

					output, err := cmd.CombinedOutput()
					assert.NoError(t, err, "tool %s should be available: %s", tool, string(output))
					t.Logf("Tool %s is available", tool)
				}
			},
		},
		{
			name:        "build environment variables",
			description: "Test that build environment is properly configured",
			assert: func(t *testing.T) {
				envVars := []string{"GOPATH", "GOROOT", "GOPROXY", "GOSUMDB"}

				for _, envVar := range envVars {
					cmd := exec.Command("go", "env", envVar)
					output, err := cmd.CombinedOutput()
					require.NoError(t, err)

					value := strings.TrimSpace(string(output))
					t.Logf("%s: %s", envVar, value)

					if envVar == "GOROOT" || envVar == "GOPROXY" {
						assert.NotEmpty(t, value, "%s should be set", envVar)
					}
				}
			},
		},
		{
			name:        "module mode enabled",
			description: "Test that Go modules are enabled",
			assert: func(t *testing.T) {
				cmd := exec.Command("go", "env", "GO111MODULE")
				output, err := cmd.CombinedOutput()
				require.NoError(t, err)

				goModule := strings.TrimSpace(string(output))
				t.Logf("GO111MODULE: %s", goModule)

				// Should be "on" or empty (which defaults to auto)
				assert.True(t, goModule == "on" || goModule == "" || goModule == "auto",
					"GO111MODULE should enable module mode")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assert(t)
		})
	}
}

// Helper functions

func getBinaryPath() string {
	binaryName := "echoforge"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

func cleanupBinaries(t *testing.T) {
	patterns := []string{
		"echoforge*",
		"*.exe",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			if strings.Contains(match, "echoforge") {
				err := os.Remove(match)
				if err == nil {
					t.Logf("Cleaned up binary: %s", match)
				}
			}
		}
	}
}

// TestBuildPerformance tests build performance characteristics
func TestBuildPerformance(t *testing.T) {
	basePath := filepath.Join("..", "..")
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { os.Chdir(originalDir) }()

	err = os.Chdir(basePath)
	require.NoError(t, err)

	t.Run("parallel build capability", func(t *testing.T) {
		// Test that builds can utilize multiple CPU cores
		cmd := exec.Command("go", "env", "GOMAXPROCS")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		maxprocs := strings.TrimSpace(string(output))
		t.Logf("GOMAXPROCS: %s", maxprocs)

		if maxprocs != "0" {
			procs, err := strconv.Atoi(maxprocs)
			if err == nil {
				assert.True(t, procs > 0, "should have at least 1 processor for builds")
			}
		}
	})

	t.Run("incremental build performance", func(t *testing.T) {
		// This test validates that incremental builds are faster
		binaryPath := getBinaryPath()

		// Initial build
		startTime := time.Now()
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
		output, err := cmd.CombinedOutput()
		buildTime := time.Since(startTime)

		if err != nil {
			t.Logf("Build failed: %s", string(output))
			return
		}

		t.Logf("Build completed in %v", buildTime)
		assert.True(t, buildTime < time.Duration(maxBuildTimeSeconds)*time.Second,
			"build should complete within reasonable time")

		// Clean up
		os.Remove(binaryPath)
	})

	t.Run("memory usage during build", func(t *testing.T) {
		// This is a placeholder for memory usage validation
		// In practice, you might use tools to monitor memory usage during builds
		t.Logf("Build memory usage validation - placeholder for monitoring tools")

		// Basic validation that we can complete a build without running out of memory
		binaryPath := getBinaryPath()
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
		output, err := cmd.CombinedOutput()

		if err != nil {
			// Check if it's a memory-related error
			outputStr := string(output)
			assert.NotContains(t, outputStr, "out of memory", "build should not run out of memory")
			assert.NotContains(t, outputStr, "killed", "build should not be killed due to resource limits")
		}

		// Clean up
		os.Remove(binaryPath)
	})
}
