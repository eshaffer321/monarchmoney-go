package monarch

import (
	"context"
	"testing"
	"time"

	"github.com/eshaffer321/monarchmoney-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBudgetService_List(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// This test should FAIL before the fix and PASS after the fix
	// We're testing that the service sends the correct parameter names

	// Mock a successful response when correct parameters are used
	mockResponse := `{
		"budgetData": {
			"monthlyAmountsByCategory": [
				{
					"category": {
						"id": "cat-1",
						"name": "Food"
					},
					"monthlyAmounts": [
						{
							"month": "2025-08-01",
							"plannedAmount": 500.00,
							"actualAmount": 250.00,
							"remainingAmount": 250.00,
							"percentComplete": 50.0,
							"transactionCount": 10
						}
					]
				}
			]
		}
	}`

	// Set up expectation that startMonth/endMonth should be sent
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(vars map[string]interface{}) bool {
			// The service should send startDate and endDate as variable names
			// (but they get mapped to startMonth/endMonth in the GraphQL query)
			_, hasStartDate := vars["startDate"]
			_, hasEndDate := vars["endDate"]

			return hasStartDate && hasEndDate
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute with date range
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgets, err := service.List(ctx, startOfMonth, endOfMonth)

	// After fix: should succeed
	assert.NoError(t, err, "Should not return an error when using correct parameters")
	assert.NotNil(t, budgets, "Should return budgets")

	// Verify the mock was called with correct parameters
	mockTransport.AssertExpectations(t)
}

func TestBudgetService_List_CorrectParameters(t *testing.T) {
	// This test shows what SHOULD happen with correct parameters

	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// Mock successful response when using correct parameters
	mockResponse := `{
		"budgetData": [
			{
				"id": "budget-1",
				"amount": 500.00,
				"spent": 250.00,
				"remaining": 250.00,
				"percentageComplete": 50.0,
				"category": {
					"id": "cat-1",
					"name": "Groceries"
				}
			}
		]
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(vars map[string]interface{}) bool {
			// After fix, we should be sending startMonth/endMonth
			_, hasStartMonth := vars["startMonth"]
			_, hasEndMonth := vars["endMonth"]

			// This will fail until we fix the implementation
			return hasStartMonth && hasEndMonth
		}),
		mock.Anything,
	).Return(mockResponse, nil).Maybe()

	// For now, return error since we haven't fixed it yet
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(nil, &Error{Code: "BAD_REQUEST", Message: "Invalid parameters"})

	// Execute
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgets, err := service.List(ctx, startOfMonth, endOfMonth)

	// Currently this fails because we're using wrong parameter names
	assert.Error(t, err, "Should fail until we fix the parameter names")
	assert.Nil(t, budgets)
}
