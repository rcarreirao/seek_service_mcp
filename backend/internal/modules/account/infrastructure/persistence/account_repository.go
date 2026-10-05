// Package persistence implements the account module's ports with GORM.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/repository"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/valueobject"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/infrastructure/persistence/model"
)

// GormAccountRepo implements repository.AccountRepository with GORM.
type GormAccountRepo struct {
	db *gorm.DB
}

var _ repository.AccountRepository = (*GormAccountRepo)(nil)

// NewGormAccountRepo wires the repository.
func NewGormAccountRepo(db *gorm.DB) *GormAccountRepo {
	return &GormAccountRepo{db: db}
}

// CreateGuest inserts account + login + guest profile in one transaction.
func (r *GormAccountRepo) CreateGuest(ctx context.Context, a *entity.Account, login *entity.AccountLogin, g *entity.Guest) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc := toAccountModel(a)
		if err := tx.Create(acc).Error; err != nil {
			return mapDup(err)
		}
		a.ID, a.CreatedAt, a.UpdatedAt = acc.ID, acc.CreatedAt, acc.UpdatedAt

		lm := toLoginModel(login, acc.ID)
		if err := tx.Create(lm).Error; err != nil {
			return mapDup(err)
		}
		login.ID, login.CreatedAt, login.UpdatedAt = lm.ID, lm.CreatedAt, lm.UpdatedAt

		gm := &model.GuestModel{AccountID: acc.ID}
		if err := tx.Create(gm).Error; err != nil {
			return mapDup(err)
		}
		g.ID, g.AccountID, g.CreatedAt, g.UpdatedAt = gm.ID, gm.AccountID, gm.CreatedAt, gm.UpdatedAt
		return nil
	})
}

// CreateOwner inserts account + login + owner profile in one transaction.
func (r *GormAccountRepo) CreateOwner(ctx context.Context, a *entity.Account, login *entity.AccountLogin, o *entity.Owner) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc := toAccountModel(a)
		if err := tx.Create(acc).Error; err != nil {
			return mapDup(err)
		}
		a.ID, a.CreatedAt, a.UpdatedAt = acc.ID, acc.CreatedAt, acc.UpdatedAt

		lm := toLoginModel(login, acc.ID)
		if err := tx.Create(lm).Error; err != nil {
			return mapDup(err)
		}
		login.ID, login.CreatedAt, login.UpdatedAt = lm.ID, lm.CreatedAt, lm.UpdatedAt

		om := &model.OwnerModel{AccountID: acc.ID}
		if err := tx.Create(om).Error; err != nil {
			return mapDup(err)
		}
		o.ID, o.AccountID, o.CreatedAt, o.UpdatedAt = om.ID, om.AccountID, om.CreatedAt, om.UpdatedAt
		return nil
	})
}

// GetAccountByID returns a non-deleted account by id.
func (r *GormAccountRepo) GetAccountByID(ctx context.Context, id uint) (*entity.Account, error) {
	var m model.AccountModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return fromAccountModel(&m), nil
}

// GetAccountByEmail returns a non-deleted account by email.
func (r *GormAccountRepo) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
	var m model.AccountModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	return fromAccountModel(&m), nil
}

// GetLoginByEmail returns the login row whose account is still active.
func (r *GormAccountRepo) GetLoginByEmail(ctx context.Context, email string) (*entity.AccountLogin, error) {
	var m model.AccountLoginModel
	err := r.db.WithContext(ctx).
		Joins("JOIN accounts ON accounts.id = account_logins.account_id").
		Where("account_logins.email = ? AND accounts.deleted_at IS NULL", email).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrAccountNotFound
		}
		return nil, err
	}
	e, _ := valueobject.NewEmail(m.Email)
	return &entity.AccountLogin{
		ID:           m.ID,
		AccountID:    m.AccountID,
		Email:        e,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

// GetProfile returns the guest or owner profile of the account.
func (r *GormAccountRepo) GetProfile(ctx context.Context, accountID uint) (*entity.Guest, *entity.Owner, error) {
	var g model.GuestModel
	gErr := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&g).Error
	if gErr == nil {
		return &entity.Guest{ID: g.ID, AccountID: g.AccountID, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt}, nil, nil
	}
	if !errors.Is(gErr, gorm.ErrRecordNotFound) {
		return nil, nil, gErr
	}

	var o model.OwnerModel
	oErr := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&o).Error
	if oErr == nil {
		return nil, &entity.Owner{ID: o.ID, AccountID: o.AccountID, Verified: o.Verified, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt}, nil
	}
	if errors.Is(oErr, gorm.ErrRecordNotFound) {
		return nil, nil, errs.ErrProfileNotFound
	}
	return nil, nil, oErr
}

// Update persists account changes.
func (r *GormAccountRepo) Update(ctx context.Context, a *entity.Account) error {
	m := toAccountModel(a)
	m.ID = a.ID
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete soft-deletes the account aggregate and renames the emails so the
// real addresses can be registered again.
func (r *GormAccountRepo) Delete(ctx context.Context, id uint) error {
	now := time.Now()
	fakeEmail := fmt.Sprintf("deleted+%d@deleted.invalid", id)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.AccountModel{}).Where("id = ?", id).
			Updates(map[string]any{"email": fakeEmail, "deleted_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errs.ErrAccountNotFound
		}
		if err := tx.Model(&model.AccountLoginModel{}).Where("account_id = ?", id).
			Updates(map[string]any{"email": fakeEmail, "deleted_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.GuestModel{}).Where("account_id = ?", id).
			Update("deleted_at", now).Error; err != nil {
			return err
		}
		return tx.Model(&model.OwnerModel{}).Where("account_id = ?", id).
			Update("deleted_at", now).Error
	})
}

func toAccountModel(a *entity.Account) *model.AccountModel {
	return &model.AccountModel{
		Name:      a.Name,
		Email:     a.Email.String(),
		Phone:     a.Phone,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func fromAccountModel(m *model.AccountModel) *entity.Account {
	e, _ := valueobject.NewEmail(m.Email)
	return &entity.Account{
		ID:        m.ID,
		Name:      m.Name,
		Email:     e,
		Phone:     m.Phone,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toLoginModel(l *entity.AccountLogin, accountID uint) *model.AccountLoginModel {
	return &model.AccountLoginModel{
		AccountID:    accountID,
		Email:        l.Email.String(),
		PasswordHash: l.PasswordHash,
	}
}

func mapDup(err error) error {
	if err != nil && isDuplicate(err) {
		return errs.ErrAccountAlreadyExists
	}
	return err
}

func isDuplicate(err error) bool {
	var dup interface{ Number() int }
	return errors.As(err, &dup) && dup.Number() == 1062
}