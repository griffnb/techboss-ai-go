package slacknotifications

import (
	"context"
	"fmt"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/slack"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/pkg/errors"
	slackapi "github.com/slack-go/slack"
)

const (
	SUBSCRIPTIONS_CHANNEL = "subscriptions"
)

// SubscriptionStarted sends a Slack notification when a new subscription is started
func SubscriptionStarted(_ context.Context, org *organization.OrganizationJoined, sub *subscription.Subscription) error {
	if tools.Empty(org) {
		return errors.Errorf("organization is required")
	}

	if tools.Empty(sub) {
		return errors.Errorf("subscription is required")
	}

	webhookURL := environment.GetConfig().Slack.APIURL
	if tools.Empty(webhookURL) {
		log.Debug("Slack webhook URL not configured, skipping subscription started notification")
		return nil
	}

	client := slack.NewWebhook(webhookURL)

	// Format the billing cycle
	billingCycle := formatBillingCycle(sub.BillingCycle.Get())

	// Format the amount
	amount := sub.Amount.Get().StringFixed(2)

	// Create the message
	msg := &slackapi.Message{}
	msg.Text = "New Subscription Started"

	// Add rich attachment with subscription details
	attachment := slackapi.Attachment{
		Color:      "good",
		Title:      "🎉 New Subscription Started",
		AuthorName: org.Name.Get(),
		Fields: []slackapi.AttachmentField{
			{
				Title: "Organization",
				Value: org.Name.Get(),
				Short: true,
			},
			{
				Title: "Plan ID",
				Value: sub.PriceOrPlanID.Get(),
				Short: true,
			},
			{
				Title: "Amount",
				Value: fmt.Sprintf("$%s/%s", amount, billingCycle),
				Short: true,
			},
			{
				Title: "Subscription ID",
				Value: sub.SubscriptionID.Get(),
				Short: true,
			},
		},
		Footer: "Asset Trading Desk Billing",
	}

	slack.AddAttachment(msg, attachment)

	// Send the message and log any errors (don't fail if Slack is down)
	if err := client.Send(msg); err != nil {
		log.Error(errors.WithStack(err), "Failed to send Slack notification for subscription started")
		return errors.WithStack(err)
	}

	return nil
}

// SubscriptionCanceled sends a Slack notification when a subscription is canceled
func SubscriptionCanceled(_ context.Context, org *organization.OrganizationJoined, sub *subscription.Subscription) error {
	if tools.Empty(org) {
		return errors.Errorf("organization is required")
	}

	if tools.Empty(sub) {
		return errors.Errorf("subscription is required")
	}

	webhookURL := environment.GetConfig().Slack.APIURL
	if tools.Empty(webhookURL) {
		log.Debug("Slack webhook URL not configured, skipping subscription canceled notification")
		return nil
	}

	client := slack.NewWebhook(webhookURL)

	// Format the billing cycle
	billingCycle := formatBillingCycle(sub.BillingCycle.Get())

	// Format the amount
	amount := sub.Amount.Get().StringFixed(2)

	// Format the end date
	endDate := "N/A"
	if sub.NextBillingTS.Get() > 0 {
		endDate = time.Unix(sub.NextBillingTS.Get(), 0).Format("2006-01-02")
	}

	// Create the message
	msg := &slackapi.Message{}
	msg.Text = "Subscription Canceled"

	// Add rich attachment with subscription details
	attachment := slackapi.Attachment{
		Color:      "warning",
		Title:      "⚠️ Subscription Canceled",
		AuthorName: org.Name.Get(),
		Fields: []slackapi.AttachmentField{
			{
				Title: "Organization",
				Value: org.Name.Get(),
				Short: true,
			},
			{
				Title: "Plan ID",
				Value: sub.PriceOrPlanID.Get(),
				Short: true,
			},
			{
				Title: "Amount",
				Value: fmt.Sprintf("$%s/%s", amount, billingCycle),
				Short: true,
			},
			{
				Title: "Ends On",
				Value: endDate,
				Short: true,
			},
			{
				Title: "Subscription ID",
				Value: sub.SubscriptionID.Get(),
				Short: false,
			},
		},
		Footer: "Asset Trading Desk Billing",
	}

	slack.AddAttachment(msg, attachment)

	// Send the message and log any errors (don't fail if Slack is down)
	if err := client.Send(msg); err != nil {
		log.Error(errors.WithStack(err), "Failed to send Slack notification for subscription canceled")
		return errors.WithStack(err)
	}

	return nil
}

// SubscriptionResumed sends a Slack notification when a subscription is resumed
func SubscriptionResumed(_ context.Context, org *organization.OrganizationJoined, sub *subscription.Subscription) error {
	if tools.Empty(org) {
		return errors.Errorf("organization is required")
	}

	if tools.Empty(sub) {
		return errors.Errorf("subscription is required")
	}

	webhookURL := environment.GetConfig().Slack.APIURL
	if tools.Empty(webhookURL) {
		log.Debug("Slack webhook URL not configured, skipping subscription resumed notification")
		return nil
	}

	client := slack.NewWebhook(webhookURL)

	// Format the billing cycle
	billingCycle := formatBillingCycle(sub.BillingCycle.Get())

	// Format the amount
	amount := sub.Amount.Get().StringFixed(2)

	// Create the message
	msg := &slackapi.Message{}
	msg.Text = "Subscription Resumed"

	// Add rich attachment with subscription details
	attachment := slackapi.Attachment{
		Color:      "good",
		Title:      "✅ Subscription Resumed",
		AuthorName: org.Name.Get(),
		Fields: []slackapi.AttachmentField{
			{
				Title: "Organization",
				Value: org.Name.Get(),
				Short: true,
			},
			{
				Title: "Plan ID",
				Value: sub.PriceOrPlanID.Get(),
				Short: true,
			},
			{
				Title: "Amount",
				Value: fmt.Sprintf("$%s/%s", amount, billingCycle),
				Short: true,
			},
			{
				Title: "Subscription ID",
				Value: sub.SubscriptionID.Get(),
				Short: true,
			},
		},
		Footer: "Asset Trading Desk Billing",
	}

	slack.AddAttachment(msg, attachment)

	// Send the message and log any errors (don't fail if Slack is down)
	if err := client.Send(msg); err != nil {
		log.Error(errors.WithStack(err), "Failed to send Slack notification for subscription resumed")
		return errors.WithStack(err)
	}

	return nil
}

// formatBillingCycle converts the billing cycle constant to a human-readable string
func formatBillingCycle(cycle subscription.BillingCycle) string {
	switch cycle {
	case subscription.BILLING_CYCLE_MONTHLY:
		return "month"
	case subscription.BILLING_CYCLE_QUARTERLY:
		return "quarter"
	case subscription.BILLING_CYCLE_ANNUALLY:
		return "year"
	default:
		return "unknown"
	}
}
