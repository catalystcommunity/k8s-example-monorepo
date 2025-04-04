package models

import (
	"time"
)

// Session maps to the sessions table in the database
type Session struct {
	SessionID string    `gorm:"primaryKey;type:uuid;default:generate_ulid()" json:"session_id"`
	UserID    string    `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime:false;default:timezone('utc', now())" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false;default:timezone('utc', now())" json:"updated_at"`
	Token     string    `gorm:"type:text;not null" json:"token"`
	Rotated   bool      `gorm:"not null;default:false" json:"rotated"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for the model
func (Session) TableName() string {
	return "sessions"
}
