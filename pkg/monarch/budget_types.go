package monarch

// BudgetData represents the budget data response
type BudgetData struct {
	MonthlyAmountsByCategory []*BudgetCategoryMonthly `json:"monthlyAmountsByCategory"`
}

// BudgetCategoryMonthly represents budget data for a category
type BudgetCategoryMonthly struct {
	Category       *TransactionCategory   `json:"category"`
	MonthlyAmounts []*BudgetMonthlyAmount `json:"monthlyAmounts"`
}

// BudgetMonthlyAmount represents budget amounts for a specific month
type BudgetMonthlyAmount struct {
	Month                       string  `json:"month"`
	PlannedCashFlowAmount       float64 `json:"plannedCashFlowAmount"`
	PlannedSetAsideAmount       float64 `json:"plannedSetAsideAmount"`
	ActualAmount                float64 `json:"actualAmount"`
	RemainingAmount             float64 `json:"remainingAmount"`
	PreviousMonthRolloverAmount float64 `json:"previousMonthRolloverAmount"`
	RolloverType                string  `json:"rolloverType"`
}
