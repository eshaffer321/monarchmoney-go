package monarch

import (
	"context"

	"github.com/pkg/errors"
)

// cashflowService implements the CashflowService interface
type cashflowService struct {
	client *Client
}

// Get retrieves cashflow data
func (s *cashflowService) Get(ctx context.Context, params *CashflowParams) (*Cashflow, error) {
	query := s.client.loadQuery("cashflow/get.graphql")

	variables := map[string]interface{}{
		"startDate": params.StartDate.Format("2006-01-02"),
		"endDate":   params.EndDate.Format("2006-01-02"),
	}

	if params.Limit > 0 {
		variables["limit"] = params.Limit
	}

	if len(params.AccountIDs) > 0 {
		variables["accountIds"] = params.AccountIDs
	}

	var result struct {
		Cashflow *Cashflow `json:"cashflow"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get cashflow")
	}

	return result.Cashflow, nil
}

// GetSummary retrieves cashflow summary
func (s *cashflowService) GetSummary(ctx context.Context, params *CashflowSummaryParams) (*CashflowSummary, error) {
	query := s.client.loadQuery("cashflow/summary.graphql")

	variables := map[string]interface{}{
		"startDate": params.StartDate.Format("2006-01-02"),
		"endDate":   params.EndDate.Format("2006-01-02"),
		"interval":  params.Interval,
	}

	if params.CategoryID != "" {
		variables["categoryId"] = params.CategoryID
	}

	if len(params.AccountsFilter) > 0 {
		variables["accountsFilter"] = params.AccountsFilter
	}

	var result struct {
		CashflowSummary *CashflowSummary `json:"cashflowSummary"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get cashflow summary")
	}

	return result.CashflowSummary, nil
}
