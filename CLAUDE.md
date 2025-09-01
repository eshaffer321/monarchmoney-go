# Claude Code Session Guide - Monarch Money Go Client

## 🎯 Project Mission
We are creating a **production-grade Go client** for the Monarch Money API that is significantly better than the existing Python implementation located at `/Users/erickshaffer/code/monarchmoney/monarchmoney/monarchmoney.py`.

## 📍 Important Locations
- **Python Reference Implementation**: `/Users/erickshaffer/code/monarchmoney/monarchmoney/monarchmoney.py`
- **Go Implementation**: `/Users/erickshaffer/code/monarchmoney-go`
- **Python Tests**: `/Users/erickshaffer/code/monarchmoney/tests/test_monarchmoney.py`

## 🏗️ Target Architecture

### Core Principles
1. **Domain-Driven Resource Pattern**: Operations grouped by resource type (Accounts, Transactions, etc.)
2. **Interface-First Design**: Every major component should be an interface for testability
3. **Context-Aware**: All operations must accept context.Context for cancellation/timeouts
4. **Type Safety**: No `interface{}` or `map[string]interface{}` in public APIs
5. **Error Wrapping**: Rich error context with stack traces
6. **Concurrent by Design**: Leverage goroutines for bulk operations
7. **Observable**: Built-in metrics, tracing, and logging hooks

### API Design Pattern
```go
// Domain-driven resource access
client := monarch.NewClient(token)

// Clean resource-based API
accounts, err := client.Accounts.List(ctx)
account, err := client.Accounts.Get(ctx, accountID)
holdings, err := client.Accounts.Holdings(ctx, accountID)
err := client.Accounts.Delete(ctx, accountID)

// Builder pattern for complex queries
txns, err := client.Transactions.Query().
    Between(start, end).
    WithTags("vacation").
    WithMinAmount(50).
    Execute(ctx)

// Async operations return jobs
job := client.Admin.RefreshAccounts(ctx, accountIDs...)
err := job.Wait(ctx, 30*time.Second)
```

### Package Structure
```
monarchmoney-go/
├── pkg/monarch/
│   ├── client.go          # Main client with service references
│   ├── accounts.go        # AccountService implementation
│   ├── transactions.go    # TransactionService implementation
│   ├── budgets.go         # BudgetService implementation
│   ├── cashflow.go        # CashflowService implementation
│   ├── admin.go           # AdminService (refresh, sync, etc.)
│   ├── types.go           # All type definitions
│   ├── errors.go          # Custom error types
│   ├── options.go         # Client configuration options
│   └── filters.go         # Type-safe filter builders
├── internal/
│   ├── transport/         # HTTP/GraphQL transport layer
│   ├── auth/              # Authentication logic
│   ├── session/           # Session persistence
│   └── cache/             # Smart caching layer
├── graphql/
│   ├── queries/           # All GraphQL query definitions
│   ├── schema.graphql     # Full schema
│   └── generated/         # Code-generated types
├── cmd/
│   ├── monarch/           # CLI tool
│   └── validator/         # Python compatibility validator
└── examples/              # Usage examples
```

## 📊 Implementation Progress

### ✅ COMPLETED
<!-- Update this section after completing each phase -->
- [x] Project structure initialization
- [x] Core interfaces defined (interfaces.go with all service contracts)
- [x] Type definitions created (types.go with all domain models)
- [x] GraphQL schema extracted and documented
- [x] Authentication system (internal/auth with Login, MFA, TOTP support)
- [x] Session management (JSON-based, not pickle)
- [x] Base HTTP/GraphQL transport layer (internal/transport)
- [x] AccountService fully implemented (all 13 methods)
- [x] Error handling system with proper error types
- [x] Client architecture with domain-driven services
- [x] Method inventory documented (METHOD_INVENTORY.md)

### 🔄 IN PROGRESS
<!-- Current work item - ONE item at a time -->
- Creating Python compatibility validator

### 📝 Method Migration Checklist
<!-- Track every method from Python client -->
#### Authentication (COMPLETED ✅)
- [x] login → Login()
- [x] interactive_login → (not implemented - CLI only)
- [x] multi_factor_authenticate → LoginWithMFA()
- [x] save_session → SaveSession()
- [x] load_session → LoadSession()

#### Accounts (COMPLETED ✅)
- [x] get_accounts → List()
- [x] get_account_type_options → GetTypes()
- [x] create_manual_account → Create()
- [x] update_account → Update()
- [x] delete_account → Delete()
- [x] request_accounts_refresh → Refresh()
- [x] request_accounts_refresh_and_wait → RefreshAndWait()
- [x] is_accounts_refresh_complete → IsRefreshComplete()
- [x] get_account_holdings → GetHoldings()
- [x] get_account_history → GetHistory()
- [x] get_recent_account_balances → GetBalances()
- [x] get_account_snapshots_by_type → GetSnapshots()

