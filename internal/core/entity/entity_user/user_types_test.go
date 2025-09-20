package entity_user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTypesTestSuite struct {
	suite.Suite
	emptyField string
	nilField   *string
}

func (s *UserTypesTestSuite) SetupTest() {
	s.emptyField = ""
}

func (s *UserTypesTestSuite) TestUserTypesNil() {
	var filter *UserFilter
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")

	filter = &UserFilter{}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")
}

func (s *UserTypesTestSuite) TestUserTypesEmail() {
	filter := &UserFilter{
		Email: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")

	filter = &UserFilter{
		Email: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")

	email := "test@example.com"
	filter = &UserFilter{
		Email: &email,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *UserTypesTestSuite) TestUserTypesID() {
	filter := &UserFilter{
		ID: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")

	filter = &UserFilter{
		ID: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")
	id := "123"
	filter = &UserFilter{
		ID: &id,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *UserTypesTestSuite) TestUserTypesUUID() {
	filter := &UserFilter{
		UUID: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")

	filter = &UserFilter{
		UUID: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")
	uuid := "123"
	filter = &UserFilter{
		UUID: &uuid,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *UserTypesTestSuite) TestUserTypesTenants() {
	filter := &UserFilter{
		Tenants: nil,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")
	list := []string{}

	filter = &UserFilter{
		Tenants: &list,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter provided")

	tenant := "my-tenant"
	list = []string{tenant}
	filter = &UserFilter{
		Tenants: &list,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func TestUserTypesTestSuite(t *testing.T) {
	suite.Run(t, new(UserTypesTestSuite))
}
