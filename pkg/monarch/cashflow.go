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

	// Build filters according to the GraphQL schema
	filters := map[string]interface{}{
		"startDate":  params.StartDate.Format("2006-01-02"),
		"endDate":    params.EndDate.Format("2006-01-02"),
		"search":     "",
		"categories": []string{},
		"accounts":   []string{},
		"tags":       []string{},
	}

	if len(params.AccountIDs) > 0 {
		filters["accounts"] = params.AccountIDs
	}

	variables := map[string]interface{}{
		"filters": filters,
	}

	// The response structure matches the GraphQL query aliases
	var result struct {
		ByCategory []struct {
			GroupBy struct {
				Category *TransactionCategory `json:"category"`
			} `json:"groupBy"`
			Summary struct {
				Sum float64 `json:"sum"`
			} `json:"summary"`
		} `json:"byCategory"`
		ByCategoryGroup []struct {
			GroupBy struct {
				CategoryGroup *CategoryGroup `json:"categoryGroup"`
			} `json:"groupBy"`
			Summary struct {
				Sum float64 `json:"sum"`
			} `json:"summary"`
		} `json:"byCategoryGroup"`
		ByMerchant []struct {
			GroupBy struct {
				Merchant *Merchant `json:"merchant"`
			} `json:"groupBy"`
			Summary struct {
				Sum float64 `json:"sum"`
			} `json:"summary"`
		} `json:"byMerchant"`
		Summary []struct {
			Summary struct {
				Sum         float64 `json:"sum"`
				SumIncome   float64 `json:"sumIncome"`
				SumExpense  float64 `json:"sumExpense"`
				Savings     float64 `json:"savings"`
				SavingsRate float64 `json:"savingsRate"`
			} `json:"summary"`
		} `json:"summary"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get cashflow")
	}

	// Transform the response into our Cashflow structure
	cashflow := &Cashflow{
		StartDate:       params.StartDate,
		EndDate:         params.EndDate,
		ByCategory:      make([]*CashflowCategory, 0),
		ByCategoryGroup: make([]*CashflowCategoryGroup, 0),
		ByMerchant:      make([]*CashflowMerchant, 0),
	}

	// Extract summary if available
	if len(result.Summary) > 0 {
		cashflow.Summary = &CashflowSummary{
			Income:      result.Summary[0].Summary.SumIncome,
			Expense:     -result.Summary[0].Summary.SumExpense, // API returns negative
			Savings:     result.Summary[0].Summary.Savings,
			SavingsRate: result.Summary[0].Summary.SavingsRate,
		}
	}

	// Transform categories
	for _, cat := range result.ByCategory {
		cashflow.ByCategory = append(cashflow.ByCategory, &CashflowCategory{
			Category: cat.GroupBy.Category,
			Amount:   cat.Summary.Sum,
		})
	}

	// Transform category groups
	for _, grp := range result.ByCategoryGroup {
		cashflow.ByCategoryGroup = append(cashflow.ByCategoryGroup, &CashflowCategoryGroup{
			CategoryGroup: grp.GroupBy.CategoryGroup,
			Amount:        grp.Summary.Sum,
		})
	}

	// Transform merchants
	for _, merch := range result.ByMerchant {
		cashflow.ByMerchant = append(cashflow.ByMerchant, &CashflowMerchant{
			Merchant: merch.GroupBy.Merchant,
			Amount:   merch.Summary.Sum,
		})
	}

	return cashflow, nil
}

// GetSummary retrieves cashflow summary
func (s *cashflowService) GetSummary(ctx context.Context, params *CashflowSummaryParams) (*CashflowSummary, error) {
	query := s.client.loadQuery("cashflow/summary.graphql")

	// Build filters
	filters := map[string]interface{}{
		"startDate":  params.StartDate.Format("2006-01-02"),
		"endDate":    params.EndDate.Format("2006-01-02"),
		"search":     "",
		"categories": []string{},
		"accounts":   []string{},
		"tags":       []string{},
	}

	if params.CategoryID != "" {
		filters["categories"] = []string{params.CategoryID}
	}

	if len(params.AccountsFilter) > 0 {
		filters["accounts"] = params.AccountsFilter
	}

	variables := map[string]interface{}{
		"filters": filters,
	}

	// The API returns an array with a single element containing the summary
	var result struct {
		Summary []struct {
			Summary struct {
				SumIncome   float64 `json:"sumIncome"`
				SumExpense  float64 `json:"sumExpense"`
				Savings     float64 `json:"savings"`
				SavingsRate float64 `json:"savingsRate"`
			} `json:"summary"`
		} `json:"summary"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get cashflow summary")
	}

	// Check if we got any results
	if len(result.Summary) == 0 {
		return &CashflowSummary{
			StartDate: params.StartDate,
			EndDate:   params.EndDate,
		}, nil
	}

	return &CashflowSummary{
		Income:      result.Summary[0].Summary.SumIncome,
		Expense:     -result.Summary[0].Summary.SumExpense, // API returns negative
		Savings:     result.Summary[0].Summary.Savings,
		SavingsRate: result.Summary[0].Summary.SavingsRate,
		StartDate:   params.StartDate,
		EndDate:     params.EndDate,
	}, nil
}
