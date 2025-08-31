# MonarchMoney Go Client

[![CI](https://github.com/eshaffer321/monarchmoney-go/actions/workflows/ci.yml/badge.svg)](https://github.com/eshaffer321/monarchmoney-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/eshaffer321/monarchmoney-go)](https://goreportcard.com/report/github.com/eshaffer321/monarchmoney-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/eshaffer321/monarchmoney-go)](https://pkg.go.dev/github.com/eshaffer321/monarchmoney-go)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A production-grade Go client library for the [Monarch Money](https://www.monarchmoney.com/) API, providing a clean, idiomatic interface for managing personal finances programmatically.

## Features

- 🔐 **Full Authentication Support** - Login, MFA, TOTP, and session management
- 📊 **Complete API Coverage** - Accounts, transactions, budgets, and more
- 🚀 **High Performance** - Concurrent operations, connection pooling, and smart caching
- 🛡️ **Production Ready** - Comprehensive error handling, retries, and rate limiting
- 🧪 **Well Tested** - Extensive test coverage with mocked responses
- 📝 **Fully Documented** - Complete GoDoc documentation and examples

## Installation

```bash
go get github.com/eshaffer321/monarchmoney-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/eshaffer321/monarchmoney-go/pkg/monarch"
)

func main() {
    // Create client
    client, err := monarch.NewClient(&monarch.ClientOptions{
        // Option 1: Use token directly
        Token: "your-auth-token",
        
        // Option 2: Login with credentials
        // Email: "your-email@example.com",
        // Password: "your-password",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Get all accounts
    accounts, err := client.Accounts.List(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, account := range accounts {
        fmt.Printf("%s: $%.2f\n", account.DisplayName, account.CurrentBalance)
    }
}
```

## Authentication

### Login with Credentials

```go
client, _ := monarch.NewClient(&monarch.ClientOptions{})

// Basic login
err := client.Auth.Login(ctx, "email@example.com", "password")

// Login with MFA
err := client.Auth.LoginWithMFA(ctx, "email@example.com", "password", "123456")

// Login with TOTP secret
err := client.Auth.LoginWithTOTP(ctx, "email@example.com", "password", "TOTP_SECRET")
```

### Session Management

```go
// Save session to file
err := client.Auth.SaveSession("~/.monarch_session.json")

// Load session from file
err := client.Auth.LoadSession("~/.monarch_session.json")

// Use session file automatically
client, _ := monarch.NewClient(&monarch.ClientOptions{
    SessionFile: "~/.monarch_session.json",
})
```

## Services

### Accounts

```go
// List all accounts
accounts, err := client.Accounts.List(ctx)

// Get account details
account, err := client.Accounts.Get(ctx, "account-id")

// Create manual account
account, err := client.Accounts.Create(ctx, &monarch.CreateAccountParams{
    Type:           monarch.AccountTypeCredit,
    Name:           "My Credit Card",
    CurrentBalance: -1500.00,
})

// Refresh accounts
job, err := client.Accounts.Refresh(ctx, "account-id-1", "account-id-2")
err = job.Wait(ctx, 30*time.Second)
```

### Transactions

```go
// Query transactions with filters
result, err := client.Transactions.Query().
    Between(startDate, endDate).
    WithAccounts("account-id").
    WithCategories("category-id").
    WithMinAmount(100).
    Execute(ctx)

// Get transaction details
transaction, err := client.Transactions.Get(ctx, "transaction-id")

// Update transaction
err := client.Transactions.Update(ctx, "transaction-id", &monarch.UpdateTransactionParams{
    Category: "new-category-id",
    Notes:    "Updated note",
})

// Stream transactions for real-time updates
txnChan, errChan := client.Transactions.Stream(ctx, startDate, endDate)
for {
    select {
    case txn := <-txnChan:
        fmt.Printf("New transaction: %s - $%.2f\n", txn.Merchant, txn.Amount)
    case err := <-errChan:
        log.Printf("Stream error: %v", err)
    }
}
```

### Budgets

```go
// Get all budgets
budgets, err := client.Budgets.List(ctx)

// Update budget amount
err := client.Budgets.SetAmount(ctx, "budget-id", 500.00)

// Get budget details with spending
budget, err := client.Budgets.Get(ctx, "budget-id")
fmt.Printf("Spent $%.2f of $%.2f\n", budget.ActualAmount, budget.PlannedAmount)
```

### Cash Flow

```go
// Get cash flow summary
summary, err := client.Cashflow.GetSummary(ctx, startDate, endDate)
fmt.Printf("Income: $%.2f, Expenses: $%.2f\n", summary.Income, summary.Expenses)

// Get detailed cash flow by category
details, err := client.Cashflow.GetByCategory(ctx, startDate, endDate)
```

## Advanced Features

### Rate Limiting

```go
import "golang.org/x/time/rate"

client, _ := monarch.NewClient(&monarch.ClientOptions{
    RateLimiter: rate.NewLimiter(rate.Every(time.Second), 10), // 10 requests/second
})
```

### Retry Configuration

```go
client, _ := monarch.NewClient(&monarch.ClientOptions{
    RetryConfig: &monarch.RetryConfig{
        MaxRetries: 3,
        RetryWait:  1 * time.Second,
        MaxWait:    10 * time.Second,
    },
})
```

### Hooks for Observability

```go
client, _ := monarch.NewClient(&monarch.ClientOptions{
    Hooks: &monarch.Hooks{
        OnRequest: func(ctx context.Context, req *http.Request) {
            log.Printf("Request: %s %s", req.Method, req.URL)
        },
        OnResponse: func(ctx context.Context, resp *http.Response, duration time.Duration) {
            log.Printf("Response: %d in %v", resp.StatusCode, duration)
        },
        OnError: func(ctx context.Context, err error) {
            log.Printf("Error: %v", err)
        },
    },
})
```

## Examples

See the [examples](examples/) directory for complete working examples:
- [Full Example](examples/full_example.go) - Demonstrates all available features

## Development

### Prerequisites

- Go 1.20 or higher
- Make

### Building

```bash
# Run tests
make test

# Build binary
make build

# Run linter
make lint

# Format code
make fmt
```

### Project Structure

```
monarchmoney-go/
├── pkg/monarch/       # Public API package
├── internal/          # Internal implementation
│   ├── auth/         # Authentication logic
│   ├── graphql/      # GraphQL queries and loader
│   └── transport/    # HTTP/GraphQL transport
├── cmd/              # Command-line tools
├── examples/         # Usage examples
└── tests/           # Integration tests
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See our [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This is an unofficial client library and is not affiliated with Monarch Money. Use at your own risk.

## Acknowledgments

- Inspired by the [Python Monarch Money client](https://github.com/hammem/monarchmoney)
- Built with modern Go best practices