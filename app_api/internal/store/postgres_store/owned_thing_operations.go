package postgres_store

import (
	"context"
	"errors"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"gorm.io/gorm"
)

// GetOwnedThingsByOwner retrieves all owned things for a given owner
func (s PostgresDbStore) GetOwnedThingsByOwner(ctx context.Context, ownerID string) ([]models.OwnedThing, error) {
	var ownedThings []models.OwnedThing
	result := GetDBFromContext(ctx).WithContext(ctx).Where("owner_id = ?", ownerID).Find(&ownedThings)
	if result.Error != nil {
		return nil, result.Error
	}
	return ownedThings, nil
}

// GetOwnedThingByID retrieves an owned thing by its ID
func (s PostgresDbStore) GetOwnedThingByID(ctx context.Context, id string) (*models.OwnedThing, error) {
	var ownedThing models.OwnedThing
	result := GetDBFromContext(ctx).WithContext(ctx).First(&ownedThing, "owned_thing_id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, result.Error
	}
	return &ownedThing, nil
}

// CreateOwnedThing creates a new owned thing
func (s PostgresDbStore) CreateOwnedThing(ctx context.Context, ownedThing *models.OwnedThing) error {
	return GetDBFromContext(ctx).WithContext(ctx).Create(ownedThing).Error
}

// UpdateOwnedThing updates an existing owned thing
func (s PostgresDbStore) UpdateOwnedThing(ctx context.Context, ownedThing *models.OwnedThing) error {
	return GetDBFromContext(ctx).WithContext(ctx).Save(ownedThing).Error
}

// DeleteOwnedThing deletes an owned thing by its ID
func (s PostgresDbStore) DeleteOwnedThing(ctx context.Context, id string) error {
	return GetDBFromContext(ctx).WithContext(ctx).Delete(&models.OwnedThing{}, "owned_thing_id = ?", id).Error
}