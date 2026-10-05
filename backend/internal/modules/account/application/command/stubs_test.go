package command

import (
	"context"
	"time"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	sharedjwt "github.com/rcarreirao/seek_service_mcp/internal/shared/jwt"
)

type stubAccountRepo struct {
	getAccountByEmail func(ctx context.Context, email string) (*entity.Account, error)
	getLoginByEmail   func(ctx context.Context, email string) (*entity.AccountLogin, error)
	getProfile        func(ctx context.Context, accountID uint) (*entity.Guest, *entity.Owner, error)
	createGuest       func(ctx context.Context, a *entity.Account, l *entity.AccountLogin, g *entity.Guest) error
	createOwner       func(ctx context.Context, a *entity.Account, l *entity.AccountLogin, o *entity.Owner) error
}

func (s *stubAccountRepo) CreateGuest(ctx context.Context, a *entity.Account, l *entity.AccountLogin, g *entity.Guest) error {
	if s.createGuest != nil {
		return s.createGuest(ctx, a, l, g)
	}
	return nil
}

func (s *stubAccountRepo) CreateOwner(ctx context.Context, a *entity.Account, l *entity.AccountLogin, o *entity.Owner) error {
	if s.createOwner != nil {
		return s.createOwner(ctx, a, l, o)
	}
	return nil
}

func (s *stubAccountRepo) GetAccountByID(context.Context, uint) (*entity.Account, error) {
	return nil, nil
}

func (s *stubAccountRepo) GetAccountByEmail(ctx context.Context, email string) (*entity.Account, error) {
	if s.getAccountByEmail != nil {
		return s.getAccountByEmail(ctx, email)
	}
	return nil, errAccountNotFoundSentinel
}

func (s *stubAccountRepo) GetLoginByEmail(ctx context.Context, email string) (*entity.AccountLogin, error) {
	if s.getLoginByEmail != nil {
		return s.getLoginByEmail(ctx, email)
	}
	return nil, errAccountNotFoundSentinel
}

func (s *stubAccountRepo) GetProfile(ctx context.Context, accountID uint) (*entity.Guest, *entity.Owner, error) {
	if s.getProfile != nil {
		return s.getProfile(ctx, accountID)
	}
	return nil, nil, errProfileNotFoundSentinel
}

func (s *stubAccountRepo) Update(context.Context, *entity.Account) error { return nil }

func (s *stubAccountRepo) Delete(context.Context, uint) error { return nil }

type stubUserRepo struct {
	getByEmail func(ctx context.Context, email string) (*entity.User, error)
}

func (s *stubUserRepo) Create(context.Context, *entity.User) error { return nil }
func (s *stubUserRepo) GetByID(context.Context, uint) (*entity.User, error) {
	return nil, nil
}
func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if s.getByEmail != nil {
		return s.getByEmail(ctx, email)
	}
	return nil, errUserNotFoundSentinel
}
func (s *stubUserRepo) Update(context.Context, *entity.User) error { return nil }
func (s *stubUserRepo) Delete(context.Context, uint) error         { return nil }

type stubHasher struct {
	hashErr error
}

func (s *stubHasher) Hash(plain string) (string, error) {
	if s.hashErr != nil {
		return "", s.hashErr
	}
	return "hash:" + plain, nil
}

func (s *stubHasher) Compare(hash, plain string) error {
	if hash != "hash:"+plain {
		return errCompareFailed
	}
	return nil
}

type stubIssuer struct {
	panelErr  error
	adminErr  error
	lastPanel sharedjwt.PanelClaims
	lastAdmin sharedjwt.AdminClaims
	lastTTL   time.Duration
}

func (s *stubIssuer) IssueAdmin(c sharedjwt.AdminClaims, ttl time.Duration) (string, error) {
	if s.adminErr != nil {
		return "", s.adminErr
	}
	s.lastAdmin, s.lastTTL = c, ttl
	return "admin-token", nil
}

func (s *stubIssuer) IssuePanel(c sharedjwt.PanelClaims, ttl time.Duration) (string, error) {
	if s.panelErr != nil {
		return "", s.panelErr
	}
	s.lastPanel, s.lastTTL = c, ttl
	return "panel-token", nil
}