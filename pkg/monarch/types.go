package monarch

import (
	"time"
)

// Account represents a financial account
type Account struct {
	ID                              string              `json:"id"`
	DisplayName                     string              `json:"displayName"`
	SyncDisabled                    bool                `json:"syncDisabled"`
	DeactivatedAt                   *time.Time          `json:"deactivatedAt,omitempty"`
	IsHidden                        bool                `json:"isHidden"`
	IsAsset                         bool                `json:"isAsset"`
	Mask                            string              `json:"mask"`
	CreatedAt                       time.Time           `json:"createdAt"`
	UpdatedAt                       time.Time           `json:"updatedAt"`
	DisplayLastUpdatedAt            time.Time           `json:"displayLastUpdatedAt"`
	CurrentBalance                  float64             `json:"currentBalance"`
	DisplayBalance                  float64             `json:"displayBalance"`
	IncludeInNetWorth               bool                `json:"includeInNetWorth"`
	HideFromList                    bool                `json:"hideFromList"`
	HideTransactionsFromReports     bool                `json:"hideTransactionsFromReports"`
	IncludeBalanceInNetWorth        bool                `json:"includeBalanceInNetWorth"`
	IncludeInGoalBalance            bool                `json:"includeInGoalBalance"`
	DataProvider                    string              `json:"dataProvider"`
	DataProviderAccountID           string              `json:"dataProviderAccountId"`
	IsManual                        bool                `json:"isManual"`
	TransactionsCount               int                 `json:"transactionsCount"`
	HoldingsCount                   int                 `json:"holdingsCount"`
	ManualInvestmentsTrackingMethod string              `json:"manualInvestmentsTrackingMethod"`
	Order                           int                 `json:"order"`
	LogoURL                         string              `json:"logoUrl"`
	Type                            *AccountTypeInfo    `json:"type"`
	Subtype                         *AccountSubtypeInfo `json:"subtype"`
	Credential                      *Credential         `json:"credential"`
	Institution                     *Institution        `json:"institution"`
}

// AccountTypeInfo represents account type information
type AccountTypeInfo struct {
	Name    string `json:"name"`
	Display string `json:"display"`
}

// AccountSubtypeInfo represents account subtype information
type AccountSubtypeInfo struct {
	Name    string `json:"name"`
	Display string `json:"display"`
}

// AccountType represents available account types
type AccountType struct {
	Type             *AccountTypeInfo      `json:"type"`
	Subtype          *AccountSubtypeInfo   `json:"subtype"`
	PossibleSubtypes []*AccountSubtypeInfo `json:"possibleSubtypes"`
}

// Credential represents account credentials
type Credential struct {
	ID                             string       `json:"id"`
	UpdateRequired                 bool         `json:"updateRequired"`
	DisconnectedFromDataProviderAt *time.Time   `json:"disconnectedFromDataProviderAt,omitempty"`
	DataProvider                   string       `json:"dataProvider"`
	Institution                    *Institution `json:"institution"`
}

// Institution represents a financial institution
type Institution struct {
	ID                 string `json:"id"`
	PlaidInstitutionID string `json:"plaidInstitutionId"`
	Name               string `json:"name"`
	Status             string `json:"status"`
	PrimaryColor       string `json:"primaryColor"`
	URL                string `json:"url"`
}

// Merchant represents a merchant
type Merchant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID                 string               `json:"id"`
	Date               time.Time            `json:"date"`
	Amount             float64              `json:"amount"`
	Pending            bool                 `json:"pending"`
	HideFromReports    bool                 `json:"hideFromReports"`
	PlaidName          string               `json:"plaidName"`
	Merchant           *Merchant            `json:"merchant"`
	Notes              string               `json:"notes"`
	HasSplits          bool                 `json:"hasSplits"`
	IsSplitTransaction bool                 `json:"isSplitTransaction"`
	IsRecurring        bool                 `json:"isRecurring"`
	NeedsReview        bool                 `json:"needsReview"`
	ReviewedAt         *time.Time           `json:"reviewedAt,omitempty"`
	ReviewedByUserID   string               `json:"reviewedByUserId"`
	CreatedAt          time.Time            `json:"createdAt"`
	UpdatedAt          time.Time            `json:"updatedAt"`
	Account            *Account             `json:"account"`
	Category           *TransactionCategory `json:"category"`
	Tags               []*Tag               `json:"tags"`
}

