package monarch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// transactionService implements the TransactionService interface
type transactionService struct {
	client     *Client
	categories TransactionCategoryService
}

// newTransactionService creates a new transaction service
func newTransactionService(client *Client) *transactionService {
	s := &transactionService{
		client: client,
	}
	s.categories = &transactionCategoryService{client: client}
	return s
}

// Query returns a transaction query builder
func (s *transactionService) Query() TransactionQueryBuilder {
	return &transactionQueryBuilder{
		client:  s.client,
		filters: make(map[string]interface{}),
		limit:   100,
		offset:  0,
	}
}

// Get retrieves a single transaction
func (s *transactionService) Get(ctx context.Context, transactionID string) (*TransactionDetails, error) {
	query := s.client.loadQuery("transactions/get.graphql")

	variables := map[string]interface{}{
		"id": transactionID,
	}

	var result struct {
		GetTransaction *TransactionDetails `json:"getTransaction"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get transaction")
	}

	if result.GetTransaction == nil {
		return nil, ErrNotFound
	}

	return result.GetTransaction, nil
}

// Create creates a new transaction
func (s *transactionService) Create(ctx context.Context, params *CreateTransactionParams) (*Transaction, error) {
	query := s.client.loadQuery("transactions/create.graphql")

	input := map[string]interface{}{
		"date":       params.Date.Format("2006-01-02"),
		"accountId":  params.AccountID,
		"amount":     params.Amount,
		"merchant":   params.Merchant,
		"categoryId": params.CategoryID,
	}

	if params.Notes != "" {
		input["notes"] = params.Notes
	}

	variables := map[string]interface{}{
		"input": input,
	}

	var result struct {
		CreateTransaction struct {
			Transaction *struct {
				ID string `json:"id"`
			} `json:"transaction"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"createTransaction"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to create transaction")
	}

	if len(result.CreateTransaction.Errors) > 0 {
		return nil, &Error{
			Code:    result.CreateTransaction.Errors[0].Code,
			Message: result.CreateTransaction.Errors[0].Message,
		}
	}

	if result.CreateTransaction.Transaction == nil {
		return nil, errors.New("no transaction returned from creation")
	}

	// Fetch the full transaction details
	details, err := s.Get(ctx, result.CreateTransaction.Transaction.ID)
	if err != nil {
		return nil, err
	}

	// Convert details to transaction
	return &Transaction{
		ID:                 details.ID,
		Date:               details.Date,
		Amount:             details.Amount,
		Pending:            details.Pending,
		HideFromReports:    details.HideFromReports,
		PlaidName:          details.PlaidName,
		Merchant:           details.Merchant,
		Notes:              details.Notes,
		HasSplits:          details.HasSplits,
		IsSplitTransaction: details.IsSplitTransaction,
		IsRecurring:        details.IsRecurring,
		NeedsReview:        details.NeedsReview,
		ReviewedAt:         details.ReviewedAt,
		CreatedAt:          details.CreatedAt,
		UpdatedAt:          details.UpdatedAt,
		Account:            details.Account,
		Category:           details.Category,
		Tags:               details.Tags,
	}, nil
}

