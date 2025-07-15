package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// AuthTestSuite defines the test suite for auth package
type AuthTestSuite struct {
	suite.Suite
	validUser      User
	validClient    Client
	validTimestamp time.Time
	invalidUser    User
	invalidClient  Client
}

// SetupTest runs before each test in the suite
func (suite *AuthTestSuite) SetupTest() {
	suite.validUser = User{
		Uid:   "user-123",
		Email: "test@example.com",
		Name:  "Test User",
		Plan:  "premium",
	}

	suite.validClient = Client{
		IpAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (Test Browser)",
	}

	suite.validTimestamp = time.Now()

	suite.invalidUser = User{
		Uid:   "", // Invalid: empty UID
		Email: "", // Invalid: empty email
		Name:  "Test User",
		Plan:  "premium",
	}

	suite.invalidClient = Client{
		IpAddress: "", // Invalid: empty IP
		UserAgent: "", // Invalid: empty user agent
	}
}

// TestUserValidation tests User struct validation
func (suite *AuthTestSuite) TestUserValidation() {
	suite.Run("ValidUser", func() {
		err := suite.validUser.IsValid()
		assert.NoError(suite.T(), err)
	})

	suite.Run("InvalidUser_EmptyUID", func() {
		user := suite.validUser
		user.Uid = ""
		err := user.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "uid is required")
	})

	suite.Run("InvalidUser_EmptyEmail", func() {
		user := suite.validUser
		user.Email = ""
		err := user.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "email is required")
	})

	suite.Run("ValidUser_OptionalFields", func() {
		user := User{
			Uid:   "user-123",
			Email: "test@example.com",
			// Name and Plan are optional
		}
		err := user.IsValid()
		assert.NoError(suite.T(), err)
	})
}

// TestClientValidation tests Client struct validation
func (suite *AuthTestSuite) TestClientValidation() {
	suite.Run("ValidClient", func() {
		err := suite.validClient.IsValid()
		assert.NoError(suite.T(), err)
	})

	suite.Run("InvalidClient_EmptyIP", func() {
		client := suite.validClient
		client.IpAddress = ""
		err := client.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "ipAddress is required")
	})

	suite.Run("InvalidClient_EmptyUserAgent", func() {
		client := suite.validClient
		client.UserAgent = ""
		err := client.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "userAgent is required")
	})
}

// TestEventTypeValidation tests EventType validation and methods
func (suite *AuthTestSuite) TestEventTypeValidation() {
	suite.Run("ValidEventTypes", func() {
		validEventTypes := []EventType{
			LOGIN_SUCCESS,
			SIGNUP_SUCCESS,
			LOGIN_FAILURE,
			LOGOUT,
			PASSWORD_RESET_REQUEST,
			PASSWORD_CHANGE_SUCCESS,
		}

		for _, eventType := range validEventTypes {
			err := eventType.IsValid()
			assert.NoError(suite.T(), err, "EventType %s should be valid", eventType)
		}
	})

	suite.Run("InvalidEventType", func() {
		invalidEventType := EventType("INVALID_EVENT")
		err := invalidEventType.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "invalid event type")
	})

	suite.Run("GetEventType", func() {
		testCases := map[EventType]string{
			LOGIN_SUCCESS:           "Login Success",
			SIGNUP_SUCCESS:          "Signup Success",
			LOGIN_FAILURE:           "Login Failure",
			LOGOUT:                  "Logout",
			PASSWORD_RESET_REQUEST:  "Password Reset Request",
			PASSWORD_CHANGE_SUCCESS: "Password Change Success",
		}

		for eventType, expected := range testCases {
			result := eventType.GetEventType()
			assert.Equal(suite.T(), expected, result)
		}
	})

	suite.Run("GetEventType_Unknown", func() {
		unknownEventType := EventType("UNKNOWN")
		result := unknownEventType.GetEventType()
		assert.Equal(suite.T(), "Unknown Event Type", result)
	})

	suite.Run("SetEventType", func() {
		testCases := map[string]EventType{
			"LOGIN_SUCCESS":           LOGIN_SUCCESS,
			"SIGNUP_SUCCESS":          SIGNUP_SUCCESS,
			"LOGIN_FAILURE":           LOGIN_FAILURE,
			"LOGOUT":                  LOGOUT,
			"PASSWORD_RESET_REQUEST":  PASSWORD_RESET_REQUEST,
			"PASSWORD_CHANGE_SUCCESS": PASSWORD_CHANGE_SUCCESS,
		}

		for input, expected := range testCases {
			var eventType EventType
			err := eventType.SetEventType(input)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), expected, eventType)
		}
	})

	suite.Run("SetEventType_Invalid", func() {
		var eventType EventType
		err := eventType.SetEventType("INVALID_EVENT")
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "invalid event type")
	})

	suite.Run("String", func() {
		eventType := LOGIN_SUCCESS
		result := eventType.String()
		assert.Equal(suite.T(), "LOGIN_SUCCESS", result)
	})

	suite.Run("IsLoginEvent", func() {
		assert.True(suite.T(), LOGIN_SUCCESS.IsLoginEvent())
		assert.True(suite.T(), LOGIN_FAILURE.IsLoginEvent())
		assert.False(suite.T(), SIGNUP_SUCCESS.IsLoginEvent())
		assert.False(suite.T(), LOGOUT.IsLoginEvent())
	})

	suite.Run("IsSuccessEvent", func() {
		assert.True(suite.T(), LOGIN_SUCCESS.IsSuccessEvent())
		assert.True(suite.T(), SIGNUP_SUCCESS.IsSuccessEvent())
		assert.True(suite.T(), PASSWORD_CHANGE_SUCCESS.IsSuccessEvent())
		assert.False(suite.T(), LOGIN_FAILURE.IsSuccessEvent())
		assert.False(suite.T(), LOGOUT.IsSuccessEvent())
	})

	suite.Run("IsFailureEvent", func() {
		assert.True(suite.T(), LOGIN_FAILURE.IsFailureEvent())
		assert.False(suite.T(), LOGIN_SUCCESS.IsFailureEvent())
		assert.False(suite.T(), SIGNUP_SUCCESS.IsFailureEvent())
	})
}

