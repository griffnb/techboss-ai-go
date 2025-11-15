package system_proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/griffnb/assettradingdesk-go/internal/controllers/helpers"
	"github.com/griffnb/assettradingdesk-go/internal/environment"
	"github.com/pkg/errors"
)

const (
	PROD        = "https://api.assettradingdesk.com"
	PROD_UI_URL = "https://assettradingdesk.com"
)

type ResultType[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Error   string `json:"error"`
}

func CallProxy(sessionKeyOrEmail, method, target, path string, body io.Reader, params url.Values, headers http.Header) (*http.Response, error) {
	// Force admin here

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	fullPath := target + "/admin" + path

	// log.Debugf("Calling proxy %s", fullPath)

	req, err := http.NewRequest(method, fullPath+"?"+params.Encode(), body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header = headers
	// req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set(helpers.API_EMAIL_HEADER, sessionKeyOrEmail)
	req.Header.Set(helpers.API_KEY_HEADER, environment.GetConfig().InternalAPIKey)
	// req.Header.Set("x-domain", environment.GetOauth().GetDomain())
	// req.Header.Set("x-domain-key", tools.Sha256(fmt.Sprintf("%s:%s", environment.GetOauth().GetDomain(), "atlas")))
	// key := environment.GetConfig().Server.AdminSessionKey
	// req.Header.Set(key, sessionKey)

	fmt.Println(fullPath)
	fmt.Printf("Headers: %+v\n", req.Header)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}

/*
func post[T any](domain, method, sessionKey, path string, body []byte, params url.Values) (*ResultType[T], error) {
	resp, err := CallProxy(sessionKey, method, domain, path, bytes.NewReader(body), params, http.Header{
		"Content-Type": []string{"application/json"},
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Error calling %s: %s", path, resp.Status)
	}

	resultWrap := &ResultType[T]{}
	err = json.NewDecoder(resp.Body).Decode(resultWrap)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resultWrap, nil

}
*/

func get[T any](domain, sessionKey, path string, params url.Values) (*ResultType[T], error) {
	resp, err := CallProxy(sessionKey, http.MethodGet, domain, path, nil, params, http.Header{
		"Content-Type": []string{"application/json"},
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Error calling %s: %s", path, resp.Status)
	}

	resultWrap := &ResultType[T]{}
	err = json.NewDecoder(resp.Body).Decode(resultWrap)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resultWrap, nil
}

type tempResult struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

func getResult(domain, sessionKeyOrEmail, path string, params url.Values, objType any) (*tempResult, error) {
	resp, err := CallProxy(sessionKeyOrEmail, http.MethodGet, domain, path, nil, params, http.Header{
		"Content-Type": []string{"application/json"},
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Error calling %s: %s", path, resp.Status)
	}

	tmp := &tempResult{}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = json.Unmarshal(bytes, tmp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = json.Unmarshal(tmp.Data, objType)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return tmp, nil
}
