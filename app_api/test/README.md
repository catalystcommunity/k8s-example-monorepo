# App API Test Infrastructure

This package provides a database testing infrastructure for app_api that:
1. Initializes the database once for all tests
2. Runs each test in a transaction that's rolled back after the test completes
3. Provides utilities for generating test data

## How It Works

The testing infrastructure:
- Sets up a global database connection once during test initialization
- Wraps each test in a transaction that is automatically rolled back
- Ensures tests don't affect each other and don't require cleanup
- Uses the same database initialization flow as the real application

## Using the Test Database

### 1. Set up TestMain in your test package

```go
func TestMain(m *testing.M) {
    // Use the TestMain from test_utils.go
    test.TestMain(m)
}
```

### 2. Write transactional tests

```go
func TestSomething(t *testing.T) {
    RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
        // Your test code here
        
        // Use store.AppStore for database operations
        testUser := CreateUser()
        err := store.AppStore.CreateUser(ctx, testUser)
        
        // Or use the transaction directly for more complex operations
        result := tx.Create(&models.YourModel{...})
        
        // Assertions
        assert.NoError(t, err)
    })
    // No cleanup needed - transaction is automatically rolled back
}
```

### 3. Configuration

By default, tests connect to:
```
postgresql://root:root@localhost:26257/app_test?sslmode=disable
```

To use a different database, set the `TEST_DB_URI` environment variable:

```bash
export TEST_DB_URI="postgresql://username:password@hostname:port/database?sslmode=disable"
```

## Test Data Generation

The `datautils.go` file provides functions to create test data:

- `CreateUser()` - Creates a test user with random data
- `CreateThingType()` - Creates a test thing type
- `CreateSession()` - Creates a test session
- `CreateOwnedThing()` - Creates a test owned thing
- `CreateBorrow()` - Creates a test borrow record

## Benefits

1. **Speed**: Database is initialized only once for all tests
2. **Isolation**: Each test runs in its own transaction
3. **No cleanup**: All changes are rolled back automatically
4. **Real database**: Tests run against a real database, not mocks
5. **Consistent**: Uses the same initialization flow as the real app