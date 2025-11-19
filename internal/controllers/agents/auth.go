package agents

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func authSession(_ http.ResponseWriter, req *http.Request) (string, int, error) {
	userSession := request.GetReqSession(req)
	client := openai.NewClient(
		option.WithAPIKey(environment.GetConfig().AIKeys.OpenAI.APIKey),
	)

	resp, err := client.Beta.ChatKit.Sessions.New(req.Context(), openai.BetaChatKitSessionNewParams{
		User: userSession.User.ID().String(),
		Workflow: openai.ChatSessionWorkflowParam{
			ID: req.URL.Query().Get("workflow_id"),
		},
	})
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[string]()
	}

	return response.Success(resp.ClientSecret)
}
