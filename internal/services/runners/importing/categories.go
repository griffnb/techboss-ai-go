package importing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

type CategoryImportRunner struct{}

func (this *CategoryImportRunner) Run(_ context.Context, _ ...string) error {
	toolsFilePath := "/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/ai_tool/tools.json"
	outputFilePath := "/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/category/categories.json"

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputFilePath)
	// nolint:gosec
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Extract categories
	if err := this.ExtractCategories(toolsFilePath, outputFilePath); err != nil {
		log.Fatalf("Failed to extract categories: %v", err)
	}

	return nil
}

// Tool represents the structure of each tool in the JSON file
type Tool struct {
	Categorization   string `json:"categorization"`
	BusinessFunction string `json:"business_function"`
}

// Categories represents the extracted unique values
type Categories struct {
	Categorizations   []string `json:"categorizations"`
	BusinessFunctions []string `json:"business_functions"`
}

// ExtractCategories reads the tools.json file and extracts unique categorization and business_function values
func (this *CategoryImportRunner) ExtractCategories(toolsFilePath, outputFilePath string) error {
	// Read the tools.json file
	// nolint:gosec
	data, err := os.ReadFile(toolsFilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Parse the JSON
	var tools []Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Use maps to store unique values
	categorizationSet := make(map[string]bool)
	businessFunctionSet := make(map[string]bool)

	// Extract unique values
	for _, tool := range tools {
		if tool.Categorization != "" {
			categorizationSet[tool.Categorization] = true
		}
		if tool.BusinessFunction != "" {
			businessFunctionSet[tool.BusinessFunction] = true
		}
	}

	// Convert maps to sorted slices
	var categorizations []string
	for cat := range categorizationSet {
		categorizations = append(categorizations, cat)
	}
	sort.Strings(categorizations)

	var businessFunctions []string
	for bf := range businessFunctionSet {
		businessFunctions = append(businessFunctions, bf)
	}
	sort.Strings(businessFunctions)

	// Create the categories structure
	categories := Categories{
		Categorizations:   categorizations,
		BusinessFunctions: businessFunctions,
	}

	// Marshal to JSON with indentation
	output, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	// Write to output file
	// nolint:gosec
	if err := os.WriteFile(outputFilePath, output, 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Successfully extracted %d categorizations and %d business functions to %s\n",
		len(categorizations), len(businessFunctions), outputFilePath)

	return nil
}
