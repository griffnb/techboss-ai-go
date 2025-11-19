package account

import (
	"fmt"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// GetName Safely gets the name from a join or combined first and last name
func (this *Account) GetName() string {
	return fmt.Sprintf("%s %s", this.FirstName.Get(), this.LastName.Get())
}

// IsInternal returns true if the account is an internal user (test user).
func (this *Account) IsInternal() bool {
	return this.TestUserType.Get() > 0
}

func (this *Account) GetAdminURL() string {
	baseAdminURL := environment.GetConfig().Server.AdminURL
	return fmt.Sprintf("%s/accounts/details/%s", baseAdminURL, this.ID())
}

// HashPassword Hashes a password for storage
func HashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Error(errors.WithStack(err))
	}
	return string(hashed)
}

// VerifyPassword verifys a hashed password
func VerifyPassword(savedPassword, enteredPassword string, accountID types.UUID) bool {
	if tools.Empty(savedPassword) {
		log.Error(errors.Errorf("Saved password is empty for account %s", accountID))
	}
	fail := bcrypt.CompareHashAndPassword([]byte(savedPassword), []byte(enteredPassword))
	if fail != nil {
		log.Info(fmt.Sprintf("Failed to verify password for account %s", accountID))
	}
	return fail == nil
}

// ToSavingUser returns a new account with the same data as this account, need to disassociate to prevent lock issues if you are saving
// with the same account as you are possibly modifying
func (this *Account) ToSavingUser() *Account {
	newAccount := New()
	copyData := this.GetDataCopy()
	newAccount.SetData(copyData)
	return newAccount
}

/*
func (this *Account) IsEmailVerified() bool {
	return this.Status.Get() != STATUS_PENDING_EMAIL_VERIFICATION
}
*/
