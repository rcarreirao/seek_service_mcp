// Package application aggregates all commands and queries of the module.
package application

import "github.com/rcarreirao/seek_service_mcp/internal/modules/account/application/command"

// Application exposes every use case of the account module.
type Application struct {
	GuestRegister *command.GuestRegisterCommand
	OwnerRegister *command.OwnerRegisterCommand
	LoginAccount  *command.LoginAccountCommand
	AdminLogin    *command.AdminLoginCommand
}

// NewApplication wires the aggregate.
func NewApplication(
	guestRegister *command.GuestRegisterCommand,
	ownerRegister *command.OwnerRegisterCommand,
	loginAccount *command.LoginAccountCommand,
	adminLogin *command.AdminLoginCommand,
) *Application {
	return &Application{
		GuestRegister: guestRegister,
		OwnerRegister: ownerRegister,
		LoginAccount:  loginAccount,
		AdminLogin:    adminLogin,
	}
}