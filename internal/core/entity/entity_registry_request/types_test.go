package entity_registry_request

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TypesTestSuite struct {
	suite.Suite
	emptyField       string
	nilField         *string
	dataPointerField *string
}

func (s *TypesTestSuite) SetupTest() {
	s.emptyField = ""
	dataField := "field data"
	s.dataPointerField = &dataField
}

// UserFilter Tests
func (s *TypesTestSuite) TestUserFilterValidate() {
	s.Run("NilFilter", func() {
		var filter *UserFilter
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("EmptyFilter", func() {
		filter := &UserFilter{}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithEmptyEmail", func() {
		filter := &UserFilter{
			Email: &s.emptyField,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithNilEmail", func() {
		filter := &UserFilter{
			Email: s.nilField,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithValidEmail", func() {
		filter := &UserFilter{
			Email: s.dataPointerField,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithEmptyID", func() {
		filter := &UserFilter{
			ID: &s.emptyField,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithNilID", func() {
		filter := &UserFilter{
			ID: s.nilField,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithValidID", func() {
		filter := &UserFilter{
			ID: s.dataPointerField,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithEmptyUUID", func() {
		filter := &UserFilter{
			UUID: &s.emptyField,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithNilUUID", func() {
		filter := &UserFilter{
			UUID: s.nilField,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithValidUUID", func() {
		filter := &UserFilter{
			UUID: s.dataPointerField,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithEmptyTenants", func() {
		emptyTenants := []string{}
		filter := &UserFilter{
			Tenants: &emptyTenants,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithNilTenants", func() {
		filter := &UserFilter{
			Tenants: nil,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithValidTenants", func() {
		tenants := []string{"tenant1", "tenant2"}
		filter := &UserFilter{
			Tenants: &tenants,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithSingleTenant", func() {
		tenants := []string{"single-tenant"}
		filter := &UserFilter{
			Tenants: &tenants,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithMultipleValidFields", func() {
		tenants := []string{"tenant1"}
		filter := &UserFilter{
			Email:   s.dataPointerField,
			ID:      s.dataPointerField,
			UUID:    s.dataPointerField,
			Tenants: &tenants,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithMixedValidAndInvalid", func() {
		filter := &UserFilter{
			Email: s.dataPointerField, // valid
			ID:    &s.emptyField,      // invalid but should still pass because Email is valid
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithAllEmptyFields", func() {
		emptyTenants := []string{}
		filter := &UserFilter{
			Email:   &s.emptyField,
			ID:      &s.emptyField,
			UUID:    &s.emptyField,
			Tenants: &emptyTenants,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithAllNilFields", func() {
		filter := &UserFilter{
			Email:   s.nilField,
			ID:      s.nilField,
			UUID:    s.nilField,
			Tenants: nil,
		}
		err := filter.Validate()
		assert.EqualError(s.T(), err, "no filter provided")
	})

	s.Run("FilterWithOnlyValidEmail", func() {
		email := "user@example.com"
		filter := &UserFilter{
			Email:   &email,
			ID:      s.nilField,
			UUID:    s.nilField,
			Tenants: nil,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithOnlyValidID", func() {
		id := "user123"
		filter := &UserFilter{
			Email:   s.nilField,
			ID:      &id,
			UUID:    s.nilField,
			Tenants: nil,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithOnlyValidUUID", func() {
		uuid := "550e8400-e29b-41d4-a716-446655440000"
		filter := &UserFilter{
			Email:   s.nilField,
			ID:      s.nilField,
			UUID:    &uuid,
			Tenants: nil,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithOnlyValidTenants", func() {
		tenants := []string{"tenant1", "tenant2", "tenant3"}
		filter := &UserFilter{
			Email:   s.nilField,
			ID:      s.nilField,
			UUID:    s.nilField,
			Tenants: &tenants,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithWhitespaceOnlyFields", func() {
		whitespaceField := "   "
		emptyTenants := []string{}
		filter := &UserFilter{
			Email:   &whitespaceField,
			ID:      &whitespaceField,
			UUID:    &whitespaceField,
			Tenants: &emptyTenants,
		}
		err := filter.Validate()
		// Whitespace fields are treated as valid because the code only checks != ""
		assert.NoError(s.T(), err)
	})

	s.Run("FilterWithTenantsContainingEmptyStrings", func() {
		tenants := []string{"", "tenant1", ""}
		filter := &UserFilter{
			Tenants: &tenants,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err) // Should pass because len(*f.Tenants) > 0
	})

	s.Run("FilterWithValidEmailAndEmptyOthers", func() {
		email := "test@example.com"
		emptyTenants := []string{}
		filter := &UserFilter{
			Email:   &email,
			ID:      &s.emptyField,
			UUID:    &s.emptyField,
			Tenants: &emptyTenants,
		}
		err := filter.Validate()
		assert.NoError(s.T(), err)
	})
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}