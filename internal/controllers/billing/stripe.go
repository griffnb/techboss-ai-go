package billing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/stripe_wrapper"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"

	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
)

type CheckoutSuccessPost struct {
	SubscriptionID string `json:"subscription_id"`
	BillingPlanID  string `json:"billing_plan_id"`
	PromoCode      string `json:"promo_code"`
}

func authStripeCheckoutSuccess(_ http.ResponseWriter, req *http.Request) (bool, int, error) {
	accountObj := helpers.GetLoadedUser(req)
	checkoutSuccess, err := request.GetJSONPostAs[*CheckoutSuccessPost](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	organizationObj, err := organization.Get(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	plan, err := billing_plan.Get(req.Context(), types.UUID(checkoutSuccess.BillingPlanID))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	err = billing.SuccessfulStripeCheckout(req.Context(), organizationObj, plan, &billing.SuccessCheckout{
		SubscriptionID: checkoutSuccess.SubscriptionID,
		PromoCode:      checkoutSuccess.PromoCode,
	}, accountObj)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	return response.Success(true)
}

type Checkout struct {
	CouponCode    string `json:"coupon_code"`
	PromotionCode string `json:"promotion_code"`
}

func authStripeCheckout(_ http.ResponseWriter, req *http.Request) (*stripe_wrapper.StripeCheckout, int, error) {
	accountObj := helpers.GetLoadedUser(req)

	if accountObj.Role.Get() < constants.ROLE_ORG_ADMIN {
		return response.Unauthorized[*stripe_wrapper.StripeCheckout]()
	}

	org, err := organization.Get(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*stripe_wrapper.StripeCheckout]()
	}

	checkoutData, err := request.GetJSONPostAs[*Checkout](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*stripe_wrapper.StripeCheckout]()
	}

	planID := chi.URLParam(req, "id")
	if tools.Empty(planID) {
		return response.PublicBadRequestError[*stripe_wrapper.StripeCheckout]()
	}

	plan, err := billing_plan.Get(req.Context(), types.UUID(planID))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*stripe_wrapper.StripeCheckout]()
	}

	if tools.Empty(plan) {
		return response.PublicBadRequestError[*stripe_wrapper.StripeCheckout]()
	}

	var promoCodeID string
	if !tools.Empty(checkoutData.PromotionCode) {
		promoCodeID, err = billing.GetPromoCodeID(req.Context(), checkoutData.PromotionCode)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicCustomError[*stripe_wrapper.StripeCheckout]("Invalid Promo Code", http.StatusBadRequest)
		}
	}

	checkoutProps, err := billing.StripeCheckout(req.Context(), org, plan, &stripe_wrapper.StripeCodes{
		PromotionCodeID: promoCodeID,
	})
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*stripe_wrapper.StripeCheckout]()
	}

	checkoutProps.PromoCode = checkoutData.PromotionCode

	return response.Success(checkoutProps)
}
