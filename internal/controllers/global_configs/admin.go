package global_configs

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/global_config"
	"github.com/pkg/errors"
)

// SetupAdmin sets up admin routes

func adminIndex(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	parameters := router.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)

	if tools.Empty(parameters.Limit) {
		parameters.Limit = constants.SYSTEM_LIMIT
	}

	coreModels, err := global_config.FindAll(req.Context(), parameters)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)

	}

	return helpers.Success(coreModels)
}

func adminGet(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	id := chi.URLParam(req, "id")
	coreModel, err := global_config.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)

	}

	return helpers.Success(coreModel)
}

func adminCreate(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	userSession := req.Context().Value(router.SessionContextKey("session")).(*session.Session)
	rawdata := router.GetJSONPostData(req)
	data := helpers.ConvertPost(rawdata)
	coreModel := global_config.New()
	coreModel.MergeData(data)
	err := coreModel.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)

	}

	return helpers.Success(coreModel)
}

func adminUpdate(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	userSession := req.Context().Value(router.SessionContextKey("session")).(*session.Session)
	rawdata := router.GetJSONPostData(req)
	data := helpers.ConvertPost(rawdata)

	bulkParams, err := helpers.BuildBulkParams(req.Context(), rawdata, TABLE_NAME)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	if tools.Empty(bulkParams) {
		idString, _ := tools.ParseString(chi.URLParam(req, "id"))
		bulkParams = &model.Options{
			Conditions: fmt.Sprintf("%s.id = :id:", TABLE_NAME),
			Params: map[string]interface{}{
				":id:": idString,
			},
		}
	}
	delete(data, "ids")
	delete(data, "query")

	if tools.Empty(bulkParams) {
		return helpers.AdminBadRequestError[any](errors.Errorf("Save Error : No IDs Passed"))
	}

	coreModels, err := global_config.FindAll(req.Context(), bulkParams)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	if tools.Empty(coreModels) {
		return helpers.Success(map[string]any{})
	}

	var wg sync.WaitGroup
	for _, coreModel := range coreModels {
		wg.Add(1)

		go func(coreModel coremodel.Model) {
			defer wg.Done()
			coreModel.MergeData(data)
			saveErr := coreModel.Save(userSession.User)
			if saveErr != nil {
				log.Error(saveErr)
				return
			}
		}(coreModel)
	}

	wg.Wait()

	if len(coreModels) > 1 {
		return helpers.Success(coreModels)
	}

	return helpers.Success(coreModels[0])
}

func adminCount(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	parameters := router.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
	count, err := global_config.FindResultsCount(req.Context(), parameters)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	return helpers.Success(count)
}
