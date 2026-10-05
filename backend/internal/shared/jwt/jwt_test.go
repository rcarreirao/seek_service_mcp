package jwt

import (
	"testing"
	"time"
)

func TestIssueAndParseRoundTrip(t *testing.T) {
	iss := NewTokenIssuer("admin-secret", "panel-secret")

	adminTok, err := iss.IssueAdmin(AdminClaims{UserID: 7}, time.Hour)
	if err != nil {
		t.Fatalf("IssueAdmin: %v", err)
	}
	admin, err := iss.ParseAdmin(adminTok)
	if err != nil {
		t.Fatalf("ParseAdmin: %v", err)
	}
	if admin.UserID != 7 {
		t.Fatalf("UserID = %d, want 7", admin.UserID)
	}

	gid := uint(3)
	panelTok, err := iss.IssuePanel(PanelClaims{
		AccountID: 9,
		GuestID:   &gid,
		Email:     "g@example.com",
		Roles:     []string{"guest"},
	}, time.Hour)
	if err != nil {
		t.Fatalf("IssuePanel: %v", err)
	}
	panel, err := iss.ParsePanel(panelTok)
	if err != nil {
		t.Fatalf("ParsePanel: %v", err)
	}
	if panel.AccountID != 9 || panel.GuestID == nil || *panel.GuestID != 3 {
		t.Fatalf("panel claims mismatch: %+v", panel)
	}
	if len(panel.Roles) != 1 || panel.Roles[0] != "guest" {
		t.Fatalf("roles = %v", panel.Roles)
	}
}

func TestWrongRealmRejected(t *testing.T) {
	iss := NewTokenIssuer("admin-secret", "panel-secret")

	panelTok, err := iss.IssuePanel(PanelClaims{AccountID: 1, Email: "x@y.zz", Roles: []string{"owner"}}, time.Hour)
	if err != nil {
		t.Fatalf("IssuePanel: %v", err)
	}
	if _, err := iss.ParseAdmin(panelTok); err == nil {
		t.Fatal("ParseAdmin accepted a panel token")
	}

	adminTok, err := iss.IssueAdmin(AdminClaims{UserID: 1}, time.Hour)
	if err != nil {
		t.Fatalf("IssueAdmin: %v", err)
	}
	if _, err := iss.ParsePanel(adminTok); err == nil {
		t.Fatal("ParsePanel accepted an admin token")
	}
}

func TestWrongSecretRejected(t *testing.T) {
	a := NewTokenIssuer("secret-a", "panel-a")
	b := NewTokenIssuer("secret-b", "panel-b")

	tok, err := a.IssueAdmin(AdminClaims{UserID: 1}, time.Hour)
	if err != nil {
		t.Fatalf("IssueAdmin: %v", err)
	}
	if _, err := b.ParseAdmin(tok); err == nil {
		t.Fatal("ParseAdmin accepted a token signed with another secret")
	}
}

func TestExpiredRejected(t *testing.T) {
	iss := NewTokenIssuer("admin-secret", "panel-secret")

	tok, err := iss.IssuePanel(PanelClaims{AccountID: 1, Email: "x@y.zz"}, -time.Minute)
	if err != nil {
		t.Fatalf("IssuePanel: %v", err)
	}
	if _, err := iss.ParsePanel(tok); err == nil {
		t.Fatal("ParsePanel accepted an expired token")
	}
}

func TestGarbageRejected(t *testing.T) {
	iss := NewTokenIssuer("admin-secret", "panel-secret")
	if _, err := iss.ParsePanel("not-a-token"); err == nil {
		t.Fatal("ParsePanel accepted garbage")
	}
}