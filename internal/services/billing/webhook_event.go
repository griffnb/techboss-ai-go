package billing

import (
	"context"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v83"
)

type WebhookEventService struct {
	Subscription *subscription.Subscription
}

// This file contains additional helper functions for the Subscription model

func (this *WebhookEventService) ProcessActive(ctx context.Context, stripeSub *stripe.Subscription) error {
	subObj, err := subscription.GetBySubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}
	if tools.Empty(subObj) {
		return errors.Errorf("subscription not found for id %s", stripeSub.ID)
	}

	this.Subscription = subObj
	return this.Subscription.ProcessActive(stripeSub)
}

func (this *WebhookEventService) ProcessCanceled(ctx context.Context, stripeSub *stripe.Subscription) error {
	subObj, err := subscription.GetBySubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}
	if tools.Empty(subObj) {
		return errors.Errorf("subscription not found for id %s", stripeSub.ID)
	}

	this.Subscription = subObj
	return this.Subscription.ProcessCanceled(stripeSub)
}

func (this *WebhookEventService) ProcessPaused(ctx context.Context, stripeSub *stripe.Subscription) error {
	subObj, err := subscription.GetBySubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}
	if tools.Empty(subObj) {
		return errors.Errorf("subscription not found for id %s", stripeSub.ID)
	}

	this.Subscription = subObj
	return this.Subscription.ProcessPaused(stripeSub)
}

func (this *WebhookEventService) ProcessUnpaid(ctx context.Context, stripeSub *stripe.Subscription) error {
	subObj, err := subscription.GetBySubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}
	if tools.Empty(subObj) {
		return errors.Errorf("subscription not found for id %s", stripeSub.ID)
	}

	this.Subscription = subObj
	err = this.Subscription.ProcessUnpaid(stripeSub)
	if err != nil {
		return err
	}
	return this.Subscription.Save(nil)
}

func (this *WebhookEventService) ProcessTrialStarted(ctx context.Context, stripeSub *stripe.Subscription) error {
	subObj, err := subscription.GetBySubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}
	if tools.Empty(subObj) {
		return errors.Errorf("subscription not found for id %s", stripeSub.ID)
	}

	this.Subscription = subObj
	return this.Subscription.ProcessTrialStarted(stripeSub)
}

func ProcessStripeEvent(ctx context.Context, event stripe.Event) error {
	service := &WebhookEventService{}
	stripeSub, err := Client().ProcessStripeEvent(ctx, event, service)
	if err != nil {
		return err
	}

	// Succeeded
	if !tools.Empty(service.Subscription) {
		subObj := service.Subscription

		if subObj.IsNew() {
			err := mergeBillingInfo(subObj, stripeSub)
			if err != nil {
				return err
			}
		}

		return subObj.Save(nil)
	}
	return nil
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
