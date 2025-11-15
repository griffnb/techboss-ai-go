package system_proxy

import (
	"net/url"
)

func RemoteGet[T any](sessionKey, path string, params url.Values) (*ResultType[T], error) {
	//if environment.IsProduction() {
	//	return get[T](STAGE, sessionKey, path, params)
	//}
	return get[T](PROD, sessionKey, path, params)
}

func RemoteGetType(sessionKey, path string, params url.Values, objType any) (*tempResult, error) {
	//if environment.IsProduction() {
	//	return getResult(STAGE, sessionKey, path, params, objType)
	//}
	return getResult(PROD, sessionKey, path, params, objType)
}

/*
func RemotePost[T any](sessionKey, path string, body []byte, params url.Values) (*ResultType[T], error) {
	if environment.IsProduction() || environment.IsLocalDev() {
		return post[T](STAGE, sessionKey, path, path, body, params)
	}
	return post[T](PROD, http.MethodPost, sessionKey, path, body, params)
}

func RemotePut[T any](sessionKey, path string, body []byte, params url.Values) (*ResultType[T], error) {
	if environment.IsProduction() || environment.IsLocalDev() {
		return post[T](STAGE, http.MethodPut, sessionKey, path, body, params)
	}
	return post[T](PROD, http.MethodPut, sessionKey, path, body, params)
}
*/