// Update updates an existing transaction
func (s *transactionService) Update(ctx context.Context, transactionID string, params *UpdateTransactionParams) (*Transaction, error) {
	query := s.client.loadQuery("transactions/update.graphql")

	input := map[string]interface{}{
		"id": transactionID,
	}

	if params.Date != nil {
		input["date"] = params.Date.Format("2006-01-02")
	}
	if params.AccountID != nil {
		input["accountId"] = *params.AccountID
	}
	if params.Amount != nil {
		input["amount"] = *params.Amount
	}
	if params.Merchant != nil {
		input["name"] = *params.Merchant  // Field name is "name" not "merchant"
	}
	if params.CategoryID != nil {
		input["category"] = *params.CategoryID  // Field name is "category" not "categoryId"
	}
	if params.Notes != nil {
		input["notes"] = *params.Notes
	}
	if params.HideFromReports != nil {
		input["hideFromReports"] = *params.HideFromReports
	}
	if params.NeedsReview != nil {
		input["needsReview"] = *params.NeedsReview
	}

	variables := map[string]interface{}{
		"input": input,
	}

	var result struct {
		UpdateTransaction struct {
			Transaction *Transaction `json:"transaction"`
			Errors      []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"updateTransaction"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to update transaction")
	}

	if len(result.UpdateTransaction.Errors) > 0 {
		return nil, &Error{
			Code:    result.UpdateTransaction.Errors[0].Code,
			Message: result.UpdateTransaction.Errors[0].Message,
		}
	}

	return result.UpdateTransaction.Transaction, nil
}

// Delete deletes a transaction
func (s *transactionService) Delete(ctx context.Context, transactionID string) error {
	query := s.client.loadQuery("transactions/delete.graphql")

	variables := map[string]interface{}{
		"id": transactionID,
	}

	var result struct {
		DeleteTransaction struct {
			Deleted bool `json:"deleted"`
			Errors  []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"deleteTransaction"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return errors.Wrap(err, "failed to delete transaction")
	}

	if len(result.DeleteTransaction.Errors) > 0 {
		return &Error{
			Code:    result.DeleteTransaction.Errors[0].Code,
			Message: result.DeleteTransaction.Errors[0].Message,
		}
	}

	if !result.DeleteTransaction.Deleted {
		return errors.New("transaction was not deleted")
	}

	return nil
}

// GetSummary retrieves transaction summary
func (s *transactionService) GetSummary(ctx context.Context) (*TransactionSummary, error) {
	query := s.client.loadQuery("transactions/summary.graphql")

	// The API expects a filters parameter, even if empty
	variables := map[string]interface{}{
		"filters": map[string]interface{}{},
	}

	// The API returns aggregates as an array
	var result struct {
		Aggregates []struct {
			Summary *TransactionSummary `json:"summary"`
		} `json:"aggregates"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get transactions summary")
	}

	if len(result.Aggregates) == 0 || result.Aggregates[0].Summary == nil {
		return nil, errors.New("no transaction summary data available")
	}

	return result.Aggregates[0].Summary, nil
}

// GetSplits retrieves transaction splits
func (s *transactionService) GetSplits(ctx context.Context, transactionID string) ([]*TransactionSplit, error) {
	query := s.client.loadQuery("transactions/get_splits.graphql")

	variables := map[string]interface{}{
		"id": transactionID,
	}

	var result struct {
		GetTransaction struct {
			ID               string              `json:"id"`
			Amount           float64             `json:"amount"`
			Category         *TransactionCategory `json:"category"`
			Merchant         *Merchant           `json:"merchant"`
			SplitTransactions []*TransactionSplit `json:"splitTransactions"`
		} `json:"getTransaction"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get transaction splits")
	}

	return result.GetTransaction.SplitTransactions, nil
}

// UpdateSplits updates transaction splits
func (s *transactionService) UpdateSplits(ctx context.Context, transactionID string, splits []*TransactionSplit) error {
	query := s.client.loadQuery("transactions/update_splits.graphql")

	// Build split data according to Python client format
	splitData := make([]map[string]interface{}, len(splits))
	for i, split := range splits {
		splitData[i] = map[string]interface{}{
			"amount":     split.Amount,
			"categoryId": split.CategoryID,
		}
		
		// Add optional fields if present
		if split.Notes != "" {
			splitData[i]["notes"] = split.Notes
		}
		if split.Merchant != nil && split.Merchant.Name != "" {
			splitData[i]["merchantName"] = split.Merchant.Name
		}
	}

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"transactionId": transactionID,
			"splitData":     splitData,
		},
	}

	// Use RawMessage for errors field to handle both single object and array
	var result struct {
		UpdateTransactionSplit struct {
			Transaction *struct {
				ID                   string `json:"id"`
				HasSplitTransactions bool   `json:"hasSplitTransactions"`
				SplitTransactions    []*TransactionSplit `json:"splitTransactions"`
			} `json:"transaction"`
			Errors json.RawMessage `json:"errors"`
		} `json:"updateTransactionSplit"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return errors.Wrap(err, "failed to update transaction splits")
	}

	// Handle errors field which can be either a single object or an array
	if len(result.UpdateTransactionSplit.Errors) > 0 {
		// Skip if it's an empty array: []
		if string(result.UpdateTransactionSplit.Errors) == "[]" {
			return nil
		}

		// Define the error structure
		type errorStruct struct {
			Message string `json:"message"`
			Code    string `json:"code"`
			FieldErrors []struct {
				Field    string   `json:"field"`
				Messages []string `json:"messages"`
			} `json:"fieldErrors"`
		}

		// Try to unmarshal as a single error object first
		var singleError errorStruct
		if err := json.Unmarshal(result.UpdateTransactionSplit.Errors, &singleError); err == nil && singleError.Message != "" {
			// Build detailed error message including field errors
			errorMsg := singleError.Message
			if len(singleError.FieldErrors) > 0 {
				errorMsg += " ("
				for i, fe := range singleError.FieldErrors {
					if i > 0 {
						errorMsg += "; "
					}
					errorMsg += fe.Field + ": " + strings.Join(fe.Messages, ", ")
				}
				errorMsg += ")"
			}
			
			return &Error{
				Code:    singleError.Code,
				Message: errorMsg,
			}
		}

		// Try to unmarshal as an array of errors
		var errorArray []errorStruct
		if err := json.Unmarshal(result.UpdateTransactionSplit.Errors, &errorArray); err == nil && len(errorArray) > 0 {
			// Use the first error from the array
			firstError := errorArray[0]
			
			// Build detailed error message including field errors
			errorMsg := firstError.Message
			if len(firstError.FieldErrors) > 0 {
				errorMsg += " ("
				for i, fe := range firstError.FieldErrors {
					if i > 0 {
						errorMsg += "; "
					}
					errorMsg += fe.Field + ": " + strings.Join(fe.Messages, ", ")
				}
				errorMsg += ")"
			}
			
			return &Error{
				Code:    firstError.Code,
				Message: errorMsg,
			}
		}
	}

	return nil
}

// Categories returns the category sub-service
func (s *transactionService) Categories() TransactionCategoryService {
	return s.categories
}

// transactionQueryBuilder implements TransactionQueryBuilder
type transactionQueryBuilder struct {
	client    *Client
	filters   map[string]interface{}
	limit     int
	offset    int
	orderBy   string
	minAmount float64
	maxAmount float64
}

// Between sets date range filter
func (b *transactionQueryBuilder) Between(start, end time.Time) TransactionQueryBuilder {
	b.filters["startDate"] = start.Format("2006-01-02")
	b.filters["endDate"] = end.Format("2006-01-02")
	return b
}

// WithAccounts filters by account IDs
func (b *transactionQueryBuilder) WithAccounts(accountIDs ...string) TransactionQueryBuilder {
	b.filters["accounts"] = accountIDs
	return b
}

// WithCategories filters by category IDs
func (b *transactionQueryBuilder) WithCategories(categoryIDs ...string) TransactionQueryBuilder {
	b.filters["categories"] = categoryIDs
	return b
}

// WithTags filters by tag IDs
func (b *transactionQueryBuilder) WithTags(tagIDs ...string) TransactionQueryBuilder {
	b.filters["tags"] = tagIDs
	return b
}

// WithMinAmount sets minimum amount filter
// NOTE: This filter is applied client-side as the GraphQL API may not support it directly
func (b *transactionQueryBuilder) WithMinAmount(amount float64) TransactionQueryBuilder {
	// Store for client-side filtering
	b.minAmount = amount
	return b
}

// WithMaxAmount sets maximum amount filter
// NOTE: This filter is applied client-side as the GraphQL API may not support it directly
func (b *transactionQueryBuilder) WithMaxAmount(amount float64) TransactionQueryBuilder {
	// Store for client-side filtering
	b.maxAmount = amount
	return b
}

// WithMerchant filters by merchant name
func (b *transactionQueryBuilder) WithMerchant(merchant string) TransactionQueryBuilder {
	b.filters["merchant"] = merchant
	return b
}

// Search sets search filter
func (b *transactionQueryBuilder) Search(query string) TransactionQueryBuilder {
	b.filters["search"] = query
	return b
}

// Limit sets result limit
func (b *transactionQueryBuilder) Limit(limit int) TransactionQueryBuilder {
	b.limit = limit
	return b
}

// Offset sets result offset
func (b *transactionQueryBuilder) Offset(offset int) TransactionQueryBuilder {
	b.offset = offset
	return b
}

// Execute runs the query
func (b *transactionQueryBuilder) Execute(ctx context.Context) (*TransactionList, error) {
	query := b.client.loadQuery("transactions/list.graphql")

	variables := map[string]interface{}{
		"offset":  b.offset,
		"limit":   b.limit,
		"filters": b.filters,
		"orderBy": "date",
	}

	var result struct {
		AllTransactions struct {
			TotalCount int            `json:"totalCount"`
			Results    []*Transaction `json:"results"`
		} `json:"allTransactions"`
	}

	if err := b.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get transactions")
	}

	// Apply client-side filtering for amount if needed
	transactions := result.AllTransactions.Results
	if b.minAmount > 0 || b.maxAmount > 0 {
		filtered := make([]*Transaction, 0, len(transactions))
		for _, txn := range transactions {
			// Convert to absolute value for comparison
			absAmount := txn.Amount
			if absAmount < 0 {
				absAmount = -absAmount
			}

			if b.minAmount > 0 && absAmount < b.minAmount {
				continue
			}
			if b.maxAmount > 0 && absAmount > b.maxAmount {
				continue
			}
			filtered = append(filtered, txn)
		}
		transactions = filtered
	}

	hasMore := (b.offset + b.limit) < result.AllTransactions.TotalCount

	return &TransactionList{
		Transactions: transactions,
		TotalCount:   result.AllTransactions.TotalCount, // Keep original count
		HasMore:      hasMore,
		NextOffset:   b.offset + b.limit,
	}, nil
}

// Stream returns results as a channel for large queries
func (b *transactionQueryBuilder) Stream(ctx context.Context) (<-chan *Transaction, <-chan error) {
	txnChan := make(chan *Transaction)
	errChan := make(chan error, 1)

	go func() {
		defer close(txnChan)
		defer close(errChan)

		offset := b.offset
		limit := b.limit
		if limit > 100 {
			limit = 100 // Use smaller batches for streaming
		}

		for {
			// Create a copy of builder with current offset
			queryBuilder := &transactionQueryBuilder{
				client:  b.client,
				filters: b.filters,
				limit:   limit,
				offset:  offset,
				orderBy: b.orderBy,
			}

			// Execute query
			result, err := queryBuilder.Execute(ctx)
			if err != nil {
				errChan <- err
				return
			}

			// Send transactions to channel
			for _, txn := range result.Transactions {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case txnChan <- txn:
				}
			}

			// Check if we have more results
			if !result.HasMore {
				break
			}

			offset = result.NextOffset
		}
	}()

	return txnChan, errChan
}

// transactionCategoryService implements TransactionCategoryService
type transactionCategoryService struct {
	client *Client
}

// List retrieves all categories
func (s *transactionCategoryService) List(ctx context.Context) ([]*TransactionCategory, error) {
	query := s.client.loadQuery("categories/list.graphql")

	var result struct {
		Categories []*TransactionCategory `json:"categories"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get transaction categories")
	}

	return result.Categories, nil
}

