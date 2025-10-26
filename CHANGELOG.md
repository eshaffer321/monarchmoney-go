# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2025-10-25

### Fixed
- Fixed `Transactions.Delete()` method returning `BAD_REQUEST` error by updating GraphQL mutation format to match Python client and Monarch API expectations. The mutation now uses `DeleteTransactionMutationInput` with `transactionId` field wrapped in `input` parameter instead of passing UUID directly.

### Added
- Added comprehensive test coverage for `Transactions.Delete()` including error handling scenarios
- Added example documentation for transaction deletion and `hideFromReports` workaround (`examples/transaction_deletion/main.go`)

### Documentation
- Documented `HideFromReports` field as an alternative to deletion for bank-imported transactions that cannot be deleted
- Added detailed examples for consolidating multi-delivery orders (e.g., Walmart split deliveries)

## [1.0.0] - 2025-10-20

### Added

**Core Services**
- Full authentication support (Login, MFA, TOTP, session management)
- Account service with listing, creation, updates, and refresh capabilities
- Transaction service with querying, filtering, updates, splits, and streaming
- Budget service with listing, updates, and goals (goalsV2) support
- Cashflow service with summary and category-level details
- Institution service for managing financial institution connections
- Subscription service for managing Monarch Money subscriptions

**Advanced Features**
- MCP (Model Context Protocol) server for AI integration with Monarch Money
- Goals tracking with rollover amounts and types
- Transaction splits support with comprehensive error handling
- Sentry integration for error tracking and monitoring
- Rate limiting support with configurable limits
- Retry logic with exponential backoff
- Hooks for observability (OnRequest, OnResponse, OnError)
- Session management with file-based persistence

**Developer Experience**
- Comprehensive test coverage with mocked responses
- Full GoDoc documentation
- CI/CD with GitHub Actions
- Code coverage reporting with Codecov
- Go Report Card integration
- Example implementations demonstrating all features
- Structured error handling with custom error types

**Infrastructure**
- GraphQL query loader for efficient query management
- Connection pooling and smart caching
- Context support throughout for cancellation and timeouts
- Custom date parsing to handle multiple API date formats

### Changed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- N/A (Initial release)

## [0.0.0] - Development

All development work leading up to the v1.0.0 release.

[Unreleased]: https://github.com/eshaffer321/monarchmoney-go/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/eshaffer321/monarchmoney-go/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/eshaffer321/monarchmoney-go/releases/tag/v1.0.0
