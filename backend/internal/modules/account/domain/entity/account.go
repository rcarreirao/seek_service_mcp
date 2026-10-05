// Package entity holds plain domain entities for the account module.
package entity

import (
	"time"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/valueobject"
)

// Account is the platform identity shared by guests and owners.
type Account struct {
	ID        uint
	Name      string
	Email     valueobject.Email
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// AccountLogin holds the credentials for an account, isolated from the
// profile so credentials never leak into profile reads.
type AccountLogin struct {
	ID           uint
	AccountID    uint
	Email        valueobject.Email
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}