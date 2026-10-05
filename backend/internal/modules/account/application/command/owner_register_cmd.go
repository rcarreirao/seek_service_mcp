package command

import (
	"context"
	"errors"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/entity"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/repository"
	sharedjwt "github.com/rcarreirao/seek_service_mcp/internal/shared/jwt"
)

// OwnerRegisterCommand registers a shop-owner account with its login and
// owner profile in a single transaction, then issues a panel token.
type OwnerRegisterCommand struct {
	accounts repository.AccountRepository
	hasher   repository.PasswordHasher
	issuer   repository.TokenIssuer
}

// NewOwnerRegisterCommand wires the command.
func NewOwnerRegisterCommand(
	accounts repository.AccountRepository,
	hasher repository.PasswordHasher,
	issuer repository.TokenIssuer,
) *OwnerRegisterCommand {
	return &OwnerRegisterCommand{accounts: accounts, hasher: hasher, issuer: issuer}
}

// Execute validates the input, registers the aggregate, and returns a
// panel token.
func (c *OwnerRegisterCommand) Execute(ctx context.Context, in dto.OwnerRegisterInputDTO) (dto.AuthOutputDTO, error) {
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
	owner := &entity.Owner{}

	if err := c.accounts.CreateOwner(ctx, account, login, owner); err != nil {
		return dto.AuthOutputDTO{}, err
	}

	token, err := c.issuer.IssuePanel(sharedjwt.PanelClaims{
		AccountID: account.ID,
		OwnerID:   &owner.ID,
		Email:     email.String(),
		Roles:     []string{"owner"},
	}, tokenTTL)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	return dto.AuthOutputDTO{Token: token, TokenType: "Bearer"}, nil
}