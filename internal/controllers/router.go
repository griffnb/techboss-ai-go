package controllers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/griffnb/techboss-ai-go/internal/controllers/accounts"
	"github.com/griffnb/techboss-ai-go/internal/controllers/admins"
	"github.com/griffnb/techboss-ai-go/internal/controllers/agents"
	"github.com/griffnb/techboss-ai-go/internal/controllers/ai"
	"github.com/griffnb/techboss-ai-go/internal/controllers/ai_tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/billing"
	"github.com/griffnb/techboss-ai-go/internal/controllers/billing_plan_prices"
	"github.com/griffnb/techboss-ai-go/internal/controllers/billing_plans"
	"github.com/griffnb/techboss-ai-go/internal/controllers/categories"
	"github.com/griffnb/techboss-ai-go/internal/controllers/change_logs"
	"github.com/griffnb/techboss-ai-go/internal/controllers/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/controllers/subscriptions"

	"github.com/griffnb/core/lib/log"
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
	billing_plan_prices.Setup(coreRouter)
	categories.Setup(coreRouter)
	leads.Setup(coreRouter)
	organizations.Setup(coreRouter)
	sandbox.Setup(coreRouter)
	subscriptions.Setup(coreRouter)
	agents.Setup(coreRouter)

	// Setup static file serving for Modal Sandbox UI
	setupStaticFiles(coreRouter)

	// Print all routes
	// printRoutes(coreRouter.Router)
}

// setupStaticFiles configures static file serving for the application.
// It serves files from the ./static directory at the /static/ path.
func setupStaticFiles(coreRouter *router.CoreRouter) {
	// Get working directory to construct absolute path
	workDir, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return
	}

	staticDir := filepath.Join(workDir, "static")

	// Check if static directory exists
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Infof("Static directory does not exist: %s", staticDir)
		return
	}

	// Create file server
	fileServer := http.FileServer(http.Dir(staticDir))

	// Add route for static files
	coreRouter.Router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	log.Infof("Static file server configured at /static/ serving from: %s", staticDir)
}
