package monarch

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Security represents a tradable security (stock, ETF, crypto, etc.)
type Security struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Ticker       string  `json:"ticker"`
	CurrentPrice float64 `json:"currentPrice"`
}

// CreateHoldingParams for creating a manual holding
type CreateHoldingParams struct {
	AccountID  string  `json:"accountId"`
	SecurityID string  `json:"securityId"`
	Quantity   float64 `json:"quantity"`
}

// SearchSecurities searches for securities by ticker or name.
// The limit parameter controls the maximum number of results returned.
func (s *accountService) SearchSecurities(ctx context.Context, query string, limit int) ([]*Security, error) {
	if limit <= 0 {
		limit = 10
	}
	gql := s.client.loadQuery("accounts/search_securities.graphql")

	variables := map[string]interface{}{
		"search":            query,
		"limit":             limit,
		"orderByPopularity": true,
	}

	var result struct {
		Securities []*Security `json:"securities"`
	}

	if err := s.client.executeGraphQL(ctx, gql, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to search securities")
	}

	return result.Securities, nil
}

// CreateHolding creates a manual investment holding in an account.
func (s *accountService) CreateHolding(ctx context.Context, params *CreateHoldingParams) (*Holding, error) {
	gql := s.client.loadQuery("accounts/create_holding.graphql")

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"accountId":  params.AccountID,
			"securityId": params.SecurityID,
			"quantity":   params.Quantity,
		},
	}

	var result struct {
		CreateManualHolding struct {
			Holding *struct {
				ID     string `json:"id"`
				Ticker string `json:"ticker"`
			} `json:"holding"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"createManualHolding"`
	}

	if err := s.client.executeGraphQL(ctx, gql, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to create holding")
	}

	if len(result.CreateManualHolding.Errors) > 0 {
		return nil, &Error{
			Code:    result.CreateManualHolding.Errors[0].Code,
			Message: result.CreateManualHolding.Errors[0].Message,
		}
	}

	if result.CreateManualHolding.Holding == nil {
		return nil, errors.New("no holding returned from creation")
	}

	return &Holding{
		ID:        result.CreateManualHolding.Holding.ID,
		AccountID: params.AccountID,
		Symbol:    result.CreateManualHolding.Holding.Ticker,
		Quantity:  params.Quantity,
	}, nil
}

// CreateHoldingByTicker looks up a security by ticker and creates a holding.
func (s *accountService) CreateHoldingByTicker(ctx context.Context, accountID, ticker string, quantity float64) (*Holding, error) {
	securities, err := s.SearchSecurities(ctx, ticker, 5)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for ticker")
	}

	if len(securities) == 0 {
		return nil, fmt.Errorf("no security found for ticker: %s", ticker)
	}

	// Require exact case-insensitive ticker match
	var securityID string
	for _, sec := range securities {
		if strings.EqualFold(sec.Ticker, ticker) {
			securityID = sec.ID
			break
		}
	}
	if securityID == "" {
		return nil, fmt.Errorf("no exact ticker match for %q in search results", ticker)
	}

	return s.CreateHolding(ctx, &CreateHoldingParams{
		AccountID:  accountID,
		SecurityID: securityID,
		Quantity:   quantity,
	})
}

// DeleteHolding deletes a manual investment holding.
func (s *accountService) DeleteHolding(ctx context.Context, holdingID string) error {
	gql := s.client.loadQuery("accounts/delete_holding.graphql")

	variables := map[string]interface{}{
		"id": holdingID,
	}

	var result struct {
		DeleteHolding struct {
			Deleted bool `json:"deleted"`
			Errors  []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"deleteHolding"`
	}

	if err := s.client.executeGraphQL(ctx, gql, variables, &result); err != nil {
		return errors.Wrap(err, "failed to delete holding")
	}

	if len(result.DeleteHolding.Errors) > 0 {
		return &Error{
			Code:    result.DeleteHolding.Errors[0].Code,
			Message: result.DeleteHolding.Errors[0].Message,
		}
	}

	if !result.DeleteHolding.Deleted {
		return errors.New("holding was not deleted")
	}

	return nil
}

// UpdateHoldingQuantity updates a holding's quantity via the updateHolding mutation.
func (s *accountService) UpdateHoldingQuantity(ctx context.Context, accountID, holdingID string, newQuantity float64) (*Holding, error) {
	gql := s.client.loadQuery("accounts/update_holding.graphql")

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"id":       holdingID,
			"quantity": newQuantity,
		},
	}

	var result struct {
		UpdateHolding struct {
			Holding *struct {
				ID       string  `json:"id"`
				Quantity float64 `json:"quantity"`
			} `json:"holding"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"updateHolding"`
	}

	if err := s.client.executeGraphQL(ctx, gql, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to update holding")
	}

	if len(result.UpdateHolding.Errors) > 0 {
		return nil, &Error{
			Code:    result.UpdateHolding.Errors[0].Code,
			Message: result.UpdateHolding.Errors[0].Message,
		}
	}

	if result.UpdateHolding.Holding == nil {
		return nil, errors.New("no holding returned from update")
	}

	return &Holding{
		ID:        result.UpdateHolding.Holding.ID,
		AccountID: accountID,
		Quantity:  result.UpdateHolding.Holding.Quantity,
	}, nil
}
