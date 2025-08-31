package monarch

import (
	"context"

	"github.com/pkg/errors"
)

// recurringService implements the RecurringService interface
type recurringService struct {
	client *Client
}

// List retrieves all recurring transactions
func (s *recurringService) List(ctx context.Context) ([]*RecurringTransaction, error) {
	query := s.client.loadQuery("recurring/list.graphql")

	var result struct {
		RecurringTransactions []*RecurringTransaction `json:"recurringTransactions"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get recurring transactions")
	}

	return result.RecurringTransactions, nil
}
