package entity

import (
	"time"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/valueobject"
)

// User is the admin-panel identity. It never crosses into the panel realm.
type User struct {
	ID           uint
	Name         string
	Email        valueobject.Email
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}