package billing

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/router/response"
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

func authStripeCheckout(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	/*
		userSession := request.GetReqSession(req)

		accountObj := helpers.GetLoadedUser(req)

		checkoutData := &Checkout{}
		err := request.GetJSONPostDataStruct(req, checkoutData)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[any]()
		}

		planID := chi.URLParam(req, "id")
		if tools.Empty(planID) {
			return response.PublicBadRequestError[any]()
		}

		var promoCodeID string
		if !tools.Empty(checkoutData.PromotionCode) {
			promoCodeID, err = billing.GetPromoCodeID(req.Context(), checkoutData.PromotionCode)
			if err != nil {
				log.ErrorContext(err, req.Context())
				return response.PublicCustomError[any]("Invalid Promo Code", http.StatusBadRequest)
			}
		}

		checkoutProps, err := billing.SetupStripeCheckout(req.Context(), &accountObj.Account, types.UUID(planID), &billing.StripeCodes{
			PromotionCodeID: promoCodeID,
		}, userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[any]()
		}

		checkoutProps.PromoCode = checkoutData.PromotionCode

		log.Debugf("Subscription Created With ID \n\n%s\n\n", checkoutProps.SubscriptionID)

		return response.Success(checkoutProps)
	*/
	return nil, 0, nil
}
