package entity_registry_request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RegistryRequestTestSuite struct {
	suite.Suite
	emptyField       string
	nilField         *string
	dataPointerField *string
}

func (s *RegistryRequestTestSuite) SetupTest() {
	s.emptyField = ""
	dataField := "field data"
	s.dataPointerField = &dataField
}

// RegistryRequest Tests
func (s *RegistryRequestTestSuite) TestRegistryRequestValidate() {
	s.Run("NilRequest", func() {
		var req *RegistryRequest
		err := req.Validate()
		assert.EqualError(s.T(), err, "registry request is nil")
	})

	s.Run("EmptyName", func() {
		req := &RegistryRequest{
			Name:     "",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "name is required")
	})

	s.Run("WhitespaceOnlyName", func() {
		req := &RegistryRequest{
			Name:     "   ",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "name is required")
	})

	s.Run("EmptyEmail", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "email is required")
	})

	s.Run("WhitespaceOnlyEmail", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "   ",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "email is required")
	})

	s.Run("InvalidEmailFormat", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "invalid-email",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "invalid email format")
	})

	s.Run("EmptyTenant", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "tenant is required")
	})

	s.Run("WhitespaceOnlyTenant", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "   ",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "tenant is required")
	})

	s.Run("InvalidTenantFormat", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "Invalid-Tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "invalid tenant format: use only lowercase letters, numbers and hyphens")
	})

	s.Run("EmptyPassword", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "password is required")
	})

	s.Run("WhitespaceOnlyPassword", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "   ",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "password is required")
	})

	s.Run("ValidRequest", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		err := req.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("ValidGoogleAuth", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "google-auth",
		}
		err := req.Validate()
		assert.NoError(s.T(), err)
	})
}

func (s *RegistryRequestTestSuite) TestRegistryRequestIsValidEmail() {
	s.Run("ValidEmails", func() {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"test+label@example.org",
			"123@domain.com",
			"user@sub.domain.com",
		}

		for _, email := range validEmails {
			req := &RegistryRequest{Email: email}
			assert.True(s.T(), req.IsValidEmail(), "Email should be valid: %s", email)
		}
	})

	s.Run("InvalidEmails", func() {
		invalidEmails := []string{
			"",
			"invalid-email",
			"@domain.com",
			"user@",
			"user@domain",
			"user.domain.com",
			"user@@domain.com",
		}

		for _, email := range invalidEmails {
			req := &RegistryRequest{Email: email}
			assert.False(s.T(), req.IsValidEmail(), "Email should be invalid: %s", email)
		}
	})
}

func (s *RegistryRequestTestSuite) TestRegistryRequestIsValidTenant() {
	s.Run("ValidTenants", func() {
		validTenants := []string{
			"test-tenant",
			"company123",
			"my-company-name",
			"abc",
			"tenant-with-numbers-123",
		}

		for _, tenant := range validTenants {
			req := &RegistryRequest{Tenant: tenant}
			assert.True(s.T(), req.IsValidTenant(), "Tenant should be valid: %s", tenant)
		}
	})

	s.Run("InvalidTenants", func() {
		invalidTenants := []string{
			"",
			"ab", // too short
			strings.Repeat("a", 51), // too long
			"Invalid-Tenant",         // uppercase
			"tenant_with_underscores",
			"tenant with spaces",
			"tenant@invalid",
			"tenant.invalid",
		}

		for _, tenant := range invalidTenants {
			req := &RegistryRequest{Tenant: tenant}
			assert.False(s.T(), req.IsValidTenant(), "Tenant should be invalid: %s", tenant)
		}
	})
}

func (s *RegistryRequestTestSuite) TestRegistryRequestIsGoogleAuth() {
	s.Run("GoogleAuthPassword", func() {
		req := &RegistryRequest{Password: "google-auth"}
		assert.True(s.T(), req.IsGoogleAuth())
	})

	s.Run("RegularPassword", func() {
		req := &RegistryRequest{Password: "password123"}
		assert.False(s.T(), req.IsGoogleAuth())
	})

	s.Run("EmptyPassword", func() {
		req := &RegistryRequest{Password: ""}
		assert.False(s.T(), req.IsGoogleAuth())
	})
}

func (s *RegistryRequestTestSuite) TestRegistryRequestGetAuthType() {
	s.Run("GoogleAuthType", func() {
		req := &RegistryRequest{Password: "google-auth"}
		assert.Equal(s.T(), "google", req.GetAuthType())
	})

	s.Run("EmailAuthType", func() {
		req := &RegistryRequest{Password: "password123"}
		assert.Equal(s.T(), "email", req.GetAuthType())
	})

	s.Run("EmptyPasswordAuthType", func() {
		req := &RegistryRequest{Password: ""}
		assert.Equal(s.T(), "email", req.GetAuthType())
	})
}

func (s *RegistryRequestTestSuite) TestRegistryRequestString() {
	s.Run("RegularAuth", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "password123",
		}
		expected := "RegistryRequest{Name: Test User, Email: test@example.com, Tenant: test-tenant, AuthType: email}"
		assert.Equal(s.T(), expected, req.String())
	})

	s.Run("GoogleAuth", func() {
		req := &RegistryRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Tenant:   "test-tenant",
			Password: "google-auth",
		}
		expected := "RegistryRequest{Name: Test User, Email: test@example.com, Tenant: test-tenant, AuthType: google}"
		assert.Equal(s.T(), expected, req.String())
	})
}

