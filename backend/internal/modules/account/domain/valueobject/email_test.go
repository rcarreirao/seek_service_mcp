package valueobject

import "testing"

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"valid simple", "guest@example.com", false},
		{"valid subdomain", "a.b+tag@mail.example.co.uk", false},
		{"empty", "", true},
		{"no at sign", "guestexample.com", true},
		{"no tld", "guest@example", true},
		{"spaces", "gues t@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NewEmail(%q) = nil error, want error", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewEmail(%q) error: %v", tt.in, err)
			}
			if email.String() != tt.in {
				t.Fatalf("String() = %q, want %q", email.String(), tt.in)
			}
		})
	}
}