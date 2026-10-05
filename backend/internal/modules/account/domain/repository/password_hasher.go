package repository

// PasswordHasher hashes and verifies passwords (bcrypt in infrastructure).
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}