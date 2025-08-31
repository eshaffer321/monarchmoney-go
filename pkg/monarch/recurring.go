package monarch

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// recurringService implements the RecurringService interface
type recurringService struct {
	client *Client
}

// List retrieves all recurring transactions
func (s *recurringService) List(ctx context.Context) ([]*RecurringTransaction, error) {
	// Use date range for next 30 days by default
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // 1 month from now

	return s.ListWithDateRange(ctx, startDate, endDate)
}

// ListWithDateRange retrieves recurring transactions for a specific date range
func (s *recurringService) ListWithDateRange(ctx context.Context, startDate, endDate time.Time) ([]*RecurringTransaction, error) {
	query := s.client.loadQuery("recurring/list.graphql")

	variables := map[string]interface{}{
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
		// Don't include filters if it's null - the API doesn't like it
	}

	var result struct {
		RecurringTransactionItems []*RecurringTransactionItem `json:"recurringTransactionItems"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get recurring transactions")
	}

	// Transform RecurringTransactionItems into RecurringTransactions
	var transactions []*RecurringTransaction
	for _, item := range result.RecurringTransactionItems {
		transaction := &RecurringTransaction{
			ID:            item.Stream.ID,
			Merchant:      item.Stream.Merchant,
			Amount:        item.Amount,
			Frequency:     item.Stream.Frequency,
			NextDate:      item.Date,
			Category:      item.Category,
			Account:       item.Account,
			IsActive:      !item.IsPast,
			IsApproximate: item.Stream.IsApproximate,
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// RecurringTransactionItem represents a single recurring transaction item from the API
type RecurringTransactionItem struct {
	Stream struct {
		ID            string    `json:"id"`
		Frequency     string    `json:"frequency"`
		Amount        float64   `json:"amount"`
		IsApproximate bool      `json:"isApproximate"`
		Merchant      *Merchant `json:"merchant"`
	} `json:"stream"`
	Date          Date                 `json:"date"`
	IsPast        bool                 `json:"isPast"`
	TransactionID *string              `json:"transactionId"`
	Amount        float64              `json:"amount"`
	AmountDiff    *float64             `json:"amountDiff"`
	Account       *Account             `json:"account"`
	Category      *TransactionCategory `json:"category"`
}