// LoginRequest Tests
func (s *RegistryRequestTestSuite) TestLoginRequestValidate() {
	s.Run("NilRequest", func() {
		var req *LoginRequest
		err := req.Validate()
		assert.EqualError(s.T(), err, "login request is nil")
	})

	s.Run("EmptyUID", func() {
		req := &LoginRequest{
			UID:   "",
			Email: "test@example.com",
			Token: "token123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "uid is required")
	})

	s.Run("WhitespaceOnlyUID", func() {
		req := &LoginRequest{
			UID:   "   ",
			Email: "test@example.com",
			Token: "token123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "uid is required")
	})

	s.Run("EmptyEmail", func() {
		req := &LoginRequest{
			UID:   "uid123",
			Email: "",
			Token: "token123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "email is required")
	})

	s.Run("WhitespaceOnlyEmail", func() {
		req := &LoginRequest{
			UID:   "uid123",
			Email: "   ",
			Token: "token123",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "email is required")
	})

	s.Run("EmptyToken", func() {
		req := &LoginRequest{
			UID:   "uid123",
			Email: "test@example.com",
			Token: "",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "token is required")
	})

	s.Run("WhitespaceOnlyToken", func() {
		req := &LoginRequest{
			UID:   "uid123",
			Email: "test@example.com",
			Token: "   ",
		}
		err := req.Validate()
		assert.EqualError(s.T(), err, "token is required")
	})

	s.Run("ValidRequest", func() {
		req := &LoginRequest{
			UID:           "uid123",
			Email:         "test@example.com",
			Token:         "token123",
			DisplayName:   "Test User",
			EmailVerified: true,
		}
		err := req.Validate()
		assert.NoError(s.T(), err)
	})

	s.Run("ValidRequestMinimal", func() {
		req := &LoginRequest{
			UID:   "uid123",
			Email: "test@example.com",
			Token: "token123",
		}
		err := req.Validate()
		assert.NoError(s.T(), err)
	})
}

func (s *RegistryRequestTestSuite) TestLoginRequestGetDisplayNameOrFallback() {
	s.Run("WithDisplayName", func() {
		req := &LoginRequest{
			DisplayName: "Test User",
			Email:       "test@example.com",
		}
		assert.Equal(s.T(), "Test User", req.GetDisplayNameOrFallback())
	})

	s.Run("WithoutDisplayNameValidEmail", func() {
		req := &LoginRequest{
			DisplayName: "",
			Email:       "test.user@example.com",
		}
		assert.Equal(s.T(), "Test.user", req.GetDisplayNameOrFallback())
	})

	s.Run("WithoutDisplayNameSimpleEmail", func() {
		req := &LoginRequest{
			DisplayName: "",
			Email:       "john@example.com",
		}
		assert.Equal(s.T(), "John", req.GetDisplayNameOrFallback())
	})

	s.Run("EmptyEmailAndDisplayName", func() {
		req := &LoginRequest{
			DisplayName: "",
			Email:       "",
		}
		assert.Equal(s.T(), "Usuário", req.GetDisplayNameOrFallback())
	})

	s.Run("EmailWithoutAtSymbol", func() {
		req := &LoginRequest{
			DisplayName: "",
			Email:       "testuser",
		}
		assert.Equal(s.T(), "Testuser", req.GetDisplayNameOrFallback())
	})

	s.Run("EmailWithEmptyBeforeAt", func() {
		req := &LoginRequest{
			DisplayName: "",
			Email:       "@example.com",
		}
		assert.Equal(s.T(), "Usuário", req.GetDisplayNameOrFallback())
	})

	s.Run("SingleCharacterEmail", func() {
		req := &LoginRequest{
			DisplayName: "",
			Email:       "a@example.com",
		}
		assert.Equal(s.T(), "A", req.GetDisplayNameOrFallback())
	})
}

func (s *RegistryRequestTestSuite) TestLoginRequestString() {
	s.Run("FullLoginRequest", func() {
		req := &LoginRequest{
			UID:           "uid123",
			Email:         "test@example.com",
			DisplayName:   "Test User",
			EmailVerified: true,
			Token:         "token123",
		}
		expected := "LoginRequest{UID: uid123, Email: test@example.com, DisplayName: Test User, EmailVerified: true}"
		assert.Equal(s.T(), expected, req.String())
	})

	s.Run("MinimalLoginRequest", func() {
		req := &LoginRequest{
			UID:           "uid123",
			Email:         "test@example.com",
			EmailVerified: false,
		}
		expected := "LoginRequest{UID: uid123, Email: test@example.com, DisplayName: , EmailVerified: false}"
		assert.Equal(s.T(), expected, req.String())
	})
}

// UserFilter Tests
func (s *RegistryRequestTestSuite) TestUserFilterValidate() {
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
}

func TestRegistryRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryRequestTestSuite))
}