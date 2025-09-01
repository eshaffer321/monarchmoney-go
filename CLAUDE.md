# MonarchMoney Go Client - Developer Guide

## 🚀 Quick Development Commands

```bash
# Run all tests (most important command)
go test ./pkg/monarch/... -v

# Run with coverage  
go test ./pkg/monarch/... -coverprofile=coverage.out && go tool cover -html=coverage.out

# Run specific service tests
go test ./pkg/monarch -run TestAccountService -v
go test ./pkg/monarch -run TestTransaction -v

# Debug a failing test
go test ./pkg/monarch -run TestSpecificMethod -v -count=1

# Check for issues
go vet ./...
```

## 🧭 How to Navigate This Codebase

### When Adding New Features
1. **Start**: `pkg/monarch/interfaces.go` - add method to relevant service interface
2. **Implement**: `pkg/monarch/{service}.go` - implement the method
3. **GraphQL**: Save query in `internal/graphql/queries/{service}/`
4. **Types**: Add to `pkg/monarch/types.go` if needed
5. **Test**: Create test in `pkg/monarch/{service}_test.go`

### Key Files by Purpose
```
pkg/monarch/
├── client.go           # Main client + service initialization
├── interfaces.go       # ALL service method signatures
├── types.go           # Data structures (40+ types)
├── errors.go          # Error handling patterns
├── date.go            # Custom date parsing (important!)
└── accounts.go         # Example service implementation
    transactions.go     
    budgets.go         
    ...

internal/graphql/queries/   # GraphQL organized by service
├── accounts/
├── transactions/  
├── budgets/
...
```

## 🔧 Development Patterns

### Adding a New Service Method
```go
// 1. Add to interface (interfaces.go)
type AccountService interface {
    NewMethod(ctx context.Context, param string) (*Result, error)
}

// 2. Implement (accounts.go) 
func (s *accountService) NewMethod(ctx context.Context, param string) (*Result, error) {
    query := s.client.loadQuery("accounts/new_method.graphql")
    
    variables := map[string]interface{}{
        "param": param,
    }
    
    var result struct {
        Data *Result `json:"data"`
    }
    
    if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
        return nil, errors.Wrap(err, "failed to execute new method")
    }
    
    return result.Data, nil
}

// 3. Test (accounts_test.go)
func TestAccountService_NewMethod(t *testing.T) {
    mockTransport := new(MockTransport)
    client := &Client{
        transport:   mockTransport,
        queryLoader: graphql.NewQueryLoader(),
        options:     &ClientOptions{},
        baseURL:     "https://api.test.com",
    }
    client.initServices()

    response := `{"data": {"field": "value"}}`
    mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(response, nil)

    result, err := client.Accounts.NewMethod(context.Background(), "test")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockTransport.AssertExpectations(t)
}
```

### GraphQL Response Handling
```go
// GraphQL responses don't always match the schema exactly
// Common patterns:

// 1. Direct field mapping
var result struct {
    Accounts []*Account `json:"accounts"`
}

// 2. Nested response (common)
var result struct {
    GetAccount *Account `json:"getAccount"`
}

// 3. Array with single object (aggregates pattern)
var result struct {
    Aggregates []struct {
        Summary *TransactionSummary `json:"summary"`
    } `json:"aggregates"`
}
```

## 🐛 Common Issues & Solutions

### Test Failures
```bash
# GraphQL field mismatch
# Look for: `json: cannot unmarshal`
# Fix: Check GraphQL response format in test vs actual API

# Mock expectations failing  
# Look for: `mock: Unexpected call`
# Fix: Verify mock.Anything vs specific matchers

# Date parsing errors
# Look for: `parsing time` errors
# Fix: Use Date type, not time.Time for API dates
```

### Authentication Issues
```bash
# Session token problems
client := monarch.NewClient("your-session-token")

# Login with MFA
session, err := client.Auth.Login(ctx, "email", "password")
if err != nil {
    // Check for MFA challenge
    if mfaErr, ok := err.(*MFARequired); ok {
        session, err = client.Auth.LoginWithMFA(ctx, "email", "password", mfaErr.Token, "123456")
    }
}
```

