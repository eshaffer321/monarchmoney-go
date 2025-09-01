package monarch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/erickshaffer/monarchmoney-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccountService_GetAggregateSnapshots(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		"aggregateSnapshots": [
			{
				"date": "2025-01-01",
				"balance": 10000.50
			},
			{
				"date": "2025-01-02",
				"balance": 10500.75
			},
			{
				"date": "2025-01-03",
				"balance": 11000.00
			}
		]
	}`

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		filters := vars["filters"].(map[string]interface{})
		return filters["startDate"] == "2025-01-01" &&
			filters["endDate"] == "2025-01-03"
	}), mock.Anything).Return(response, nil)

	params := &AggregateSnapshotsParams{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	snapshots, err := client.Accounts.GetAggregateSnapshots(context.Background(), params)

	assert.NoError(t, err)
	assert.Len(t, snapshots, 3)
	assert.Equal(t, "2025-01-01", snapshots[0].Date)
	assert.Equal(t, 10000.50, snapshots[0].Balance)
	assert.Equal(t, "2025-01-03", snapshots[2].Date)
	assert.Equal(t, 11000.00, snapshots[2].Balance)
	
	mockTransport.AssertExpectations(t)
}

func TestAccountService_GetAggregateSnapshots_WithAccountType(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"aggregateSnapshots": [
				{
					"date": "2025-01-01",
					"balance": 5000.00
				}
			]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		filters := vars["filters"].(map[string]interface{})
		return filters["accountType"] == "INVESTMENT"
	}), mock.Anything).Return(response, nil)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	params := &AggregateSnapshotsParams{
		StartDate:   &startDate,
		AccountType: "INVESTMENT",
	}

	snapshots, err := client.Accounts.GetAggregateSnapshots(context.Background(), params)

	assert.NoError(t, err)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, 5000.00, snapshots[0].Balance)
	
	mockTransport.AssertExpectations(t)
}

func TestAccountService_GetAggregateSnapshots_DefaultStartDate(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"aggregateSnapshots": []
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		filters := vars["filters"].(map[string]interface{})
		// Should default to 150 years ago
		startDate := filters["startDate"].(string)
		return strings.HasPrefix(startDate, "18") // Year should be in 1800s
	}), mock.Anything).Return(response, nil)

	// Call with nil params to test default behavior
	snapshots, err := client.Accounts.GetAggregateSnapshots(context.Background(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, snapshots)
	assert.Len(t, snapshots, 0)
	
	mockTransport.AssertExpectations(t)
}

func TestAccountService_UploadBalanceHistory(t *testing.T) {
	// Create a test HTTP server to handle the multipart upload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/account-balance-history/upload/", r.URL.Path)
		
		// Verify content type is multipart
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		
		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10MB
		assert.NoError(t, err)
		
		// Verify files field
		file, header, err := r.FormFile("files")
		assert.NoError(t, err)
		assert.Equal(t, "upload.csv", header.Filename)
		file.Close()
		
		// Verify account mapping field
		mapping := r.FormValue("account_files_mapping")
		assert.Contains(t, mapping, "test-account-123")
		assert.Contains(t, mapping, "upload.csv")
		
		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create client with test server URL
	client := &Client{
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     server.URL,
		httpClient:  http.DefaultClient,
	}
	client.initServices()

	csvContent := `Date,Balance
2025-01-01,1000.00
2025-01-02,1100.00
2025-01-03,1200.00`

	err := client.Accounts.UploadBalanceHistory(context.Background(), "test-account-123", csvContent)
	assert.NoError(t, err)
}

func TestAccountService_UploadBalanceHistory_EmptyParams(t *testing.T) {
	client := &Client{
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Test with empty account ID
	err := client.Accounts.UploadBalanceHistory(context.Background(), "", "csv content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accountID and csvContent cannot be empty")

	// Test with empty CSV content
	err = client.Accounts.UploadBalanceHistory(context.Background(), "account-123", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accountID and csvContent cannot be empty")
}

func TestAccountService_UploadBalanceHistory_ServerError(t *testing.T) {
	// Create a test HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid CSV format"))
	}))
	defer server.Close()

	client := &Client{
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     server.URL,
		httpClient:  http.DefaultClient,
	}
	client.initServices()

	err := client.Accounts.UploadBalanceHistory(context.Background(), "test-account", "invalid csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 400")
	assert.Contains(t, err.Error(), "Invalid CSV format")
}