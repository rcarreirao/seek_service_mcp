// Package repository declares the ports (interfaces) of the account module.
package repository

import (
	"context"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
)

// AccountRepository persists the account aggregate (account + login +
// role profile). Registration writes all rows in one transaction.
type AccountRepository interface {
	CreateGuest(ctx context.Context, a *entity.Account, login *entity.AccountLogin, g *entity.Guest) error
	CreateOwner(ctx context.Context, a *entity.Account, login *entity.AccountLogin, o *entity.Owner) error
	GetAccountByID(ctx context.Context, id uint) (*entity.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error)
	GetLoginByEmail(ctx context.Context, email string) (*entity.AccountLogin, error)
	// GetProfile returns the guest or the owner profile of the account
	// (at most one non-nil); errs.ErrProfileNotFound when the account has
	// neither (invariant violation).
	GetProfile(ctx context.Context, accountID uint) (*entity.Guest, *entity.Owner, error)
	Update(ctx context.Context, a *entity.Account) error
	// Delete soft-deletes the account and renames its email so the address
	// can be registered again.
	Delete(ctx context.Context, id uint) error
}