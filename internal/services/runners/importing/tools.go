package importing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/ai_tool"
	"github.com/griffnb/techboss-ai-go/internal/models/category"
	"github.com/pkg/errors"
)

// ToolData represents the structure of tools from tools.json
type ToolData struct {
	ID               int             `json:"id"`
	WebsiteURL       string          `json:"website_url"`
	HeroSection      HeroSection     `json:"hero_section"`
	Overview         Overview        `json:"overview"`
	FeaturesAndCaps  FeaturesAndCaps `json:"features_and_capabilities"`
	UseCases         UseCases        `json:"use_cases"`
	PricingAndPlans  PricingAndPlans `json:"pricing_and_plans"`
	TargetAudience   string          `json:"target_audience"`
	Tags             *string         `json:"tags"`
	CreatedAt        string          `json:"created_at"`
	ToolName         string          `json:"tool_name"`
	Categorization   string          `json:"categorization"`
	BusinessFunction string          `json:"business_function"`
	Affiliate        string          `json:"affiliate"`
	IsFeatured       bool            `json:"is_featured"`
	Embedding        *string         `json:"embedding"`
	UpdatedAt        *string         `json:"updated_at"`
}

type HeroSection struct {
	Logo        *string `json:"logo"`
	Tagline     string  `json:"tagline"`
	ToolName    string  `json:"toolName"`
	Description string  `json:"description"`
}

type Overview struct {
	KeyBenefits  []string `json:"keyBenefits"`
	Introduction string   `json:"introduction"`
}

type FeaturesAndCaps struct {
	HowItWorks   string        `json:"howItWorks"`
	CoreFeatures []CoreFeature `json:"coreFeatures"`
}

type CoreFeature struct {
	Feature     string `json:"feature"`
	Description string `json:"description"`
}

type UseCases struct {
	Applications []string `json:"applications"`
}

type PricingAndPlans struct {
	FreeTier       bool    `json:"freeTier"`
	PricingRange   *string `json:"pricingRange"`
	PricingOptions string  `json:"pricingOptions"`
}

type ToolImportRunner struct{}

func (this *ToolImportRunner) Run(_ context.Context, args ...string) error {
	if len(args) < 1 {
		return errors.New("please provide the path to tools.json")
	}

	toolsJSONPath := args[0]

	err := importTools(toolsJSONPath)
	if err != nil {
		return errors.WithMessagef(err, "failed to import tools from %s", toolsJSONPath)
	}

	fmt.Println("Import completed successfully!")

	return nil
}

func importTools(filePath string) error {
	// Read the tools.json file
	// nolint:gosec
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON
	var tools []ToolData
	err = json.Unmarshal(data, &tools)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Printf("Found %d tools to import\n", len(tools))

	// Load existing categories for matching
	categoryMap, businessFunctionMap, err := loadCategoryMaps()
	if err != nil {
		return fmt.Errorf("failed to load category maps: %w", err)
	}

	// Import each tool
	for i, tool := range tools {
		fmt.Printf("Importing tool %d/%d: %s\n", i+1, len(tools), tool.ToolName)

		err := importSingleTool(tool, categoryMap, businessFunctionMap)
		if err != nil {
			log.Error(err)
			// Continue with next tool instead of stopping
			continue
		}
	}

	return nil
}

func loadCategoryMaps() (map[string]string, map[string]string, error) {
	ctx := context.Background()

	// Load all categories to create lookup maps
	allCategories, err := category.FindAll(ctx, model.NewOptions())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load categories: %w", err)
	}

	categoryMap := make(map[string]string)
	businessFunctionMap := make(map[string]string)

	for _, cat := range allCategories {
		// For matching, use the category name to generate slug and compare
		slug := category.GenerateSlug(cat.Name.Get())

		// Determine if this is a categorization or business function based on parent
		if cat.ParentCategoryID.Get() != "" {
			// This is a child category, determine which parent
			parent, err := category.Get(ctx, cat.ParentCategoryID.Get())
			if err != nil {
				continue
			}

			parentSlug := parent.Slug.Get()

			switch parentSlug {
			case "categorizations":
				categoryMap[slug] = cat.ID().String()
			case "business-functions":
				businessFunctionMap[slug] = cat.ID().String()
			}

		}
	}

	return categoryMap, businessFunctionMap, nil
}

func importSingleTool(tool ToolData, categoryMap, businessFunctionMap map[string]string) error {
	// Create a new AI tool
	aiTool := ai_tool.New()

	// Map basic fields
	aiTool.Name.Set(tool.ToolName)
	aiTool.Description.Set(tool.HeroSection.Description)
	aiTool.WebsiteURL.Set(tool.WebsiteURL)

	// Set affiliate URL if affiliate is true
	if tool.Affiliate == "true" {
		aiTool.AffiliateURL.Set(tool.WebsiteURL)
	}

	// Set featured status
	if tool.IsFeatured {
		aiTool.IsFeatured.Set(1)
	} else {
		aiTool.IsFeatured.Set(0)
	}

	// Map category IDs
	if tool.Categorization != "" {
		categorySlug := category.GenerateSlug(tool.Categorization)
		if categoryID, exists := categoryMap[categorySlug]; exists {
			aiTool.CategoryID.Set(types.UUID(categoryID))
		}
	}

	if tool.BusinessFunction != "" {
		businessFunctionSlug := category.GenerateSlug(tool.BusinessFunction)
		if businessFunctionID, exists := businessFunctionMap[businessFunctionSlug]; exists {
			aiTool.BusinessFunctionCategoryID.Set(types.UUID(businessFunctionID))
		}
	}

	// Create MetaData structure
	metaData := &ai_tool.MetaData{
		Logo:             getStringValue(tool.HeroSection.Logo),
		Tagline:          tool.HeroSection.Tagline,
		KeyBenefits:      tool.Overview.KeyBenefits,
		Introduction:     tool.Overview.Introduction,
		HowItWorks:       tool.FeaturesAndCaps.HowItWorks,
		CoreFeatures:     convertCoreFeatures(tool.FeaturesAndCaps.CoreFeatures),
		Applications:     tool.UseCases.Applications,
		FreeTier:         tool.PricingAndPlans.FreeTier,
		PricingRange:     getStringValue(tool.PricingAndPlans.PricingRange),
		PricingOptions:   tool.PricingAndPlans.PricingOptions,
		TargetAudience:   tool.TargetAudience,
		Categorization:   tool.Categorization,
		BusinessFunction: tool.BusinessFunction,
	}

	aiTool.MetaData.Set(metaData)

	// Save the AI tool
	err := aiTool.Save(nil)
	if err != nil {
		return errors.WithMessagef(err, "failed to save tool %s", tool.ToolName)
	}

	return nil
}

func convertCoreFeatures(features []CoreFeature) []*ai_tool.CoreFeature {
	result := make([]*ai_tool.CoreFeature, len(features))
	for i, feature := range features {
		result[i] = &ai_tool.CoreFeature{
			Feature:     feature.Feature,
			Description: feature.Description,
		}
	}
	return result
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
