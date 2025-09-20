package entitytenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TenantTestSuite struct {
	suite.Suite
	emptyField       string
	nilField         *string
	dataPointerField *string
}

func (s *TenantTestSuite) SetupTest() {
	s.emptyField = ""
	dataField := "field data"
	s.dataPointerField = &dataField
}

func (s *TenantTestSuite) TestTenantFilter() {

	var filter *TenantFilter
	err := filter.Validate()
	assert.EqualError(s.T(), err, "tenant filter is nil")

	filter = &TenantFilter{}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")
}

func (s *TenantTestSuite) TestTenantFilterName() {

	filter := &TenantFilter{
		Name: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Name: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Name: s.dataPointerField,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *TenantTestSuite) TestTenantFilterDescription() {

	filter := &TenantFilter{
		Description: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Description: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Description: s.dataPointerField,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *TenantTestSuite) TestTenantFilterDisplayName() {

	filter := &TenantFilter{
		DisplayName: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		DisplayName: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		DisplayName: s.dataPointerField,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *TenantTestSuite) TestTenantFilterSlug() {

	filter := &TenantFilter{
		Slug: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Slug: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Slug: s.dataPointerField,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *TenantTestSuite) TestTenantFilterOwner() {

	filter := &TenantFilter{
		Owner: &s.emptyField,
	}
	err := filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Owner: s.nilField,
	}
	err = filter.Validate()
	assert.EqualError(s.T(), err, "no filter found")

	filter = &TenantFilter{
		Owner: s.dataPointerField,
	}
	err = filter.Validate()
	assert.NoError(s.T(), err)
}

func (s *TenantTestSuite) TestNewTenant() {
	name := "Test Tenant"
	slug := "test-tenant"
	owner := "owner@example.com"
	displayName := "Test Tenant Display"
	description := "This is a test tenant"

	tenant, err := NewTenant(name, slug, owner, displayName, description)
	assert.NotNil(s.T(), tenant)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), name, tenant.Name)
	assert.Equal(s.T(), slug, tenant.Slug)
	assert.Equal(s.T(), owner, tenant.Owner)
	assert.Equal(s.T(), displayName, tenant.DisplayName)
	assert.Equal(s.T(), description, tenant.Description)
}

func (s *TenantTestSuite) TestNewTenant_Failed() {
	s.Run("InvalidOwner", func() {
		name := "Test Tenant"
		slug := "test-tenant"
		invalidOwner := "example.com"
		displayName := "Test Tenant Display"
		description := "This is a test tenant"

		tenant, err := NewTenant(name, slug, invalidOwner, displayName, description)
		assert.Nil(s.T(), tenant)
		assert.EqualError(s.T(), err, "invalid tenant owner")
	})

	s.Run("InvalidName", func() {
		name := ""
		slug := "test-tenant"
		owner := "owner@example.com"
		displayName := "Test Tenant Display"
		description := "This is a test tenant"

		tenant, err := NewTenant(name, slug, owner, displayName, description)
		assert.Nil(s.T(), tenant)
		assert.EqualError(s.T(), err, "invalid tenant name")
	})
}

func (s *TenantTestSuite) TestValidate() {
	s.Run("NilTenant", func() {
		var tenant *Tenant
		err := tenant.Validate()
		assert.EqualError(s.T(), err, "tenant is nil")
	})

	s.Run("ValidTenant", func() {
		name := "My Tenant"
		slug := "test-tenant"
		owner := "owner@example.com"
		displayName := "Test Tenant Display"
		description := "This is a test tenant"

		tenant, err := NewTenant(name, slug, owner, displayName, description)
		assert.NotNil(s.T(), tenant)
		assert.NoError(s.T(), err)
	})

	s.Run("EmptyDisplayName", func() {
		name := "My Tenant"
		slug := "test-tenant"
		owner := "owner@example.com"
		description := "This is a test tenant"

		tenant, err := NewTenant(name, slug, owner, "", description)
		assert.NotNil(s.T(), tenant)
		assert.NoError(s.T(), err)
	})

	s.Run("EmptySlug", func() {
		name := "My Tenant"
		owner := "owner@example.com"
		displayName := "Test Tenant Display"
		description := "This is a test tenant"

		tenant, err := NewTenant(name, "", owner, displayName, description)
		assert.NotNil(s.T(), tenant)
		assert.NoError(s.T(), err)
	})
}

func (s *TenantTestSuite) TestIsValidName() {
	s.Run("NilTenant", func() {
		var tenant *Tenant
		b := tenant.IsValidName()
		assert.False(s.T(), b)
	})

	s.Run("EmptyName", func() {
		tenant := &Tenant{Name: ""}
		b := tenant.IsValidName()
		assert.False(s.T(), b)
	})

	s.Run("ValidName", func() {
		tenant := &Tenant{Name: "My Tenant"}
		b := tenant.IsValidName()
		assert.True(s.T(), b)
	})
}

func (s *TenantTestSuite) TestIsValidOwner() {
	s.Run("NilTenant", func() {
		var tenant *Tenant
		b := tenant.IsValidOwner()
		assert.False(s.T(), b)
	})

	s.Run("InvalidEmail", func() {
		tenant := &Tenant{Owner: "example.com"}
		b := tenant.IsValidOwner()
		assert.False(s.T(), b)
	})

	s.Run("ValidEmail", func() {
		tenant := &Tenant{Owner: "user@example.com"}
		b := tenant.IsValidOwner()
		assert.True(s.T(), b)
	})
}

func (s *TenantTestSuite) TestGetters() {
	name := "My Tenant"
	slug := "test-tenant"
	owner := "user@example.com"
	displayName := "Test Tenant Display"
	description := "This is a test tenant"

	tenant := &Tenant{
		Name:        name,
		Slug:        slug,
		Owner:       owner,
		DisplayName: displayName,
		Description: description,
	}

	s.Run("GetName", func() {
		assert.Equal(s.T(), name, tenant.GetName())
	})

	s.Run("GetSlug", func() {
		assert.Equal(s.T(), slug, tenant.GetSlug())
	})

	s.Run("GetOwner", func() {
		assert.Equal(s.T(), owner, tenant.GetOwner())
	})

	s.Run("GetDisplayName", func() {
		assert.Equal(s.T(), displayName, tenant.GetDisplayName())
	})

	s.Run("GetDescription", func() {
		assert.Equal(s.T(), description, tenant.GetDescription())
	})

	s.Run("GettersWithNilTenant", func() {
		var nilTenant *Tenant
		assert.Empty(s.T(), nilTenant.GetName())
		assert.Empty(s.T(), nilTenant.GetSlug())
		assert.Empty(s.T(), nilTenant.GetOwner())
		assert.Empty(s.T(), nilTenant.GetDisplayName())
		assert.Empty(s.T(), nilTenant.GetDescription())
	})
}

func TestTenantTestSuite(t *testing.T) {
	suite.Run(t, new(TenantTestSuite))
}
