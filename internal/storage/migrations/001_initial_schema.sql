-- Initial schema for Circles.DIY
-- Based on ERD.md

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table (authentication level)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    account_status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_account_status ON users(account_status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- External identities (DIDs, federated identities)
CREATE TABLE external_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_type VARCHAR(50) NOT NULL,
    owner_id UUID NOT NULL,
    did VARCHAR(500) UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    provider VARCHAR(100) NOT NULL,
    proof TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_external_identities_owner ON external_identities(owner_type, owner_id);
CREATE INDEX idx_external_identities_did ON external_identities(did);

-- WebAuthn credentials
CREATE TABLE webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id VARCHAR(500) UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    credential_type VARCHAR(50) NOT NULL,
    sign_count INTEGER NOT NULL DEFAULT 0,
    transports TEXT,
    aaguid VARCHAR(100),
    attestation_format VARCHAR(50),
    attestation_object BYTEA,
    friendly_name VARCHAR(255),
    is_synced BOOLEAN NOT NULL DEFAULT FALSE,
    is_backup BOOLEAN NOT NULL DEFAULT FALSE,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);

CREATE INDEX idx_webauthn_credentials_user_id ON webauthn_credentials(user_id);
CREATE INDEX idx_webauthn_credentials_credential_id ON webauthn_credentials(credential_id);

-- Devices
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(500) UNIQUE NOT NULL,
    friendly_name VARCHAR(255),
    device_type VARCHAR(50),
    os VARCHAR(100),
    browser VARCHAR(100),
    is_trusted BOOLEAN NOT NULL DEFAULT FALSE,
    last_verified_at TIMESTAMP,
    first_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_device_fingerprint ON devices(device_fingerprint);

-- Device credentials (links devices to WebAuthn credentials)
CREATE TABLE device_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    credential_id UUID NOT NULL REFERENCES webauthn_credentials(id) ON DELETE CASCADE,
    linked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    UNIQUE(device_id, credential_id)
);

CREATE INDEX idx_device_credentials_device_id ON device_credentials(device_id);
CREATE INDEX idx_device_credentials_credential_id ON device_credentials(credential_id);

-- Profiles (social identity level)
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    handle VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    display_name VARCHAR(255),
    bio TEXT,
    avatar_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);
CREATE INDEX idx_profiles_handle ON profiles(handle);
CREATE INDEX idx_profiles_deleted_at ON profiles(deleted_at);

-- Sessions
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    active_profile_id UUID REFERENCES profiles(id) ON DELETE SET NULL,
    authenticated_credential_ids TEXT,
    session_token VARCHAR(500) UNIQUE NOT NULL,
    refresh_token VARCHAR(500) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    auth_level INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_device_id ON sessions(device_id);
CREATE INDEX idx_sessions_active_profile_id ON sessions(active_profile_id);
CREATE INDEX idx_sessions_session_token ON sessions(session_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_revoked_at ON sessions(revoked_at);

-- Recovery methods (backup codes, recovery emails)
CREATE TABLE recovery_methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    encrypted_secret TEXT NOT NULL,
    salt VARCHAR(100) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    remaining_uses INTEGER,
    verified_at TIMESTAMP,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);

CREATE INDEX idx_recovery_methods_user_id ON recovery_methods(user_id);

-- Authentication challenges (for WebAuthn and magic links)
CREATE TABLE authentication_challenges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    challenge VARCHAR(500) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    required_credential_id VARCHAR(500),
    options JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP
);

CREATE INDEX idx_authentication_challenges_challenge ON authentication_challenges(challenge);
CREATE INDEX idx_authentication_challenges_user_id ON authentication_challenges(user_id);
CREATE INDEX idx_authentication_challenges_expires_at ON authentication_challenges(expires_at);

-- Device verification
CREATE TABLE device_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL,
    verification_code VARCHAR(100) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_device_verifications_device_id ON device_verifications(device_id);

-- Profile settings
CREATE TABLE profile_settings (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    location VARCHAR(255),
    website TEXT,
    interests JSONB,
    social_links JSONB,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- User moderation
CREATE TABLE user_moderation (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    trust_score DECIMAL(5,2) NOT NULL DEFAULT 100.0,
    warning_count INTEGER NOT NULL DEFAULT 0,
    violation_count INTEGER NOT NULL DEFAULT 0,
    banned_until TIMESTAMP,
    notes TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Trigger to update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_external_identities_updated_at BEFORE UPDATE ON external_identities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
