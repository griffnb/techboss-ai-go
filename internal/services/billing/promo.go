package billing

import "context"

func GetPromoCodeID(ctx context.Context, promoCode string) (string, error) {
	return Client().GetPromoCodeID(ctx, promoCode)
}
