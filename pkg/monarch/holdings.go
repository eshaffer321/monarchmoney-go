package monarch

import (
	"context"
	"fmt"

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
func (s *accountService) SearchSecurities(ctx context.Context, query string) ([]*Security, error) {
	gql := s.client.loadQuery("accounts/search_securities.graphql")

	variables := map[string]interface{}{
		"search":            query,
		"limit":             5,
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
	securities, err := s.SearchSecurities(ctx, ticker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for ticker")
	}

	if len(securities) == 0 {
		return nil, fmt.Errorf("no security found for ticker: %s", ticker)
	}

	// Find exact ticker match
	var securityID string
	for _, sec := range securities {
		if sec.Ticker == ticker {
			securityID = sec.ID
			break
		}
	}
	if securityID == "" {
		// Fall back to first result
		securityID = securities[0].ID
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

// UpdateHoldingQuantity updates a holding's quantity by deleting and recreating it.
// The Monarch API does not support direct quantity updates on holdings.
func (s *accountService) UpdateHoldingQuantity(ctx context.Context, accountID, holdingID string, newQuantity float64) (*Holding, error) {
	// Get the current holding to preserve the security info
	holdings, err := s.GetHoldings(ctx, accountID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get holdings")
	}

	var targetHolding *Holding
	for _, h := range holdings {
		if h.ID == holdingID {
			targetHolding = h
			break
		}
	}

	if targetHolding == nil {
		return nil, fmt.Errorf("holding not found: %s", holdingID)
	}

	// Delete the old holding
	if err := s.DeleteHolding(ctx, holdingID); err != nil {
		return nil, errors.Wrap(err, "failed to delete old holding")
	}

	// Recreate with the new quantity using the ticker
	return s.CreateHoldingByTicker(ctx, accountID, targetHolding.Symbol, newQuantity)
}
