package monarch

import (
	"context"
	"testing"

	"github.com/eshaffer321/monarchmoney-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestClient(transport *MockTransport) *Client {
	return &Client{
		transport:   transport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
}

func TestAccountService_SearchSecurities(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	responseJSON := `{
		"securities": [
			{"id": "sec-123", "name": "Bitcoin", "ticker": "BTC", "currentPrice": 94200.50}
		]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(responseJSON, nil)

	securities, err := client.Accounts.SearchSecurities(context.Background(), "BTC")
	require.NoError(t, err)
	require.Len(t, securities, 1)
	assert.Equal(t, "sec-123", securities[0].ID)
	assert.Equal(t, "BTC", securities[0].Ticker)
	assert.Equal(t, "Bitcoin", securities[0].Name)
	assert.Equal(t, 94200.50, securities[0].CurrentPrice)
}

func TestAccountService_CreateHolding(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	responseJSON := `{
		"createManualHolding": {
			"holding": {"id": "hold-456", "ticker": "BTC"},
			"errors": []
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(responseJSON, nil)

	holding, err := client.Accounts.CreateHolding(context.Background(), &CreateHoldingParams{
		AccountID:  "acc-789",
		SecurityID: "sec-123",
		Quantity:   0.5,
	})
	require.NoError(t, err)
	assert.Equal(t, "hold-456", holding.ID)
	assert.Equal(t, "BTC", holding.Symbol)
	assert.Equal(t, "acc-789", holding.AccountID)
	assert.Equal(t, 0.5, holding.Quantity)
}

func TestAccountService_CreateHolding_Error(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	responseJSON := `{
		"createManualHolding": {
			"holding": null,
			"errors": [{"message": "invalid security", "code": "INVALID"}]
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(responseJSON, nil)

	_, err := client.Accounts.CreateHolding(context.Background(), &CreateHoldingParams{
		AccountID:  "acc-789",
		SecurityID: "bad-id",
		Quantity:   1.0,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid security")
}

func TestAccountService_DeleteHolding(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	responseJSON := `{
		"deleteHolding": {
			"deleted": true,
			"errors": []
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(responseJSON, nil)

	err := client.Accounts.DeleteHolding(context.Background(), "hold-456")
	require.NoError(t, err)
}

func TestAccountService_DeleteHolding_NotDeleted(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	responseJSON := `{
		"deleteHolding": {
			"deleted": false,
			"errors": []
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(responseJSON, nil)

	err := client.Accounts.DeleteHolding(context.Background(), "hold-456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not deleted")
}

func TestAccountService_CreateHoldingByTicker(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	// First call: search securities
	searchResponse := `{
		"securities": [
			{"id": "sec-btc", "name": "Bitcoin", "ticker": "BTC", "currentPrice": 94200.00}
		]
	}`

	// Second call: create holding
	createResponse := `{
		"createManualHolding": {
			"holding": {"id": "hold-new", "ticker": "BTC"},
			"errors": []
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(searchResponse, nil).Once()
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(createResponse, nil).Once()

	holding, err := client.Accounts.CreateHoldingByTicker(context.Background(), "acc-789", "BTC", 1.5)
	require.NoError(t, err)
	assert.Equal(t, "hold-new", holding.ID)
	assert.Equal(t, "BTC", holding.Symbol)
	assert.Equal(t, 1.5, holding.Quantity)
}

func TestAccountService_CreateHoldingByTicker_NotFound(t *testing.T) {
	mockTransport := new(MockTransport)
	client := newTestClient(mockTransport)
	client.Accounts = &accountService{client: client}

	searchResponse := `{"securities": []}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(searchResponse, nil)

	_, err := client.Accounts.CreateHoldingByTicker(context.Background(), "acc-789", "NOTREAL", 1.0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no security found")
}
