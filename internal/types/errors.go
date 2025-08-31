package types

import "fmt"

// Error represents an API error
type Error struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"statusCode"`
	Details    map[string]interface{} `json:"details,omitempty"`
	RequestID  string                 `json:"requestId,omitempty"`
	Err        error                  `json:"-"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("error: %s", e.Code)
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrors represents multiple GraphQL errors
type GraphQLErrors struct {
	Errors []*GraphQLError `json:"errors"`
}

// Error implements the error interface
func (e *GraphQLErrors) Error() string {
	if len(e.Errors) == 0 {
		return "GraphQL error"
	}
	return e.Errors[0].Message
}
