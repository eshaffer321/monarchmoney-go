package monarch

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// accountService implements the AccountService interface
type accountService struct {
	client *Client
}

// List retrieves all accounts
func (s *accountService) List(ctx context.Context) ([]*Account, error) {
	query := s.client.loadQuery("accounts/list.graphql")

	var result struct {
		Accounts []*Account `json:"accounts"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	return result.Accounts, nil
}

// Get retrieves a single account by ID
func (s *accountService) Get(ctx context.Context, accountID string) (*Account, error) {
	accounts, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrNotFound
}

// Create creates a new manual account
func (s *accountService) Create(ctx context.Context, params *CreateAccountParams) (*Account, error) {
	query := s.client.loadQuery("accounts/create.graphql")

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"type":              params.AccountType,
			"subtype":           params.AccountSubtype,
			"includeInNetWorth": params.IncludeInNetWorth,
			"name":              params.AccountName,
			"displayBalance":    params.CurrentBalance,
		},
	}

	var result struct {
		CreateManualAccount struct {
			Account *struct {
				ID string `json:"id"`
			} `json:"account"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"createManualAccount"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to create account")
	}

	if len(result.CreateManualAccount.Errors) > 0 {
		return nil, &Error{
			Code:    result.CreateManualAccount.Errors[0].Code,
			Message: result.CreateManualAccount.Errors[0].Message,
		}
	}

	if result.CreateManualAccount.Account == nil {
		return nil, errors.New("no account returned from creation")
	}

	// Fetch the full account details
	return s.Get(ctx, result.CreateManualAccount.Account.ID)
}

