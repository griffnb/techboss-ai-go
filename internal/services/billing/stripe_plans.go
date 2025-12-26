package billing

import (
	"context"

	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/stripe_wrapper"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/tools/ptr"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func CreateStripePrice(
	ctx context.Context,
	stripeProductID string,
	planPrice *billing_plan_price.BillingPlanPrice,
	savingUser coremodel.Model,
) error {
	if tools.Empty(stripeProductID) {
		return errors.Errorf("Cannot create Stripe price, missing product ID")
	}

	price := planPrice.Price.Get()
	priceInCents := price.Mul(decimal.NewFromInt(100)).IntPart() // Convert to cents
	stripePrice, err := Client().CreatePrice(ctx, stripeProductID, priceInCents, stripe_wrapper.RecurringInterval(planPrice.BillingCycle.Get().ToStripe()))
	if err != nil {
		return err
	}

	planPrice.StripePriceID.Set(stripePrice.ID)

	return planPrice.Save(savingUser)
}

func UpdateStripePrice(ctx context.Context, planPrice *billing_plan_price.BillingPlanPrice, savingUser coremodel.Model) error {
	priceID := planPrice.StripePriceID.Get()

	if tools.Empty(priceID) {
		return errors.Errorf("Cannot update Stripe price, missing price ID")
	}

	price := planPrice.Price.Get()
	priceInCents := price.Mul(decimal.NewFromInt(100)).IntPart() // Convert to cents
	_, err := Client().UpdatePrice(ctx, priceID, "USD", priceInCents)
	if err != nil {
		return err
	}

	return planPrice.Save(savingUser)
}

func CreateStripeProduct(ctx context.Context, plan *billing_plan.BillingPlan, savingUser coremodel.Model) error {
	stripeProduct, err := Client().CreateProduct(ctx, stripe_wrapper.ProductCreateParams{
		Name:        plan.Name.Get(),
		Description: plan.Description.Get(),
	})
	if err != nil {
		return err
	}

	plan.StripeProductID.Set(stripeProduct.ID)
	return plan.Save(savingUser)
}

func UpdateStripeProduct(ctx context.Context, plan *billing_plan.BillingPlan, savingUser coremodel.Model) error {
	productID := plan.StripeProductID.Get()
	if tools.Empty(productID) {
		return errors.Errorf("Cannot update Stripe product, missing product ID")
	}

	_, err := Client().UpdateProduct(ctx, productID, stripe_wrapper.ProductUpdateParams{
		Name:        ptr.To(plan.Name.Get()),
		Description: ptr.To(plan.Description.Get()),
	})
	if err != nil {
		return err
	}

	return plan.Save(savingUser)
}
