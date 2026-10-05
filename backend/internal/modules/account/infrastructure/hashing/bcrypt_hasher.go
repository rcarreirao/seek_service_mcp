// Package hashing provides password hashing implementations.
package hashing

import "golang.org/x/crypto/bcrypt"

// BcryptHasher hashes and verifies passwords with bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher builds a hasher with the given bcrypt cost (0 = default).
func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// Hash returns the bcrypt hash of the plain password.
func (h *BcryptHasher) Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare reports whether plain matches the bcrypt hash.
func (h *BcryptHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}