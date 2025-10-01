package sendpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type AddContactPhoneRequest struct {
	Phone string `json:"phone"`
}

type AddContactPhoneResponse struct {
	Data *ContactPhone `json:"data,omitempty"`
}

// AddContactPhone adds a phone number to the contact
// https://sendpulse.com/integrations/api/crm#/Contact%20phone%20number/post_contacts__contactId__phones
func (this *APIClient) AddContactPhone(ctx context.Context, id int64, request *AddContactPhoneRequest) (*AddContactPhoneResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &AddContactPhoneResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, fmt.Sprintf("/contacts/%d/phones", id)).
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

type UpdateContactPhoneRequest struct {
	Phone string `json:"phone"`
}

type UpdateContactPhoneResponse struct {
	Data *ContactPhone `json:"data,omitempty"`
}

// UpdateContactPhone updates the phone number of the contact
// https://sendpulse.com/integrations/api/crm#/Contact%20phone%20number/put_contacts__contactId__phones__phoneId_
func (this *APIClient) UpdateContactPhone(
	ctx context.Context,
	id int64,
	phoneID int64,
	request *UpdateContactPhoneRequest,
) (*UpdateContactPhoneResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &UpdateContactPhoneResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%d/phones/%d", id, phoneID)).
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

// DeleteContactPhone removes the phone number from the contact
// https://sendpulse.com/integrations/api/crm#/Contact%20phone%20number/delete_contacts__contactId__phones__phoneId_
func (this *APIClient) DeleteContactPhone(ctx context.Context, id int64, phoneID int64) error {
	token, err := this.GetToken()
	if err != nil {
		return err
	}

	result := new(any) // delete endpoints typically return empty response

	req := this.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d/phones/%d", id, phoneID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
