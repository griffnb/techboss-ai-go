package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

type TSResponse struct {
	Fields      map[string]*model.TSProperties `json:"fields"`
	TSCode      string                         `json:"ts_code"`
	PublicTypes string                         `json:"public_types"`
}

func TSValidation(tableName string) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		tablePtrs := environment.GetDBClient(environment.CLIENT_DEFAULT).GetTablePtrs()
		fields := model.GetTSFieldValidation(tablePtrs[tableName])
		tsCode, _ := model.GenerateTypeScriptCode(tablePtrs[tableName], &base.Structure{})
		publicTypes := model.GeneratePublicTypeScriptModel(tablePtrs[tableName], tableName)

		response.JSONDataResponseWrapper(res, req, TSResponse{
			Fields:      fields,
			TSCode:      tsCode,
			PublicTypes: publicTypes,
		})
	})
}

func TSDynamoValidation(ptr any) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fields := model.GetTSFieldValidation(ptr)
		response.JSONDataResponseWrapper(res, req, TSResponse{
			Fields: fields,
		})
	})
}
