.PHONY: stripe-login
stripe-login: ## Login to Stripe
	stripe login

.PHONY: stripe-webhook
stripe-webhook: ## Start Stripe webhook listener
	stripe listen --forward-to http://localhost:8080/billing/stripe/webhook