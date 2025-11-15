package system_proxy

import (
	"net/url"
)

func ProdGet[T any](sessionKey, path string, params url.Values) (*ResultType[T], error) {
	return get[T](PROD, sessionKey, path, params)
}

/*
func ProdPost[T any](sessionKey, path string, body []byte, params url.Values) (*ResultType[T], error) {
	return post[T](PROD, http.MethodPost, sessionKey, path, body, params)
}

func ProdPut[T any](sessionKey, path string, body []byte, params url.Values) (*ResultType[T], error) {
	return post[T](PROD, http.MethodPut, sessionKey, path, body, params)
}
*/
