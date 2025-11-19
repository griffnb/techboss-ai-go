package sendpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type AddContactMessengerRequest struct {
	TypeID        int    `json:"typeId"`
	Login         string `json:"login"`
	IsMainChatbot bool   `json:"isMainChatbot,omitempty"`
}

type AddContactMessengerResponse struct {
	Data *ContactMessenger `json:"data,omitempty"`
}

// AddContactMessenger adds messenger to contact
// https://sendpulse.com/integrations/api/crm#/Contacts%20messengers/post_contacts__contactId__messengers
func (this *APIClient) AddContactMessenger(ctx context.Context, id int64, request *AddContactMessengerRequest) (*AddContactMessengerResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &AddContactMessengerResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, fmt.Sprintf("/contacts/%d/messengers", id)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type UpdateContactMessengerRequest struct {
	TypeID        int    `json:"typeId"`
	Login         string `json:"login"`
	IsMainChatbot bool   `json:"isMainChatbot,omitempty"`
}

type UpdateContactMessengerResponse struct {
	Data *ContactMessenger `json:"data,omitempty"`
}

// UpdateContactMessenger updates information about messenger of the contact
// https://sendpulse.com/integrations/api/crm#/Contacts%20messengers/put_contacts__contactId__messengers__messengerId_
func (this *APIClient) UpdateContactMessenger(
	ctx context.Context,
	id int64,
	messengerID int64,
	request *UpdateContactMessengerRequest,
) (*UpdateContactMessengerResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &UpdateContactMessengerResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%d/messengers/%d", id, messengerID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteContactMessengers deletes messenger information of the contact
// https://sendpulse.com/integrations/api/crm#/Contacts%20messengers/delete_contacts__contactId__messengers__messengerId_
func (this *APIClient) DeleteContactMessengers(ctx context.Context, id int64, messengerID int64) error {
	token, err := this.GetToken()
	if err != nil {
		return err
	}

	req := this.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d/messengers/%d", id, messengerID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = this.Call(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
