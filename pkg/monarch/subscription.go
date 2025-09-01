package monarch

import (
	"context"

	"github.com/pkg/errors"
)

// subscriptionService implements the SubscriptionService interface
type subscriptionService struct {
	client *Client
}

// GetDetails retrieves subscription details
func (s *subscriptionService) GetDetails(ctx context.Context) (*SubscriptionDetails, error) {
	query := s.client.loadQuery("subscription/details.graphql")

	var result struct {
		Subscription *SubscriptionDetails `json:"subscription"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get subscription details")
	}

	if result.Subscription == nil {
		return nil, errors.New("no subscription found")
	}

	return result.Subscription, nil
}