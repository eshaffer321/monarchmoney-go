package monarch

import (
	"context"
	"testing"

	"github.com/erickshaffer/monarchmoney-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionService_GetSplits(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response
	response := `{
		
			"getTransaction": {
				"id": "test-txn-123",
				"amount": 100.50,
				"category": {
					"id": "cat-1",
					"name": "Groceries"
				},
				"merchant": {
					"id": "merch-1",
					"name": "Test Store"
				},
				"splitTransactions": [
					{
						"id": "split-1",
						"amount": 50.25,
						"notes": "Half for groceries",
						"category": {
							"id": "cat-1",
							"name": "Groceries"
						},
						"merchant": {
							"id": "merch-1",
							"name": "Test Store"
						}
					},
					{
						"id": "split-2",
						"amount": 50.25,
						"notes": "Half for household",
						"category": {
							"id": "cat-2",
							"name": "Household"
						}
					}
				]
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits, err := client.Transactions.GetSplits(context.Background(), "test-txn-123")

	assert.NoError(t, err)
	assert.Len(t, splits, 2)
	assert.Equal(t, "split-1", splits[0].ID)
	assert.Equal(t, 50.25, splits[0].Amount)
	assert.Equal(t, "Half for groceries", splits[0].Notes)
	assert.Equal(t, "Groceries", splits[0].Category.Name)
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_New(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response
	response := `{
		
			"updateTransactionSplit": {
				"transaction": {
					"id": "test-txn-123",
					"hasSplitTransactions": true,
					"splitTransactions": [
						{
							"id": "split-1",
							"amount": 60.00,
							"notes": "Updated split"
						},
						{
							"id": "split-2",
							"amount": 40.50
						}
					]
				},
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits := []*TransactionSplit{
		{
			Amount:     60.00,
			CategoryID: "cat-1",
			Notes:      "Updated split",
			Merchant:   &Merchant{Name: "Store A"},
		},
		{
			Amount:     40.50,
			CategoryID: "cat-2",
		},
	}

	err := client.Transactions.UpdateSplits(context.Background(), "test-txn-123", splits)
	assert.NoError(t, err)
	
	// Verify the mutation was called with correct parameters
	mockTransport.AssertCalled(t, "Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		input := vars["input"].(map[string]interface{})
		splitData := input["splitData"].([]map[string]interface{})
		return len(splitData) == 2 && 
			splitData[0]["amount"] == 60.00 &&
			splitData[0]["merchantName"] == "Store A"
	}), mock.Anything)
}

func TestTransactionService_GetSummary(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response - note that aggregates is an array
	response := `{
		
			"aggregates": [{
				"summary": {
					"avg": 31.47,
					"count": 5498,
					"max": 25000.00,
					"maxExpense": -25000.00,
					"sum": 172968.22,
					"sumIncome": 651023.79,
					"sumExpense": -478055.57,
					"first": "2021-10-22",
					"last": "2025-08-30"
				}
			}]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	summary, err := client.Transactions.GetSummary(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 5498, summary.Count)
	assert.Equal(t, 31.47, summary.Avg)
	assert.Equal(t, 172968.22, summary.Sum)
	assert.Equal(t, 651023.79, summary.SumIncome)
	assert.Equal(t, -478055.57, summary.SumExpense)
	assert.Equal(t, "2021-10-22", summary.First)
	assert.Equal(t, "2025-08-30", summary.Last)
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_GetSummary_EmptyResult(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock empty response
	response := `{
		
			"aggregates": []
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	summary, err := client.Transactions.GetSummary(context.Background())

	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "no transaction summary data available")
	
	mockTransport.AssertExpectations(t)
}