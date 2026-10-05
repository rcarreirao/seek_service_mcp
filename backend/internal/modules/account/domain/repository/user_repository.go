package repository

import (
	"context"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
)

// UserRepository persists admin-panel users (admin realm only).
type UserRepository interface {
	Create(ctx context.Context, u *entity.User) error
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, u *entity.User) error
	Delete(ctx context.Context, id uint) error
}