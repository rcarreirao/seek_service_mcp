// Package valueobject holds typed domain value objects.
package valueobject

import (
	"errors"
	"regexp"
)

// ErrInvalidEmail is returned by NewEmail for malformed addresses.
var ErrInvalidEmail = errors.New("valueobject: invalid email address")

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is a validated email address value object.
type Email struct {
	value string
}

// NewEmail validates s and builds an Email.
func NewEmail(s string) (Email, error) {
	if !emailPattern.MatchString(s) {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: s}, nil
}

// String returns the underlying address.
func (e Email) String() string { return e.value }