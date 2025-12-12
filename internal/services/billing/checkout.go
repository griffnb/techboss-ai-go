package billing

import (
	"context"

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
	PromoCode string
}

func SuccessfulStripeCheckout(
	ctx context.Context,
	org *organization.Organization,
	plan *billing_plan.BillingPlan,
	checkoutValues *SuccessCheckout,
	savingUser coremodel.Model,
) error {
	log.Debugf("--------Processing successful checkout for organization %s with plan %s------------", org.ID().String(), plan.ID().String())
	subObj, err := subscription.GetByOrganizationAndPlanID(ctx, org.ID(), plan.ID())
	if err != nil {
		return err
	}

	// Create new subscription if one does not exist, webhook possibly could have come in first
	if tools.Empty(subObj) {
		subObj = subscription.New()
		subObj.Status.Set(subscription.STATUS_PENDING)
		subObj.OrganizationID.Set(org.ID())
		subObj.BillingProvider.Set(subscription.BILLING_PROVIDER_STRIPE)
		subObj.BillingCycle.Set(plan.BillingCycle.Get())
		subObj.CouponCode.Set(checkoutValues.PromoCode)
		subObj.PriceOrPlanID.Set(plan.Properties.GetI().StripePriceID)
		subObj.Amount.Set(plan.Price.Get())
	}

	subObj.BillingPlanID.Set(plan.ID())

	stripeSubscription, err := Client().GetSubscriptionByCustomer(ctx, org.StripeID.Get())
	if err != nil {
		return err
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

	org.BillingPlanID.Set(plan.ID())
	return org.Save(savingUser)
}

func ProcessStripeCancel(ctx context.Context, sub *subscription.Subscription, savingUser coremodel.Model) error {
	// Cancel it now but set end date to the current term end

	err := Client().Cancel(ctx, sub.SubscriptionID.Get())
	if err != nil {
		return err
	}

	sub.Status.Set(subscription.STATUS_CANCELING)
	sub.EndTS.Set(sub.NextBillingTS.Get())
	return sub.Save(savingUser)
}

// TODO
func ProcessStripeResume(ctx context.Context, sub *subscription.Subscription, savingUser coremodel.Model) error {
	err := Client().Resume(ctx, sub.SubscriptionID.Get())
	if err != nil {
		return err
	}

	sub.Status.Set(subscription.STATUS_ACTIVE)
	return sub.Save(savingUser)
}

func ProcessStripePlanChange(
	ctx context.Context,
	org *organization.Organization,
	currentSub *subscription.Subscription,
	newPlan *billing_plan.BillingPlan,
	savingUser coremodel.Model,
) error {
	stripeSub, err := Client().GetSubscriptionByID(ctx, currentSub.SubscriptionID.Get())
	if err != nil {
		return err
	}

	itemID := stripeSub.Items.Data[0].ID

	err = Client().Change(ctx, currentSub.SubscriptionID.Get(), itemID, newPlan.Properties.GetI().StripePriceID)
	if err != nil {
		return err
	}

	// Disable current subscription and create a new one
	currentSub.Status.Set(subscription.STATUS_DISABLED)
	err = currentSub.Save(savingUser)
	if err != nil {
		return err
	}

	// Create new subscription record to track new plan
	newSub := subscription.New()
	newSub.OrganizationID.Set(currentSub.OrganizationID.Get())
	newSub.BillingProvider.Set(subscription.BILLING_PROVIDER_STRIPE)
	newSub.BillingCycle.Set(newPlan.BillingCycle.Get())
	newSub.SubscriptionID.Set(currentSub.SubscriptionID.Get())
	newSub.Status.Set(subscription.STATUS_ACTIVE)
	newSub.PriceOrPlanID.Set(newPlan.Properties.GetI().StripePriceID)
	newSub.Amount.Set(newPlan.Price.Get())
	newSub.BillingPlanID.Set(newPlan.ID())

	err = newSub.Save(savingUser)
	if err != nil {
		return err
	}

	org.BillingPlanID.Set(newPlan.ID())
	err = org.Save(savingUser)
	if err != nil {
		return err
	}

	return nil
}
