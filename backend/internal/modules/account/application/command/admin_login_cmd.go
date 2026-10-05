package command

import (
	"context"
	"errors"

	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/dto"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/errs"
	"github.com/rcarreirao/seek_service_mcp/internal/modules/account/domain/repository"
	sharedjwt "github.com/rcarreirao/seek_service_mcp/internal/shared/jwt"
)

// AdminLoginCommand authenticates an admin-panel user and issues an
// admin-realm token.
type AdminLoginCommand struct {
	users   repository.UserRepository
	hasher  repository.PasswordHasher
	issuer  repository.TokenIssuer
}

// NewAdminLoginCommand wires the command.
func NewAdminLoginCommand(
	users repository.UserRepository,
	hasher repository.PasswordHasher,
	issuer repository.TokenIssuer,
) *AdminLoginCommand {
	return &AdminLoginCommand{users: users, hasher: hasher, issuer: issuer}
}

// Execute authenticates the admin credentials and returns an admin token.
func (c *AdminLoginCommand) Execute(ctx context.Context, in dto.AdminLoginInputDTO) (dto.AuthOutputDTO, error) {
	user, err := c.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return dto.AuthOutputDTO{}, errs.ErrInvalidCredentials
		}
		return dto.AuthOutputDTO{}, err
	}

	if err := c.hasher.Compare(user.PasswordHash, in.Password); err != nil {
		return dto.AuthOutputDTO{}, errs.ErrInvalidCredentials
	}

	token, err := c.issuer.IssueAdmin(sharedjwt.AdminClaims{UserID: user.ID}, tokenTTL)
	if err != nil {
		return dto.AuthOutputDTO{}, err
	}

	return dto.AuthOutputDTO{Token: token, TokenType: "Bearer"}, nil
}