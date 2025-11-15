// Package testing_service contains the builder struct to create and save objects for testing purposes.  used for creating entire sets of seed data vs one off seed data on the models themselves.
package testing_service

import (
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/griffnb/techboss-ai-go/internal/models/account"

	"github.com/griffnb/techboss-ai-go/internal/models/organization"
)

// Builder is a helper struct to create and save objects for testing purposes
// Start from the lowest objects and go up, i.e. add acount first then family then organization etc.
type Builder struct {
	Account      *account.Account
	Organization *organization.Organization

	allObjects []coremodel.Model
}

// New creates a new Builder object
func New() *Builder {
	return &Builder{
		allObjects: make([]coremodel.Model, 0),
	}
}

// SaveAll saves all objects created by the builder
func (this *Builder) SaveAll() error {
	for _, obj := range this.allObjects {
		if err := obj.Save(nil); err != nil {
			return err
		}
	}

	return nil
}

// CleanupAll runs a function against all models, should be the testtools.CleanupModel(obj) function
func (this *Builder) CleanupAll(run func(model coremodel.Model)) {
	for _, obj := range this.allObjects {
		run(obj)
	}
}

// WithAccount creates an account object and assigns it to the builder
// Run First
func (this *Builder) WithAccount(accountObj ...*account.Account) *Builder {
	if len(accountObj) > 0 {
		this.Account = accountObj[0]
	} else {
		this.Account = account.TESTCreateAccount()
	}

	this.allObjects = append(this.allObjects, this.Account)
	return this
}

// WithOrganization creates an organization object and assigns it to the builder
// Run Fourth - Relies on account and family
func (this *Builder) WithOrganization(organizationObj ...*organization.Organization) *Builder {
	if len(organizationObj) > 0 {
		this.Organization = organizationObj[0]
	} else {
		this.Organization = organization.TESTCreateOrganization()
	}

	if this.Account != nil {
		this.Account.OrganizationID.Set(this.Organization.ID())
	}

	this.allObjects = append(this.allObjects, this.Organization)

	return this
}
