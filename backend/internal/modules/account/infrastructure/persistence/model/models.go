// Package model holds the GORM persistence models of the account module.
// Explicit column tags are mandatory so precision/lengths never fall back
// to driver defaults.
package model

import (
	"time"

	"gorm.io/gorm"
)

// AccountModel maps the accounts table.
type AccountModel struct {
	ID        uint           `gorm:"primaryKey;column:id"`
	Name      string         `gorm:"column:name;type:varchar(190);not null"`
	Email     string         `gorm:"column:email;type:varchar(190);not null;uniqueIndex"`
	Phone     string         `gorm:"column:phone;type:varchar(50)"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

// TableName implements the table naming convention.
func (AccountModel) TableName() string { return "accounts" }

// AccountLoginModel maps the account_logins table.
type AccountLoginModel struct {
	ID           uint           `gorm:"primaryKey;column:id"`
	AccountID    uint           `gorm:"column:account_id;not null;uniqueIndex"`
	Email        string         `gorm:"column:email;type:varchar(190);not null;uniqueIndex"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(190);not null"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

// TableName implements the table naming convention.
func (AccountLoginModel) TableName() string { return "account_logins" }

// GuestModel maps the guests table.
type GuestModel struct {
	ID        uint           `gorm:"primaryKey;column:id"`
	AccountID uint           `gorm:"column:account_id;not null;uniqueIndex"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

// TableName implements the table naming convention.
func (GuestModel) TableName() string { return "guests" }

// OwnerModel maps the owners table.
type OwnerModel struct {
	ID        uint           `gorm:"primaryKey;column:id"`
	AccountID uint           `gorm:"column:account_id;not null;uniqueIndex"`
	Verified  bool           `gorm:"column:verified;not null;default:false"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

// TableName implements the table naming convention.
func (OwnerModel) TableName() string { return "owners" }

// UserModel maps the users table (admin realm).
type UserModel struct {
	ID           uint           `gorm:"primaryKey;column:id"`
	Name         string         `gorm:"column:name;type:varchar(190);not null"`
	Email        string         `gorm:"column:email;type:varchar(190);not null;uniqueIndex"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(190);not null"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

// TableName implements the table naming convention.
func (UserModel) TableName() string { return "users" }