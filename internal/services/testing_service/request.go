package testing_service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"

	"github.com/pkg/errors"
)

type TestRequest[T any] struct {
	Request  *http.Request
	Admin    *admin.Admin
	Account  *account.Account
	Response *http.Response
	Recorder *httptest.ResponseRecorder
}

// NewTestRequest creates a new TestRequest with the given method, URL, and optional JSON body
func NewTestRequest[T any](method, url string, body any) (*TestRequest[T], error) {
	var req *http.Request
	var err error

	if body != nil {
		jsonData, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, errors.Wrap(marshalErr, "failed to marshal request body")
		}
		req = httptest.NewRequest(method, url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	return &TestRequest[T]{
		Request:  req,
		Recorder: httptest.NewRecorder(),
	}, err
}

// NewGETRequest creates a GET request
func NewGETRequest[T any](url string) (*TestRequest[T], error) {
	return NewTestRequest[T]("GET", url, nil)
}

// NewPOSTRequest creates a POST request with JSON body
func NewPOSTRequest[T any](url string, body any) (*TestRequest[T], error) {
	return NewTestRequest[T]("POST", url, body)
}

// NewPUTRequest creates a PUT request with JSON body
func NewPUTRequest[T any](url string, body any) (*TestRequest[T], error) {
	return NewTestRequest[T]("PUT", url, body)
}

// NewDELETERequest creates a DELETE request
func NewDELETERequest[T any](url string) (*TestRequest[T], error) {
	return NewTestRequest[T]("DELETE", url, nil)
}

// Do executes the request against a local handler (for unit tests)
func (this *TestRequest[T]) Do(handler response.StandardRequest[T]) (T, int, error) {
	if this.Recorder == nil {
		this.Recorder = httptest.NewRecorder()
	}

	return handler(this.Recorder, this.Request)
}

func (this *TestRequest[T]) WithAdmin(adminObj *admin.Admin) error {
	if adminObj == nil {
		adminObj = admin.New()
		adminObj.Role.Set(constants.ROLE_ADMIN)
		adminObj.Email.Set(tools.RandString(10) + "@example.com")
		err := adminObj.Save(nil)
		if err != nil {
			return err
		}
	}

	this.Admin = adminObj

	// create a session and set its value as the same as the token
	userSession := session.New("").WithUser(adminObj)
	err := userSession.Save()
	if err != nil {
		return err
	}

	// Add session to request context (for middleware/controller access)
	ctx := context.WithValue(this.Request.Context(), router.SessionContextKey("session"), userSession)
	this.Request = this.Request.WithContext(ctx)

	// Add headers or cookies if your middleware needs them
	this.Request.Header.Set(environment.GetConfig().Server.AdminSessionKey, userSession.Key)

	return nil
}

func (this *TestRequest[T]) WithAccount(accountObj *account.Account) error {
	if accountObj == nil {
		accountObj = account.New()
		accountObj.Role.Set(constants.ROLE_ADMIN)
		accountObj.Email.Set(tools.RandString(10) + "@example.com")
		err := accountObj.Save(nil)
		if err != nil {
			return err
		}
	}

	this.Account = accountObj

	// create a session and set its value as the same as the token
	userSession := session.New("").WithUser(accountObj)
	userSession.Key = "key"
	err := userSession.Save()
	if err != nil {
		return err
	}

	// Add session to request context (for middleware/controller access)
	ctx := context.WithValue(this.Request.Context(), router.SessionContextKey("session"), userSession)
	this.Request = this.Request.WithContext(ctx)

	// Add headers or cookies if your middleware needs them
	this.Request.Header.Set(environment.GetConfig().Server.SessionKey, userSession.Key)

	return nil
}

// WithQueryParam adds a query parameter to the request URL
func (this *TestRequest[T]) WithQueryParam(key, value string) *TestRequest[T] {
	q := this.Request.URL.Query()
	q.Add(key, value)
	this.Request.URL.RawQuery = q.Encode()
	return this
}

// WithHeader adds a header to the request
func (this *TestRequest[T]) WithHeader(key, value string) *TestRequest[T] {
	this.Request.Header.Set(key, value)
	return this
}
