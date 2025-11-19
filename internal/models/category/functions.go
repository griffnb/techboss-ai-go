package category

import (
	"regexp"
	"strings"
)

// This file contains additional helper functions for the Category model

// GenerateSlug creates a URL-friendly slug from a category name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}
