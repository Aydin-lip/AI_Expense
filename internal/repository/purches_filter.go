package repository

import "time"

type PurchaseFilter struct {
	UserIDs   []uint
	Usernames []string
	Category  []string
	MinAmount *int
	MaxAmount *int
	DateFrom  *time.Time
	DateTo    *time.Time
}
