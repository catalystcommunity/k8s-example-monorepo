package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Create a test context with transaction
func createTestContextWithTx(tx *gorm.DB) context.Context {
	return context.WithValue(context.Background(), postgres_store.GetTxContextKey(), tx)
}

// TestTransactionMiddleware tests that the transaction middleware correctly handles commits and rollbacks
func TestTransactionMiddleware(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a test context with our transaction
		txCtx := context.WithValue(ctx, postgres_store.GetTxContextKey(), tx)
		
		// Get the app's HTTP mux for testing
		router := GetTestMux()
		
		// Create a thing type for testing
		jsonStr := `{
			"name": "Test Thing Type",
			"description": "A thing type created for transaction middleware test"
		}`
		
		// Use our txCtx with the transaction
		req, err := http.NewRequestWithContext(txCtx, "POST", "/api/thing-types", strings.NewReader(jsonStr))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		// Create a response recorder
		rr := httptest.NewRecorder()
		
		// Serve the request
		router.ServeHTTP(rr, req)
		
		// Check the status code is what we expect
		assert.Equal(t, http.StatusCreated, rr.Code)
		
		// Get the thing type directly using our transaction to verify it was created
		var thingType models.ThingType
		err = tx.Where("name = ?", "Test Thing Type").First(&thingType).Error
		assert.NoError(t, err, "Thing type should be created in the transaction")
		
		// Verify the thing type properties
		assert.Equal(t, "Test Thing Type", thingType.Name)
		assert.Equal(t, "A thing type created for transaction middleware test", thingType.Description)
	})
}