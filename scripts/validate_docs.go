package main

import (
	"fmt"
	"os"
)

// Simple documentation validation script
func main() {
	fmt.Println("🔍 Validating Echoforge Documentation...")

	// Change to project root
	if err := os.Chdir(".."); err != nil {
		fmt.Printf("Error changing to project root: %v\n", err)
		os.Exit(1)
	}

	// Define required files
	requiredFiles := []string{
		"docs/api/openapi.yaml",
		"docs/guides/site-extension/manga-site-setup.md",
		"docs/guides/site-extension/blog-site-setup.md",
		"docs/guides/site-extension/portfolio-site-setup.md",
		"docs/architecture/system-architecture.md",
		"docs/architecture/data-flow.md",
		"docs/architecture/deployment-architecture.md",
	}

	allExist := true

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("❌ Missing: %s\n", file)
			allExist = false
		} else {
			// Check file size to ensure it's not empty
			stat, _ := os.Stat(file)
			if stat.Size() < 100 {
				fmt.Printf("⚠️  Too small: %s (%d bytes)\n", file, stat.Size())
				allExist = false
			} else {
				fmt.Printf("✅ Valid: %s (%d bytes)\n", file, stat.Size())
			}
		}
	}

	if allExist {
		fmt.Println("\n🎉 All documentation files are present and have content!")

		// Check for key content in OpenAPI
		if content, err := os.ReadFile("docs/api/openapi.yaml"); err == nil {
			contentStr := string(content)
			if len(contentStr) > 10000 {
				fmt.Println("✅ OpenAPI specification appears comprehensive (>10KB)")
			} else {
				fmt.Println("⚠️  OpenAPI specification might be incomplete (<10KB)")
			}
		}

		os.Exit(0)
	} else {
		fmt.Println("\n❌ Some documentation files are missing or incomplete!")
		os.Exit(1)
	}
}
