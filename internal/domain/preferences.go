package domain

import (
	"encoding/json"
	"time"
)

// UserPreferences represents a user's base preferences (accessibility, comfort)
// These preferences apply across all of the user's profiles
type UserPreferences struct {
	UserID    string          `json:"user_id"`
	BaseTheme json.RawMessage `json:"base_theme"` // JSONB - base accessibility/comfort settings
	UpdatedAt time.Time       `json:"updated_at"`
}

// ProfilePreferences represents profile-specific preference overrides (identity, branding)
// These override the user's base preferences for a specific profile
// Null ThemeOverrides means inherit user's base theme
type ProfilePreferences struct {
	ProfileID      string          `json:"profile_id"`
	ThemeOverrides json.RawMessage `json:"theme_overrides,omitempty"` // JSONB - optional identity/branding overrides
	UpdatedAt      time.Time       `json:"updated_at"`
}

// BaseThemeSettings represents user-level base theme preferences
// These are accessibility and comfort settings that apply to all profiles
type BaseThemeSettings struct {
	Mode             string `json:"mode"`              // "light", "dark", "system"
	Radius           string `json:"radius"`            // "0", "6", "12", "32" (border radius in px)
	FontSize         string `json:"font_size,omitempty"`         // "small", "medium", "large" (future)
	Contrast         string `json:"contrast,omitempty"`          // "normal", "high" (future)
	MotionPreference string `json:"motion_preference,omitempty"` // "full", "reduced" (future)
	Density          string `json:"density,omitempty"`           // "comfortable", "compact", "spacious" (future)
}

// ProfileThemeOverrides represents profile-specific theme customizations
// These are identity and branding settings that override base theme for a specific profile
type ProfileThemeOverrides struct {
	AccentColor  string            `json:"accent_color,omitempty"`  // Custom brand color (hex)
	FontFamily   string            `json:"font_family,omitempty"`   // Custom typography (future)
	CustomColors map[string]string `json:"custom_colors,omitempty"` // Advanced color overrides (future)
	// Infinitely extensible via JSONB
}

// EffectiveTheme represents the merged result of base theme + profile overrides
// This is what gets applied to the UI for a specific profile context
type EffectiveTheme struct {
	BaseThemeSettings
	ProfileThemeOverrides
}

// Theme mode constants
const (
	ThemeModeLight  = "light"
	ThemeModeDark   = "dark"
	ThemeModeSystem = "system"
)

// Theme radius constants (in pixels)
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

// DefaultBaseTheme returns the default base theme settings
func DefaultBaseTheme() *BaseThemeSettings {
	return &BaseThemeSettings{
		Mode:   DefaultThemeMode,
		Radius: DefaultThemeRadius,
	}
}

// DefaultEffectiveTheme returns the default effective theme (no overrides)
func DefaultEffectiveTheme() *EffectiveTheme {
	return &EffectiveTheme{
		BaseThemeSettings: *DefaultBaseTheme(),
	}
}
