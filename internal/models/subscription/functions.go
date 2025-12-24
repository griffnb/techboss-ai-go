package subscription

import (
	"github.com/griffnb/core/lib/tools"
	"github.com/stripe/stripe-go/v83"
)

// This file contains additional helper functions for the Subscription model

func (this *Subscription) ProcessActive(sub *stripe.Subscription) error {
	if !tools.Empty(sub.CanceledAt) {
		this.Status.Set(STATUS_CANCELING)
		this.EndTS.Set(sub.CancelAt)
	} else {
		this.Status.Set(STATUS_ACTIVE)
	}

	return nil
}

func (this *Subscription) ProcessCanceled(_ *stripe.Subscription) error {
	this.Status.Set(STATUS_CANCELLED)
	return nil
}

func (this *Subscription) ProcessPaused(_ *stripe.Subscription) error {
	this.Status.Set(STATUS_CANCELLED)
	return nil
}

func (this *Subscription) ProcessUnpaid(_ *stripe.Subscription) error {
	this.Status.Set(STATUS_UNPAID_CANCELED)
	return nil
}

func (this *Subscription) ProcessTrialStarted(sub *stripe.Subscription) error {
	if !tools.Empty(sub.CanceledAt) {
		this.Status.Set(STATUS_CANCELING)
		this.EndTS.Set(sub.CancelAt)
	} else {
		this.Status.Set(STATUS_TRIALING)
	}
	return nil
}
