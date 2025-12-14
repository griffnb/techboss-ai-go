package billing

import (
	"context"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/stripe_wrapper"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v83"
)

func StripeCheckout(
	ctx context.Context,
	org *organization.Organization,
	planPrice *billing_plan_price.BillingPlanPrice,
	promoCodeIDs *stripe_wrapper.StripeCodes,
) (*stripe_wrapper.StripeCheckout, error) {
	customer := &stripe_wrapper.Customer{
		ReturnURL: environment.GetConfig().Server.AppURL,
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

		if !environment.IsProduction() {
			err = Client().AddTestPaymentMethod(ctx, stripeCustomer.ID)
			if err != nil {
				return nil, err
			}
			log.Debugf("Added test payment method to Stripe customer %s", stripeCustomer.ID)
		}

	} else {
		customer.ID = org.StripeID.Get()
	}

	return Client().SetupStripeCheckoutSession(ctx, planPrice.StripePriceID.Get(), customer, promoCodeIDs)
}

type SuccessCheckout struct {
	PromoCode string
}

func SuccessfulStripeCheckout(
	ctx context.Context,
	org *organization.Organization,
	planPrice *billing_plan_price.BillingPlanPrice,
	checkoutValues *SuccessCheckout,
	savingUser coremodel.Model,
) (*subscription.Subscription, error) {
	log.Debugf("--------Processing successful checkout for organization %s with plan %s------------", org.ID().String(), planPrice.ID().String())
	subObj, err := subscription.GetByOrganizationAndPlanPriceID(ctx, org.ID(), planPrice.ID())
	if err != nil {
		return nil, err
	}

	// Create new subscription if one does not exist, webhook possibly could have come in first
	if tools.Empty(subObj) {
		subObj = subscription.New()
		subObj.Status.Set(subscription.STATUS_PENDING)
		subObj.OrganizationID.Set(org.ID())
		subObj.BillingProvider.Set(subscription.BILLING_PROVIDER_STRIPE)
		subObj.BillingCycle.Set(planPrice.BillingCycle.Get())
		subObj.CouponCode.Set(checkoutValues.PromoCode)
		subObj.StripePriceID.Set(planPrice.StripePriceID.Get())
		subObj.StripeCustomerID.Set(org.StripeID.Get())
		subObj.Amount.Set(planPrice.Price.Get())
	}

	subObj.BillingPlanPriceID.Set(planPrice.ID())
	stripeSubscription, err := Client().GetSubscriptionByCustomer(ctx, org.StripeID.Get())
	if err != nil {
		return nil, err
	}

	log.Debugf("--------Subscription Status on success call------------ %s", stripeSubscription.Status)

	if tools.Empty(stripeSubscription) {
		return nil, errors.Errorf("could not find stripe subscription for organization %s", org.ID().String())
	}

	if subObj.StripeSubscriptionID.IsEmpty() {
		subObj.StripeSubscriptionID.Set(stripeSubscription.ID)
	}

	if !tools.Empty(stripeSubscription) && stripeSubscription.Status == stripe.SubscriptionStatusActive {
		err := mergeBillingInfo(ctx, subObj, stripeSubscription)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	org.BillingPlanPriceID.Set(planPrice.ID())
	err = org.Save(savingUser)
	if err != nil {
		return nil, err
	}

	return subObj, nil
}

func ProcessStripeCancel(ctx context.Context, sub *subscription.Subscription, savingUser coremodel.Model) error {
	// Cancel it now but set end date to the current term end

	err := Client().Cancel(ctx, sub.StripeSubscriptionID.Get())
	if err != nil {
		return err
	}

	sub.Status.Set(subscription.STATUS_CANCELING)
	sub.EndTS.Set(sub.NextBillingTS.Get())
	return sub.Save(savingUser)
}

// TODO
func ProcessStripeResume(ctx context.Context, sub *subscription.Subscription, savingUser coremodel.Model) error {
	err := Client().Resume(ctx, sub.StripeSubscriptionID.Get())
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
	newPlanPrice *billing_plan_price.BillingPlanPrice,
	savingUser coremodel.Model,
) error {
	stripeSub, err := Client().GetSubscriptionByID(ctx, currentSub.StripeSubscriptionID.Get())
	if err != nil {
		return err
	}

	itemID := stripeSub.Items.Data[0].ID

	err = Client().Change(ctx, currentSub.StripeSubscriptionID.Get(), itemID, newPlanPrice.StripePriceID.Get(), true)
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
	newSub.BillingCycle.Set(newPlanPrice.BillingCycle.Get())
	newSub.StripeSubscriptionID.Set(currentSub.StripeSubscriptionID.Get())
	newSub.Status.Set(subscription.STATUS_ACTIVE)
	newSub.StripePriceID.Set(newPlanPrice.StripePriceID.Get())
	newSub.StripeCustomerID.Set(currentSub.StripeCustomerID.Get())
	newSub.Amount.Set(newPlanPrice.Price.Get())
	newSub.BillingPlanPriceID.Set(newPlanPrice.ID())
	newSub.BillingInfo.Set(currentSub.BillingInfo.GetI())

	err = newSub.Save(savingUser)
	if err != nil {
		return err
	}

	org.BillingPlanPriceID.Set(newPlanPrice.ID())
	err = org.Save(savingUser)
	if err != nil {
		return err
	}

	return nil
}