// Update updates an existing account
func (s *accountService) Update(ctx context.Context, accountID string, params *UpdateAccountParams) (*Account, error) {
	query := s.client.loadQuery("accounts/update.graphql")

	input := map[string]interface{}{
		"id": accountID,
	}

	if params.DisplayName != nil {
		input["name"] = *params.DisplayName
	}
	if params.CurrentBalance != nil {
		input["displayBalance"] = *params.CurrentBalance
	}
	if params.IncludeInNetWorth != nil {
		input["includeInNetWorth"] = *params.IncludeInNetWorth
	}
	if params.HideFromList != nil {
		input["hideFromSummaryList"] = *params.HideFromList
	}
	if params.HideTransactionsFromReports != nil {
		input["hideTransactionsFromReports"] = *params.HideTransactionsFromReports
	}

	variables := map[string]interface{}{
		"input": input,
	}

	var result struct {
		UpdateAccount struct {
			Account *Account `json:"account"`
			Errors  []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"updateAccount"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to update account")
	}

	if len(result.UpdateAccount.Errors) > 0 {
		return nil, &Error{
			Code:    result.UpdateAccount.Errors[0].Code,
			Message: result.UpdateAccount.Errors[0].Message,
		}
	}

	return result.UpdateAccount.Account, nil
}

// Delete deletes an account
func (s *accountService) Delete(ctx context.Context, accountID string) error {
	query := s.client.loadQuery("accounts/delete.graphql")

	variables := map[string]interface{}{
		"id": accountID,
	}

	var result struct {
		DeleteAccount struct {
			Deleted bool `json:"deleted"`
			Errors  []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"deleteAccount"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return errors.Wrap(err, "failed to delete account")
	}

	if len(result.DeleteAccount.Errors) > 0 {
		return &Error{
			Code:    result.DeleteAccount.Errors[0].Code,
			Message: result.DeleteAccount.Errors[0].Message,
		}
	}

	if !result.DeleteAccount.Deleted {
		return errors.New("account was not deleted")
	}

	return nil
}

// GetTypes returns available account types and subtypes
func (s *accountService) GetTypes(ctx context.Context) ([]*AccountType, error) {
	query := s.client.loadQuery("accounts/types.graphql")

	var result struct {
		AccountTypeOptions []*AccountType `json:"accountTypeOptions"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get account types")
	}

	return result.AccountTypeOptions, nil
}

// GetBalances retrieves recent balance history
func (s *accountService) GetBalances(ctx context.Context, startDate *time.Time) ([]*AccountBalance, error) {
	if startDate == nil {
		defaultStart := time.Now().AddDate(0, 0, -31)
		startDate = &defaultStart
	}

	query := s.client.loadQuery("accounts/balances.graphql")

	variables := map[string]interface{}{
		"startDate": startDate.Format("2006-01-02"),
	}

	var result struct {
		Accounts []struct {
			ID             string                   `json:"id"`
			RecentBalances []map[string]interface{} `json:"recentBalances"`
		} `json:"accounts"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get account balances")
	}

	// Transform the result into AccountBalance structs
	var balances []*AccountBalance
	for _, account := range result.Accounts {
		for _, balance := range account.RecentBalances {
			if dateStr, ok := balance["date"].(string); ok {
				if balanceVal, ok := balance["balance"].(float64); ok {
					date, _ := time.Parse("2006-01-02", dateStr)
					balances = append(balances, &AccountBalance{
						AccountID: account.ID,
						Date:      date,
						Balance:   balanceVal,
					})
				}
			}
		}
	}

	return balances, nil
}

// GetSnapshots retrieves account snapshots by type
func (s *accountService) GetSnapshots(ctx context.Context, params *SnapshotParams) ([]*AccountSnapshot, error) {
	if params.Timeframe != "year" && params.Timeframe != "month" {
		return nil, fmt.Errorf("invalid timeframe: %s (must be 'year' or 'month')", params.Timeframe)
	}

	query := s.client.loadQuery("accounts/snapshots.graphql")

	variables := map[string]interface{}{
		"startDate": params.StartDate.Format("2006-01-02"),
		"timeframe": params.Timeframe,
	}

	var result struct {
		SnapshotsByAccountType []struct {
			Month          string  `json:"month"`
			AccountType    string  `json:"accountType"`
			AccountSubtype string  `json:"accountSubtype"`
			Sum            float64 `json:"sum"`
		} `json:"snapshotsByAccountType"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get account snapshots")
	}

	// Transform results
	var snapshots []*AccountSnapshot
	for _, s := range result.SnapshotsByAccountType {
		year := 0
		if len(s.Month) >= 4 {
			_, _ = fmt.Sscanf(s.Month[:4], "%d", &year)
		}

		snapshots = append(snapshots, &AccountSnapshot{
			Month:      s.Month,
			Year:       year,
			Type:       s.AccountType,
			Subtype:    s.AccountSubtype,
			TotalValue: s.Sum,
		})
	}

	return snapshots, nil
}

// GetHistory retrieves full account history
func (s *accountService) GetHistory(ctx context.Context, accountID string) (*AccountHistory, error) {
	query := s.client.loadQuery("accounts/history.graphql")

	variables := map[string]interface{}{
		"accountId": accountID,
	}

	var result struct {
		Account struct {
			ID             string `json:"id"`
			BalanceHistory []struct {
				Date    string  `json:"date"`
				Balance float64 `json:"balance"`
			} `json:"balanceHistory"`
		} `json:"account"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get account history")
	}

	history := &AccountHistory{
		AccountID: accountID,
		Balances:  make([]*BalanceEntry, 0),
	}

	for _, entry := range result.Account.BalanceHistory {
		date, _ := time.Parse("2006-01-02", entry.Date)
		history.Balances = append(history.Balances, &BalanceEntry{
			Date:    date,
			Balance: entry.Balance,
			Synced:  true,
		})
	}

	return history, nil
}

// GetHoldings retrieves investment holdings for an account
func (s *accountService) GetHoldings(ctx context.Context, accountID string) ([]*Holding, error) {
	query := s.client.loadQuery("accounts/holdings.graphql")

	variables := map[string]interface{}{
		"accountId": accountID,
	}

	var result struct {
		Account struct {
			ID       string `json:"id"`
			Holdings struct {
				Edges []struct {
					Node struct {
						ID        string    `json:"id"`
						Symbol    string    `json:"symbol"`
						Quantity  float64   `json:"quantity"`
						Price     float64   `json:"price"`
						Value     float64   `json:"value"`
						CostBasis float64   `json:"costBasis"`
						UpdatedAt time.Time `json:"updatedAt"`
						Holding   struct {
							Name string `json:"name"`
						} `json:"holding"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"holdings"`
		} `json:"account"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get account holdings")
	}

	var holdings []*Holding
	for _, edge := range result.Account.Holdings.Edges {
		holdings = append(holdings, &Holding{
			ID:        edge.Node.ID,
			AccountID: accountID,
			Symbol:    edge.Node.Symbol,
			Name:      edge.Node.Holding.Name,
			Quantity:  edge.Node.Quantity,
			Price:     edge.Node.Price,
			Value:     edge.Node.Value,
			CostBasis: edge.Node.CostBasis,
			UpdatedAt: edge.Node.UpdatedAt,
		})
	}

	return holdings, nil
}

// Refresh triggers a refresh for specified accounts
func (s *accountService) Refresh(ctx context.Context, accountIDs ...string) (RefreshJob, error) {
	query := s.client.loadQuery("accounts/refresh.graphql")

	variables := map[string]interface{}{
		"accountIds": accountIDs,
	}

	var result struct {
		RequestAccountsRefresh bool `json:"requestAccountsRefresh"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to request accounts refresh")
	}

	if !result.RequestAccountsRefresh {
		return nil, errors.New("refresh request was not accepted")
	}

	// Create a refresh job with proper initialization
	job := newRefreshJob(s.client, accountIDs)

	return job, nil
}

// RefreshAndWait triggers refresh and waits for completion
func (s *accountService) RefreshAndWait(ctx context.Context, timeout time.Duration, accountIDs ...string) error {
	job, err := s.Refresh(ctx, accountIDs...)
	if err != nil {
		return err
	}

	return job.Wait(ctx, timeout)
}

// IsRefreshComplete checks if refresh is complete for accounts
func (s *accountService) IsRefreshComplete(ctx context.Context, accountIDs ...string) (bool, error) {
	query := s.client.loadQuery("accounts/is_refresh_complete.graphql")

	variables := map[string]interface{}{
		"accountIds": accountIDs,
	}

	var result struct {
		Accounts []struct {
			ID         string `json:"id"`
			Credential *struct {
				UpdateRequired                 bool       `json:"updateRequired"`
				DisconnectedFromDataProviderAt *time.Time `json:"disconnectedFromDataProviderAt"`
			} `json:"credential"`
		} `json:"accounts"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return false, errors.Wrap(err, "failed to check refresh status")
	}

	// Check if any account is still updating
	for _, account := range result.Accounts {
		if account.Credential != nil && account.Credential.UpdateRequired {
			return false, nil
		}
	}

	return true, nil
}
