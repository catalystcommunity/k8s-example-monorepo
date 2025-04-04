package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/handlers"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// executeRequest is a test helper to execute HTTP requests
func executeRequest(ctx context.Context, method, path string, body []byte) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequestWithContext(ctx, method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequestWithContext(ctx, method, path, nil)
	}

	rr := httptest.NewRecorder()
	router := handlers.GetAppMux()
	router.ServeHTTP(rr, req)
	return rr
}

// TestGetAllThingTypes tests the GetAllThingTypes handler
func TestGetAllThingTypes(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create test data using DataSetup maps
		thingType1, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)
		thingType2, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)
		thingType3, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Execute request
		rr := executeRequest(ctx, "GET", "/api/thing-types", nil)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)

		// Parse response
		var responseThingTypes []models.ThingType
		err = json.Unmarshal(rr.Body.Bytes(), &responseThingTypes)
		require.NoError(t, err)

		// Verify we have at least our 3 thing types
		assert.GreaterOrEqual(t, len(responseThingTypes), 3)

		// Verify our created thing types are in the response
		foundTypes := 0
		for _, tt := range responseThingTypes {
			if tt.ThingTypeID == thingType1.ThingTypeID || tt.ThingTypeID == thingType2.ThingTypeID || tt.ThingTypeID == thingType3.ThingTypeID {
				foundTypes++
			}
		}
		assert.Equal(t, 3, foundTypes, "All created thing types should be found in the response")
	})
}

// TestGetThingType tests the GetThingType handler
func TestGetThingType(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create test data using DataSetup map
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Execute request
		rr := executeRequest(ctx, "GET", "/api/thing-types/"+thingType.ThingTypeID, nil)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)

		// Parse response
		var responseThingType models.ThingType
		err = json.Unmarshal(rr.Body.Bytes(), &responseThingType)
		require.NoError(t, err)

		// Verify retrieved thing type matches created thing type
		assert.Equal(t, thingType.ThingTypeID, responseThingType.ThingTypeID)
		assert.Equal(t, thingType.Name, responseThingType.Name)
		assert.Equal(t, thingType.Description, responseThingType.Description)
	})
}

// TestGetThingTypeNotFound tests the GetThingType handler with a non-existent ID
func TestGetThingTypeNotFound(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Execute request with a valid UUID format but non-existent ID
		rr := executeRequest(ctx, "GET", "/api/thing-types/00000000-0000-0000-0000-000000000000", nil)

		// Check response
		assert.Equal(t, http.StatusNotFound, rr.Code)

		// Verify error message
		assert.Contains(t, rr.Body.String(), "not_found")
	})
}

// TestCreateThingType tests the CreateThingType handler
func TestCreateThingType(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create test data
		thingType := models.ThingType{
			Name:        "Test Thing Type",
			Description: "This is a test thing type",
		}

		// Marshal data
		body, _ := json.Marshal(thingType)

		// Execute request
		rr := executeRequest(ctx, "POST", "/api/thing-types", body)

		// Check response
		assert.Equal(t, http.StatusCreated, rr.Code)

		// Parse response
		var responseThingType models.ThingType
		err := json.Unmarshal(rr.Body.Bytes(), &responseThingType)
		require.NoError(t, err)

		// Verify created thing type
		assert.NotEmpty(t, responseThingType.ThingTypeID)
		assert.Equal(t, thingType.Name, responseThingType.Name)
		assert.Equal(t, thingType.Description, responseThingType.Description)

		// Verify the thing type was persisted
		storedThingType, err := store.AppStore.GetThingTypeByID(ctx, responseThingType.ThingTypeID)
		require.NoError(t, err)
		assert.NotNil(t, storedThingType)
		assert.Equal(t, responseThingType.ThingTypeID, storedThingType.ThingTypeID)
		assert.Equal(t, thingType.Name, storedThingType.Name)
		assert.Equal(t, thingType.Description, storedThingType.Description)
	})
}

// TestCreateThingTypeDuplicate tests the CreateThingType handler with a duplicate name
func TestCreateThingTypeDuplicate(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create first thing type using DataSetup map
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Try creating another thing type with the same name
		duplicateThingType := models.ThingType{
			Name:        thingType.Name, // Use same name to trigger unique constraint
			Description: "This is a duplicate thing type",
		}

		// Marshal data
		body, _ := json.Marshal(duplicateThingType)

		// Execute request
		rr := executeRequest(ctx, "POST", "/api/thing-types", body)

		// Check response - should fail due to unique constraint
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, strings.ToLower(rr.Body.String()), "internal_error")
	})
}

