package domain

import "time"

// User represents a user in the system
// This is the shared domain model that can be used across all packages
type User struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	AccountStatus string     `json:"account_status"`
	ThemeMode     string     `json:"theme_mode"`   // light, dark, system
	ThemeRadius   string     `json:"theme_radius"` // 0, 6, 12, 32
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

// Theme mode constants
const (
	ThemeModeLight  = "light"
	ThemeModeDark   = "dark"
	ThemeModeSystem = "system"
)

// Theme radius constants
const (
	ThemeRadiusNone   = "0"
	ThemeRadiusSmall  = "6"
	ThemeRadiusMedium = "12"
	ThemeRadiusLarge  = "32"
)

// Default theme values
const (
	DefaultThemeMode   = ThemeModeSystem
	DefaultThemeRadius = ThemeRadiusNone
)
