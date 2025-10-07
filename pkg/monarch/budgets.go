package monarch

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// budgetService implements the BudgetService interface
type budgetService struct {
	client *Client
}

// List retrieves budgets for a date range
func (s *budgetService) List(ctx context.Context, startDate, endDate time.Time) ([]*Budget, error) {
	query := s.client.loadQuery("budgets/list.graphql")

	variables := map[string]interface{}{
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
	}

	var result struct {
		BudgetData *BudgetData `json:"budgetData"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get budgets")
	}

	// Convert the nested structure to flat Budget list
	// For now, return empty since the structure is different
	// TODO: Properly convert BudgetData to []*Budget
	if result.BudgetData == nil {
		return []*Budget{}, nil
	}

	// Convert to Budget format
	var budgets []*Budget
	for _, catBudget := range result.BudgetData.MonthlyAmountsByCategory {
		if catBudget.Category == nil {
			continue
		}

		// Create a budget entry for each month
		for _, monthly := range catBudget.MonthlyAmounts {
			budget := &Budget{
				CategoryID:     catBudget.Category.ID,
				Category:       catBudget.Category,
				Amount:         monthly.PlannedCashFlowAmount,
				Spent:          -monthly.ActualAmount, // Actual is negative for expenses
				Remaining:      monthly.RemainingAmount,
				Rollover:       monthly.RolloverType != "",
				RolloverType:   monthly.RolloverType,
				RolloverAmount: monthly.PreviousMonthRolloverAmount,
			}

			// Calculate percentage
			if budget.Amount > 0 {
				budget.PercentageComplete = (budget.Spent / budget.Amount) * 100
			}

			budgets = append(budgets, budget)
		}
	}

	return budgets, nil
}

// SetAmount sets budget amount
func (s *budgetService) SetAmount(ctx context.Context, budgetID string, amount float64, rollover bool, startDate time.Time) error {
	query := s.client.loadQuery("budgets/set_amount.graphql")

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"budgetId":  budgetID,
			"amount":    amount,
			"rollover":  rollover,
			"startDate": startDate.Format("2006-01-02"),
		},
	}

	var result struct {
		SetBudgetAmount struct {
			Budget *struct {
				ID       string  `json:"id"`
				Amount   float64 `json:"amount"`
				Rollover bool    `json:"rollover"`
			} `json:"budget"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"setBudgetAmount"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return errors.Wrap(err, "failed to set budget amount")
	}

	if len(result.SetBudgetAmount.Errors) > 0 {
		return &Error{
			Code:    result.SetBudgetAmount.Errors[0].Code,
			Message: result.SetBudgetAmount.Errors[0].Message,
		}
	}

	return nil
}
