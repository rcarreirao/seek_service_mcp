package hashing

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasherRoundTrip(t *testing.T) {
	h := NewBcryptHasher(bcrypt.MinCost)

	hash, err := h.Hash("s3cret-pass")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "s3cret-pass" || hash == "" {
		t.Fatalf("Hash returned plaintext or empty: %q", hash)
	}
	if err := h.Compare(hash, "s3cret-pass"); err != nil {
		t.Fatalf("Compare(valid): %v", err)
	}
	if err := h.Compare(hash, "wrong"); err == nil {
		t.Fatal("Compare(invalid) = nil error")
	}
}