package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleHTTPError_ServerError_IncludesResponseBody(t *testing.T) {
	transport := &GraphQLTransport{}

	tests := []struct {
		name           string
		statusCode     int
		responseBody   []byte
		expectedInMsg  string
		expectedCode   string
	}{
		{
			name:          "525 SSL Handshake Failed with HTML body",
			statusCode:    525,
			responseBody:  []byte(`<html><body>SSL Handshake Failed</body></html>`),
			expectedInMsg: "525",
			expectedCode:  "SERVER_ERROR",
		},
		{
			name:          "500 with JSON error message",
			statusCode:    500,
			responseBody:  []byte(`{"error": "Internal server error", "message": "Database connection failed"}`),
			expectedInMsg: "Database connection failed",
			expectedCode:  "SERVER_ERROR",
		},
		{
			name:          "502 Bad Gateway with empty body",
			statusCode:    502,
			responseBody:  []byte{},
			expectedInMsg: "502",
			expectedCode:  "SERVER_ERROR",
		},
		{
			name:          "503 Service Unavailable",
			statusCode:    503,
			responseBody:  []byte(`Service temporarily unavailable`),
			expectedInMsg: "503",
			expectedCode:  "SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transport.handleHTTPError(tt.statusCode, tt.responseBody)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedInMsg, "error should contain status code or message")

			// For JSON responses with message field, should include the message
			if tt.statusCode == 500 && len(tt.responseBody) > 0 {
				assert.Contains(t, err.Error(), "Database connection failed", "should include parsed error message")
			}
		})
	}
}

func TestHandleHTTPError_ServerError_IncludesStatusCodeDescription(t *testing.T) {
	transport := &GraphQLTransport{}

	tests := []struct {
		name         string
		statusCode   int
		expectedDesc string
	}{
		{"500 Internal Server Error", 500, "Internal Server Error"},
		{"502 Bad Gateway", 502, "Bad Gateway"},
		{"503 Service Unavailable", 503, "Service Unavailable"},
		{"525 SSL Handshake Failed", 525, "SSL Handshake Failed"},
		{"526 Invalid SSL Certificate", 526, "Invalid SSL Certificate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transport.handleHTTPError(tt.statusCode, []byte(`error page`))

			assert.Error(t, err)
			errMsg := err.Error()
			assert.Contains(t, errMsg, tt.expectedDesc, "error should include human-readable description")
		})
	}
}
