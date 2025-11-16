package billing

import (
	"context"

	"github.com/CrowdShield/go-core/lib/stripe_wrapper"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/stripe/stripe-go/v83"
)

type WebhookEventService struct{}

func (this *WebhookEventService) GetBySubscriptionID(ctx context.Context, subscriptionID string) (stripe_wrapper.SubscriptionObject, error) {
	return nil, nil
}

func ProcessStripeEvent(ctx context.Context, event stripe.Event) error {
	service := &WebhookEventService{}
	sub, stripeSub, err := Client().ProcessStripeEvent(ctx, event, service)
	if err != nil {
		return err
	}

	subObj := sub.(*subscription.Subscription)

	if subObj.IsNew() {
		err := mergeBillingInfo(subObj, stripeSub)
		if err != nil {
			return err
		}
	}

	return subObj.Save(nil)
}

func mergeBillingInfo(subObj *subscription.Subscription, stripeSub *stripe.Subscription) error {
	subInfo, err := Client().GetSubscriptionInfo(stripeSub)
	if err != nil {
		return err
	}

	if subInfo.BillingInfo != nil {
		subObj.BillingInfo.Set(&subscription.BillingInfo{
			CardType:     subInfo.BillingInfo.CardType,
			CardLast4:    subInfo.BillingInfo.CardLast4,
			CardExpMonth: subInfo.BillingInfo.CardExpMonth,
			CardExpYear:  subInfo.BillingInfo.CardExpYear,
			CardAddress1: subInfo.BillingInfo.CardAddress1,
			CardAddress2: subInfo.BillingInfo.CardAddress2,
			CardCity:     subInfo.BillingInfo.CardCity,
			CardState:    subInfo.BillingInfo.CardState,
			CardZip:      subInfo.BillingInfo.CardZip,
			CardCountry:  subInfo.BillingInfo.CardCountry,
		})
	}

	return nil
}
