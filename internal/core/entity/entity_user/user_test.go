package entity_user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	user *User
}

func (s *UserTestSuite) SetupTest() {
	uid := "123"
	email := "test@example.com"
	displayName := "Test User"
	emailVerified := true
	token := "token123"
	user, err := NewUser(uid, email, displayName, emailVerified, token, nil, nil)
	assert.NoError(s.T(), err)
	s.user = user
}

func (s *UserTestSuite) TestNewUser() {
	s.T().Run("should create a new user", func(t *testing.T) {
		uid := "123"
		email := "test@example.com"
		displayName := "Test User"
		emailVerified := true
		token := "token123"

		user, err := NewUser(uid, email, displayName, emailVerified, token, nil, nil)
		assert.NotNil(t, user)
		assert.NoError(t, err)
		assert.Equal(t, uid, user.UID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, displayName, user.DisplayName)
		assert.Equal(t, emailVerified, user.EmailVerified)
		assert.Equal(t, token, user.Token)
	})

	s.T().Run("should return error for invalid email", func(t *testing.T) {
		user, err := NewUser("123", "", "Test User", true, "token123", nil, nil)
		assert.Nil(t, user)
		assert.EqualError(t, err, "invalid email: invalid email")
	})
}

func (s *UserTestSuite) TestValidate() {
	var user *User
	err := user.Validate()
	assert.EqualError(s.T(), err, "user is nil")

	_, err = NewUser("", "test@example.com", "Test User", true, "token123", nil, nil)
	assert.EqualError(s.T(), err, "invalid uid: cannot be empty")

	_, err = NewUser("123", "invalid-email", "Test User", true, "token123", nil, nil)
	assert.EqualError(s.T(), err, "invalid email: invalid email")

	_, err = NewUser("123", "test@example.com", "Test User", true, "", nil, nil)
	assert.EqualError(s.T(), err, "invalid token: cannot be empty")

	err = s.user.Validate()
	assert.NoError(s.T(), err)
}

func (s *UserTestSuite) TestIsEmail() {
	var user *User

	assert.False(s.T(), user.IsEmail())
	assert.True(s.T(), s.user.IsEmail())
	user = &User{Email: "invalid-email"}
	assert.False(s.T(), user.IsEmail())
}

func (s *UserTestSuite) TestIsUID() {
	var user *User
	assert.False(s.T(), user.IsUID())
	assert.True(s.T(), s.user.IsUID())
}

func (s *UserTestSuite) TestIsToken() {
	var user *User
	assert.False(s.T(), user.IsToken())
	assert.True(s.T(), s.user.IsToken())
}

func (s *UserTestSuite) TestGetters() {
	var user *User
	assert.Empty(s.T(), user.GetUID())
	assert.Empty(s.T(), user.GetEmail())
	assert.Empty(s.T(), user.GetDisplayName())
	assert.Empty(s.T(), user.GetToken())
	assert.False(s.T(), user.GetEmailVerified())

	assert.Equal(s.T(), "123", s.user.GetUID())
	assert.Equal(s.T(), "test@example.com", s.user.GetEmail())
	assert.Equal(s.T(), "Test User", s.user.GetDisplayName())
	assert.Equal(s.T(), "token123", s.user.GetToken())
	assert.True(s.T(), s.user.GetEmailVerified())
}

func (s *UserTestSuite) TestTenantMethods() {
	var user *User
	assert.EqualError(s.T(), user.AppendTenant(""), "user is nil")
	assert.EqualError(s.T(), user.RemoveTenant(""), "user is nil")

	err := s.user.AppendTenant("")
	assert.EqualError(s.T(), err, "invalid tenant id")

	err = s.user.RemoveTenant("")
	assert.EqualError(s.T(), err, "invalid tenant id")

	err = s.user.RemoveTenant("tenant1")
	assert.EqualError(s.T(), err, "tenant not found")

	err = s.user.AppendTenant("tenant1")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), s.user.Tenants, 1)
	assert.Contains(s.T(), s.user.Tenants, "tenant1")

	err = s.user.AppendTenant("tenant1")
	assert.EqualError(s.T(), err, "tenant already exists for this user")
	assert.Len(s.T(), s.user.Tenants, 1)
	assert.Contains(s.T(), s.user.Tenants, "tenant1")

	err = s.user.RemoveTenant("tenant1")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), s.user.Tenants, 0)

	// Add multiple tenants
	err = s.user.AppendTenant("tenant1")
	assert.NoError(s.T(), err)
	err = s.user.AppendTenant("tenant2")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), s.user.Tenants, 2)

	// Remove one of them
	err = s.user.RemoveTenant("tenant1")
	assert.NoError(s.T(), err)

	// Check the result
	assert.Len(s.T(), s.user.Tenants, 1)
	assert.NotContains(s.T(), s.user.Tenants, "tenant1")
	assert.Contains(s.T(), s.user.Tenants, "tenant2")

}

func (s *UserTestSuite) TestClaimMethods() {
	var user *User
	assert.EqualError(s.T(), user.AppendClaim("", nil), "user is nil")
	assert.EqualError(s.T(), user.RemoveClaim(""), "user is nil")
	assert.EqualError(s.T(), user.UpdateClaim("", nil), "user is nil")

	err := s.user.AppendClaim("", nil)
	assert.EqualError(s.T(), err, "invalid key")

	err = s.user.AppendClaim("role", nil)
	assert.EqualError(s.T(), err, "invalid value")

	err = s.user.RemoveClaim("")
	assert.EqualError(s.T(), err, "invalid key")

	err = s.user.RemoveClaim("role")
	assert.EqualError(s.T(), err, "claim not found")

	err = s.user.UpdateClaim("role", "admin")
	assert.EqualError(s.T(), err, "claim not found")

	err = s.user.AppendClaim("role", "user")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), s.user.Claims, 1)
	assert.Equal(s.T(), "user", s.user.Claims["role"])

	err = s.user.UpdateClaim("role", "admin")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), s.user.Claims, 1)
	assert.Equal(s.T(), "admin", s.user.Claims["role"])

	err = s.user.RemoveClaim("role")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), s.user.Claims, 0)

	err = s.user.UpdateClaim("", "admin")
	assert.EqualError(s.T(), err, "invalid key")

	err = s.user.UpdateClaim("role", "")
	assert.EqualError(s.T(), err, "invalid value")

	err = s.user.UpdateClaim("role", nil)
	assert.EqualError(s.T(), err, "invalid value")
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
