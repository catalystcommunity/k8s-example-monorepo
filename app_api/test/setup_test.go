package test

import (
	"os"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/cmd"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
)

// TestMain for all tests in the test package
func TestMain(m *testing.M) {
	// Configure test database URI for the running container
	testDbUri := "postgresql://devuser:devpass@localhost:5432/monodemopg?sslmode=disable"
	os.Setenv("TEST_DB_URI", testDbUri)
	
	// Also set the main DB_URI for migrations
	config.DbUri = testDbUri
	os.Setenv("DB_URI", testDbUri)
	
	// Configure test environment to never commit transactions
	os.Setenv("COMMIT_ON_SUCCESS", "false")
	
	// Simply run migrations before tests
	// This is safe because goose tracks applied migrations and won't rerun them
	err := cmd.RunMigrations()
	if err != nil {
		panic("Failed to run migrations: " + err.Error())
	}
	
	// Initialize the test database connections
	initTestDB()
	if initErr != nil {
		panic("Failed to initialize test database: " + initErr.Error())
	}
	
	// Run the tests
	code := m.Run()
	
	// Clean up the database connection
	if cleanupFunc != nil {
		cleanupFunc()
	}
	
	// Exit with the test status code
	os.Exit(code)
}