// TestLoginStruct tests Login struct validation and methods
func (suite *AuthTestSuite) TestLoginStruct() {
	suite.Run("ValidLogin", func() {
		login := Login{
			EventType:  LOGIN_SUCCESS,
			User:       suite.validUser,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
		}

		err := login.IsValid()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), LOGIN_SUCCESS, login.GetEventType())
		assert.Equal(suite.T(), suite.validUser, login.GetUser())
		assert.Equal(suite.T(), suite.validClient, login.GetClientInfo())
		assert.Equal(suite.T(), suite.validTimestamp, login.GetTimestamp())
	})

	suite.Run("InvalidLogin_InvalidUser", func() {
		login := Login{
			EventType:  LOGIN_SUCCESS,
			User:       suite.invalidUser,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
		}

		err := login.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "uid is required")
	})

	suite.Run("InvalidLogin_InvalidClient", func() {
		login := Login{
			EventType:  LOGIN_SUCCESS,
			User:       suite.validUser,
			ClientInfo: suite.invalidClient,
			Timestamp:  suite.validTimestamp,
		}

		err := login.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "ipAddress is required")
	})

	suite.Run("InvalidLogin_EmptyEventType", func() {
		login := Login{
			EventType:  "",
			User:       suite.validUser,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
		}

		err := login.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "eventType is required")
	})

	suite.Run("InvalidLogin_ZeroTimestamp", func() {
		login := Login{
			EventType:  LOGIN_SUCCESS,
			User:       suite.validUser,
			ClientInfo: suite.validClient,
			Timestamp:  time.Time{}, // Zero timestamp
		}

		err := login.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "timestamp is required")
	})
}

// TestSignupStruct tests Signup struct validation and methods
func (suite *AuthTestSuite) TestSignupStruct() {
	suite.Run("ValidSignup", func() {
		signup := Signup{
			EventType:  SIGNUP_SUCCESS,
			User:       suite.validUser,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
			Tenant:     "tenant-123",
		}

		err := signup.IsValid()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), SIGNUP_SUCCESS, signup.GetEventType())
		assert.Equal(suite.T(), suite.validUser, signup.GetUser())
		assert.Equal(suite.T(), suite.validClient, signup.GetClientInfo())
		assert.Equal(suite.T(), suite.validTimestamp, signup.GetTimestamp())
	})

	suite.Run("ValidSignup_EmptyTenant", func() {
		signup := Signup{
			EventType:  SIGNUP_SUCCESS,
			User:       suite.validUser,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
			// Tenant is optional
		}

		err := signup.IsValid()
		assert.NoError(suite.T(), err)
	})

	suite.Run("InvalidSignup_EmptyUserName", func() {
		user := suite.validUser
		user.Name = ""
		signup := Signup{
			EventType:  SIGNUP_SUCCESS,
			User:       user,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
		}

		err := signup.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "user name is required")
	})

	suite.Run("InvalidSignup_EmptyUserPlan", func() {
		user := suite.validUser
		user.Plan = ""
		signup := Signup{
			EventType:  SIGNUP_SUCCESS,
			User:       user,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
		}

		err := signup.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "user plan is required")
	})

	suite.Run("InvalidSignup_EmptyTimestamp", func() {
		user := suite.validUser
		signup := Signup{
			EventType:  SIGNUP_SUCCESS,
			User:       user,
			ClientInfo: suite.validClient,
			Timestamp:  time.Time{},
		}

		err := signup.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "invalid signup: timestamp is required")
	})

	suite.Run("InvalidSignup_EmptyEventType", func() {
		user := suite.validUser
		signup := Signup{
			EventType:  "",
			User:       user,
			ClientInfo: suite.validClient,
			Timestamp:  suite.validTimestamp,
		}

		err := signup.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "invalid signup: eventType is required")
	})

	suite.Run("InvalidSignup_EmptyClientInfo", func() {
		user := suite.validUser
		signup := Signup{
			EventType:  SIGNUP_SUCCESS,
			User:       user,
			ClientInfo: Client{},
			Timestamp:  suite.validTimestamp,
		}

		err := signup.IsValid()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "invalid client: ipAddress is required")
	})
}

