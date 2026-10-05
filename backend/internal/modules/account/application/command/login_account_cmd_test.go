package command

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/valueobject"
)

func TestLoginAccountGuestSuccess(t *testing.T) {
	email := mustEmail(t, "ana@example.com")
	accounts := &stubAccountRepo{
		getLoginByEmail: func(context.Context, string) (*entity.AccountLogin, error) {
			return &entity.AccountLogin{AccountID: 11, Email: email, PasswordHash: "hash:password1"}, nil
		},
		getAccountByEmail: func(context.Context, string) (*entity.Account, error) {
			return &entity.Account{ID: 11, Email: email}, nil
		},
		getProfile: func(context.Context, uint) (*entity.Guest, *entity.Owner, error) {
			return &entity.Guest{ID: 13, AccountID: 11}, nil, nil
		},
	}
	issuer := &stubIssuer{}
	cmd := NewLoginAccountCommand(accounts, &stubHasher{}, issuer)

	out, err := cmd.Execute(context.Background(), dto.LoginInputDTO{Email: "ana@example.com", Password: "password1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Token != "panel-token" {
		t.Fatalf("token = %q", out.Token)
	}
	if issuer.lastPanel.AccountID != 11 || issuer.lastPanel.GuestID == nil || *issuer.lastPanel.GuestID != 13 {
		t.Fatalf("claims = %+v", issuer.lastPanel)
	}
	if issuer.lastPanel.Email != "ana@example.com" {
		t.Fatalf("email claim = %q", issuer.lastPanel.Email)
	}
}

func TestLoginAccountOwnerSuccess(t *testing.T) {
	email := mustEmail(t, "bia@example.com")
	accounts := &stubAccountRepo{
		getLoginByEmail: func(context.Context, string) (*entity.AccountLogin, error) {
			return &entity.AccountLogin{AccountID: 21, Email: email, PasswordHash: "hash:password1"}, nil
		},
		getAccountByEmail: func(context.Context, string) (*entity.Account, error) {
			return &entity.Account{ID: 21, Email: email}, nil
		},
		getProfile: func(context.Context, uint) (*entity.Guest, *entity.Owner, error) {
			return nil, &entity.Owner{ID: 23, AccountID: 21}, nil
		},
	}
	issuer := &stubIssuer{}
	cmd := NewLoginAccountCommand(accounts, &stubHasher{}, issuer)

	_, err := cmd.Execute(context.Background(), dto.LoginInputDTO{Email: "bia@example.com", Password: "password1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if issuer.lastPanel.OwnerID == nil || *issuer.lastPanel.OwnerID != 23 {
		t.Fatalf("claims = %+v", issuer.lastPanel)
	}
	if len(issuer.lastPanel.Roles) != 1 || issuer.lastPanel.Roles[0] != "owner" {
		t.Fatalf("roles = %v", issuer.lastPanel.Roles)
	}
}

func TestLoginAccountInvalidCredentials(t *testing.T) {
	email := mustEmail(t, "ana@example.com")
	accounts := &stubAccountRepo{
		getLoginByEmail: func(context.Context, string) (*entity.AccountLogin, error) {
			return &entity.AccountLogin{AccountID: 11, Email: email, PasswordHash: "hash:password1"}, nil
		},
		getAccountByEmail: func(context.Context, string) (*entity.Account, error) {
			return &entity.Account{ID: 11, Email: email}, nil
		},
		getProfile: func(context.Context, uint) (*entity.Guest, *entity.Owner, error) {
			return &entity.Guest{ID: 13, AccountID: 11}, nil, nil
		},
	}
	cmd := NewLoginAccountCommand(accounts, &stubHasher{}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.LoginInputDTO{Email: "ana@example.com", Password: "wrong-pass"})
	if !errors.Is(err, errs.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginAccountUnknownEmail(t *testing.T) {
	cmd := NewLoginAccountCommand(&stubAccountRepo{}, &stubHasher{}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.LoginInputDTO{Email: "ghost@example.com", Password: "password1"})
	if !errors.Is(err, errs.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials (no user enumeration)", err)
	}
}

func mustEmail(t *testing.T, s string) valueobject.Email {
	t.Helper()
	e, err := valueobject.NewEmail(s)
	if err != nil {
		t.Fatalf("email: %v", err)
	}
	return e
}