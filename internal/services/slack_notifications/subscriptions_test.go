package slacknotifications

import (
	"context"
	"testing"

	"github.com/CrowdShield/go-core/lib/testtools"
	"github.com/CrowdShield/go-core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/griffnb/techboss-ai-go/internal/services/testing_service"
	"github.com/shopspring/decimal"
)

func init() {
	system_testing.BuildSystem()
}

func TestSubscriptionStarted(t *testing.T) {
	t.Run("successful notification with complete subscription data", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}
		orgJoined.Name.Set("Acme Corporation")

		sub := subscription.New()
		sub.SubscriptionID.Set("sub_123abc456def")
		sub.BillingCycle.Set(subscription.BILLING_CYCLE_ANNUALLY)
		sub.Amount.Set(decimal.NewFromFloat(99.99))
		sub.PriceOrPlanID.Set("price_premium_annual")

		// Act
		err = SubscriptionStarted(context.Background(), orgJoined, sub)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("successful notification with monthly billing", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}
		orgJoined.Name.Set("Test Company")

		sub := subscription.New()
		sub.SubscriptionID.Set("sub_monthly123")
		sub.BillingCycle.Set(subscription.BILLING_CYCLE_MONTHLY)
		sub.Amount.Set(decimal.NewFromFloat(9.99))
		sub.PriceOrPlanID.Set("price_basic_monthly")

		// Act
		err = SubscriptionStarted(context.Background(), orgJoined, sub)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("handles nil organization gracefully", func(t *testing.T) {
		// Arrange
		sub := subscription.New()
		sub.SubscriptionID.Set("sub_test")

		// Act
		err := SubscriptionStarted(context.Background(), nil, sub)

		// Assert - should not panic and should return error
		assert.NEmpty(t, err)
	})

	t.Run("handles nil subscription gracefully", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}

		// Act
		err = SubscriptionStarted(context.Background(), orgJoined, nil)

		// Assert - should not panic and should return error
		assert.NEmpty(t, err)
	})
}

func TestSubscriptionCanceled(t *testing.T) {
	t.Run("successful cancellation notification", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}
		orgJoined.Name.Set("Canceling Company")

		sub := subscription.New()
		sub.SubscriptionID.Set("sub_cancel123")
		sub.BillingCycle.Set(subscription.BILLING_CYCLE_ANNUALLY)
		sub.Amount.Set(decimal.NewFromFloat(99.99))
		sub.PriceOrPlanID.Set("price_premium_annual")
		sub.NextBillingTS.Set(1704067200) // Some future timestamp

		// Act
		err = SubscriptionCanceled(context.Background(), orgJoined, sub)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("handles nil organization gracefully", func(t *testing.T) {
		// Arrange
		sub := subscription.New()
		sub.SubscriptionID.Set("sub_test")

		// Act
		err := SubscriptionCanceled(context.Background(), nil, sub)

		// Assert
		assert.NEmpty(t, err)
	})

	t.Run("handles nil subscription gracefully", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}

		// Act
		err = SubscriptionCanceled(context.Background(), orgJoined, nil)

		// Assert
		assert.NEmpty(t, err)
	})
}

func TestSubscriptionResumed(t *testing.T) {
	t.Run("successful resume notification", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}
		orgJoined.Name.Set("Resuming Company")

		sub := subscription.New()
		sub.SubscriptionID.Set("sub_resume123")
		sub.BillingCycle.Set(subscription.BILLING_CYCLE_MONTHLY)
		sub.Amount.Set(decimal.NewFromFloat(19.99))
		sub.PriceOrPlanID.Set("price_pro_monthly")

		// Act
		err = SubscriptionResumed(context.Background(), orgJoined, sub)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("handles nil organization gracefully", func(t *testing.T) {
		// Arrange
		sub := subscription.New()
		sub.SubscriptionID.Set("sub_test")

		// Act
		err := SubscriptionResumed(context.Background(), nil, sub)

		// Assert
		assert.NEmpty(t, err)
	})

	t.Run("handles nil subscription gracefully", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithOrganization()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// nolint:govet // Intentionally copying for test purposes
		orgJoined := &organization.OrganizationJoined{
			Organization: *builder.Organization,
		}

		// Act
		err = SubscriptionResumed(context.Background(), orgJoined, nil)

		// Assert
		assert.NEmpty(t, err)
	})
}
