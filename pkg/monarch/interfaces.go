package monarch

import (
	"context"
	"time"
)

// AccountService handles all account-related operations
type AccountService interface {
	// List retrieves all accounts
	List(ctx context.Context) ([]*Account, error)

	// Get retrieves a single account by ID
	Get(ctx context.Context, accountID string) (*Account, error)

	// Create creates a new manual account
	Create(ctx context.Context, params *CreateAccountParams) (*Account, error)

	// Update updates an existing account
	Update(ctx context.Context, accountID string, params *UpdateAccountParams) (*Account, error)

	// Delete deletes an account
	Delete(ctx context.Context, accountID string) error

	// GetTypes returns available account types and subtypes
	GetTypes(ctx context.Context) ([]*AccountType, error)

	// GetBalances retrieves recent balance history
	GetBalances(ctx context.Context, startDate *time.Time) ([]*AccountBalance, error)

	// GetSnapshots retrieves account snapshots by type
	GetSnapshots(ctx context.Context, params *SnapshotParams) ([]*AccountSnapshot, error)

	// GetHistory retrieves full account history
	GetHistory(ctx context.Context, accountID string) (*AccountHistory, error)

	// GetHoldings retrieves investment holdings for an account
	GetHoldings(ctx context.Context, accountID string) ([]*Holding, error)

	// Refresh triggers a refresh for specified accounts
	Refresh(ctx context.Context, accountIDs ...string) (RefreshJob, error)

	// RefreshAndWait triggers refresh and waits for completion
	RefreshAndWait(ctx context.Context, timeout time.Duration, accountIDs ...string) error

	// IsRefreshComplete checks if refresh is complete for accounts
	IsRefreshComplete(ctx context.Context, accountIDs ...string) (bool, error)
}

// TransactionService handles all transaction-related operations
type TransactionService interface {
	// Query returns a transaction query builder
	Query() TransactionQueryBuilder

	// Get retrieves a single transaction
	Get(ctx context.Context, transactionID string) (*TransactionDetails, error)

	// Create creates a new transaction
	Create(ctx context.Context, params *CreateTransactionParams) (*Transaction, error)

	// Update updates an existing transaction
	Update(ctx context.Context, transactionID string, params *UpdateTransactionParams) (*Transaction, error)

	// Delete deletes a transaction
	Delete(ctx context.Context, transactionID string) error

	// GetSummary retrieves transaction summary
	GetSummary(ctx context.Context) (*TransactionSummary, error)

	// GetSplits retrieves transaction splits
	GetSplits(ctx context.Context, transactionID string) ([]*TransactionSplit, error)

	// UpdateSplits updates transaction splits
	UpdateSplits(ctx context.Context, transactionID string, splits []*TransactionSplit) error

	// Categories returns the category sub-service
	Categories() TransactionCategoryService
}

// TransactionCategoryService handles transaction categories
type TransactionCategoryService interface {
	// List retrieves all categories
	List(ctx context.Context) ([]*TransactionCategory, error)

	// Create creates a new category
	Create(ctx context.Context, params *CreateCategoryParams) (*TransactionCategory, error)

	// Delete deletes a category
	Delete(ctx context.Context, categoryID string) error

	// DeleteMultiple deletes multiple categories
	DeleteMultiple(ctx context.Context, categoryIDs ...string) error

	// GetGroups retrieves category groups
	GetGroups(ctx context.Context) ([]*CategoryGroup, error)
}

// TransactionQueryBuilder builds transaction queries
type TransactionQueryBuilder interface {
	// Filter methods
	Between(start, end time.Time) TransactionQueryBuilder
	WithAccounts(accountIDs ...string) TransactionQueryBuilder
	WithCategories(categoryIDs ...string) TransactionQueryBuilder
	WithTags(tagIDs ...string) TransactionQueryBuilder
	WithMinAmount(amount float64) TransactionQueryBuilder
	WithMaxAmount(amount float64) TransactionQueryBuilder
	WithMerchant(merchant string) TransactionQueryBuilder
	Search(query string) TransactionQueryBuilder
	Limit(limit int) TransactionQueryBuilder
	Offset(offset int) TransactionQueryBuilder

	// Execute runs the query
	Execute(ctx context.Context) (*TransactionList, error)

	// Stream returns results as a channel for large queries
	Stream(ctx context.Context) (<-chan *Transaction, <-chan error)
}

// TagService handles transaction tags
type TagService interface {
	// List retrieves all tags
	List(ctx context.Context) ([]*Tag, error)

	// Create creates a new tag
	Create(ctx context.Context, name, color string) (*Tag, error)

	// SetTransactionTags sets tags on a transaction
	SetTransactionTags(ctx context.Context, transactionID string, tagIDs ...string) error
}

// BudgetService handles budget operations
type BudgetService interface {
	// List retrieves budgets for a date range
	List(ctx context.Context, startDate, endDate time.Time) ([]*Budget, error)

	// SetAmount sets budget amount
	SetAmount(ctx context.Context, budgetID string, amount float64, rollover bool, startDate time.Time) error
}

// CashflowService handles cashflow analysis
type CashflowService interface {
	// Get retrieves cashflow data
	Get(ctx context.Context, params *CashflowParams) (*Cashflow, error)

	// GetSummary retrieves cashflow summary
	GetSummary(ctx context.Context, params *CashflowSummaryParams) (*CashflowSummary, error)
}

// RecurringService handles recurring transactions
type RecurringService interface {
	// List retrieves all recurring transactions
	List(ctx context.Context) ([]*RecurringTransaction, error)
}

// InstitutionService handles financial institutions
type InstitutionService interface {
	// List retrieves connected institutions
	List(ctx context.Context) ([]*Institution, error)
}

// AdminService handles administrative operations
type AdminService interface {
	// GetSubscription retrieves subscription details
	GetSubscription(ctx context.Context) (*Subscription, error)

	// UploadBalanceHistory uploads CSV balance history
	UploadBalanceHistory(ctx context.Context, accountID string, csvData []byte) error
}

// AuthService handles authentication
type AuthService interface {
	// Login performs authentication
	Login(ctx context.Context, email, password string) error

	// LoginWithMFA performs login with MFA
	LoginWithMFA(ctx context.Context, email, password, mfaCode string) error

	// LoginWithTOTP performs login with TOTP secret
	LoginWithTOTP(ctx context.Context, email, password, totpSecret string) error

	// GetSession returns the current session
	GetSession() (*Session, error)

	// SaveSession saves session to file
	SaveSession(path string) error

	// LoadSession loads session from file
	LoadSession(path string) error
}

// RefreshJob represents an async refresh operation
type RefreshJob interface {
	// ID returns the job ID
	ID() string

	// Status returns the current status
	Status() RefreshStatus

	// Wait waits for the job to complete
	Wait(ctx context.Context, timeout time.Duration) error

	// IsComplete checks if the job is complete
	IsComplete(ctx context.Context) (bool, error)

	// Cancel cancels the job
	Cancel(ctx context.Context) error

	// GetProgress returns the progress of individual accounts
	GetProgress() map[string]bool

	// GetMetrics returns job metrics
	GetMetrics() RefreshJobMetrics
}
