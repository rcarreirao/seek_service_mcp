package command

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
)

func TestOwnerRegisterSuccess(t *testing.T) {
	accounts := &stubAccountRepo{
		createOwner: func(_ context.Context, a *entity.Account, l *entity.AccountLogin, o *entity.Owner) error {
			a.ID, l.ID, o.ID, o.AccountID = 21, 22, 23, 21
			return nil
		},
	}
	issuer := &stubIssuer{}
	cmd := NewOwnerRegisterCommand(accounts, &stubHasher{}, issuer)

	out, err := cmd.Execute(context.Background(), dto.OwnerRegisterInputDTO{
		Name: "Bia", Email: "bia@example.com", Password: "password1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Token != "panel-token" {
		t.Fatalf("token = %q", out.Token)
	}
	if issuer.lastPanel.OwnerID == nil || *issuer.lastPanel.OwnerID != 23 || issuer.lastPanel.AccountID != 21 {
		t.Fatalf("claims = %+v", issuer.lastPanel)
	}
	if len(issuer.lastPanel.Roles) != 1 || issuer.lastPanel.Roles[0] != "owner" {
		t.Fatalf("roles = %v", issuer.lastPanel.Roles)
	}
}

func TestOwnerRegisterDuplicateEmail(t *testing.T) {
	accounts := &stubAccountRepo{
		getAccountByEmail: func(context.Context, string) (*entity.Account, error) {
			return &entity.Account{ID: 5}, nil
		},
	}
	cmd := NewOwnerRegisterCommand(accounts, &stubHasher{}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.OwnerRegisterInputDTO{
		Name: "Bia", Email: "bia@example.com", Password: "password1",
	})
	if !errors.Is(err, errs.ErrAccountAlreadyExists) {
		t.Fatalf("err = %v, want ErrAccountAlreadyExists", err)
	}
}

func TestOwnerRegisterValidation(t *testing.T) {
	cmd := NewOwnerRegisterCommand(&stubAccountRepo{}, &stubHasher{}, &stubIssuer{})

	_, err := cmd.Execute(context.Background(), dto.OwnerRegisterInputDTO{
		Name: "Bia", Email: "bad-email", Password: "password1",
	})
	if !errors.Is(err, errs.ErrInvalidEmail) {
		t.Fatalf("err = %v, want ErrInvalidEmail", err)
	}
}