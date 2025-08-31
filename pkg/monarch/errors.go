package monarch

import (
	"errors"
	"fmt"
)

var (
	// ErrNotAuthenticated is returned when authentication is required
	ErrNotAuthenticated = errors.New("not authenticated")

	// ErrMFARequired is returned when MFA is required
	ErrMFARequired = errors.New("multi-factor authentication required")

	// ErrLoginFailed is returned when login fails
	ErrLoginFailed = errors.New("login failed")

	// ErrSessionExpired is returned when session has expired
	ErrSessionExpired = errors.New("session expired")

	// ErrRateLimited is returned when rate limited
	ErrRateLimited = errors.New("rate limited")

	// ErrTimeout is returned on timeout
	ErrTimeout = errors.New("request timeout")

	// ErrNotFound is returned when resource not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidRequest is returned for invalid requests
	ErrInvalidRequest = errors.New("invalid request")

	// ErrServerError is returned for server errors
	ErrServerError = errors.New("server error")

	// ErrRefreshInProgress is returned when refresh is already in progress
	ErrRefreshInProgress = errors.New("refresh already in progress")

	// ErrRefreshTimeout is returned when refresh times out
	ErrRefreshTimeout = errors.New("refresh timeout")
)

// Error represents an API error
type Error struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"statusCode"`
	Details    map[string]interface{} `json:"details,omitempty"`
	RequestID  string                 `json:"requestId,omitempty"`
	Err        error                  `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// Is checks if the error matches target
func (e *Error) Is(target error) bool {
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}

	t, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.Code == t.Code
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Error implements the error interface
func (e *GraphQLError) Error() string {
	return e.Message
}

// GraphQLErrors represents multiple GraphQL errors
type GraphQLErrors struct {
	Errors []*GraphQLError `json:"errors"`
}

// Error implements the error interface
func (e *GraphQLErrors) Error() string {
	if len(e.Errors) == 0 {
		return "unknown GraphQL error"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d GraphQL errors occurred", len(e.Errors))
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []*ValidationError `json:"errors"`
}

// Error implements the error interface
func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d validation errors occurred", len(e.Errors))
}

// NewError creates a new API error
func NewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsAuthError checks if error is authentication related
func IsAuthError(err error) bool {
	return errors.Is(err, ErrNotAuthenticated) ||
		errors.Is(err, ErrMFARequired) ||
		errors.Is(err, ErrLoginFailed) ||
		errors.Is(err, ErrSessionExpired)
}

// IsRetryable checks if error is retryable
func IsRetryable(err error) bool {
	if errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrServerError) {
		return true
	}

	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500 || apiErr.StatusCode == 429
	}

	return false
}