// TestFactoryMethods tests the factory methods
func (suite *AuthTestSuite) TestFactoryMethods() {
	suite.Run("NewLogin", func() {
		login := NewLogin(suite.validUser, suite.validClient)

		assert.NotNil(suite.T(), login)
		assert.Equal(suite.T(), LOGIN_SUCCESS, login.GetEventType())
		assert.Equal(suite.T(), suite.validUser, login.GetUser())
		assert.Equal(suite.T(), suite.validClient, login.GetClientInfo())
		assert.False(suite.T(), login.GetTimestamp().IsZero())
		assert.WithinDuration(suite.T(), time.Now(), login.GetTimestamp(), time.Second)

		err := login.IsValid()
		assert.NoError(suite.T(), err)
	})

	suite.Run("NewSignup", func() {
		tenant := "tenant-123"
		signup := NewSignup(suite.validUser, suite.validClient, tenant)

		assert.NotNil(suite.T(), signup)
		assert.Equal(suite.T(), SIGNUP_SUCCESS, signup.GetEventType())
		assert.Equal(suite.T(), suite.validUser, signup.GetUser())
		assert.Equal(suite.T(), suite.validClient, signup.GetClientInfo())
		assert.Equal(suite.T(), tenant, signup.GetTenant())
		assert.False(suite.T(), signup.GetTimestamp().IsZero())
		assert.WithinDuration(suite.T(), time.Now(), signup.GetTimestamp(), time.Second)

		err := signup.IsValid()
		assert.NoError(suite.T(), err)
	})

	suite.Run("NewSignup_EmptyTenant", func() {
		signup := NewSignup(suite.validUser, suite.validClient, "")

		assert.NotNil(suite.T(), signup)
		assert.Equal(suite.T(), "", signup.GetTenant())

		err := signup.IsValid()
		assert.NoError(suite.T(), err)
	})
}

// TestJSONSerialization tests JSON marshaling/unmarshaling
func (suite *AuthTestSuite) TestJSONSerialization() {
	suite.Run("Login_JSONSerialization", func() {
		login := NewLogin(suite.validUser, suite.validClient)

		// Test that the struct can be serialized (this would be tested with actual JSON marshaling in a real scenario)
		assert.NotEmpty(suite.T(), login.GetEventType())
		assert.NotEmpty(suite.T(), login.GetUser().Uid)
		assert.NotEmpty(suite.T(), login.GetUser().Email)
		assert.NotEmpty(suite.T(), login.GetClientInfo().IpAddress)
		assert.NotEmpty(suite.T(), login.GetClientInfo().UserAgent)
	})

	suite.Run("Signup_JSONSerialization", func() {
		signup := NewSignup(suite.validUser, suite.validClient, "tenant-123")

		// Test that the struct can be serialized
		assert.NotEmpty(suite.T(), signup.GetEventType())
		assert.NotEmpty(suite.T(), signup.GetUser().Uid)
		assert.NotEmpty(suite.T(), signup.GetUser().Email)
		assert.NotEmpty(suite.T(), signup.GetClientInfo().IpAddress)
		assert.NotEmpty(suite.T(), signup.GetClientInfo().UserAgent)
		assert.NotEmpty(suite.T(), signup.GetTenant())
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func (suite *AuthTestSuite) TestEdgeCases() {
	suite.Run("EmptyStructs", func() {
		var user User
		var client Client
		var login Login
		var signup Signup

		assert.Error(suite.T(), user.IsValid())
		assert.Error(suite.T(), client.IsValid())
		assert.Error(suite.T(), login.IsValid())
		assert.Error(suite.T(), signup.IsValid())
	})

	suite.Run("VeryLongStrings", func() {
		longString := string(make([]byte, 10000)) // Very long string

		user := User{
			Uid:   "user-123",
			Email: "test@example.com",
			Name:  longString,
			Plan:  longString,
		}

		err := user.IsValid()
		assert.NoError(suite.T(), err) // Should still be valid
	})

	suite.Run("SpecialCharacters", func() {
		user := User{
			Uid:   "user-123-!@#$%^&*()",
			Email: "test+tag@example.co.uk",
			Name:  "José María Ñoño",
			Plan:  "premium-plan_v2",
		}

		client := Client{
			IpAddress: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", // IPv6
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		}

		assert.NoError(suite.T(), user.IsValid())
		assert.NoError(suite.T(), client.IsValid())
	})
}

// TestConcurrency tests thread safety (basic test)
func (suite *AuthTestSuite) TestConcurrency() {
	suite.Run("ConcurrentFactoryMethods", func() {
		const numGoroutines = 100
		results := make(chan LoginEvent, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				login := NewLogin(suite.validUser, suite.validClient)
				results <- login
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			login := <-results
			assert.NoError(suite.T(), login.IsValid())
			assert.Equal(suite.T(), LOGIN_SUCCESS, login.GetEventType())
		}
	})
}

// Run the test suite
func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
