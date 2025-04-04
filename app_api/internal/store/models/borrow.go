package models

import (
	"time"
)

// Borrow maps to the borrows table in the database
type Borrow struct {
	BorrowID      string     `gorm:"primaryKey;type:uuid;default:generate_ulid()" json:"borrow_id"`
	CreatedAt     time.Time  `gorm:"autoCreateTime:false;default:timezone('utc', now())" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime:false;default:timezone('utc', now())" json:"updated_at"`
	BorrowerID    string     `gorm:"type:uuid;not null" json:"borrower_id"`
	OwnedThingID  string     `gorm:"type:uuid;not null" json:"owned_thing_id"`
	BorrowedAt    time.Time  `gorm:"default:timezone('utc', now())" json:"borrowed_at"`
	BorrowedUntil *time.Time `gorm:"type:timestamp" json:"borrowed_until,omitempty"`
	ReturnedAt    *time.Time `gorm:"type:timestamp" json:"returned_at,omitempty"`
	Reposessed    bool       `gorm:"default:false" json:"reposessed"`

	// Relationships
	BorrowerRef User       `gorm:"foreignKey:BorrowerID" json:"borrower,omitempty"`
	OwnedThing  OwnedThing `gorm:"foreignKey:OwnedThingID" json:"owned_thing,omitempty"`
}

// TableName specifies the table name for the model
func (Borrow) TableName() string {
	return "borrows"
}
