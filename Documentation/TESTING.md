# Testing Guide

This document describes the testing setup, patterns, and best practices for the transaction-split-go project.

## Overview

The project uses [Testify](https://github.com/stretchr/testify) for unit testing. Testify provides assertion libraries, mocking capabilities, and test suites.

## Dependencies

The following Testify packages are used:
- `github.com/stretchr/testify/assert` - Assertions with helpful error messages
- `github.com/stretchr/testify/require` - Assertions that stop test execution on failure
- `github.com/stretchr/testify/mock` - Mock objects for testing
- `github.com/stretchr/testify/suite` - Test suite support

## Running Tests

### Run all tests
```bash
make test
```

### Run tests with coverage
```bash
make test-coverage
```

### Generate HTML coverage report
```bash
make test-coverage-html
```

This generates `coverage.html` which you can open in a browser to see coverage details.

### Run tests for a specific package
```bash
go test ./internal/handlers -v
go test ./internal/services -v
go test ./internal/pkg/utils -v
```

## Test Structure

### Test File Naming

Test files follow the Go convention: `*_test.go` files are in the same package as the code they test.

### Mock Store

The `internal/mocks/store_mock.go` file contains a mock implementation of the `db.Store` interface. This allows handlers to be tested without a real database connection.

The `MockStore` fully implements the `db.Store` interface, which means it can be used directly without any workarounds. This was made possible by removing unused database queries from the codebase (specifically the `getTransactionsByGroup` query that was never called and had an unexported method).

### Test Helper Functions

Located in `internal/handlers/helpers_test.go`:
- `createRequest()` - Creates HTTP requests for testing
- `createRequestWithPath()` - Creates HTTP requests with path parameters
- `validateJSONResponse()` - Validates JSON response format and content
- `storeAsInterface()` - Converts MockStore to db.Store interface (simple type conversion)

## Writing Tests

### Handler Test Pattern

Handler tests follow this pattern:

```go
func TestHandlerFunction(t *testing.T) {
    tests := []struct {
        name           string
        setupMock      func(*mocks.MockStore)
        requestSetup   func() *http.Request
        expectedStatus int
        expectedBody   interface{}
    }{
        {
            name: "success case",
            setupMock: func(ms *mocks.MockStore) {
                // Set up mock expectations
                ms.On("MethodName", mock.Anything, arg).Return(result, nil)
            },
            requestSetup: func() *http.Request {
                // Create request
                return httptest.NewRequest("GET", "/path", nil)
            },
            expectedStatus: http.StatusOK,
            expectedBody: map[string]interface{}{
                // Expected response
            },
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockStore := mocks.NewMockStore(t)
            tt.setupMock(mockStore)

            req := tt.requestSetup()
            rr := httptest.NewRecorder()

            handler := handlerFunction(storeAsInterface(mockStore))
            handler(rr, req)

            assert.Equal(t, tt.expectedStatus, rr.Code)
            // Validate response body
            mockStore.AssertExpectations(t)
        })
    }
}
```

### Mock Store Usage

```go
mockStore := mocks.NewMockStore(t)

// Set up expectations
mockStore.On("GetUserByID", mock.Anything, int64(1)).Return(user, nil)

// Use the mock
handler := getUserByID(storeAsInterface(mockStore))

// Verify expectations were met
mockStore.AssertExpectations(t)
```

### Service Test Pattern

Service tests (without database dependencies) follow a simpler pattern:

```go
func TestServiceFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       InputType
        expectError bool
        validate    func(t *testing.T, result ResultType)
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ServiceFunction(tt.input)

            if tt.expectError {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                if tt.validate != nil {
                    tt.validate(t, result)
                }
            }
        })
    }
}
```

## Test Coverage Areas

### Handlers (`internal/handlers/`)

- **Helper Functions**: ParsePathInt64, ParseQueryInt32, ParseLimitOffset, DecodeJSONBody, WriteJSONResponse, HandleDBError, HandleDBListError
- **User Handlers**: listUsers, getUserByID, createUser, updateUser, deleteUser, getUserBalances, getTransactionsByUserNested
- **Group Handlers**: listGroups, getGroupByID, createGroup, updateGroup, deleteGroup
- **Group Member Handlers**: listGroupMembersByGroupID, getGroupMemberByID, createGroupMember, updateGroupMember, deleteGroupMember
- **Transaction Handlers**: listTransactions, getTransactionByID, createTransaction, updateTransaction, deleteTransaction, getSplitsByTransactionNested
- **Split Handlers**: listSplits, getSplitsByTransactionID, getSplitByID, createSplit, updateSplit, deleteSplit

### Services (`internal/services/`)

- **Debt Simplification**: SimplifyDebts, splitNetBalances
  - Tests cover: zero-sum validation, payment calculations, edge cases, decimal precision

### Utilities (`internal/pkg/utils/`)

- **Decimal Utils**: SumNetBalances
  - Tests cover: positive/negative balances, empty slices, single values, decimal precision

## Best Practices

1. **Use table-driven tests**: Group related test cases in a slice of structs
2. **Test both success and error cases**: Include happy paths and error scenarios
3. **Validate response structure**: Check HTTP status codes and response formats
4. **Verify mock expectations**: Call `AssertExpectations()` to ensure mocks were used correctly
5. **Use descriptive test names**: Test names should clearly describe what is being tested
6. **Keep tests independent**: Each test should be able to run in isolation
7. **Test edge cases**: Include empty inputs, boundary values, and error conditions

## Example: Adding a New Handler Test

1. Create or update the test file (e.g., `internal/handlers/my_handler_test.go`)
2. Import required packages:
   ```go
   import (
       "net/http"
       "net/http/httptest"
       "testing"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/require"
       // ... other imports
   )
   ```
3. Write test cases following the handler test pattern
4. Run tests: `go test ./internal/handlers -v`

## Troubleshooting

### Mock Store Interface Issues

If you encounter compile errors about the MockStore not implementing db.Store:
- Ensure all methods in the `db.Store` interface are implemented in `MockStore`
- If sqlc generates new query methods, update the mock to include them
- Check that `internal/mocks/store_mock.go` has all methods from the `db.Querier` interface and transaction methods

### Test Failures

- Check that mock expectations match actual calls
- Verify request setup (path values, query parameters, body)
- Ensure response assertions match actual response format

## Future Improvements

- Add integration tests with a test database
- Increase test coverage for remaining handlers
- Add benchmarks for performance-critical functions
- Consider adding table-driven test generators

