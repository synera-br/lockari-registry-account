package entity_user

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserResponseTestSuite struct {
	suite.Suite
}

func (s *UserTestSuite) TestNewUserResponse() {
	uid := "123"
	email := "test@example.com"
	displayName := "Test User"
	emailVerified := true
	token := "token123"
	id := uuid.New().String()

	user, err := NewUser(uid, email, displayName, emailVerified, token, nil, nil)
	assert.NotNil(s.T(), user)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), uid, user.UID)
	assert.Equal(s.T(), email, user.Email)
	assert.Equal(s.T(), displayName, user.DisplayName)
	assert.Equal(s.T(), emailVerified, user.EmailVerified)
	assert.Equal(s.T(), token, user.Token)

	userResponse, err := NewUserResponse("", user)
	assert.Nil(s.T(), userResponse)
	assert.EqualError(s.T(), err, "invalid id")

	userResponse, err = NewUserResponse(id, user)
	assert.NotNil(s.T(), userResponse)
	assert.NoError(s.T(), err)
}

func (s *UserResponseTestSuite) TestGetID() {

	var userResponse *UserResponse
	assert.Empty(s.T(), userResponse.GetID())

	user := &User{UID: "123", Email: "test@example.com", DisplayName: "Test User", EmailVerified: true, Token: "token123"}
	resp := &UserResponse{ID: "resp-1", User: user}
	assert.Equal(s.T(), "resp-1", resp.GetID())

	resp = &UserResponse{ID: "", User: user}
	assert.Equal(s.T(), "", resp.GetID())

	resp = &UserResponse{ID: "resp-2", User: nil}
	assert.Equal(s.T(), "resp-2", resp.GetID())
}

func (s *UserResponseTestSuite) TestGetUser() {
	var userResponse *UserResponse
	assert.Nil(s.T(), userResponse.GetUser())

	user := &User{UID: "123", Email: "test@example.com", DisplayName: "Test User", EmailVerified: true, Token: "token123"}
	resp := &UserResponse{ID: "resp-1", User: user}
	assert.NotNil(s.T(), resp.GetUser())
	assert.Equal(s.T(), user, resp.GetUser())

	resp = &UserResponse{ID: "resp-2", User: nil}
	assert.Nil(s.T(), resp.GetUser())
}

func (s *UserResponseTestSuite) TestIsID() {

	var userResponse *UserResponse
	assert.False(s.T(), userResponse.IsID())

	user := &User{UID: "123", Email: "test@example.com", DisplayName: "Test User", EmailVerified: true, Token: "token123"}
	resp := &UserResponse{ID: "resp-1", User: user}
	assert.Equal(s.T(), "resp-1", resp.GetID())
}

func (s *UserResponseTestSuite) TestValidate() {

	var userResponse *UserResponse
	assert.EqualError(s.T(), userResponse.Validate(), "invalid user response")

	user := &User{UID: "123", Email: "test@example.com", DisplayName: "Test User", EmailVerified: true, Token: "token123"}
	resp := &UserResponse{ID: "", User: user}
	assert.EqualError(s.T(), resp.Validate(), "invalid id")

	user = &User{UID: "", Email: "test@example.com", DisplayName: "Test User", EmailVerified: true, Token: "token123"}
	resp = &UserResponse{ID: "123", User: user}
	assert.EqualError(s.T(), resp.Validate(), "invalid uid: cannot be empty")
}

func TestUserResponseTestSuite(t *testing.T) {
	suite.Run(t, new(UserResponseTestSuite))
}
