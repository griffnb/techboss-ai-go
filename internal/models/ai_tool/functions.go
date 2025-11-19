package ai_tool

import (
	"strings"

	"github.com/griffnb/core/lib/model/fields"
)

// This file contains additional helper functions for the AiTool model

// GenerateSearchBlob creates a text blob from all searchable fields for tsvector indexing
func (this *AiTool) GenerateSearchBlob() string {
	var searchParts []string

	// Add basic fields
	if this.Name != nil && this.Name.Get() != "" {
		searchParts = append(searchParts, this.Name.Get())
	}
	if this.Description != nil && this.Description.Get() != "" {
		searchParts = append(searchParts, this.Description.Get())
	}
	if this.WebsiteURL != nil && this.WebsiteURL.Get() != "" {
		searchParts = append(searchParts, this.WebsiteURL.Get())
	}

	// Add metadata fields if available
	if this.MetaData != nil {
		metadata, err := this.MetaData.Get()
		if err == nil && metadata != nil {
			if metadata.Tagline != "" {
				searchParts = append(searchParts, metadata.Tagline)
			}
			if metadata.Introduction != "" {
				searchParts = append(searchParts, metadata.Introduction)
			}
			if metadata.HowItWorks != "" {
				searchParts = append(searchParts, metadata.HowItWorks)
			}
			if metadata.PricingRange != "" {
				searchParts = append(searchParts, metadata.PricingRange)
			}
			if metadata.PricingOptions != "" {
				searchParts = append(searchParts, metadata.PricingOptions)
			}
			if metadata.TargetAudience != "" {
				searchParts = append(searchParts, metadata.TargetAudience)
			}

			// Add key benefits
			if len(metadata.KeyBenefits) > 0 {
				searchParts = append(searchParts, strings.Join(metadata.KeyBenefits, " "))
			}

			// Add applications
			if len(metadata.Applications) > 0 {
				searchParts = append(searchParts, strings.Join(metadata.Applications, " "))
			}

			// Add core features
			if len(metadata.Features) > 0 {
				for _, feature := range metadata.Features {
					if feature != nil {
						if feature.Feature != "" {
							searchParts = append(searchParts, feature.Feature)
						}
						if feature.Description != "" {
							searchParts = append(searchParts, feature.Description)
						}
					}
				}
			}
		}
	}

	// Join all parts with spaces and clean up extra whitespace
	searchBlob := strings.Join(searchParts, " ")
	searchBlob = strings.TrimSpace(searchBlob)

	// Replace multiple spaces with single spaces
	for strings.Contains(searchBlob, "  ") {
		searchBlob = strings.ReplaceAll(searchBlob, "  ", " ")
	}

	return searchBlob
}

// UpdateSearchVector updates the search_blob_tsv field with a tsvector representation
func (this *AiTool) UpdateSearchVector() {
	searchBlob := this.GenerateSearchBlob()
	if searchBlob != "" {
		// Set the search blob as text - PostgreSQL will convert it to tsvector
		// when we use to_tsvector() in the database query
		if this.SearchBlobTSV == nil {
			this.SearchBlobTSV = &fields.StringField{}
		}
		this.SearchBlobTSV.Set(searchBlob)
	}
}
