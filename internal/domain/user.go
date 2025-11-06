package domain

import "time"

// User represents a user in the system
// This is the shared domain model that can be used across all packages
type User struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	AccountStatus string     `json:"account_status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// AccountStatus constants
const (
	AccountStatusActive    = "active"
	AccountStatusSuspended = "suspended"
	AccountStatusPending   = "pending"
)

// IsActive returns true if the user account is active
func (u *User) IsActive() bool {
	return u.AccountStatus == AccountStatusActive && u.DeletedAt == nil
}

// IsSuspended returns true if the user account is suspended
func (u *User) IsSuspended() bool {
	return u.AccountStatus == AccountStatusSuspended
}
