package accounts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/account_service"
)

func authDelete(_ http.ResponseWriter, req *http.Request) (*account.AccountJoined, int, error) {
	//if helpers.IsSuperUpdate(req) {
	//	return helpers.PublicCustomErrorV2[*account.AccountJoinedSession]("not allowed to update as super user", http.StatusBadRequest)
	//}
	userSession := request.GetReqSession(req)
	id := chi.URLParam(req, "id")

	accountObj, err := account.GetRestrictedJoined(
		req.Context(),
		types.UUID(id),
		userSession.User,
	)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}

	accountObj.Status.Set(account.STATUS_USER_DELETED)
	err = accountObj.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}

	return response.Success[*account.AccountJoined](accountObj)
}

func authResendInvite(_ http.ResponseWriter, req *http.Request) (*account.AccountJoined, int, error) {
	//if helpers.IsSuperUpdate(req) {
	//	return helpers.PublicCustomErrorV2[*account.AccountJoinedSession]("not allowed to update as super user", http.StatusBadRequest)
	//}
	userSession := request.GetReqSession(req)
	id := chi.URLParam(req, "id")

	primaryAccount := helpers.GetLoadedUser(req)

	accountToInvite, err := account.GetRestrictedJoined(
		req.Context(),
		types.UUID(id),
		userSession.User,
	)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()

	}

	err = account_service.SendInviteEmail(req.Context(), &accountToInvite.Account, &primaryAccount.Account, userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}

	if accountToInvite.Status.Get() == account.STATUS_PENDING_INVITE {
		accountToInvite.Status.Set(account.STATUS_PENDING_EMAIL_VERIFICATION)
		err = accountToInvite.Save(userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*account.AccountJoined]()
		}
	}

	return response.Success(accountToInvite)
}

func authCancelInvite(_ http.ResponseWriter, req *http.Request) (*account.AccountJoined, int, error) {
	//if helpers.IsSuperUpdate(req) {
	//	return helpers.PublicCustomErrorV2[*account.AccountJoined]("not allowed to update as super user", http.StatusBadRequest)
	//}
	userSession := request.GetReqSession(req)
	id := chi.URLParam(req, "id")
	invitedAccount, err := account.GetRestrictedJoined(
		req.Context(),
		types.UUID(id),
		userSession.User,
	)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()

	}

	props, err := invitedAccount.Properties.Get()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}

	if !tools.Empty(props.InviteKey) {
		inviteSession := session.Load(props.InviteKey)
		if !tools.Empty(inviteSession) {
			err := inviteSession.Invalidate()
			log.ErrorContext(err, req.Context())
		}
	}

	invitedAccount.Status.Set(account.STATUS_PENDING_INVITE)
	err = invitedAccount.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}

	return response.Success(invitedAccount)
}
