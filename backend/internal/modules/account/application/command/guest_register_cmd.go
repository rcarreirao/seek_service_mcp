// Package command holds the write use cases of the account module.
package command

import (
	"context"
	"errors"
	"time"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/repository"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/valueobject"
	sharedjwt "github.com/rcarreirao/seek_service_mcp/internal/shared/jwt"
)

// tokenTTL is the panel/admin token lifetime.
const tokenTTL = 24 * time.Hour

// GuestRegisterCommand registers a customer account with its login and
// guest profile in a single transaction, then issues a panel token.
type GuestRegisterCommand struct {
	accounts repository.AccountRepository
	hasher   repository.PasswordHasher
	issuer   repository.TokenIssuer
}

// NewGuestRegisterCommand wires the command.
func NewGuestRegisterCommand(
	accounts repository.AccountRepository,
	hasher repository.PasswordHasher,
	issuer repository.TokenIssuer,
) *GuestRegisterCommand {
	return &GuestRegisterCommand{accounts: accounts, hasher: hasher, issuer: issuer}
}

// Execute validates the input, registers the aggregate, and returns a
// panel token.
func (c *GuestRegisterCommand) Execute(ctx context.Context, in dto.GuestRegisterInputDTO) (dto.AuthOutputDTO, error) {
	email, name, password, err := validateRegistration(in.Name, in.Email, in.Password)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	if _, err := c.accounts.GetAccountByEmail(ctx, email.String()); err == nil {
		return dto.AuthOutputDTO{}, errs.ErrAccountAlreadyExists
	} else if !errors.Is(err, errs.ErrAccountNotFound) {
		return dto.AuthOutputDTO{}, err
	}

	hash, err := c.hasher.Hash(password)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	account := &entity.Account{Name: name, Email: email, Phone: in.Phone}
	login := &entity.AccountLogin{Email: email, PasswordHash: hash}
	guest := &entity.Guest{}

	if err := c.accounts.CreateGuest(ctx, account, login, guest); err != nil {
		return dto.AuthOutputDTO{}, err
	}

	token, err := c.issuer.IssuePanel(sharedjwt.PanelClaims{
		AccountID: account.ID,
		GuestID:   &guest.ID,
		Email:     email.String(),
		Roles:     []string{"guest"},
	}, tokenTTL)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	return dto.AuthOutputDTO{Token: token, TokenType: "Bearer"}, nil
}

// validateRegistration applies the shared registration invariants.
func validateRegistration(name, rawEmail, password string) (valueobject.Email, string, string, error) {
	if name == "" {
		return valueobject.Email{}, "", "", errs.ErrNameRequired
	}
	email, err := valueobject.NewEmail(rawEmail)
	if err != nil {
		return valueobject.Email{}, "", "", errs.ErrInvalidEmail
	}
	if len(password) < 8 {
		return valueobject.Email{}, "", "", errs.ErrPasswordTooShort
	}
	return email, name, password, nil
}