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

// GormUserRepo implements repository.UserRepository with GORM.
type GormUserRepo struct {
	db *gorm.DB
}

var _ repository.UserRepository = (*GormUserRepo)(nil)

// NewGormUserRepo wires the repository.
func NewGormUserRepo(db *gorm.DB) *GormUserRepo {
	return &GormUserRepo{db: db}
}

// Create inserts an admin user.
func (r *GormUserRepo) Create(ctx context.Context, u *entity.User) error {
	m := &model.UserModel{
		Name:         u.Name,
		Email:        u.Email.String(),
		PasswordHash: u.PasswordHash,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		if isDuplicate(err) {
			return errs.ErrUserAlreadyExists
		}
		return err
	}
	u.ID, u.CreatedAt, u.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

// GetByID returns a non-deleted user by id.
func (r *GormUserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return fromUserModel(&m), nil
}

// GetByEmail returns a non-deleted user by email.
func (r *GormUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
		return nil, err
	}
	return fromUserModel(&m), nil
}

// Update persists user changes.
func (r *GormUserRepo) Update(ctx context.Context, u *entity.User) error {
	m := &model.UserModel{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email.String(),
		PasswordHash: u.PasswordHash,
	}
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete soft-deletes a user and renames its email.
func (r *GormUserRepo) Delete(ctx context.Context, id uint) error {
	fakeEmail := fmt.Sprintf("deleted+%d@deleted.invalid", id)
	res := r.db.WithContext(ctx).Model(&model.UserModel{}).Where("id = ?", id).
		Updates(map[string]any{"email": fakeEmail, "deleted_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.ErrUserNotFound
	}
	return nil
}

func fromUserModel(m *model.UserModel) *entity.User {
	e, _ := valueobject.NewEmail(m.Email)
	return &entity.User{
		ID:           m.ID,
		Name:         m.Name,
		Email:        e,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}