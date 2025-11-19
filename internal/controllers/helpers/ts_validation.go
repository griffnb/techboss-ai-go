package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func TSValidation(tableName string) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		tablePtrs := environment.GetDBClient(environment.CLIENT_DEFAULT).GetTablePtrs()
		fields := model.GetTSFieldValidation(tablePtrs[tableName])
		response.JSONDataResponseWrapper(res, req, fields)
	})
}

func TSDynamoValidation(ptr any) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fields := model.GetTSFieldValidation(ptr)
		response.JSONDataResponseWrapper(res, req, fields)
	})
}
