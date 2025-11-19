package controllers

import (
	"github.com/griffnb/techboss-ai-go/internal/controllers/accounts"
	"github.com/griffnb/techboss-ai-go/internal/controllers/admins"
	"github.com/griffnb/techboss-ai-go/internal/controllers/agents"
	"github.com/griffnb/techboss-ai-go/internal/controllers/ai"
	"github.com/griffnb/techboss-ai-go/internal/controllers/ai_tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/billing"
	"github.com/griffnb/techboss-ai-go/internal/controllers/billing_plans"
	"github.com/griffnb/techboss-ai-go/internal/controllers/categories"
	"github.com/griffnb/techboss-ai-go/internal/controllers/change_logs"

	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/techboss-ai-go/internal/controllers/global_configs"
	"github.com/griffnb/techboss-ai-go/internal/controllers/leads"
	"github.com/griffnb/techboss-ai-go/internal/controllers/login"
	"github.com/griffnb/techboss-ai-go/internal/controllers/logs"
	"github.com/griffnb/techboss-ai-go/internal/controllers/organizations"
	"github.com/griffnb/techboss-ai-go/internal/controllers/utilities"
)

// Setup Adds the controllers to the router
func Setup(coreRouter *router.CoreRouter) {
	admins.Setup(coreRouter)

	change_logs.Setup(coreRouter)
	global_configs.Setup(coreRouter)
	login.Setup(coreRouter)
	logs.Setup(coreRouter)

	utilities.Setup(coreRouter)

	ai.Setup(coreRouter)
	agents.Setup(coreRouter)
	accounts.Setup(coreRouter)
	ai_tools.Setup(coreRouter)
	billing.Setup(coreRouter)
	billing_plans.Setup(coreRouter)
	categories.Setup(coreRouter)
	leads.Setup(coreRouter)
	organizations.Setup(coreRouter)
	agents.Setup(coreRouter)

	// Print all routes
	// printRoutes(coreRouter.Router)
}
