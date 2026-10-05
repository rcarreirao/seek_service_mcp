package command

import (
	"context"
	"errors"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/repository"
	sharedjwt "github.com/rcarreirao/seek_service_mcp/internal/shared/jwt"
)

// LoginAccountCommand authenticates a guest or owner against its login
// and issues a panel token with the role profile ids embedded.
type LoginAccountCommand struct {
	accounts repository.AccountRepository
	hasher   repository.PasswordHasher
	issuer   repository.TokenIssuer
}

// NewLoginAccountCommand wires the command.
func NewLoginAccountCommand(
	accounts repository.AccountRepository,
	hasher repository.PasswordHasher,
	issuer repository.TokenIssuer,
) *LoginAccountCommand {
	return &LoginAccountCommand{accounts: accounts, hasher: hasher, issuer: issuer}
}

// Execute authenticates the credentials and returns a panel token.
func (c *LoginAccountCommand) Execute(ctx context.Context, in dto.LoginInputDTO) (dto.AuthOutputDTO, error) {
	login, err := c.accounts.GetLoginByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, errs.ErrAccountNotFound) {
			return dto.AuthOutputDTO{}, errs.ErrInvalidCredentials
		}
		return dto.AuthOutputDTO{}, err
	}

	if err := c.hasher.Compare(login.PasswordHash, in.Password); err != nil {
		return dto.AuthOutputDTO{}, errs.ErrInvalidCredentials
	}

	account, err := c.accounts.GetAccountByEmail(ctx, in.Email)
	if err != nil {
		return dto.AuthOutputDTO{}, errs.ErrInvalidCredentials
	}

	guest, owner, err := c.accounts.GetProfile(ctx, account.ID)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	claims := sharedjwt.PanelClaims{
		AccountID: account.ID,
		Email:     account.Email.String(),
	}
	switch {
	case guest != nil:
		claims.GuestID = &guest.ID
		claims.Roles = []string{"guest"}
	case owner != nil:
		claims.OwnerID = &owner.ID
		claims.Roles = []string{"owner"}
	default:
		return dto.AuthOutputDTO{}, errs.ErrProfileNotFound
	}

	token, err := c.issuer.IssuePanel(claims, tokenTTL)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	return dto.AuthOutputDTO{Token: token, TokenType: "Bearer"}, nil
}