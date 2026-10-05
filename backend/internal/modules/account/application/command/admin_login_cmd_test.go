package command

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
)

func TestAdminLoginSuccess(t *testing.T) {
	email := mustEmail(t, "admin@test.com")
	users := &stubUserRepo{
		getByEmail: func(context.Context, string) (*entity.User, error) {
			return &entity.User{ID: 1, Email: email, PasswordHash: "hash:admin123"}, nil
		},
	}
	issuer := &stubIssuer{}
	cmd := NewAdminLoginCommand(users, &stubHasher{}, issuer)

	out, err := cmd.Execute(context.Background(), dto.AdminLoginInputDTO{Email: "admin@test.com", Password: "admin123"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Token != "admin-token" {
		t.Fatalf("token = %q", out.Token)
	}
	if issuer.lastAdmin.UserID != 1 {
		t.Fatalf("claims = %+v", issuer.lastAdmin)
	}
}

func TestAdminLoginInvalidCredentials(t *testing.T) {
	email := mustEmail(t, "admin@test.com")
	users := &stubUserRepo{
		getByEmail: func(context.Context, string) (*entity.User, error) {
			return &entity.User{ID: 1, Email: email, PasswordHash: "hash:admin123"}, nil
		},
	}
	cmd := NewAdminLoginCommand(users, &stubHasher{}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.AdminLoginInputDTO{Email: "admin@test.com", Password: "nope"})
	if !errors.Is(err, errs.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAdminLoginUnknownEmail(t *testing.T) {
	cmd := NewAdminLoginCommand(&stubUserRepo{}, &stubHasher{}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.AdminLoginInputDTO{Email: "ghost@test.com", Password: "admin123"})
	if !errors.Is(err, errs.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}