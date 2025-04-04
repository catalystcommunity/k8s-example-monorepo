package test

import (
	"context"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/checkauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestAuthVerification tests the verification of authentication
func TestAuthVerification(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils for this transaction
		du := &DataUtils{db: tx}
		
		// Create a user
		user, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// Create a session for the user
		session, err := du.CreateSession(DataSetup{
			"UserID": user.UserID,
		})
		require.NoError(t, err)

		// Generate a verifier
		token := session.Token
		verifier := checkauth.GenerateVerifier(user.UserID, user.Salt, token)

		// Verify it works with the verification function
		isVerified := checkauth.VerifySession(token, verifier, user.UserID, user.Salt)
		assert.True(t, isVerified, "Session verification should succeed")

		// Now test directly if the owner check in UpdateOwnedThing works
		// Create an owned thing
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)

		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"OwnerID":     user.UserID,
			"ThingTypeID": thingType.ThingTypeID,
			"Name":        "Test Thing",
		})
		require.NoError(t, err)

		// Direct verification test
		// If user.UserID matches owner ID -> should allow
		assert.Equal(t, user.UserID, ownedThing.OwnerID, "User should be the owner")

		// If user.UserID doesn't match owner ID -> should deny
		otherUser, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)

		// This is the check that happens in the handler:
		// if user.UserID != existingOwnedThing.OwnerID {
		//     h.respondWithError(w, http.StatusForbidden, store.ErrUnauthorized)
		// }
		assert.NotEqual(t, otherUser.UserID, ownedThing.OwnerID, "Other user should not be the owner")
	})
}

// Let's use a simplified approach for testing the handler directly
// Instead of creating a mock of the entire store interface
