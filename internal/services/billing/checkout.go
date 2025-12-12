package billing

import (
	"context"
	"fmt"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/stripe_wrapper"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v83"
)

func StripeCheckout(
	ctx context.Context,
	org *organization.Organization,
	plan *billing_plan.BillingPlan,
	promoCodeIDs *stripe_wrapper.StripeCodes,
) (*stripe_wrapper.StripeCheckout, error) {
	customer := &stripe_wrapper.Customer{
		ReturnURL: "http://localhost:5173",
	}
	fmt.Printf("\n\n--------Creating checkout for organization %s with plan %s------------\n\n", org.ID().String(), plan.ID().String())

	if org.StripeID.IsEmpty() {
		stripeCustomer, err := Client().CreateCustomer(ctx, org.Properties.GetI().BillingEmail, map[string]string{
			"organization_id": org.ID().String(),
		})
		if err != nil {
			return nil, err
		}
		log.Debugf("Created Stripe customer %s for organization %s", stripeCustomer.ID, org.ID().String())
		org.StripeID.Set(stripeCustomer.ID)
		err = org.Save(nil)
		if err != nil {
			return nil, err
		}
		customer.ID = stripeCustomer.ID
	} else {
		customer.ID = org.StripeID.Get()
	}

	return Client().SetupStripeCheckoutSession(ctx, plan.Properties.GetI().StripePriceID, customer, promoCodeIDs)
}

type SuccessCheckout struct {
	SubscriptionID string
	PromoCode      string
}

func SuccessfulStripeCheckout(
	ctx context.Context,
	org *organization.Organization,
	plan *billing_plan.BillingPlan,
	checkoutValues *SuccessCheckout,
	savingUser coremodel.Model,
) error {
	var subObj *subscription.Subscription
	var err error
	fmt.Printf("\n\n--------Processing successful checkout for organization %s with plan %s------------\n\n", org.ID().String(), plan.ID().String())

	if !tools.Empty(checkoutValues.SubscriptionID) {
		subObj, err = subscription.GetBySubscriptionID(ctx, checkoutValues.SubscriptionID)
		if err != nil {
			return err
		}
	} else {
		subObj, err = subscription.GetByOrganizationAndPlanID(ctx, org.ID(), plan.ID())
		if err != nil {
			return err
		}
	}

	// Possible maybe for webhook to create it first?
	if tools.Empty(subObj) {
		subObj = subscription.New()
		subObj.Status.Set(subscription.STATUS_PENDING)
		subObj.OrganizationID.Set(org.ID())
		subObj.BillingProvider.Set(subscription.BILLING_PROVIDER_STRIPE)
		subObj.BillingCycle.Set(plan.BillingCycle.Get())
		if !tools.Empty(checkoutValues.SubscriptionID) {
			subObj.SubscriptionID.Set(checkoutValues.SubscriptionID)
		}
		subObj.CouponCode.Set(checkoutValues.PromoCode)
		subObj.PriceOrPlanID.Set(plan.Properties.GetI().StripePriceID)
		subObj.Amount.Set(plan.Price.Get())
	}

	var stripeSubscription *stripe.Subscription

	subObj.BillingPlanID.Set(plan.ID())
	if !tools.Empty(checkoutValues.SubscriptionID) {
		stripeSubscription, err = Client().GetSubscriptionByID(ctx, checkoutValues.SubscriptionID)
		if err != nil {
			return err
		}
	} else {
		stripeSubscription, err = Client().GetSubscriptionByCustomer(ctx, org.StripeID.Get())
		if err != nil {
			return err
		}
	}

	log.Debugf("--------Subscription Status on success call------------ %s", stripeSubscription.Status)

	if tools.Empty(stripeSubscription) {
		return errors.Errorf("could not find stripe subscription for organization %s", org.ID().String())
	}

	if subObj.SubscriptionID.IsEmpty() {
		subObj.SubscriptionID.Set(stripeSubscription.ID)
	}

	if !tools.Empty(stripeSubscription) && stripeSubscription.Status == stripe.SubscriptionStatusActive {
		err := mergeBillingInfo(subObj, stripeSubscription)
		if err != nil {
			return err
		}
		subObj.Status.Set(subscription.STATUS_ACTIVE)

		/*
			go func() {
				err := slacknotifications.SubscriptionStarted(context.Background(), org, subObj)
				if err != nil {
					log.Error(err)
				}
			}()
		*/

	}

	err = subObj.Save(savingUser)
	if err != nil {
		return err
	}

	org.BillingPlanID.Set(subObj.BillingPlanID.Get())
	return org.Save(savingUser)
}

// TODO
func ProcessStripeCancel(_ context.Context, _ *organization.Organization, _ *subscription.Subscription, _ coremodel.Model) error {
	return nil
}

// TODO
func ProcessStripeResume(_ context.Context, _ *organization.Organization, _ *subscription.Subscription, _ coremodel.Model) error {
	return nil
}
