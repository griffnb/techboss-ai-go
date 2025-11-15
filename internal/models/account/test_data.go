package account

import (
	"fmt"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common"
)

func TESTCreateAccount() *Account {
	name := common.GenerateRandomName()
	accountObj := New()
	common.GenerateURN(accountObj)
	accountObj.Email.Set(fmt.Sprintf("%s@%s.com", tools.RandString(10), tools.RandString(10)))
	accountObj.FirstName.Set(name.FirstName)
	accountObj.LastName.Set(name.LastName + "Test")

	return accountObj
}
