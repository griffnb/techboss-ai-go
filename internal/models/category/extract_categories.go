package category

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Tool represents the structure of each tool in the JSON file
type Tool struct {
	Categorization    string `json:"categorization"`
	BusinessFunction  string `json:"business_function"`
}

// Categories represents the extracted unique values
type Categories struct {
	Categorizations   []string `json:"categorizations"`
	BusinessFunctions []string `json:"business_functions"`
}

// ExtractCategories reads the tools.json file and extracts unique categorization and business_function values
func ExtractCategories(toolsFilePath, outputFilePath string) error {
	// Read the tools.json file
	data, err := os.ReadFile(toolsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read tools file: %w", err)
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
	if err := os.WriteFile(outputFilePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Successfully extracted %d categorizations and %d business functions to %s\n",
		len(categorizations), len(businessFunctions), outputFilePath)

	return nil
}