#### Transactions (COMPLETED ✅)
- [x] get_transactions → Query().Execute()
- [x] get_transactions_summary → GetSummary()
- [x] create_transaction → Create()
- [x] update_transaction → Update()
- [x] delete_transaction → Delete()
- [x] get_transaction_details → Get()
- [x] get_transaction_splits → GetSplits()
- [x] update_transaction_splits → UpdateSplits()
- [x] get_transaction_categories → Categories().List()
- [x] create_transaction_category → Categories().Create()
- [x] delete_transaction_category → Categories().Delete()
- [x] get_transaction_category_groups → Categories().GetGroups()
- [x] get_transaction_tags → Tags.List()
- [x] create_transaction_tag → Tags.Create()
- [x] set_transaction_tags → Tags.SetTransactionTags()

#### Budgets (COMPLETED ✅)
- [x] get_budgets → List()
- [x] set_budget_amount → SetAmount()

#### Cashflow (COMPLETED ✅)
- [x] get_cashflow → Get()
- [x] get_cashflow_summary → GetSummary()

#### Additional Methods (COMPLETED ✅)
- [x] get_subscription_details → Subscription.GetDetails()
- [x] get_aggregate_snapshots → Accounts.GetAggregateSnapshots()
- [x] upload_account_balance_history → Accounts.UploadBalanceHistory()
- [x] get_recurring_transactions → Recurring.List()
- [x] get_institutions → Institutions.List()

## 🚀 Next Steps for New Session
<!-- ALWAYS UPDATE THIS SECTION BEFORE ENDING A SESSION -->

### Immediate Next Task:
1. Create comprehensive unit tests for all new methods
2. Add integration tests with mocked responses
3. Ensure test coverage meets 70% threshold
4. Document any API differences from Python client

### Context for Next Session:
- All major methods from Python client are now implemented
- Transaction splits, categories, and tags are fully functional
- Subscription details and aggregate snapshots are complete
- Balance history upload uses multipart form data
- Need to focus on testing and documentation
- Python client has poor error handling - we've improved it
- Session management uses JSON instead of pickle
- All GraphQL queries should be saved in graphql/queries/

## 🔧 Development Guidelines

### Test-Driven Development (TDD) Approach
**IMPORTANT**: We follow strict TDD practices in this repository:

1. **Always create a test FIRST to reproduce a bug**
   - Write a test that demonstrates the bug
   - Run the test - it should FAIL
   - Fix the bug in the source code
   - Re-run the test - it should PASS
   - This ensures bugs are properly captured and prevented from regression

2. **Example TDD workflow for bug fixes:**
```bash
# 1. Write failing test
vim pkg/monarch/accounts_test.go  # Add test case for the bug

# 2. Run test - should fail
go test ./pkg/monarch -run TestAccountService_BugCase -v

# 3. Fix the bug
vim pkg/monarch/accounts.go  # Fix the implementation

# 4. Run test - should pass
go test ./pkg/monarch -run TestAccountService_BugCase -v
```

### For Every Method Implementation:
1. **Write** the test first (TDD)
2. **Read** the Python implementation for reference
3. **Extract** the GraphQL query to `graphql/queries/`
4. **Define** types in `pkg/monarch/types.go`
5. **Implement** with proper error handling
6. **Test** with unit and integration tests
7. **Validate** against Python client output
8. **Document** any behavioral differences

### Testing Strategy:
```go
// Every method needs:
// 1. Unit test with mocked transport
// 2. Integration test with recorded responses  
// 3. Compatibility test against Python
// 4. Benchmark comparing to Python
```

### Session Handoff Protocol:
Before ending any session:
1. Commit all working code
2. Update COMPLETED section
3. Move current task to COMPLETED or document blockers
4. Update NEXT STEPS with specific instructions
5. Note any important discoveries or decisions

## 🎯 Success Metrics

### Must Have:
- [ ] 100% API coverage from Python client
- [ ] All tests passing
- [ ] 10x performance improvement
- [ ] Zero memory leaks
- [ ] Concurrent operations support

### Nice to Have:
- [ ] CLI tool
- [ ] Prometheus metrics
- [ ] OpenTelemetry tracing
- [ ] Circuit breaker
- [ ] Response caching

## 🐛 Known Issues / Decisions
<!-- Document any important findings or decisions made -->
- Python uses pickle for session storage → Use JSON in Go
- Python has weak error handling → Implement proper error types
- Python mixes async/sync → Go will be fully concurrent
- Python uses Dict[str, Any] → Strong typing throughout

## 📚 References
- [Monarch Money API](https://api.monarchmoney.com/graphql) (requires auth)
- [Python Client Repo](https://github.com/hammem/monarchmoney)
- GraphQL Best Practices for Go
- [genqlient Documentation](https://github.com/Khan/genqlient)

---

**Remember**: This is a systematic rewrite. Every decision should improve upon the Python version. When in doubt, check the Python implementation at `/Users/erickshaffer/code/monarchmoney/monarchmoney/monarchmoney.py`.

**Session Continuity**: Always assume the next session has no context. Document everything.
