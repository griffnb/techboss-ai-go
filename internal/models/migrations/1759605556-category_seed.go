package migrations

import (
	"encoding/json"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/category"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:    1759605556,
		Table: category.TABLE,
		PostMigrationTransform: func() error {
			// Category data from categories.json
			categoriesData := `{
  "categorizations": [
    "AI Agents / Multi-agent Systems",
    "AI Governance",
    "AI Search Engines",
    "AI for Presentations & Slides",
    "Accounting & Finance",
    "Accounting Automation",
    "Audio & Music Generation",
    "Automated Testing",
    "Automation & Workflow Builders",
    "Avatar & Character Generation",
    "Business Function",
    "Chatbots & Virtual Assistants",
    "Coding & Development",
    "Compliance & Regulation",
    "Compliance Management",
    "Content Creation",
    "Content Generation",
    "Content Marketing",
    "Customer Support",
    "Data Analysis & Visualization",
    "Design & Creativity",
    "Document & PDF AI",
    "Document & PDF AI (reading, summarizing, querying)",
    "Education & Learning",
    "Email & Copywriting",
    "Face Swap / Deepfake",
    "Finance & Investing",
    "Financial",
    "Gaming",
    "Healthcare",
    "Human Resources & Recruitment",
    "Image Generation",
    "Legal & Contract Analysis",
    "Marketing & Sales",
    "Media & Publishing",
    "Meeting Summarizers & Note-takers",
    "Multi-modal AI Tools",
    "Multi-modal AI Tools (text+image+audio)",
    "Pricing Optimization",
    "Productivity",
    "Real Estate",
    "Risk Solutions",
    "SEO Optimization",
    "Sales",
    "Sales Engagement",
    "Security & Cybersecurity",
    "Social Media Management",
    "Social Media Tools",
    "Startups & Product Development",
    "Text Generation",
    "Trading & Investing",
    "Travel & Booking",
    "Travel & Tourism",
    "Travel Planning",
    "Video Generation",
    "Voice AI Platform",
    "Voice Generation",
    "Web Design & Development",
    "Web Development",
    "Website Generation",
    "Workflow Automation",
    "Writing & Editing",
    "eCommerce & Retail"
  ],
  "business_functions": [
    "Back Office",
    "Client Service & Support",
    "Compliance",
    "Compliance & Data Privacy",
    "Compliance Management",
    "Compliance and Risk Management",
    "Creative",
    "Creative & Design",
    "Creative and Design",
    "Customer Service & Support",
    "Design & Creative",
    "Growth & Marketing",
    "Hiring & Recruitment",
    "Investment Management",
    "Job Search & Recruitment",
    "Marketing & Sales",
    "Operations",
    "Product Management",
    "Risk Management",
    "Sales",
    "Technology & IT",
    "Training & Development",
    "Writing & Editing"
  ]
}`

			var data struct {
				Categorizations   []string `json:"categorizations"`
				BusinessFunctions []string `json:"business_functions"`
			}

			err := json.Unmarshal([]byte(categoriesData), &data)
			if err != nil {
				return fmt.Errorf("failed to unmarshal categories data: %w", err)
			}

			// Create parent categories
			categorizationsParent := category.New()
			categorizationsParent.Name.Set("Categorizations")
			categorizationsParent.Slug.Set("categorizations")
			categorizationsParent.Description.Set("AI tool categorizations")
			categorizationsParent.Status.Set(constants.STATUS_ACTIVE)
			err = categorizationsParent.Save(nil)
			if err != nil {
				return fmt.Errorf("failed to create categorizations parent: %w", err)
			}

			businessFunctionsParent := category.New()
			businessFunctionsParent.Name.Set("Business Functions")
			businessFunctionsParent.Slug.Set("business-functions")
			businessFunctionsParent.Description.Set("Business function categories")
			businessFunctionsParent.Status.Set(constants.STATUS_ACTIVE)
			err = businessFunctionsParent.Save(nil)
			if err != nil {
				return fmt.Errorf("failed to create business functions parent: %w", err)
			}

			// Insert categorizations as child categories
			for _, catName := range data.Categorizations {
				catObj := category.New()
				catObj.Name.Set(catName)
				catObj.Slug.Set(category.GenerateSlug(catName))
				catObj.Description.Set("")
				catObj.ParentCategoryID.Set(categorizationsParent.ID())
				catObj.Status.Set(constants.STATUS_ACTIVE)
				err := catObj.Save(nil)
				if err != nil {
					return fmt.Errorf("failed to create categorization '%s': %w", catName, err)
				}
			}

			// Insert business functions as child categories
			for _, funcName := range data.BusinessFunctions {
				funcObj := category.New()
				funcObj.Name.Set(funcName)
				funcObj.Slug.Set(category.GenerateSlug(funcName))
				funcObj.Description.Set("")
				funcObj.ParentCategoryID.Set(businessFunctionsParent.ID())
				funcObj.Status.Set(constants.STATUS_ACTIVE)
				err := funcObj.Save(nil)
				if err != nil {
					return fmt.Errorf("failed to create business function '%s': %w", funcName, err)
				}
			}

			return nil
		},
	})
}