### GraphQL Debugging
```go
// Enable GraphQL request logging (set in client options)
client := monarch.NewClientWithOptions("token", &monarch.ClientOptions{
    Debug: true,  // Logs all GraphQL requests
})
```

## 🧪 Testing Patterns

### TDD for Bug Fixes (IMPORTANT!)
When you find a bug, **always write a failing test first**:

```bash
# 1. Write a test that reproduces the bug
func TestAccountService_BugFix_Issue123(t *testing.T) {
    // Setup that reproduces the problematic scenario
    mockTransport := new(MockTransport)
    client := setupTestClient(mockTransport)
    
    // Mock the exact response that causes the bug
    response := `{"problematic": "response"}`
    mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(response, nil)
    
    // Call the method that's broken
    result, err := client.Accounts.ProblematicMethod(ctx, "test")
    
    // This should fail before the fix
    assert.NoError(t, err)
    assert.Equal(t, "expected", result.Field)
}

# 2. Run the test - it should FAIL
go test ./pkg/monarch -run TestAccountService_BugFix_Issue123 -v

# 3. Fix the bug in the implementation
# Edit pkg/monarch/accounts.go

# 4. Run the test again - it should PASS  
go test ./pkg/monarch -run TestAccountService_BugFix_Issue123 -v

# 5. Run all tests to ensure no regression
go test ./pkg/monarch/... -v
```

### Standard Test Structure
```go
func TestServiceName_MethodName(t *testing.T) {
    // Setup mock transport
    mockTransport := new(MockTransport)
    client := &Client{
        transport:   mockTransport,
        queryLoader: graphql.NewQueryLoader(),
        options:     &ClientOptions{},
        baseURL:     "https://api.test.com",
    }
    client.initServices()

    // Mock GraphQL response (NO "data" wrapper needed)
    response := `{
        "fieldName": "value"
    }`

    // Setup mock expectation
    mockTransport.On("Execute", 
        mock.Anything,           // context
        mock.Anything,           // query string  
        mock.Anything,           // variables
        mock.Anything,           // result pointer
    ).Return(response, nil)

    // Execute method
    result, err := client.ServiceName.MethodName(ctx, "param")

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockTransport.AssertExpectations(t)
}
```

### Testing with Specific Parameters
```go
mockTransport.On("Execute", 
    mock.Anything, 
    mock.Anything,
    mock.MatchedBy(func(vars map[string]interface{}) bool {
        return vars["accountId"] == "test-123"
    }),
    mock.Anything,
).Return(response, nil)
```

## 🔍 Key Architectural Decisions

### Why These Patterns Exist
- **Interface-first**: All services are interfaces for easy testing/mocking
- **Context everywhere**: All methods accept context.Context for cancellation
- **Structured errors**: Custom error types with codes, not generic errors  
- **GraphQL transport**: Single HTTP client with GraphQL query loading
- **No "data" wrapper in tests**: MonarchMoney API doesn't use standard GraphQL response format
- **Custom Date type**: API returns dates in multiple formats, needs custom parsing

### Python Client Differences
- **Sessions**: JSON files instead of Python pickle
- **Error handling**: Structured errors instead of generic exceptions
- **Types**: Strong typing instead of `Dict[str, Any]`
- **Concurrency**: Goroutines instead of asyncio

## 🔗 Key Resources

- **Python reference**: [Original Python client](https://github.com/hammem/monarchmoney) - check `monarchmoney/monarchmoney.py`
- **Example usage**: `examples/full_example.go`
- **GraphQL queries**: `internal/graphql/queries/`
- **All interfaces**: `pkg/monarch/interfaces.go`

## 🚨 Before You Start Coding

1. **Run the tests**: `go test ./pkg/monarch/... -v` (should all pass)
2. **For bugs**: **ALWAYS write a failing test first** to reproduce the issue (TDD)
3. **For new features**: Add to interface first, then implement, then test  
4. **Check coverage**: Current coverage is ~37%, don't make it worse
5. **Check Python client**: When in doubt, see how Python version works
6. **Use existing patterns**: Don't invent new ways of doing things

---

**Most Important**: This codebase has consistent patterns. Follow them, don't create new ones.