package github_installations

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/github_installation"
	"github.com/griffnb/techboss-ai-go/internal/services/github_service"
	"github.com/pkg/errors"
)

// WebhookPayload represents the GitHub installation webhook payload
type WebhookPayload struct {
	Action       string               `json:"action"`
	Installation *WebhookInstallation `json:"installation"`
	Repositories []WebhookRepository  `json:"repositories"`
	Sender       *WebhookSender       `json:"sender"`
}

// WebhookInstallation represents the installation object in the webhook
type WebhookInstallation struct {
	ID                  string          `json:"id"`
	Account             *WebhookAccount `json:"account"`
	RepositorySelection string          `json:"repository_selection"`
	Permissions         map[string]any  `json:"permissions"`
	AppSlug             string          `json:"app_slug"`
	SuspendedAt         *string         `json:"suspended_at"`
}

// WebhookAccount represents the account object in the installation
type WebhookAccount struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"`
}

// WebhookRepository represents a repository in the webhook
type WebhookRepository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// WebhookSender represents the sender of the webhook
type WebhookSender struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

// WebhookCallback handles GitHub installation webhooks
// @Public
// @Summary GitHub Installation Webhook
// @Description Processes GitHub App installation webhooks (created, deleted, suspend, unsuspend)
// @Tags GitHub
// @Accept json
// @Produce json
// @Param X-Hub-Signature-256 header string true "GitHub webhook signature"
// @Success 200 {object} response.JSONDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /admin/github_installation/callback [post]
func WebhookCallback(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.ErrorContext(err, ctx)
		response.ErrorWrapper(res, req, "failed to read request body", http.StatusBadRequest)
		return
	}

	// Get webhook signature from header
	signature := req.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		response.ErrorWrapper(res, req, "missing webhook signature", http.StatusUnauthorized)
		return
	}

	// Get webhook secret from environment config
	webhookSecret := environment.GetConfig().Github.WebhookSecret
	if webhookSecret == "" {
		log.ErrorContext(errors.New("GitHub webhook secret not configured"), ctx)
		response.ErrorWrapper(res, req, "webhook secret not configured", http.StatusInternalServerError)
		return
	}

	// Validate webhook signature
	if !github_service.ValidateWebhookSignature(body, signature, webhookSecret) {
		response.ErrorWrapper(res, req, "invalid webhook signature", http.StatusUnauthorized)
		return
	}

	// Parse JSON payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.ErrorContext(err, ctx)
		response.ErrorWrapper(res, req, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.Installation == nil {
		response.ErrorWrapper(res, req, "missing installation data", http.StatusBadRequest)
		return
	}

	// Handle webhook events
	switch payload.Action {
	case "created":
		err = handleInstallationCreated(ctx, &payload)
	case "deleted":
		err = handleInstallationDeleted(ctx, &payload)
	case "suspend":
		err = handleInstallationSuspended(ctx, &payload)
	case "unsuspend":
		err = handleInstallationUnsuspended(ctx, &payload)
	default:
		// Unknown action - log but don't error
		log.Info("unknown webhook action: " + payload.Action)
	}

	if err != nil {
		log.ErrorContext(err, ctx)
		response.ErrorWrapper(res, req, "failed to process webhook", http.StatusInternalServerError)
		return
	}

	response.JSONDataResponseWrapper(res, req, "success")
}

// handleInstallationCreated creates a new GithubInstallation record
func handleInstallationCreated(ctx context.Context, payload *WebhookPayload) error {
	installation := github_installation.New()

	installation.InstallationID.Set(payload.Installation.ID)
	installation.GithubAccountID.Set(tools.ParseStringI(payload.Installation.Account.ID))
	installation.GithubAccountName.Set(payload.Installation.Account.Login)
	installation.RepositoryAccess.Set(payload.Installation.RepositorySelection)
	installation.Permissions.Set(payload.Installation.Permissions)
	installation.AppSlug.Set(payload.Installation.AppSlug)
	installation.Suspended.Set(0)

	// Set account type based on GitHub account type
	if payload.Installation.Account.Type == "Organization" {
		installation.GithubAccountType.Set(github_installation.ACCOUNT_TYPE_ORGANIZATION)
	} else {
		installation.GithubAccountType.Set(github_installation.ACCOUNT_TYPE_USER)
	}

	return installation.SaveWithContext(ctx, nil)
}

// handleInstallationDeleted marks a GithubInstallation as deleted
func handleInstallationDeleted(ctx context.Context, payload *WebhookPayload) error {
	installation, err := github_installation.GetByInstallationID(ctx, payload.Installation.ID)
	if err != nil {
		return err
	}

	// Set status to deleted - the beforeSave hook will automatically set Deleted=1
	installation.Status.Set(constants.STATUS_DELETED)
	return installation.SaveWithContext(ctx, nil)
}

// handleInstallationSuspended sets the suspended flag
func handleInstallationSuspended(ctx context.Context, payload *WebhookPayload) error {
	installation, err := github_installation.GetByInstallationID(ctx, payload.Installation.ID)
	if err != nil {
		return err
	}

	installation.Suspended.Set(1)
	return installation.SaveWithContext(ctx, nil)
}

// handleInstallationUnsuspended clears the suspended flag
func handleInstallationUnsuspended(ctx context.Context, payload *WebhookPayload) error {
	installation, err := github_installation.GetByInstallationID(ctx, payload.Installation.ID)
	if err != nil {
		return err
	}

	installation.Suspended.Set(0)
	return installation.SaveWithContext(ctx, nil)
}
