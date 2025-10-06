package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Postman collection builder - placeholder implementation")

	// Get the repository root
	repoRoot, err := getRepoRoot()
	if err != nil {
		log.Fatalf("Failed to find repository root: %v", err)
	}

	postmanDir := filepath.Join(repoRoot, "docs", "postman")
	fmt.Printf("Building Postman collection in: %s\n", postmanDir)

	// TODO: Implement Postman collection generation logic
	// This will be completed in T020
	fmt.Println("✓ Postman collection builder initialized")
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
