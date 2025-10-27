package helpers

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/std_errors"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/maps"
	"github.com/CrowdShield/go-core/lib/tools/slice"
	"github.com/pkg/errors"
)

func Success[T any](data T) (T, int, error) {
	return data, http.StatusOK, nil
}

func AdminBadRequestError[T any](err error) (T, int, error) {
	return *new(T), http.StatusBadRequest, err
}

func Unauthorized[T any]() (T, int, error) {
	return *new(T), http.StatusUnauthorized, std_errors.Public("Unauthorized")
}

func PublicBadRequestError[T any](err ...error) (T, int, error) {
	defaultErr := std_errors.Public("Internal Error")
	if len(err) > 0 {
		defaultErr = err[0]
	}
	return *new(T), http.StatusBadRequest, defaultErr
}

func PublicCustomError[T any](msg string, code int) (T, int, error) {
	return *new(T), code, std_errors.Public(msg)
}

func PublicNotFoundError[T any]() (T, int, error) {
	return *new(T), http.StatusNotFound, std_errors.Public("Not Found")
}

// StandardRequestWrapper wraps data in simple JSON responses
func StandardRequestWrapper[T any](fn func(res http.ResponseWriter, req *http.Request) (T, int, error)) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		result, code, err := fn(res, req)
		if err != nil {
			ErrorWrapper(res, req, err.Error(), code)
			return
		}

		JSONDataResponseWrapper(res, req, result)
	})
}

// JSONDataResponseWrapper wrapper to convert responses to JSONData api response
func JSONDataResponseWrapper(res http.ResponseWriter, req *http.Request, data any) {
	sanitizedResponse, err := maps.RecursiveJSON(ToJSONDataResponseMap(data))
	if err != nil {
		ErrorWrapper(res, req, err.Error(), http.StatusInternalServerError)
		return
	}
	router.JSONDataResponse(req.Context(), res, sanitizedResponse)
}

func ToJSONDataResponseMap(data any) any {
	modelType, isModel := data.(coremodel.Model)
	if isModel {
		return modelType.ToJSON()
	}

	rv := reflect.ValueOf(data)
	if rv.Kind() == reflect.Slice {
		responseSlice := make([]any, 0)
		err := slice.IterateReflect(data, func(_ int, val any) {
			modelType, isModel := val.(coremodel.Model)
			if isModel {
				responseSlice = append(responseSlice, modelType.ToJSON())
			} else {
				responseSlice = append(responseSlice, val)
			}
		})
		if err != nil {
			log.Error(err)
		}

		return responseSlice

	}

	return data
}

// ErrorWrapper wrapper to convert responses
func ErrorWrapper(res http.ResponseWriter, _ *http.Request, message string, code int) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	jsonResponse := map[string]any{
		"success": false,
		"error":   message,
	}
	encoder := json.NewEncoder(res)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(jsonResponse)
	if err != nil {
		log.Error(errors.WithStack(err))
	}
}

// ConvertPost converts raw post data and JSONAPI post data to a common format
func ConvertPost(data map[string]any) map[string]any {
	if !tools.Empty(data["data"]) {
		switch typedData := data["data"].(type) {
		case []any:
			return data
		case map[string]any:
			return typedData
		}
	}
	return data
}
