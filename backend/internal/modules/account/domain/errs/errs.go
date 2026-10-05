// Package errs holds typed errors for the account module.
package errs

import "errors"

var (
	// ErrAccountNotFound is returned when no (non-deleted) account matches.
	ErrAccountNotFound = errors.New("account not found")
	// ErrAccountAlreadyExists is returned when the email is taken.
	ErrAccountAlreadyExists = errors.New("account already exists")
	// ErrProfileNotFound is returned when an account has no guest/owner profile.
	ErrProfileNotFound = errors.New("account profile not found")
	// ErrUserNotFound is returned when no (non-deleted) admin user matches.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when the admin email is taken.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned on failed authentication.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrNameRequired is returned when a name is empty.
	ErrNameRequired = errors.New("name is required")
	// ErrPasswordTooShort is returned when a password is under 8 characters.
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrInvalidEmail is returned when the email fails domain validation.
	ErrInvalidEmail = errors.New("invalid email address")
)