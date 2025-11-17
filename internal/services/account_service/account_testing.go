package account_service

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/organization_service"
)

type TestUserInput struct {
	OrganizationID types.UUID `json:"organization_id"`
}

func CreateTestUser(ctx context.Context, input *TestUserInput, savingUser coremodel.Model) (*account.Account, error) {
	accountObj := account.New()
	// Simple flag to make a daniels law valid user, NJ state + NJ law enforcement occupation

	accountObj.TestUserType.Set(1)
	randomName := common.GenerateRandomName()
	accountObj.FirstName.Set(randomName.FirstName)
	accountObj.LastName.Set(randomName.LastName)
	accountObj.Email.Set(fmt.Sprintf("%s.%s_%s@test.com", randomName.FirstName, randomName.LastName, tools.RandString(6)))

	accountObj.Set("password", fmt.Sprintf("%s_testpassword!", accountObj.FirstName.Get()))

	if !tools.Empty(input.OrganizationID) {
		accountObj.OrganizationID.Set(input.OrganizationID)
		accountObj.Role.Set(constants.ROLE_USER)
	} else {
		accountObj.Role.Set(constants.ROLE_ORG_OWNER)
	}

	err := accountObj.Save(savingUser)
	if err != nil {
		return nil, err
	}

	org, err := organization_service.CreateOrganization(ctx, accountObj)
	if err != nil {
		return nil, err
	}

	accountObj.OrganizationID.Set(org.ID())
	err = accountObj.Save(savingUser)
	if err != nil {
		return nil, err
	}

	return accountObj, nil
}
