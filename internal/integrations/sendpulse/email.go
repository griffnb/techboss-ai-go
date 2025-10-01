package sendpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type AddContactEmailRequest struct {
	Emails []EmailRequest `json:"emails"`
}

type EmailRequest struct {
	Email  string `json:"email"`
	IsMain bool   `json:"isMain,omitempty"`
}

type AddContactEmailResponse struct {
	Data []*ContactEmail `json:"data,omitempty"`
}

// AddContactEmail adds an email address to the contact
// https://sendpulse.com/integrations/api/crm#/Contact%20email%20addresses/post_contacts__contactId__emails
func (this *APIClient) AddContactEmail(ctx context.Context, id int64, request *AddContactEmailRequest) (*AddContactEmailResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &AddContactEmailResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, fmt.Sprintf("/contacts/%d/emails", id)).
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

type UpdateContactEmailRequest struct {
	Email string `json:"email"`
}

type UpdateContactEmailResponse struct {
	Data *ContactEmail `json:"data,omitempty"`
}

// UpdateContactEmail updates the email address of the contact
// https://sendpulse.com/integrations/api/crm#/Contact%20email%20addresses/put_contacts__contactId__emails__emailId_
func (this *APIClient) UpdateContactEmail(
	ctx context.Context,
	id int64,
	emailID int64,
	request *UpdateContactEmailRequest,
) (*UpdateContactEmailResponse, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := &UpdateContactEmailResponse{}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%d/emails/%d", id, emailID)).
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

// DeleteContactEmail removes the email address from the contact
// https://sendpulse.com/integrations/api/crm#/Contact%20email%20addresses/delete_contacts__contactId__emails__emailId_
func (this *APIClient) DeleteContactEmail(ctx context.Context, id int64, emailID int64) error {
	token, err := this.GetToken()
	if err != nil {
		return err
	}

	result := new(any) // delete endpoints typically return empty response

	req := this.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d/emails/%d", id, emailID)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithSuccessResult(result)

	_, err = this.Call(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
