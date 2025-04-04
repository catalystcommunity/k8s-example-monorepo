package postgres_store

import (
	"context"
	"errors"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"gorm.io/gorm"
)

// GetThingTypes retrieves all thing types
func (s PostgresDbStore) GetThingTypes(ctx context.Context) ([]models.ThingType, error) {
	var thingTypes []models.ThingType
	result := GetDBFromContext(ctx).WithContext(ctx).Find(&thingTypes)
	if result.Error != nil {
		return nil, result.Error
	}
	return thingTypes, nil
}

// GetThingTypeByID retrieves a thing type by its ID
func (s PostgresDbStore) GetThingTypeByID(ctx context.Context, id string) (*models.ThingType, error) {
	var thingType models.ThingType
	result := GetDBFromContext(ctx).WithContext(ctx).First(&thingType, "thing_type_id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, result.Error
	}
	return &thingType, nil
}

// CreateThingType creates a new thing type
func (s PostgresDbStore) CreateThingType(ctx context.Context, thingType *models.ThingType) error {
	return GetDBFromContext(ctx).WithContext(ctx).Create(thingType).Error
}

// UpdateThingType updates an existing thing type
func (s PostgresDbStore) UpdateThingType(ctx context.Context, thingType *models.ThingType) error {
	return GetDBFromContext(ctx).WithContext(ctx).Save(thingType).Error
}

// DeleteThingType deletes a thing type by its ID
func (s PostgresDbStore) DeleteThingType(ctx context.Context, id string) error {
	return GetDBFromContext(ctx).WithContext(ctx).Delete(&models.ThingType{}, "thing_type_id = ?", id).Error
}