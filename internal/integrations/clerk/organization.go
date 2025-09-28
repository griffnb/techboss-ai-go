package clerk

import (
	"context"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/organization"
	"github.com/pkg/errors"
)

func (this *APIClient) CreateOrganization(ctx context.Context, params *organization.CreateParams) (*clerk.Organization, error) {
	org, err := organization.Create(ctx, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return org, nil
}

func (this *APIClient) GetOrganizations(ctx context.Context, params *organization.ListParams) (*clerk.OrganizationList, error) {
	result, err := organization.List(ctx, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

func (this *APIClient) GetOrganization(ctx context.Context, id string) (*clerk.Organization, error) {
	org, err := organization.Get(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return org, nil
}

func (this *APIClient) UpdateOrganization(ctx context.Context, id string, params *organization.UpdateParams) (any, error) {
	org, err := organization.Update(ctx, id, params)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return org, nil
}

func (this *APIClient) DeleteOrganization(ctx context.Context, id string) (any, error) {
	org, err := organization.Delete(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return org, nil
}
