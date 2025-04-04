package models

import (
	"time"

	"github.com/lib/pq"
)

// UserRole represents the role enum type from the database
type UserRole string

const (
	UserRoleUser    UserRole = "user"
	UserRoleSupport UserRole = "support"
	UserRoleAdmin   UserRole = "admin"
)

// User maps to the users table in the database
type User struct {
	UserID    string         `gorm:"primaryKey;type:uuid;default:generate_ulid()" json:"user_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime:false;default:timezone('utc', now())" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:false;default:timezone('utc', now())" json:"updated_at"`
	Username  string         `gorm:"type:text;not null" json:"username"`
	Email     string         `gorm:"type:text;not null" json:"email"`
	Password  []byte         `gorm:"type:bytea;not null" json:"-"`
	Salt      []byte         `gorm:"type:bytea;not null" json:"-"`
	Roles     pq.StringArray `gorm:"type:user_role[];default:ARRAY['user'::user_role];not null" json:"roles"`
}

// TableName specifies the table name for the model
func (User) TableName() string {
	return "users"
}
