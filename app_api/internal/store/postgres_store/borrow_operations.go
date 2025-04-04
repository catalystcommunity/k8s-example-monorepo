package postgres_store

import (
	"context"
	"errors"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"gorm.io/gorm"
)

// GetBorrowsByBorrower retrieves all borrows for a given borrower
func (s PostgresDbStore) GetBorrowsByBorrower(ctx context.Context, borrowerID string) ([]models.Borrow, error) {
	var borrows []models.Borrow
	result := GetDBFromContext(ctx).WithContext(ctx).Where("borrower_id = ?", borrowerID).Find(&borrows)
	if result.Error != nil {
		return nil, result.Error
	}
	return borrows, nil
}

// GetBorrowsByOwnedThing retrieves all borrows for a given owned thing
func (s PostgresDbStore) GetBorrowsByOwnedThing(ctx context.Context, ownedThingID string) ([]models.Borrow, error) {
	var borrows []models.Borrow
	result := GetDBFromContext(ctx).WithContext(ctx).Where("owned_thing_id = ?", ownedThingID).Find(&borrows)
	if result.Error != nil {
		return nil, result.Error
	}
	return borrows, nil
}

// GetBorrowByID retrieves a borrow by its ID
func (s PostgresDbStore) GetBorrowByID(ctx context.Context, id string) (*models.Borrow, error) {
	var borrow models.Borrow
	result := GetDBFromContext(ctx).WithContext(ctx).First(&borrow, "borrow_id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, result.Error
	}
	return &borrow, nil
}

// CreateBorrow creates a new borrow
func (s PostgresDbStore) CreateBorrow(ctx context.Context, borrow *models.Borrow) error {
	return GetDBFromContext(ctx).WithContext(ctx).Create(borrow).Error
}

// UpdateBorrow updates an existing borrow
func (s PostgresDbStore) UpdateBorrow(ctx context.Context, borrow *models.Borrow) error {
	return GetDBFromContext(ctx).WithContext(ctx).Save(borrow).Error
}

// DeleteBorrow deletes a borrow by its ID
func (s PostgresDbStore) DeleteBorrow(ctx context.Context, id string) error {
	return GetDBFromContext(ctx).WithContext(ctx).Delete(&models.Borrow{}, "borrow_id = ?", id).Error
}