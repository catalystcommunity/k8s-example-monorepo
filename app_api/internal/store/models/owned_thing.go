package models

import (
	"time"
)

// OwnedThing maps to the owned_things table in the database
type OwnedThing struct {
	OwnedThingID string    `gorm:"primaryKey;type:uuid;default:generate_ulid()" json:"owned_thing_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime:false;default:timezone('utc', now())" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime:false;default:timezone('utc', now())" json:"updated_at"`
	Name         string    `gorm:"type:text" json:"name"`
	ThingTypeID  string    `gorm:"type:uuid;not null" json:"thing_type_id"`
	OwnerID      string    `gorm:"type:uuid;not null" json:"owner_id"`

	// Relationships
	ThingType ThingType `gorm:"foreignKey:ThingTypeID" json:"thing_type,omitempty"`
	Owner     User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Borrows   []Borrow  `gorm:"foreignKey:OwnedThingID" json:"borrows,omitempty"`
}

// TableName specifies the table name for the model
func (OwnedThing) TableName() string {
	return "owned_things"
}
