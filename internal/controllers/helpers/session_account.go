package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

func GetLoadedUser(req *http.Request) *account.Account {
	userSession := request.GetReqSession(req)
	return userSession.LoadedUser.(*account.Account)
}

func getAccountSession(req *http.Request) (*session.Session, error) {
	accountSession := getCustomAccountSession(req)
	if !tools.Empty(accountSession) {
		return accountSession, nil
	}

	return nil, nil
}

func loadAccount(req *http.Request, accountSession *session.Session) (*account.Account, error) {
	// TODO might need to cache this
	accnt, err := account.Get(req.Context(), accountSession.User.ID())
	if err != nil {
		return nil, err
	}

	return accnt, nil
}
