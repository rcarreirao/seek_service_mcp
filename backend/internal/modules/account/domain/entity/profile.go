package entity

import "time"

// Guest is the 1:1 role profile of a customer account.
type Guest struct {
	ID        uint
	AccountID uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Owner is the 1:1 role profile of a shop-owner account.
type Owner struct {
	ID        uint
	AccountID uint
	Verified  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}