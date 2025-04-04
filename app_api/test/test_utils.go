package test

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/handlers"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dbTimeFormat = "2006-01-02T15:04:05.999999Z"
const sfCreatedAtTimeFormat = "2006-01-02T15:04:05.000+0000"

// Default test database URI
const defaultTestDbUri = "postgresql://root:root@localhost:26257/app_test?sslmode=disable"

var (
	// Global DB connection
	testDB *gorm.DB
	// Global cleanup function
	cleanupFunc func()
	// Once to ensure we only initialize the DB once
	initOnce sync.Once
	// Initialize error
	initErr error
)

// initTestDB initializes the test database once for all tests
func initTestDB() {
	initOnce.Do(func() {
		// Get database URI from environment or use default test URI
		testDbUri := os.Getenv("TEST_DB_URI")
		if testDbUri == "" {
			testDbUri = defaultTestDbUri
		}

		// Set the database URI for the application config
		config.DbUri = testDbUri

		// Set the store implementation
		store.AppStore = postgres_store.PostgresStore

		// Initialize the store
		var cleanup func()
		cleanup, initErr = store.AppStore.Initialize()
		if initErr != nil {
			return
		}

		// Store cleanup function
		cleanupFunc = cleanup

		// Create a direct DB connection for transactions
		testDB, initErr = gorm.Open(postgres.Open(testDbUri), &gorm.Config{})
	})
}

// RunTransactionalTest runs a test function within a transaction that gets rolled back
// This ensures tests don't affect each other and don't require cleanup
// This function guarantees the transaction will be rolled back even if the test panics
func RunTransactionalTest(t *testing.T, testFunc func(ctx context.Context, tx *gorm.DB)) {
	// Make sure DB is initialized
	initTestDB()
	if initErr != nil {
		t.Fatalf("Failed to initialize test database: %v", initErr)
		return // Return to avoid further execution with an uninitialized DB
	}

	// Start a transaction
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", tx.Error)
		return // Return to avoid further execution with a failed transaction
	}

	// Ensure transaction is rolled back after test
	// This will execute even if the test function panics or calls t.Fatal
	defer func() {
		// Catch any panics from the test
		if r := recover(); r != nil {
			// Roll back the transaction
			tx.Rollback()
			// Re-panic to preserve the original panic behavior
			panic(r)
		}

		// If we didn't panic, still roll back
		tx.Rollback()
	}()

	// Create a context with transaction info
	ctx := context.WithValue(context.Background(), postgres_store.GetTxContextKey(), tx)

	// Run the test function within the transaction
	testFunc(ctx, tx)
}

// testMain can be used in test packages to set up and tear down the global test database
// This is renamed to avoid conflict with the TestMain in setup_test.go
func testMain(m *testing.M) {
	// Initialize DB before running tests
	initTestDB()
	if initErr != nil {
		// If initialization failed, report and exit
		panic("Failed to initialize test database: " + initErr.Error())
	}

	// Run all tests
	code := m.Run()

	// Clean up the database connection
	if cleanupFunc != nil {
		cleanupFunc()
	}

	// Exit with the test status code
	os.Exit(code)
}

// GetTestContext returns a context suitable for use in tests
func GetTestContext() context.Context {
	return context.Background()
}

// GetTestMux returns the application's HTTP mux for use in tests
// This uses the same server configuration as the actual application
func GetTestMux() *http.ServeMux {
	// Get the application mux
	mux := handlers.GetAppMux()

	// Create a wrapper mux
	debugMux := http.NewServeMux()

	// Handle all paths
	debugMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create a wrapper for the response writer to capture status
		rw := &responseWriterWrapper{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Call the original mux
		mux.ServeHTTP(rw, r)
	})

	return debugMux
}

// responseWriterWrapper captures the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code
func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
