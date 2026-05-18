package types

import (
	"errors"
	"time"
)

const (
	// DefaultBaseURL is the default Monarch Money API base URL
	DefaultBaseURL = "https://api.monarch.com"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// UserAgent is the user agent string
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"
)

// Common errors
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

	// ErrServerError is returned for server errors
	ErrServerError = errors.New("server error")
)
