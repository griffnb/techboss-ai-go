package sendpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// TODO https://sendpulse.com/integrations/api/crm#/Contact%20email%20addresses/post_contacts__contactId__emails
func (this *APIClient) AddContactEmail(ctx context.Context, id int64, email string) (any, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := new(any)                           // todo
	body, err := json.Marshal(map[string]string{ // todo make struct

	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, fmt.Sprintf("/contacts/%d/emails", id)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx,
		req,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TODO https://sendpulse.com/integrations/api/crm#/Contact%20email%20addresses/put_contacts__contactId__emails__emailId_
func (this *APIClient) UpdateContactEmail(ctx context.Context, id int64, emailId int64, email string) (any, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := new(any)                           // todo
	body, err := json.Marshal(map[string]string{ // todo make struct

	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%d/emails/%d", id, emailId)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithBody(body).
		WithSuccessResult(result)

	_, err = this.Call(ctx,
		req,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// https://sendpulse.com/integrations/api/crm#/Contact%20email%20addresses/delete_contacts__contactId__emails__emailId_
func (this *APIClient) DeleteContactEmail(ctx context.Context, id int64, emailId int64) (any, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := new(any) // todo

	req := this.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d/emails/%d", id, emailId)).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		WithHeader("Content-Type", "application/json").
		WithSuccessResult(result)

	_, err = this.Call(ctx,
		req,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
