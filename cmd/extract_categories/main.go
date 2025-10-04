package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/griffnb/techboss-ai-go/internal/models/category"
)

func main() {
	toolsFilePath := "/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/ai_tool/tools.json"
	outputFilePath := "/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/category/categories.json"

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputFilePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Extract categories
	if err := category.ExtractCategories(toolsFilePath, outputFilePath); err != nil {
		log.Fatalf("Failed to extract categories: %v", err)
	}
}