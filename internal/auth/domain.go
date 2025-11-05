package auth

import (
	"encoding/base64"

	"circles.diy/internal/domain"
	"circles.diy/internal/domain/auth"
	"github.com/fxamacker/webauthn"
)

// User wraps domain.User with authentication-specific functionality
// Implements the webauthn.User interface for fxamacker/webauthn
type User struct {
	*domain.User
}

// NewUser creates an auth.User from a domain.User
func NewUser(u *domain.User) *User {
	return &User{User: u}
}

// ToDomain returns the underlying domain.User
func (u *User) ToDomain() *domain.User {
	return u.User
}

// ToWebAuthnUser converts to fxamacker webauthn.User struct
// This is an adapter method to bridge domain entities with the WebAuthn library
func (u *User) ToWebAuthnUser(credentials []auth.WebAuthnCredential) *webauthn.User {
	credentialIDs := make([][]byte, len(credentials))
	for i, cred := range credentials {
		if credID, err := base64.URLEncoding.DecodeString(cred.CredentialID); err == nil {
			credentialIDs[i] = credID
		}
	}

	return &webauthn.User{
		ID:            []byte(u.ID),
		Name:          u.Username,
		DisplayName:   u.Username,
		Icon:          "", // Optional
		CredentialIDs: credentialIDs,
	}
}

// ToWebAuthnCredential converts domain credential to fxamacker/webauthn.Credential
// This is an adapter method to bridge domain entities with the WebAuthn library
func ToWebAuthnCredential(c *auth.WebAuthnCredential) *webauthn.Credential {
	pubKey, _ := base64.StdEncoding.DecodeString(c.PublicKey)

	// Create credential with the proper fxamacker API
	cred, _, _ := webauthn.ParseCredential(pubKey)
	if cred == nil {
		// Fallback to empty credential
		cred = &webauthn.Credential{}
	}

	return cred
}
