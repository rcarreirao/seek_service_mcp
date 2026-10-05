// Package dto holds the input/output data transfer objects of the account
// module. Execute methods take exactly one input DTO (low-parameter rule).
package dto

// GuestRegisterInputDTO is the input for guest registration.
type GuestRegisterInputDTO struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

// OwnerRegisterInputDTO is the input for owner registration.
type OwnerRegisterInputDTO struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

// LoginInputDTO is the input for panel login.
type LoginInputDTO struct {
	Email    string
	Password string
}

// AdminLoginInputDTO is the input for admin-realm login.
type AdminLoginInputDTO struct {
	Email    string
	Password string
}

// AuthOutputDTO is the shared authentication result (token projection).
type AuthOutputDTO struct {
	Token     string
	TokenType string
}