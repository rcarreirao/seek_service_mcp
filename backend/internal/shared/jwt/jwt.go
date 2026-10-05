// Package jwt provides typed JWT claims and an issuer for the two
// isolated realms (admin and panel). Secrets come from environment config.
package jwt

import (
	"errors"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned for malformed, expired, or wrong-realm tokens.
var ErrInvalidToken = errors.New("jwt: invalid or wrong-realm token")

const (
	issuerAdmin = "seek-admin"
	issuerPanel = "seek-panel"
)

// AdminClaims is the typed claim set of the admin realm.
type AdminClaims struct {
	UserID uint `json:"user_id"`
	golangjwt.RegisteredClaims
}

// PanelClaims is the typed claim set of the panel realm. Profile ids are
// embedded at login so ownership checks never need extra lookups.
type PanelClaims struct {
	AccountID uint     `json:"account_id"`
	GuestID   *uint    `json:"guest_id,omitempty"`
	OwnerID   *uint    `json:"owner_id,omitempty"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	golangjwt.RegisteredClaims
}

// TokenIssuer issues and parses realm-typed tokens with HS256.
type TokenIssuer struct {
	adminSecret []byte
	panelSecret []byte
}

// NewTokenIssuer builds an issuer for both realms.
func NewTokenIssuer(adminSecret, panelSecret string) *TokenIssuer {
	return &TokenIssuer{
		adminSecret: []byte(adminSecret),
		panelSecret: []byte(panelSecret),
	}
}

// IssueAdmin signs admin claims with the admin secret.
func (t *TokenIssuer) IssueAdmin(c AdminClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	c.Issuer = issuerAdmin
	c.IssuedAt = golangjwt.NewNumericDate(now)
	c.ExpiresAt = golangjwt.NewNumericDate(now.Add(ttl))
	return golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, c).SignedString(t.adminSecret)
}

// IssuePanel signs panel claims with the panel secret.
func (t *TokenIssuer) IssuePanel(c PanelClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	c.Issuer = issuerPanel
	c.IssuedAt = golangjwt.NewNumericDate(now)
	c.ExpiresAt = golangjwt.NewNumericDate(now.Add(ttl))
	return golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, c).SignedString(t.panelSecret)
}

// ParseAdmin validates a token against the admin realm.
func (t *TokenIssuer) ParseAdmin(token string) (AdminClaims, error) {
	var c AdminClaims
	if err := t.parse(token, t.adminSecret, issuerAdmin, &c); err != nil {
		return AdminClaims{}, err
	}
	return c, nil
}

// ParsePanel validates a token against the panel realm.
func (t *TokenIssuer) ParsePanel(token string) (PanelClaims, error) {
	var c PanelClaims
	if err := t.parse(token, t.panelSecret, issuerPanel, &c); err != nil {
		return PanelClaims{}, err
	}
	return c, nil
}

func (t *TokenIssuer) parse(token string, secret []byte, issuer string, out golangjwt.Claims) error {
	parsed, err := golangjwt.ParseWithClaims(token, out, func(tok *golangjwt.Token) (any, error) {
		if _, ok := tok.Method.(*golangjwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	}, golangjwt.WithIssuer(issuer))
	if err != nil || !parsed.Valid {
		return ErrInvalidToken
	}
	return nil
}