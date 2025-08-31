package monarch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/erickshaffer/monarchmoney-go/internal/graphql"
	"github.com/erickshaffer/monarchmoney-go/internal/transport"
	internalTypes "github.com/erickshaffer/monarchmoney-go/internal/types"
)

const (
	// DefaultBaseURL is the default Monarch Money API base URL
	DefaultBaseURL = "https://api.monarchmoney.com"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// UserAgent is the user agent string
	UserAgent = "monarch-go/1.0.0"
)

// Client is the main Monarch Money API client
type Client struct {
	// Service interfaces
	Accounts     AccountService
	Transactions TransactionService
	Tags         TagService
	Budgets      BudgetService
	Cashflow     CashflowService
	Recurring    RecurringService
	Institutions InstitutionService
	Admin        AdminService
	Auth         AuthService

	// Internal fields
	baseURL     string
	httpClient  *http.Client
	transport   Transport
	options     *ClientOptions
	session     *Session
	queryLoader *graphql.QueryLoader
}

// ClientOptions configures the client
type ClientOptions struct {
	// BaseURL overrides the default API base URL
	BaseURL string

	// HTTPClient allows using a custom HTTP client
	HTTPClient *http.Client

	// Timeout sets the HTTP client timeout
	Timeout time.Duration

	// Token provides direct authentication token
	Token string

	// SessionFile path for session persistence
	SessionFile string

	// Logger for debug logging
	Logger Logger

	// RetryConfig configures retry behavior
	RetryConfig *internalTypes.RetryConfig

	// RateLimiter for rate limiting
	RateLimiter RateLimiter

	// Hooks for observability
	Hooks *internalTypes.Hooks
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// Transport handles HTTP/GraphQL communication
type Transport interface {
	Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error
	SetAuth(token string)
	SetSession(session *internalTypes.Session)
}

// NewClient creates a new Monarch Money client
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		opts = &ClientOptions{}
	}

	// Set defaults
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	if opts.Timeout > 0 {
		opts.HTTPClient.Timeout = opts.Timeout
	}

	// Create transport using the internal package
	transportOpts := &transport.Options{
		BaseURL:     opts.BaseURL,
		HTTPClient:  opts.HTTPClient,
		RetryConfig: opts.RetryConfig,
		Logger:      opts.Logger,
		Hooks:       opts.Hooks,
	}
	trans := transport.NewGraphQLTransport(transportOpts)

	// Set auth if token provided
	if opts.Token != "" {
		trans.SetAuth(opts.Token)
	}

	// Create client
	c := &Client{
		baseURL:     opts.BaseURL,
		httpClient:  opts.HTTPClient,
		transport:   trans,
		options:     opts,
		queryLoader: graphql.NewQueryLoader(),
	}

	// Initialize services
	c.initServices()

	// Load session if file specified
	if opts.SessionFile != "" {
		if err := c.loadSession(opts.SessionFile); err != nil && opts.Logger != nil {
			opts.Logger.Warn("Failed to load session", "error", err)
		}
	}

	return c, nil
}

// NewClientWithToken creates a client with an auth token
func NewClientWithToken(token string) (*Client, error) {
	return NewClient(&ClientOptions{
		Token: token,
	})
}

// initServices initializes all service implementations
func (c *Client) initServices() {
	// Create service implementations
	c.Accounts = &accountService{client: c}
	c.Transactions = newTransactionService(c)
	c.Tags = &tagService{client: c}
	c.Budgets = &budgetService{client: c}
	c.Cashflow = &cashflowService{client: c}
	c.Recurring = &recurringService{client: c}
	c.Institutions = &institutionService{client: c}
	c.Admin = &adminService{client: c}
	c.Auth = newAuthService(c)
}

// WithContext returns a new client with the given context
func (c *Client) WithContext(ctx context.Context) *Client {
	// This allows for context-specific client instances
	// Useful for request-scoped configuration
	newClient := *c
	return &newClient
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.transport.SetAuth(token)
	if c.session == nil {
		c.session = &Session{}
	}
	c.session.Token = token
}

// GetSession returns the current session
func (c *Client) GetSession() *Session {
	return c.session
}

// loadQuery loads a GraphQL query from the embedded filesystem
func (c *Client) loadQuery(queryPath string) string {
	query, err := c.queryLoader.Load(queryPath)
	if err != nil {
		// This should never happen in production as queries are embedded
		panic(fmt.Sprintf("failed to load query %s: %v", queryPath, err))
	}
	return query
}

// loadSession loads session from file
func (c *Client) loadSession(path string) error {
	if c.Auth != nil {
		return c.Auth.LoadSession(path)
	}
	return nil
}

// executeGraphQL executes a GraphQL query
func (c *Client) executeGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	// Add hooks
	if c.options.Hooks != nil && c.options.Hooks.OnRequest != nil {
		// Create pseudo request for hook
		req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/graphql", nil)
		c.options.Hooks.OnRequest(ctx, req)
	}

	// Rate limiting
	if c.options.RateLimiter != nil {
		if err := c.options.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}
	}

	// Execute query
	start := time.Now()
	err := c.transport.Execute(ctx, query, variables, result)
	duration := time.Since(start)

	// Response hook
	if c.options.Hooks != nil && c.options.Hooks.OnResponse != nil {
		// Create pseudo response for hook
		resp := &http.Response{StatusCode: 200}
		if err != nil {
			resp.StatusCode = 500
		}
		c.options.Hooks.OnResponse(ctx, resp, duration)
	}

	// Error hook
	if err != nil && c.options.Hooks != nil && c.options.Hooks.OnError != nil {
		c.options.Hooks.OnError(ctx, err)
	}

	return err
}
