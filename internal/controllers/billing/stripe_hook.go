package billing

/*
// Test by activating the local cli to reroute, be sure to set the webhook key from the client
func openStripeHook(res http.ResponseWriter, req *http.Request) {
	// store event with event ID as the PK

	webHookKey := environment.GetConfig().Stripe.WebhookKey

	event, err := parseRequest(res, req, webHookKey)
	if err != nil {
		log.Error(err)
		response.ErrorWrapper(res, req, err.Error(), http.StatusBadRequest)
		return
	}

	bgContext := context.WithoutCancel(req.Context())
	go func() {
		err := billing.ProcessStripeEvent(bgContext, event)
		if err != nil {
			log.ErrorContext(err, bgContext)
			return
		}
	}()

	helpers.JSONDataResponseWrapper(res, req, "success")
}

// ParseRequest parses the request and gives back the raw stripe event
func parseRequest(w http.ResponseWriter, r *http.Request, secret string) (stripe.Event, error) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return stripe.Event{}, errors.WithStack(err)
	}
	return GetEvent(payload, r.Header.Get("Stripe-Signature"), secret)
}

// GetEvent gets the event from the payload
func getEvent(payload []byte, sigHeader, secret string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, sigHeader, secret)
	if err != nil {
		return stripe.Event{}, errors.WithStack(err)
	}

	return event, nil
}
*/
