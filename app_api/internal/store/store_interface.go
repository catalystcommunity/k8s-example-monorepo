package store

import (
	"context"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"gorm.io/gorm"
)

var AppStore Store

// GetDB returns the database connection
func GetDB() *gorm.DB {
	// This is a convenience function to access the DB from other packages
	// It's used by the transaction middleware
	if store, ok := AppStore.(interface{ GetDB() *gorm.DB }); ok {
		return store.GetDB()
	}
	return nil
}

type Store interface {
	Initialize() (deferredFunc func(), err error)

	// ThingType operations
	GetThingTypes(ctx context.Context) ([]models.ThingType, error)
	GetThingTypeByID(ctx context.Context, id string) (*models.ThingType, error)
	CreateThingType(ctx context.Context, thingType *models.ThingType) error
	UpdateThingType(ctx context.Context, thingType *models.ThingType) error
	DeleteThingType(ctx context.Context, id string) error

	// OwnedThing operations
	GetOwnedThingsByOwner(ctx context.Context, ownerID string) ([]models.OwnedThing, error)
	GetOwnedThingByID(ctx context.Context, id string) (*models.OwnedThing, error)
	CreateOwnedThing(ctx context.Context, ownedThing *models.OwnedThing) error
	UpdateOwnedThing(ctx context.Context, ownedThing *models.OwnedThing) error
	DeleteOwnedThing(ctx context.Context, id string) error

	// Borrow operations
	GetBorrowsByBorrower(ctx context.Context, borrowerID string) ([]models.Borrow, error)
	GetBorrowsByOwnedThing(ctx context.Context, ownedThingID string) ([]models.Borrow, error)
	GetBorrowByID(ctx context.Context, id string) (*models.Borrow, error)
	CreateBorrow(ctx context.Context, borrow *models.Borrow) error
	UpdateBorrow(ctx context.Context, borrow *models.Borrow) error
	DeleteBorrow(ctx context.Context, id string) error
}
