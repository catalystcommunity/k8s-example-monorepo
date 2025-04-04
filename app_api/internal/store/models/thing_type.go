package models

import (
	"time"
)

// ThingType maps to the thing_types table in the database
type ThingType struct {
	ThingTypeID string    `gorm:"primaryKey;type:uuid;default:generate_ulid()" json:"thing_type_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime:false;default:timezone('utc', now())" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:false;default:timezone('utc', now())" json:"updated_at"`
	Name        string    `gorm:"type:text;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`

	// Relationships (virtual)
	OwnedThings []OwnedThing `gorm:"foreignKey:ThingTypeID" json:"owned_things,omitempty"`
}

// TableName specifies the table name for the model
func (ThingType) TableName() string {
	return "thing_types"
}