// TestCreateThingTypeInvalidInput tests the CreateThingType handler with invalid input
func TestCreateThingTypeInvalidInput(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create invalid test data (missing required name)
		thingType := models.ThingType{
			Name:        "", // Empty name should fail validation
			Description: "This is an invalid thing type",
		}

		// Marshal data
		body, _ := json.Marshal(thingType)

		// Execute request
		rr := executeRequest(ctx, "POST", "/api/thing-types", body)

		// Check response
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid_input")
	})
}

// TestUpdateThingType tests the UpdateThingType handler
func TestUpdateThingType(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create test data using DataSetup map
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Update data
		updatedThingType := models.ThingType{
			Name:        "Updated Name",
			Description: "Updated Description",
		}

		// Marshal data
		body, _ := json.Marshal(updatedThingType)

		// Execute request
		rr := executeRequest(ctx, "PUT", "/api/thing-types/"+thingType.ThingTypeID, body)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)

		// Parse response
		var responseThingType models.ThingType
		err = json.Unmarshal(rr.Body.Bytes(), &responseThingType)
		require.NoError(t, err)

		// Verify updated fields
		assert.Equal(t, thingType.ThingTypeID, responseThingType.ThingTypeID)
		assert.Equal(t, updatedThingType.Name, responseThingType.Name)
		assert.Equal(t, updatedThingType.Description, responseThingType.Description)

		// Verify the thing type was updated in the database
		var storedThingType models.ThingType
		err = tx.First(&storedThingType, "thing_type_id = ?", thingType.ThingTypeID).Error
		require.NoError(t, err)
		assert.Equal(t, updatedThingType.Name, storedThingType.Name)
		assert.Equal(t, updatedThingType.Description, storedThingType.Description)
	})
}

// TestUpdateThingTypeDuplicate tests the UpdateThingType handler with a duplicate name
func TestUpdateThingTypeDuplicate(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create two thing types using DataSetup maps
		thingType1, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)
		thingType2, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Try to update thingType2 with thingType1's name
		updateData := models.ThingType{
			Name: thingType1.Name, // Use same name to trigger unique constraint
		}

		// Marshal data
		body, _ := json.Marshal(updateData)

		// Execute request
		rr := executeRequest(ctx, "PUT", "/api/thing-types/"+thingType2.ThingTypeID, body)

		// Check response - should fail due to unique constraint
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, strings.ToLower(rr.Body.String()), "internal_error")
	})
}

// TestUpdateThingTypeNotFound tests the UpdateThingType handler with a non-existent ID
func TestUpdateThingTypeNotFound(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create request with a valid UUID format but non-existent ID
		updateData := models.ThingType{
			Name:        "Updated Name",
			Description: "Updated Description",
		}

		// Marshal data
		body, _ := json.Marshal(updateData)

		// Execute request
		rr := executeRequest(ctx, "PUT", "/api/thing-types/00000000-0000-0000-0000-000000000000", body)

		// Check response
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "not_found")
	})
}

// TestDeleteThingType tests the DeleteThingType handler
func TestDeleteThingType(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create test data using DataSetup map
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Execute request
		rr := executeRequest(ctx, "DELETE", "/api/thing-types/"+thingType.ThingTypeID, nil)

		// Check response
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify the thing type was deleted
		var deletedThingType models.ThingType
		err = tx.First(&deletedThingType, "thing_type_id = ?", thingType.ThingTypeID).Error
		assert.Error(t, err) // Should get an error as the record should be deleted
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}

// TestDeleteThingTypeNotFound tests the DeleteThingType handler with a non-existent ID
func TestDeleteThingTypeNotFound(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Execute request with a valid UUID format but non-existent ID
		rr := executeRequest(ctx, "DELETE", "/api/thing-types/00000000-0000-0000-0000-000000000000", nil)

		// Check response
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "not_found")
	})
}

// TestDeleteThingTypeWithDependencies tests trying to delete a thing type that has owned things
func TestDeleteThingTypeWithDependencies(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create test data using DataSetup map
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create an owned thing that references this thing type
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"ThingTypeID": thingType.ThingTypeID,
		})
		require.NoError(t, err)

		// Execute request to delete the thing type
		rr := executeRequest(ctx, "DELETE", "/api/thing-types/"+thingType.ThingTypeID, nil)

		// The cascade delete should allow this, so it should succeed
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify owned thing was also deleted due to cascade
		var deletedOwnedThing models.OwnedThing
		err = tx.First(&deletedOwnedThing, "owned_thing_id = ?", ownedThing.OwnedThingID).Error
		assert.Error(t, err) // Should get an error as the record should be deleted
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}
