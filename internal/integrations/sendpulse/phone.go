package sendpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (this *APIClient) AddContactPhone(ctx context.Context, id int64, phone string) (any, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := new(any)                           // todo
	body, err := json.Marshal(map[string]string{ // todo make struct
		"phone": phone,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPost, fmt.Sprintf("/contacts/%d/phones", id)).
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

func (this *APIClient) UpdateContactPhone(ctx context.Context, id int64, phoneId int64, phone string) (any, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := new(any)                           // todo
	body, err := json.Marshal(map[string]string{ // todo make struct
		"phone": phone,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req := this.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%d/phones/%d", id, phoneId)).
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

func (this *APIClient) DeleteContactPhone(ctx context.Context, id int64, phoneId int64) (any, error) {
	token, err := this.GetToken()
	if err != nil {
		return nil, err
	}

	result := new(any) // todo

	req := this.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%d/phones/%d", id, phoneId)).
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
