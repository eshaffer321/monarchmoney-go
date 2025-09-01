package monarch

import (
	"context"
	"testing"

	"github.com/erickshaffer/monarchmoney-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionService_GetDetails(t *testing.T) {
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
		
			"subscription": {
				"id": "sub-123",
				"paymentSource": "STRIPE",
				"referralCode": "REF123",
				"isOnFreeTrial": false,
				"hasPremiumEntitlement": true
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	details, err := client.Subscription.GetDetails(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, details)
	assert.Equal(t, "sub-123", details.ID)
	assert.Equal(t, "STRIPE", details.PaymentSource)
	assert.Equal(t, "REF123", details.ReferralCode)
	assert.False(t, details.IsOnFreeTrial)
	assert.True(t, details.HasPremiumEntitlement)
	
	mockTransport.AssertExpectations(t)
}

func TestSubscriptionService_GetDetails_NoSubscription(t *testing.T) {
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
		
			"subscription": null
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	details, err := client.Subscription.GetDetails(context.Background())

	assert.Error(t, err)
	assert.Nil(t, details)
	assert.Contains(t, err.Error(), "no subscription found")
	
	mockTransport.AssertExpectations(t)
}

func TestSubscriptionService_GetDetails_FreeTrial(t *testing.T) {
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
		
			"subscription": {
				"id": "sub-456",
				"paymentSource": "",
				"referralCode": "",
				"isOnFreeTrial": true,
				"hasPremiumEntitlement": false
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	details, err := client.Subscription.GetDetails(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, details)
	assert.Equal(t, "sub-456", details.ID)
	assert.Equal(t, "", details.PaymentSource)
	assert.True(t, details.IsOnFreeTrial)
	assert.False(t, details.HasPremiumEntitlement)
	
	mockTransport.AssertExpectations(t)
}