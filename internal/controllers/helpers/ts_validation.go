package helpers

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func TSValidation(tableName string) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		tablePtrs := environment.GetDBClient(environment.CLIENT_DEFAULT).GetTablePtrs()
		fields := model.GetTSFieldValidation(tablePtrs[tableName])
		JSONDataResponseWrapper(res, req, fields)
	})
}

func TSDynamoValidation(ptr any) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fields := model.GetTSFieldValidation(ptr)
		JSONDataResponseWrapper(res, req, fields)
	})
}
