package subscription

import (
	"github.com/stripe/stripe-go/v83"
)

// This file contains additional helper functions for the Subscription model

func (this *Subscription) ProcessActive(stripeSub *stripe.Subscription) error {
	this.Status.Set(STATUS_ACTIVE)
	return nil
}

func (this *Subscription) ProcessCanceled(stripeSub *stripe.Subscription) error {
	this.Status.Set(STATUS_CANCELLED)
	return nil
}

func (this *Subscription) ProcessPaused(stripeSub *stripe.Subscription) error {
	this.Status.Set(STATUS_CANCELLED)
	return nil
}

func (this *Subscription) ProcessUnpaid(stripeSub *stripe.Subscription) error {
	this.Status.Set(STATUS_UNPAID_CANCELED)
	return nil
}

func (this *Subscription) ProcessTrialStarted(stripeSub *stripe.Subscription) error {
	this.Status.Set(STATUS_ACTIVE)
	return nil
}
