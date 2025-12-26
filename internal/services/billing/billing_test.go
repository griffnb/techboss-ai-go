package billing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"
	"github.com/stripe/stripe-go/v83"
)

const (
	TEST_PRODUCT_ID       = "prod_Tb8fWVKj6Te8Fv"            // Starter plan
	TEST_YEARLY_PRICE_ID  = "price_1SdwchQhrGqsVrZnrRKBHunY" // Yearly
	TEST_MONTHLY_PRICE_ID = "price_1SdwcDQhrGqsVrZnj0Mi5W3I" // Monthly
)

func init() {
	system_testing.BuildSystem()
}

func TestBilling(t *testing.T) {
	if !billing.Configured() {
		t.Skip("Skipping billing tests; Stripe not configured")
	}
	ctx := context.Background()

	org := organization.TESTCreateOrganization()
	err := org.Save(nil)
	if err != nil {
		t.Fatalf("Failed to save test organization: %v", err)
	}
	defer testtools.CleanupModel(org)

	stripeCustomer, err := billing.Client().CreateCustomer(ctx, fmt.Sprintf("%s@%s.com", tools.RandString(10), tools.RandString(5)), nil)
	if err != nil {
		t.Fatalf("Failed to create Stripe customer: %v", err)
	}
	err = billing.Client().AddTestPaymentMethod(ctx, stripeCustomer.ID)
	if err != nil {
		t.Fatalf("Failed to add test payment method: %v", err)
	}

	price := billing_plan_price.New()
	price.StripePriceID.Set(TEST_MONTHLY_PRICE_ID)
	err = price.Save(nil)
	if err != nil {
		t.Fatalf("Failed to save test billing plan price: %v", err)
	}
	defer testtools.CleanupModel(price)

	_, err = billing.SuccessfulStripeCheckout(ctx, org, price, &billing.SuccessCheckout{}, nil)
	if err != nil {
		t.Fatalf("Failed to process successful checkout: %v", err)
	}

	sub, err := billing.Client().CreateSubscriptionWithParams(ctx, &stripe.SubscriptionCreateParams{
		Customer: stripe.String(stripeCustomer.ID),
		Items: []*stripe.SubscriptionCreateItemParams{
			{
				Price:    stripe.String(TEST_MONTHLY_PRICE_ID),
				Quantity: stripe.Int64(1),
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	log.PrintEntity(sub, "Created subscription")
}