// TransactionDetails includes additional transaction information
type TransactionDetails struct {
	*Transaction
	OriginalMerchant    string               `json:"originalMerchant"`
	OriginalCategoryID  string               `json:"originalCategoryId"`
	OriginalCategory    *TransactionCategory `json:"originalCategory"`
	Splits              []*TransactionSplit  `json:"splits"`
	SimilarTransactions []*Transaction       `json:"similarTransactions"`
}

// TransactionSplit represents a split transaction
type TransactionSplit struct {
	ID         string               `json:"id"`
	Amount     float64              `json:"amount"`
	Merchant   *Merchant            `json:"merchant"`
	Notes      string               `json:"notes"`
	Category   *TransactionCategory `json:"category"`
	CategoryID string               `json:"categoryId"`
}

// TransactionCategory represents a transaction category
type TransactionCategory struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Icon             string         `json:"icon"`
	Order            int            `json:"order"`
	SystemCategory   string         `json:"systemCategory"`
	IsSystemCategory bool           `json:"isSystemCategory"`
	IsDisabled       bool           `json:"isDisabled"`
	Group            *CategoryGroup `json:"group"`
	GroupID          string         `json:"groupId"`
}

// CategoryGroup represents a category group
type CategoryGroup struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}

// Tag represents a transaction tag
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Order int    `json:"order"`
}

// Budget represents a budget entry
type Budget struct {
	ID                 string               `json:"id"`
	CategoryID         string               `json:"categoryId"`
	Category           *TransactionCategory `json:"category"`
	Amount             float64              `json:"amount"`
	Rollover           bool                 `json:"rollover"`
	StartDate          time.Time            `json:"startDate"`
	EndDate            time.Time            `json:"endDate"`
	Spent              float64              `json:"spent"`
	Remaining          float64              `json:"remaining"`
	PercentageComplete float64              `json:"percentageComplete"`
}

// Cashflow represents cashflow data
type Cashflow struct {
	StartDate   time.Time           `json:"startDate"`
	EndDate     time.Time           `json:"endDate"`
	Income      float64             `json:"income"`
	Expenses    float64             `json:"expenses"`
	NetCashflow float64             `json:"netCashflow"`
	ByCategory  []*CashflowCategory `json:"byCategory"`
}

// CashflowCategory represents cashflow by category
type CashflowCategory struct {
	Category *TransactionCategory `json:"category"`
	Amount   float64              `json:"amount"`
	Count    int                  `json:"count"`
}

// CashflowSummary represents cashflow summary
type CashflowSummary struct {
	Interval  string              `json:"interval"`
	Summaries []*CashflowInterval `json:"summaries"`
}

// CashflowInterval represents cashflow for an interval
type CashflowInterval struct {
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Income      float64   `json:"income"`
	Expenses    float64   `json:"expenses"`
	NetCashflow float64   `json:"netCashflow"`
}

