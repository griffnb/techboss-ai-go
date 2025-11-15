package billing

/*
func authCancel(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	userSession := helpers.GetReqSession(req)
	accountObj := helpers.GetLoadedUser(req)

	subscriptionInfo, err := family_subscription.GetJoinedActiveByFamilyID(req.Context(), accountObj.FamilyID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.PublicBadRequestError()
	}

	if tools.Empty(subscriptionInfo) {
		log.ErrorContext(errors.Errorf("failed to get active subscription for family %s", accountObj.FamilyID.Get()), req.Context())
		return helpers.PublicBadRequestError()
	}

		fullAccount, err := account.GetJoinedFull(req.Context(), accountObj.ID())
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.PublicBadRequestError()
		}
		err = billing.ProcessStripeCancel(req.Context(), fullAccount, &subscriptionInfo.FamilySubscription, userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.PublicBadRequestError()
		}
	}

	return helpers.Success(subscriptionInfo)
}


func authResume(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	userSession := helpers.GetReqSession(req)

	accountObj := helpers.GetLoadedUser(req)

	subscriptionInfo, err := family_subscription.GetJoinedActiveByFamilyID(req.Context(), accountObj.FamilyID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.PublicBadRequestError()
	}

	if tools.Empty(subscriptionInfo) {
		log.ErrorContext(errors.Errorf("failed to get active subscription for family %s", accountObj.FamilyID.Get()), req.Context())
		return helpers.PublicBadRequestError()
	}

	switch subscriptionInfo.BillingProvider.Get() {
	case constants.BILLING_PROVIDER_CHARGEBEE:

		err = billing.ProcessChargebeeResume(req.Context(), &subscriptionInfo.FamilySubscription, userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.PublicBadRequestError()
		}

	case constants.BILLING_PROVIDER_STRIPE:
		fullAccount, err := account.GetJoinedFull(req.Context(), accountObj.ID())
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.PublicBadRequestError()
		}
		err = billing.ProcessStripeResume(req.Context(), fullAccount, &subscriptionInfo.FamilySubscription, userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.PublicBadRequestError()
		}

	}

	return helpers.Success(subscriptionInfo)
}
/*
func authPortal(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	accountObj := helpers.GetLoadedUser(req)

	portal, err := billing.StripePortal(req.Context(), &accountObj.Account)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.PublicBadRequestError()
	}

	return helpers.Success(portal)
}
*/
