package repository

import (
	"time"

	sharedjwt "github.com/rcarreirao/seek_service_mcp/internal/shared/jwt"
)

// TokenIssuer issues realm-typed JWTs. Parsing/middleware lives in the
// adapters; use cases only issue.
type TokenIssuer interface {
	IssueAdmin(claims sharedjwt.AdminClaims, ttl time.Duration) (string, error)
	IssuePanel(claims sharedjwt.PanelClaims, ttl time.Duration) (string, error)
}