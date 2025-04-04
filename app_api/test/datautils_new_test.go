package test

import (
	"context"
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateUserMap(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils instance for this transaction
		du := &DataUtils{db: tx}
		
		user, err := du.CreateUser(DataSetup{
			"Username": "testuser",
			"Email":    "test@example.com",
		})
	
		require.NoError(t, err)
		assert.NotEmpty(t, user.UserID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NotEmpty(t, user.Password)
		assert.NotEmpty(t, user.Salt)
		assert.Contains(t, user.Roles, string(models.UserRoleUser))
	})
}

func TestCreateSessionMap(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils instance for this transaction
		du := &DataUtils{db: tx}
		
		// Test with explicit user ID
		user, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)
		
		session, err := du.CreateSession(DataSetup{
			"UserID": user.UserID,
			"Token":  "testtoken",
		})
	
		require.NoError(t, err)
		assert.NotEmpty(t, session.SessionID)
		assert.Equal(t, user.UserID, session.UserID)
		assert.Equal(t, "testtoken", session.Token)
		assert.False(t, session.Rotated)
		
		// Test with automatic user creation
		session2, err := du.CreateSession(DataSetup{})
		require.NoError(t, err)
		assert.NotEmpty(t, session2.SessionID)
		assert.NotEmpty(t, session2.UserID)
		assert.NotEmpty(t, session2.Token)
	})
}

func TestCreateThingTypeMap(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils instance for this transaction
		du := &DataUtils{db: tx}
		
		thingType, err := du.CreateThingType(DataSetup{
			"Name":        "Test Thing Type",
			"Description": "A test description",
		})
	
		require.NoError(t, err)
		assert.NotEmpty(t, thingType.ThingTypeID)
		assert.Equal(t, "Test Thing Type", thingType.Name)
		assert.Equal(t, "A test description", thingType.Description)
	})
}

func TestCreateOwnedThingMap(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils instance for this transaction
		du := &DataUtils{db: tx}
		
		// Test with explicit foreign keys
		thingType, err := du.CreateThingType(DataSetup{})
		require.NoError(t, err)
		
		user, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)
		
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"Name":        "Test Owned Thing",
			"ThingTypeID": thingType.ThingTypeID,
			"OwnerID":     user.UserID,
		})
	
		require.NoError(t, err)
		assert.NotEmpty(t, ownedThing.OwnedThingID)
		assert.Equal(t, "Test Owned Thing", ownedThing.Name)
		assert.Equal(t, thingType.ThingTypeID, ownedThing.ThingTypeID)
		assert.Equal(t, user.UserID, ownedThing.OwnerID)
		
		// Test with automatic creation of foreign keys
		ownedThing2, err := du.CreateOwnedThing(DataSetup{})
		require.NoError(t, err)
		assert.NotEmpty(t, ownedThing2.OwnedThingID)
		assert.NotEmpty(t, ownedThing2.ThingTypeID)
		assert.NotEmpty(t, ownedThing2.OwnerID)
	})
}

func TestCreateBorrowMap(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils instance for this transaction
		du := &DataUtils{db: tx}
		
		// Test with explicit foreign keys
		user, err := du.CreateUser(DataSetup{})
		require.NoError(t, err)
		
		ownedThing, err := du.CreateOwnedThing(DataSetup{})
		require.NoError(t, err)
		
		borrow, err := du.CreateBorrow(DataSetup{
			"BorrowerID":   user.UserID,
			"OwnedThingID": ownedThing.OwnedThingID,
		})
	
		require.NoError(t, err)
		assert.NotEmpty(t, borrow.BorrowID)
		assert.Equal(t, user.UserID, borrow.BorrowerID)
		assert.Equal(t, ownedThing.OwnedThingID, borrow.OwnedThingID)
		assert.False(t, borrow.BorrowedAt.IsZero())
		
		// Test with automatic creation of foreign keys
		borrow2, err := du.CreateBorrow(DataSetup{})
		require.NoError(t, err)
		assert.NotEmpty(t, borrow2.BorrowID)
		assert.NotEmpty(t, borrow2.BorrowerID)
		assert.NotEmpty(t, borrow2.OwnedThingID)
	})
}

// Test transactional versions of the functions
func TestCreateDataWithTransactions(t *testing.T) {
	RunTransactionalTest(t, func(ctx context.Context, tx *gorm.DB) {
		// Create a new DataUtils instance for this transaction
		du := &DataUtils{db: tx}
		
		// Create a thing type
		thingType, err := du.CreateThingType(DataSetup{
			"Name":        "Transactional Thing Type",
			"Description": "Created in a transaction",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, thingType.ThingTypeID)
		
		// Create a user
		user, err := du.CreateUser(DataSetup{
			"Username": "transuser",
			"Email":    "trans@example.com",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, user.UserID)
		
		// Create an owned thing
		ownedThing, err := du.CreateOwnedThing(DataSetup{
			"Name":        "Transactional Thing",
			"ThingTypeID": thingType.ThingTypeID,
			"OwnerID":     user.UserID,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, ownedThing.OwnedThingID)
		
		// Create a borrow
		borrow, err := du.CreateBorrow(DataSetup{
			"BorrowerID":   user.UserID,
			"OwnedThingID": ownedThing.OwnedThingID,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, borrow.BorrowID)
		
		// Verify the items were saved to the database in this transaction
		var dbBorrow models.Borrow
		err = tx.First(&dbBorrow, "borrow_id = ?", borrow.BorrowID).Error
		require.NoError(t, err)
		assert.Equal(t, borrow.BorrowID, dbBorrow.BorrowID)
		assert.Equal(t, user.UserID, dbBorrow.BorrowerID)
		assert.Equal(t, ownedThing.OwnedThingID, dbBorrow.OwnedThingID)
	})
}