// Holding represents an investment holding
type Holding struct {
	ID        string    `json:"id"`
	AccountID string    `json:"accountId"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
	Value     float64   `json:"value"`
	CostBasis float64   `json:"costBasis"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AccountBalance represents account balance at a point in time
type AccountBalance struct {
	AccountID string    `json:"accountId"`
	Date      time.Time `json:"date"`
	Balance   float64   `json:"balance"`
}

// AccountSnapshot represents account snapshot
type AccountSnapshot struct {
	Month        string  `json:"month"`
	Year         int     `json:"year"`
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype"`
	TotalValue   float64 `json:"totalValue"`
	AccountCount int     `json:"accountCount"`
}

// AccountHistory represents account balance history
type AccountHistory struct {
	AccountID string          `json:"accountId"`
	Balances  []*BalanceEntry `json:"balances"`
}

// BalanceEntry represents a balance at a point in time
type BalanceEntry struct {
	Date    time.Time `json:"date"`
	Balance float64   `json:"balance"`
	Synced  bool      `json:"synced"`
}

// RecurringTransaction represents a recurring transaction
type RecurringTransaction struct {
	ID        string               `json:"id"`
	Merchant  *Merchant            `json:"merchant"`
	Amount    float64              `json:"amount"`
	Frequency string               `json:"frequency"`
	NextDate  time.Time            `json:"nextDate"`
	LastDate  time.Time            `json:"lastDate"`
	Category  *TransactionCategory `json:"category"`
	Account   *Account             `json:"account"`
	IsActive  bool                 `json:"isActive"`
}

// Subscription represents subscription details
type Subscription struct {
	ID          string     `json:"id"`
	PlanType    string     `json:"planType"`
	Status      string     `json:"status"`
	StartDate   time.Time  `json:"startDate"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	TrialEndsAt *time.Time `json:"trialEndsAt,omitempty"`
	CanceledAt  *time.Time `json:"canceledAt,omitempty"`
	Features    []string   `json:"features"`
}

// TransactionSummary represents transaction summary
type TransactionSummary struct {
	TotalCount      int     `json:"totalCount"`
	TotalIncome     float64 `json:"totalIncome"`
	TotalExpenses   float64 `json:"totalExpenses"`
	AverageIncome   float64 `json:"averageIncome"`
	AverageExpenses float64 `json:"averageExpenses"`
}

// TransactionList represents paginated transaction results
type TransactionList struct {
	Transactions []*Transaction `json:"transactions"`
	TotalCount   int            `json:"totalCount"`
	HasMore      bool           `json:"hasMore"`
	NextOffset   int            `json:"nextOffset"`
}

// Session represents an authenticated session
type Session struct {
	Token      string    `json:"token"`
	UserID     string    `json:"userId"`
	Email      string    `json:"email"`
	ExpiresAt  time.Time `json:"expiresAt"`
	DeviceUUID string    `json:"deviceUuid"`
}

// Parameter structures

// CreateAccountParams for creating manual accounts
type CreateAccountParams struct {
	AccountType       string  `json:"accountType"`
	AccountSubtype    string  `json:"accountSubtype"`
	IsAsset           bool    `json:"isAsset"`
	AccountName       string  `json:"accountName"`
	CurrentBalance    float64 `json:"currentBalance"`
	IncludeInNetWorth bool    `json:"includeInNetWorth"`
}

// UpdateAccountParams for updating accounts
type UpdateAccountParams struct {
	DisplayName                 *string  `json:"displayName,omitempty"`
	IncludeInNetWorth           *bool    `json:"includeInNetWorth,omitempty"`
	HideFromList                *bool    `json:"hideFromList,omitempty"`
	HideTransactionsFromReports *bool    `json:"hideTransactionsFromReports,omitempty"`
	CurrentBalance              *float64 `json:"currentBalance,omitempty"`
}

// CreateTransactionParams for creating transactions
type CreateTransactionParams struct {
	Date       time.Time `json:"date"`
	AccountID  string    `json:"accountId"`
	Amount     float64   `json:"amount"`
	Merchant   *Merchant `json:"merchant"`
	CategoryID string    `json:"categoryId"`
	Notes      string    `json:"notes"`
}

// UpdateTransactionParams for updating transactions
type UpdateTransactionParams struct {
	Date            *time.Time `json:"date,omitempty"`
	AccountID       *string    `json:"accountId,omitempty"`
	Amount          *float64   `json:"amount,omitempty"`
	Merchant        *string    `json:"merchant,omitempty"`
	CategoryID      *string    `json:"categoryId,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	HideFromReports *bool      `json:"hideFromReports,omitempty"`
	NeedsReview     *bool      `json:"needsReview,omitempty"`
}

// CreateCategoryParams for creating categories
type CreateCategoryParams struct {
	Name             string `json:"name"`
	GroupID          string `json:"groupId"`
	RollupCategoryID string `json:"rollupCategoryId,omitempty"`
	Icon             string `json:"icon"`
}

// SnapshotParams for account snapshots
type SnapshotParams struct {
	StartDate    time.Time `json:"startDate"`
	Timeframe    string    `json:"timeframe"` // "year" or "month"
	AccountTypes []string  `json:"accountTypes,omitempty"`
	GroupBy      string    `json:"groupBy,omitempty"`
}

// CashflowParams for cashflow queries
type CashflowParams struct {
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	Limit      int       `json:"limit,omitempty"`
	AccountIDs []string  `json:"accountIds,omitempty"`
}

// CashflowSummaryParams for cashflow summary
type CashflowSummaryParams struct {
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	Interval       string    `json:"interval"` // "day", "week", "month", "year"
	CategoryID     string    `json:"categoryId,omitempty"`
	AccountsFilter []string  `json:"accountsFilter,omitempty"`
}
