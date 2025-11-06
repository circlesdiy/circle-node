-- Migration: 004_user_and_profile_preferences
-- Description: Add user_preferences and profile_preferences tables for hybrid theme system
-- Created: 2025-11-06

-- User base preferences table (accessibility, comfort settings)
-- One-to-one with users, contains base theme that applies across all profiles
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    base_theme JSONB NOT NULL DEFAULT '{"mode":"system","radius":"0"}',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for querying theme settings
CREATE INDEX idx_user_preferences_base_theme ON user_preferences USING GIN (base_theme);

-- Trigger to automatically update updated_at timestamp
CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Profile-specific preference overrides table (identity, branding)
-- One-to-one with profiles, contains optional theme overrides
-- NULL theme_overrides means inherit user's base theme
CREATE TABLE profile_preferences (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    theme_overrides JSONB DEFAULT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for querying theme overrides
CREATE INDEX idx_profile_preferences_theme_overrides ON profile_preferences USING GIN (theme_overrides);

-- Trigger to automatically update updated_at timestamp
CREATE TRIGGER update_profile_preferences_updated_at
    BEFORE UPDATE ON profile_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE user_preferences IS 'User-level base preferences (accessibility, comfort) that apply across all profiles';
COMMENT ON TABLE profile_preferences IS 'Profile-specific preference overrides (identity, branding) that override user base preferences';
COMMENT ON COLUMN user_preferences.base_theme IS 'Base theme settings (mode, radius, accessibility) stored as JSONB for extensibility';
COMMENT ON COLUMN profile_preferences.theme_overrides IS 'Profile-specific theme overrides (colors, branding) stored as JSONB. NULL means inherit user base theme';
