package monarch

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// GetCashflowSimple retrieves basic cashflow summary
func (s *cashflowService) GetSimple(ctx context.Context, startDate, endDate time.Time) (*CashflowSummary, error) {
	query := s.client.loadQuery("cashflow/test.graphql")

	variables := map[string]interface{}{
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
	}

	var result struct {
		Aggregates []struct {
			Summary struct {
				SumIncome   float64 `json:"sumIncome"`
				SumExpense  float64 `json:"sumExpense"`
				Savings     float64 `json:"savings"`
				SavingsRate float64 `json:"savingsRate"`
			} `json:"summary"`
		} `json:"aggregates"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get simple cashflow")
	}

	if len(result.Aggregates) == 0 {
		return &CashflowSummary{
			StartDate: startDate,
			EndDate:   endDate,
		}, nil
	}

	return &CashflowSummary{
		Income:      result.Aggregates[0].Summary.SumIncome,
		Expense:     -result.Aggregates[0].Summary.SumExpense,
		Savings:     result.Aggregates[0].Summary.Savings,
		SavingsRate: result.Aggregates[0].Summary.SavingsRate,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}
