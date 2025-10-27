package helpers

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/pkg/errors"
)

func loadAccount(req *http.Request, accountSession *session.Session) (*account.Account, error) {
	// TODO might need to cache this
	accnt, err := account.Get(req.Context(), accountSession.User.ID())
	if err != nil {
		return nil, err
	}

	if tools.Empty(accnt) {
		return nil, errors.Errorf("userid not found %s", accountSession.User.ID())
	}

	return accnt, nil
}