// Create creates a new category
func (s *transactionCategoryService) Create(ctx context.Context, params *CreateCategoryParams) (*TransactionCategory, error) {
	query := s.client.loadQuery("categories/create.graphql")

	// Build input according to Python client
	input := map[string]interface{}{
		"name":  params.Name,
		"group": params.GroupID,
		"icon":  params.Icon,
	}

	// Default icon if not provided
	if params.Icon == "" {
		input["icon"] = "❓" // Unicode for question mark
	}

	// Add rollover settings if provided
	if params.RolloverEnabled {
		input["rolloverEnabled"] = true
		input["rolloverType"] = params.RolloverType
		if params.RolloverType == "" {
			input["rolloverType"] = "monthly"
		}
		if !params.RolloverStartMonth.IsZero() {
			input["rolloverStartMonth"] = params.RolloverStartMonth.Format("2006-01-02")
		} else {
			// Default to first of current month
			now := time.Now()
			input["rolloverStartMonth"] = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		}
	}

	variables := map[string]interface{}{
		"input": input,
	}

	var result struct {
		CreateCategory struct {
			Category *TransactionCategory `json:"category"`
			Errors   []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"createCategory"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to create category")
	}

	if len(result.CreateCategory.Errors) > 0 {
		return nil, &Error{
			Code:    result.CreateCategory.Errors[0].Code,
			Message: result.CreateCategory.Errors[0].Message,
		}
	}

	return result.CreateCategory.Category, nil
}

// Delete deletes a category
func (s *transactionCategoryService) Delete(ctx context.Context, categoryID string) error {
	query := s.client.loadQuery("categories/delete.graphql")

	variables := map[string]interface{}{
		"id": categoryID,
	}

	var result struct {
		DeleteCategory struct {
			Deleted bool `json:"deleted"`
			Errors  []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"deleteCategory"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return errors.Wrap(err, "failed to delete category")
	}

	if len(result.DeleteCategory.Errors) > 0 {
		return &Error{
			Code:    result.DeleteCategory.Errors[0].Code,
			Message: result.DeleteCategory.Errors[0].Message,
		}
	}

	if !result.DeleteCategory.Deleted {
		return errors.New("category was not deleted")
	}

	return nil
}

// DeleteMultiple deletes multiple categories
func (s *transactionCategoryService) DeleteMultiple(ctx context.Context, categoryIDs ...string) error {
	for _, id := range categoryIDs {
		if err := s.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete category %s: %w", id, err)
		}
	}
	return nil
}

// GetGroups retrieves category groups
func (s *transactionCategoryService) GetGroups(ctx context.Context) ([]*CategoryGroup, error) {
	query := s.client.loadQuery("categories/groups.graphql")

	var result struct {
		CategoryGroups []*CategoryGroup `json:"categoryGroups"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get category groups")
	}

	return result.CategoryGroups, nil
}
