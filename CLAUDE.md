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

#### Transactions (11 methods)
- [ ] get_transactions
- [ ] get_transactions_summary
- [ ] create_transaction
- [ ] update_transaction
- [ ] delete_transaction
- [ ] get_transaction_details
- [ ] get_transaction_splits
- [ ] update_transaction_splits
- [ ] get_transaction_categories
- [ ] create_transaction_category
- [ ] get_transaction_tags

#### Budgets (3 methods)
- [ ] get_budgets
- [ ] set_budget_amount
- [ ] get_cashflow

#### Additional (25+ methods)
<!-- Full list to be populated from Python analysis -->

## 🚀 Next Steps for New Session
<!-- ALWAYS UPDATE THIS SECTION BEFORE ENDING A SESSION -->

### Immediate Next Task:
1. Implement TransactionService with all 13 methods
2. Create transaction query builder for advanced filtering
3. Implement transaction categories sub-service
4. Extract and save all transaction GraphQL queries

### Context for Next Session:
- AccountService is fully implemented and can be used as reference
- Transaction service needs builder pattern for complex queries
- Python client has poor error handling - we've improved it
- Session management uses JSON instead of pickle
- All GraphQL queries should be saved in graphql/queries/

## 🔧 Development Guidelines

### For Every Method Implementation:
1. **Read** the Python implementation first
2. **Extract** the GraphQL query to `graphql/queries/`
3. **Define** types in `pkg/monarch/types.go`
4. **Implement** with proper error handling
5. **Test** with unit and integration tests
6. **Validate** against Python client output
7. **Document** any behavioral differences

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
