package utils

import (
	"errors"
	"fmt"
)

var (
	// Entity Errors

	// Common Errors
	UserNotFoundInProvider = errors.New("user not found in provider")
	UserNotFound           = errors.New("user not found")
	UserInvalidUser        = errors.New("invalid user")
	UserAlreadyExists      = errors.New("user already exists")
	UserErrorToCreate      = errors.New("error to create user")
	UserErrorToUpdate      = errors.New("error to update user")
	UserErrorToDelete      = errors.New("error to delete user")

	// TENANT
	ErrInvalidTenant              = errors.New("invalid tenant")
	ErrTenantNotFound             = errors.New("tenant not found")
	ErrTenantAlreadyExists        = errors.New("tenant already exists")
	ErrInvalidTenantID            = errors.New("invalid tenant id")
	ErrInvalidTenantName          = errors.New("invalid tenant name")
	ErrInvalidTenantOwner         = errors.New("invalid tenant owner")
	ErrCannotCreateTenant         = errors.New("cannot create tenant")
	ErrCannotUpdateTenant         = errors.New("cannot update tenant")
	ErrCannotDeleteTenant         = errors.New("cannot delete tenant")
	ErrCannotGetTenant            = errors.New("cannot get tenant")
	ErrCannotListTenants          = errors.New("cannot list tenants")
	ErrTenantUnauthorized         = errors.New("tenant unauthorized")
	ErrTenantForbidden            = errors.New("tenant forbidden")
	ErrTenantConflict             = errors.New("tenant conflict")
	ErrTenantBadRequest           = errors.New("tenant bad request")
	ErrTenantInternal             = errors.New("tenant internal error")
	ErrInvalidTenantResponse      = errors.New("invalid tenant response")
	ErrCannotCreateTenantResponse = errors.New("cannot create tenant response")
	ErrCannotGetTenantResponse    = errors.New("cannot get tenant response")
	ErrCannotListTenantResponses  = errors.New("cannot list tenant responses")
	ErrCannotUpdateTenantResponse = errors.New("cannot update tenant response")
	ErrCannotDeleteTenantResponse = errors.New("cannot delete tenant response")
	ErrInvalidTenantService       = errors.New("invalid tenant service")

	// REGISTRY REQUEST
	// REGISTRY REQUEST SERVICE
	ServiceRegistryRequestInvalidRepository    = errors.New("repository cannot nil in service")
	ServiceRegistryRequestInvalidAuthClient    = errors.New("authentication provider cannot nil in service")
	ServiceRegistryRequestInvalidMQ            = errors.New("message queue cannot nil in service")
	ServiceRegistryRequestInvalidCache         = errors.New("cache cannot nil in service")
	ServiceRegistryRequestInvalidTenant        = errors.New("cache cannot nil in service")
	ServiceRegistryRequestInvalidObservability = errors.New("observability cannot nil in service")
	ServiceRegistryRequestInvalidLogger        = errors.New("logger cannot nil in service")

	// REGISTRY REQUEST HANDLER
	ServiceRegistryRequestInvalidWebServer              = errors.New("handler server cannot nil in service")
	ServiceRegistryRequestInvalidUser                   = errors.New("handler user cannot nil in service")
	ServiceRegistryRequestInvalidUserRecord             = errors.New("handler user record cannot nil in service")
	ServiceRegistryRequestInvalidAuthenticationProvider = errors.New("handler authentication provider cannot nil in service")
	ServiceRegistryRequestInvalidTenantProvider         = errors.New("handler tenant provider cannot nil in service")
	ServiceRegistryRequestInvalidUserProvider           = errors.New("handler user provider cannot nil in service")

	// Common Errors
	RegistryRequestUserNotFoundInProvider = errors.New("user not found in provider")
	RegistryRequestUserNotFound           = errors.New("user not found")
	RegistryRequestUserInvalidUser        = errors.New("invalid user")

	// USER
	ErrInvalidUser              = errors.New("invalid user")
	ErrUserNotFound             = errors.New("user not found")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrInvalidUserID            = errors.New("invalid user id")
	ErrInvalidUserEmail         = errors.New("invalid user email")
	ErrInvalidUserToken         = errors.New("invalid user token")
	ErrInvalidUserClaims        = errors.New("invalid user claims")
	ErrInvalidUserTenants       = errors.New("invalid user tenants")
	ErrCannotCreateUser         = errors.New("cannot create user")
	ErrCannotUpdateUser         = errors.New("cannot update user")
	ErrCannotDeleteUser         = errors.New("cannot delete user")
	ErrCannotGetUser            = errors.New("cannot get user")
	ErrCannotListUsers          = errors.New("cannot list users")
	ErrUserUnauthorized         = errors.New("user unauthorized")
	ErrUserForbidden            = errors.New("user forbidden")
	ErrUserConflict             = errors.New("user conflict")
	ErrUserBadRequest           = errors.New("user bad request")
	ErrUserInternal             = errors.New("user internal error")
	ErrInvalidUserResponse      = errors.New("invalid user response")
	ErrCannotCreateUserResponse = errors.New("cannot create user response")
	ErrCannotGetUserResponse    = errors.New("cannot get user response")
	ErrCannotListUserResponses  = errors.New("cannot list user responses")
	ErrCannotUpdateUserResponse = errors.New("cannot update user response")
	ErrCannotDeleteUserResponse = errors.New("cannot delete user response")
	ErrInvalidUserService       = errors.New("invalid user service")
)

const (
	UserCreatedWithSuccess = "user created with success"
)

func CustomError(err string) error {
	return errors.New(err)
}

func CustomErrorWithMessage(message string, err error) error {
	return fmt.Errorf("%s: %w", message, err)
}
