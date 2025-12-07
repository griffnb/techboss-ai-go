package agents

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func authSession(_ http.ResponseWriter, req *http.Request) (string, int, error) {
	userObj := helpers.GetLoadedUser(req)
	client := openai.NewClient(
		option.WithAPIKey(environment.GetConfig().AIKeys.OpenAI.APIKey),
	)

	org, err := organization.Get(req.Context(), userObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[string]()
	}

	resp, err := client.Beta.ChatKit.Sessions.New(req.Context(), openai.BetaChatKitSessionNewParams{
		User: userObj.ID().String(),

		Workflow: openai.ChatSessionWorkflowParam{
			ID: req.URL.Query().Get("workflow_id"),
			StateVariables: map[string]openai.ChatSessionWorkflowParamStateVariableUnion{
				"organization_id": {
					OfString: openai.String(userObj.OrganizationID.Get().String()),
				},
				"account_id": {
					OfString: openai.String(userObj.ID().String()),
				},
				"business_information": {
					OfString: openai.String(org.MetaData.GetI().GetOnboardAnswersString()),
				},
			},
		},
	})
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[string]()
	}

	return response.Success(resp.ClientSecret)
}
