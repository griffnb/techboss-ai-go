package testing_service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
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
func NewTestRequest[T any](method, path string, params url.Values, body any) (*TestRequest[T], error) {
	var req *http.Request
	var err error

	if params != nil {
		path = path + "?" + params.Encode()
	}

	if body != nil {
		jsonData, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return nil, errors.Wrap(marshalErr, "failed to marshal request body")
		}

		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	return &TestRequest[T]{
		Request:  req,
		Recorder: httptest.NewRecorder(),
	}, err
}

// NewGETRequest creates a GET request
func NewGETRequest[T any](path string, params url.Values) (*TestRequest[T], error) {
	return NewTestRequest[T]("GET", path, params, nil)
}

// NewPOSTRequest creates a POST request with JSON body
func NewPOSTRequest[T any](path string, params url.Values, body any) (*TestRequest[T], error) {
	return NewTestRequest[T]("POST", path, params, body)
}

// NewPUTRequest creates a PUT request with JSON body
func NewPUTRequest[T any](path string, params url.Values, body any) (*TestRequest[T], error) {
	return NewTestRequest[T]("PUT", path, params, body)
}

// NewDELETERequest creates a DELETE request
func NewDELETERequest[T any](path string, params url.Values) (*TestRequest[T], error) {
	return NewTestRequest[T]("DELETE", path, params, nil)
}

// Do executes the request against a local handler (for unit tests)
func (this *TestRequest[T]) Do(handler func(res http.ResponseWriter, req *http.Request) (T, int, error)) (T, int, error) {
	if this.Recorder == nil {
		this.Recorder = httptest.NewRecorder()
	}

	return handler(this.Recorder, this.Request)
}

// DoRaw executes the request against a raw void handler that writes directly to ResponseWriter.
// Used for handlers like file download endpoints that return raw content instead of JSON.
// Returns the ResponseRecorder so you can inspect headers, status code, and body.
func (this *TestRequest[T]) DoRaw(handler func(w http.ResponseWriter, req *http.Request)) (*httptest.ResponseRecorder, error) {
	if this.Recorder == nil {
		this.Recorder = httptest.NewRecorder()
	}

	handler(this.Recorder, this.Request)
	return this.Recorder, nil
}

func (this *TestRequest[T]) WithAdmin(adminObj ...*admin.Admin) error {
	if len(adminObj) == 0 {
		adminObj = append(adminObj, admin.New())
		adminObj[0].Role.Set(constants.ROLE_ADMIN)
		adminObj[0].Email.Set(tools.RandString(10) + "@example.com")
		err := adminObj[0].Save(nil)
		if err != nil {
			return err
		}
	}

	this.Admin = adminObj[0]

	// create a session and set its value as the same as the token
	userSession := session.New("").WithUser(adminObj[0])
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

func (this *TestRequest[T]) WithAccount(accountObj ...*account.Account) error {
	if len(accountObj) == 0 {
		accountObj = append(accountObj, account.New())
		accountObj[0].Role.Set(constants.ROLE_ORG_OWNER)
		accountObj[0].Email.Set(tools.RandString(10) + "@atlas.net")
		err := accountObj[0].Save(nil)
		if err != nil {
			return err
		}
	}

	this.Account = accountObj[0]

	// create a session and set its value as the same as the token
	userSession := session.New("").WithUser(accountObj[0])
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
