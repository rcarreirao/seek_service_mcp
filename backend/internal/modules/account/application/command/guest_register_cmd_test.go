package command

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
)

var (
	errAccountNotFoundSentinel = errs.ErrAccountNotFound
	errProfileNotFoundSentinel = errs.ErrProfileNotFound
	errUserNotFoundSentinel    = errs.ErrUserNotFound
	errCompareFailed           = errors.New("compare failed")
)

func TestGuestRegisterSuccess(t *testing.T) {
	accounts := &stubAccountRepo{
		createGuest: func(_ context.Context, a *entity.Account, l *entity.AccountLogin, g *entity.Guest) error {
			a.ID = 11
			l.ID = 12
			g.ID = 13
			g.AccountID = 11
			return nil
		},
	}
	issuer := &stubIssuer{}
	cmd := NewGuestRegisterCommand(accounts, &stubHasher{}, issuer)

	out, err := cmd.Execute(context.Background(), dto.GuestRegisterInputDTO{
		Name: "Ana", Email: "ana@example.com", Phone: "999", Password: "password1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Token != "panel-token" || out.TokenType != "Bearer" {
		t.Fatalf("output = %+v", out)
	}
	if issuer.lastPanel.AccountID != 11 || issuer.lastPanel.GuestID == nil || *issuer.lastPanel.GuestID != 13 {
		t.Fatalf("claims = %+v", issuer.lastPanel)
	}
	if len(issuer.lastPanel.Roles) != 1 || issuer.lastPanel.Roles[0] != "guest" {
		t.Fatalf("roles = %v", issuer.lastPanel.Roles)
	}
}

func TestGuestRegisterValidation(t *testing.T) {
	cmd := NewGuestRegisterCommand(&stubAccountRepo{}, &stubHasher{}, &stubIssuer{})

	tests := []struct {
		name    string
		in      dto.GuestRegisterInputDTO
		wantErr error
	}{
		{"missing name", dto.GuestRegisterInputDTO{Email: "a@b.cc", Password: "password1"}, errs.ErrNameRequired},
		{"invalid email", dto.GuestRegisterInputDTO{Name: "A", Email: "nope", Password: "password1"}, errs.ErrInvalidEmail},
		{"short password", dto.GuestRegisterInputDTO{Name: "A", Email: "a@b.cc", Password: "short"}, errs.ErrPasswordTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cmd.Execute(context.Background(), tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGuestRegisterDuplicateEmail(t *testing.T) {
	accounts := &stubAccountRepo{
		getAccountByEmail: func(context.Context, string) (*entity.Account, error) {
			return &entity.Account{ID: 1}, nil
		},
	}
	issuer := &stubIssuer{}
	cmd := NewGuestRegisterCommand(accounts, &stubHasher{}, issuer)

	_, err := cmd.Execute(context.Background(), dto.GuestRegisterInputDTO{
		Name: "Ana", Email: "ana@example.com", Password: "password1",
	})
	if !errors.Is(err, errs.ErrAccountAlreadyExists) {
		t.Fatalf("err = %v, want ErrAccountAlreadyExists", err)
	}
	if issuer.lastPanel.AccountID != 0 {
		t.Fatal("token issued despite duplicate email")
	}
}

func TestGuestRegisterRepoFailurePropagates(t *testing.T) {
	repoErr := errors.New("db down")
	accounts := &stubAccountRepo{
		createGuest: func(context.Context, *entity.Account, *entity.AccountLogin, *entity.Guest) error {
			return repoErr
		},
	}
	issuer := &stubIssuer{}
	cmd := NewGuestRegisterCommand(accounts, &stubHasher{}, issuer)

	_, err := cmd.Execute(context.Background(), dto.GuestRegisterInputDTO{
		Name: "Ana", Email: "ana@example.com", Password: "password1",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("err = %v, want repo error", err)
	}
	if issuer.lastPanel.AccountID != 0 {
		t.Fatal("token issued despite repo failure")
	}
}

func TestGuestRegisterHashFailurePropagates(t *testing.T) {
	hashErr := errors.New("hasher broken")
	cmd := NewGuestRegisterCommand(&stubAccountRepo{}, &stubHasher{hashErr: hashErr}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.GuestRegisterInputDTO{
		Name: "Ana", Email: "ana@example.com", Password: "password1",
	})
	if !errors.Is(err, hashErr) {
		t.Fatalf("err = %v, want hash error", err)
	}
}