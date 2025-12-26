package subscription

import "github.com/griffnb/techboss-ai-go/internal/constants"

const (
	STATUS_PENDING         constants.Status = 1
	STATUS_ACTIVE          constants.Status = 100
	STATUS_TRIALING        constants.Status = 101
	STATUS_CANCELING       constants.Status = 102
	STATUS_DISABLED        constants.Status = 200
	STATUS_CANCELLED       constants.Status = 201
	STATUS_UNPAID_CANCELED constants.Status = 202
	STATUS_DELETED         constants.Status = 300
)
