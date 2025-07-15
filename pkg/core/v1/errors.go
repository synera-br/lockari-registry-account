package corev1

import "fmt"

const (
	ContextError                    = "context error: %s"
	ContextCancelled                = "context cancelled: %w"
	ContextDeadlineExceeded         = "context deadline exceeded: %s"
	EncryptionError                 = "encryption error: %s"
	DecryptionError                 = "decryption error: %s"
	SerializationError              = "serialization error: %s"
	DeserializationError            = "deserialization error: %s"
	NotImplemented                  = "not implemented: %s"
	NotSupported                    = "not supported: %s"
	InvalidArgument                 = "invalid argument: %s"
	InvalidType                     = "invalid type: %s"
	InvalidValue                    = "invalid value: %s"
	InvalidField                    = "invalid field: %s"
	InvalidOperation                = "invalid operation: %s"
	InvalidState                    = "invalid state: %s"
	InvalidConfiguration            = "invalid configuration: %s"
	InvalidCredentials              = "invalid credentials: %s"
	InvalidToken                    = "invalid token: %s"
	InvalidSignature                = "invalid signature: %s"
	InvalidRequest                  = "invalid request: %s"
	InvalidResponse                 = "invalid response: %s"
	InvalidParameter                = "invalid parameter: %s"
	InvalidHeader                   = "invalid header: %s"
	InvalidQuery                    = "invalid query: %s"
	InvalidPath                     = "invalid path: %s"
	InvalidMethod                   = "invalid method: %s"
	InvalidInput                    = "invalid input: %s"
	DatabaseError                   = "database error: %s"
	Unauthorized                    = "unauthorized: %s"
	NotFound                        = "not found: %s"
	GenericError                    = "generic error: %w"
	ValidationError                 = "validation error: %s"
	ConversionError                 = "conversion error: %s"
	EmptyResult                     = "empty result: %s"
	RepositoryNotFound              = "repository not found: %s"
	RepositoryAlreadyExists         = "repository already exists: %s"
	RepositoryError                 = "repository error: %s"
	RepositoryCreateError           = "repository create error: %s"
	RepositoryUpdateError           = "repository update error: %s"
	RepositoryDeleteError           = "repository delete error: %s"
	RepositoryGetError              = "repository get error: %s"
	RepositoryListError             = "repository list error: %s"
	RepositoryQueryError            = "repository query error: %s"
	RepositoryTransactionError      = "repository transaction error: %s"
	RepositoryBatchError            = "repository batch error: %s"
	RepositoryBatchCreateError      = "repository batch create error: %s"
	RepositoryBatchUpdateError      = "repository batch update error: %s"
	RepositoryBatchDeleteError      = "repository batch delete error: %s"
	RepositoryBatchGetError         = "repository batch get error: %s"
	RepositoryBatchListError        = "repository batch list error: %s"
	RepositoryBatchQueryError       = "repository batch query error: %s"
	RepositoryBatchTransactionError = "repository batch transaction error: %s"
	ServiceNotFoundError            = "service not found error: %s"
	ServiceAlreadyExistsError       = "service already exists error: %s"
	ServiceError                    = "service error: %s"
	ServiceCreateError              = "service create error: %s"
	ServiceUpdateError              = "service update error: %s"
	ServiceDeleteError              = "service delete error: %s"
	ServiceGetError                 = "service get error: %s"
	ServiceListError                = "service list error: %s"
	ServiceQueryError               = "service query error: %s"
	ServiceTransactionError         = "service transaction error: %s"
	ServiceGenericError             = "service generic error: %s"
	UnauthorizedError               = "unauthorized error: %s"
	ForbiddenError                  = "forbidden error: %s"
	InternalServerError             = "internal server error: %s"
	NotImplementedError             = "not implemented error: %s"
	NotSupportedError               = "not supported error: %s"
	ConflictError                   = "conflict error: %s"
	TenantNotFoundError             = "tenant not found error: %s"
	TenantAlreadyExistsError        = "tenant already exists error: %s"
	TenantAlreadyExists             = "tenant already exists"
	TenantError                     = "tenant error: %s"
	TenantCreateError               = "tenant create error: %s"
	TenantUpdateError               = "tenant update error: %s"
	TenantDeleteError               = "tenant delete error: %s"
	TenantGetError                  = "tenant get error: %s"
	TenantListError                 = "tenant list error: %s"
	TenantQueryError                = "tenant query error: %s"
	TenantTransactionError          = "tenant transaction error: %s"
	TenantGenericError              = "tenant generic error: %s"
)

func ErrRepositoryNotFound(repository string) error {
	return fmt.Errorf("repository '%s' not found", repository)
}

func ErrServiceNotFound(service string) error {
	return fmt.Errorf("service '%s' not found", service)
}

func ErrHandlerNotFound(handler string) error {
	return fmt.Errorf("handler '%s' not found", handler)
}

func ErrInvalidRequest(message string) error {
	return fmt.Errorf("invalid request: %s", message)
}

func ErrInvalidResponse(message string) error {
	return fmt.Errorf("invalid response: %s", message)
}

func ErrInternalServer(message string) error {
	return fmt.Errorf("internal server error: %s", message)
}

func ErrUnauthorized(message string) error {
	return fmt.Errorf("unauthorized: %s", message)
}

func ErrForbidden(message string) error {
	return fmt.Errorf("forbidden: %s", message)
}

func ErrNotFound(message string) error {
	return fmt.Errorf("not found: %s", message)
}

func ErrConflict(message string) error {
	return fmt.Errorf("conflict: %s", message)
}

func ErrMethodNotAllowed(message string) error {
	return fmt.Errorf("method not allowed: %s", message)
}

func ErrNotImplemented(message string) error {
	return fmt.Errorf("not implemented: %s", message)
}

func ErrServiceUnavailable(message string) error {
	return fmt.Errorf("service unavailable: %s", message)
}

func ErrGatewayTimeout(message string) error {
	return fmt.Errorf("gateway timeout: %s", message)
}

func ErrBadRequest(message string) error {
	return fmt.Errorf("bad request: %s", message)
}

func ErrTooManyRequests(message string) error {
	return fmt.Errorf("too many requests: %s", message)
}

func ErrRequestTimeout(message string) error {
	return fmt.Errorf("request timeout: %s", message)
}

func ErrRequestEntityTooLarge(message string) error {
	return fmt.Errorf("request entity too large: %s", message)
}

func ErrGenericError(message string) error {
	return fmt.Errorf("%s", message)
}
