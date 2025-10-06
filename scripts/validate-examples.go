package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Example validation script - placeholder implementation")

	// Get the repository root
	repoRoot, err := getRepoRoot()
	if err != nil {
		log.Fatalf("Failed to find repository root: %v", err)
	}

	docsDir := filepath.Join(repoRoot, "docs")
	fmt.Printf("Validating examples in: %s\n", docsDir)

	// TODO: Implement example validation logic
	// This will be completed in T024
	fmt.Println("✓ Example validator initialized")
}

func getRepoRoot() (string, error) {
	// Find the repository root by looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repository root not found")
		}
		dir = parent
	}
}
