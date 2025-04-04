package test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// This function has been moved to auth_test_utils.go

// TestUpdateOwnedThingByNonOwner tests that a thing cannot be modified by anyone except its owner
func TestUpdateOwnedThingByNonOwner(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}

		// Create two users: owner and requester
		owner, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		requester, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create an owned thing owned by the owner
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"OwnerID":     owner.UserID,
			"ThingTypeID": thingType.ThingTypeID,
			"Name":        "Original Name",
		})
		require.NoError(t, err)

		// Try to update the owned thing as the requester (non-owner)
		updateData := map[string]string{
			"name": "Updated Name",
		}
		body, _ := json.Marshal(updateData)

		// Execute request as the requester
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodPut, "/api/owned-things/"+ownedThing.OwnedThingID, body, requester)

		// Should return a 403 Forbidden error
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "forbidden")
	})
}

// TestCreateDuplicateOwnedThing tests that a thing cannot have the same name and thing_type for the same owner
func TestCreateDuplicateOwnedThing(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create a user
		user, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create first owned thing
		_, err = du.CreateOwnedThing(DataSetup{
			"OwnerID":     user.UserID,
			"ThingTypeID": thingType.ThingTypeID,
			"Name":        "My Thing",
		})
		require.NoError(t, err)

		// Try to create another owned thing with the same name and thing type for the same owner
		createData := map[string]string{
			"name":          "My Thing", // Same name
			"thing_type_id": thingType.ThingTypeID,
			"owner_id":      user.UserID,
		}
		body, _ := json.Marshal(createData)

		// Execute request as the user
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodPost, "/api/owned-things", body, user)

		// Should return a 409 Conflict error
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "already_exists")
	})
}

// TestUpdateToDuplicateOwnedThing tests that a thing cannot be updated to match another thing's name and type
func TestUpdateToDuplicateOwnedThing(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create a user
		user, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create two owned things with different names
		_, err = du.CreateOwnedThing(DataSetup{
			"OwnerID":     user.UserID,
			"ThingTypeID": thingType.ThingTypeID,
			"Name":        "First Thing",
		})
		require.NoError(t, err)

		ownedThing2, err := du.CreateOwnedThing(DataSetup{
			"OwnerID":     user.UserID,
			"ThingTypeID": thingType.ThingTypeID,
			"Name":        "Second Thing",
		})
		require.NoError(t, err)

		// Try to update the second owned thing to have the same name as the first
		updateData := map[string]string{
			"name": "First Thing", // Match first thing's name
		}
		body, _ := json.Marshal(updateData)

		// Execute request as the user
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodPut, "/api/owned-things/"+ownedThing2.OwnedThingID, body, user)

		// Should return a 409 Conflict error
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "already_exists")
	})
}

// TestCreateOwnedThingForOtherUser tests that a thing cannot be created to be owned by anyone but the requesting user
func TestCreateOwnedThingForOtherUser(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create two users: requester and other user
		requester, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		otherUser, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Try to create an owned thing for the other user
		createData := map[string]string{
			"name":          "New Thing",
			"thing_type_id": thingType.ThingTypeID,
			"owner_id":      otherUser.UserID, // Try to create for another user
		}
		body, _ := json.Marshal(createData)

		// Execute request as the requester
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodPost, "/api/owned-things", body, requester)

		// Should return a 403 Forbidden error
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "forbidden")
	})
}

// TestDeleteOwnedThingNonOwnerNonAdmin tests that a thing cannot be deleted by a non-owner non-admin user
func TestDeleteOwnedThingNonOwnerNonAdmin(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create two users: owner and requester (non-admin)
		owner, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		nonAdminUser, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create an owned thing owned by the owner
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"OwnerID":     owner.UserID,
			"ThingTypeID": thingType.ThingTypeID,
		})
		require.NoError(t, err)

		// Try to delete the owned thing as a non-owner non-admin
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodDelete, "/api/owned-things/"+ownedThing.OwnedThingID, nil, nonAdminUser)

		// Should return a 403 Forbidden error
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "forbidden")
	})
}

// TestDeleteOwnedThingByOwner tests that a thing can be deleted by its owner
func TestDeleteOwnedThingByOwner(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create a user
		owner, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create an owned thing
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"OwnerID":     owner.UserID,
			"ThingTypeID": thingType.ThingTypeID,
		})
		require.NoError(t, err)

		// Delete the owned thing as the owner
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodDelete, "/api/owned-things/"+ownedThing.OwnedThingID, nil, owner)

		// Should succeed
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify the thing was deleted
		var deletedThing models.OwnedThing
		err = tx.First(&deletedThing, "owned_thing_id = ?", ownedThing.OwnedThingID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

// TestDeleteOwnedThingByAdmin tests that a thing can be deleted by an admin user
func TestDeleteOwnedThingByAdmin(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create an owner
		owner, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create an admin user
		adminUser, err := CreateAdminUser(tx)
		require.NoError(t, err)

		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		// Create an owned thing
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"OwnerID":     owner.UserID,
			"ThingTypeID": thingType.ThingTypeID,
		})
		require.NoError(t, err)

		// Delete the owned thing as the admin
		rr := ExecuteAuthenticatedRequest(ctx, tx, http.MethodDelete, "/api/owned-things/"+ownedThing.OwnedThingID, nil, adminUser)

		// Should succeed
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify the thing was deleted
		var deletedThing models.OwnedThing
		err = tx.First(&deletedThing, "owned_thing_id = ?", ownedThing.OwnedThingID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}
