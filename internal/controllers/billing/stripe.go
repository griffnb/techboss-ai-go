package billing

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router/request"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/stripe_wrapper"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"

	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
)

type CheckoutSuccessPost struct {
	HostedPageID   string `json:"hosted_page_id"`
	SubscriptionID string `json:"subscription_id"`
	PromoCode      string `json:"promo_code"`
}

func authStripeCheckoutSuccess(_ http.ResponseWriter, req *http.Request) (bool, int, error) {
	/*
		userSession := request.GetReqSession(req)

		checkoutSuccess := &CheckoutSuccessPost{}
		err := request.GetJSONPostDataStruct(req, checkoutSuccess)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[bool]()
		}

		accountObj, err := account.Get(req.Context(), userSession.User.ID())
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[bool]()
		}
		currentPlan, err := billing_plan.GetJoinedPlanForFamily(req.Context(), accountObj.FamilyID.Get())
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[bool]()
		}

		if !tools.Empty(checkoutSuccess.SubscriptionID) {
			err = billing.ProcessStripeSuccessCheckout(req.Context(), accountObj, &plan.OrganizationSubscriptionPlan, &billing.SuccessCheckout{
				SubscriptionID: checkoutSuccess.SubscriptionID,
				PromoCode:      checkoutSuccess.PromoCode,
			}, userSession.User)
			if err != nil {
				log.ErrorContext(err, req.Context())
				return response.PublicBadRequestError[bool]()
			}
		} else if !tools.Empty(checkoutSuccess.HostedPageID) {

			err = billing.ProcessChargebeeSuccessCheckout(req.Context(), &accountObj.Account, checkoutSuccess.HostedPageID, types.UUID(checkoutSuccess.OrganizationPlanID), userSession.User)
			if err != nil {
				log.ErrorContext(err, req.Context())
				return response.PublicBadRequestError[bool]()
			}
		}
	*/

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

	log.Debugf("Subscription Created With ID \n\n%s\n\n", checkoutProps.SubscriptionID)

	return response.Success(checkoutProps)
}
