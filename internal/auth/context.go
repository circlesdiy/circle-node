package auth

import (
	"context"

	"circles.diy/internal/domain"
)

type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "session"
)

// WithUser adds a user to the request context
// Accepts *domain.User to avoid exposing auth-specific types
func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUser retrieves the user from the request context
// Returns *domain.User to avoid circular dependencies
func GetUser(ctx context.Context) *domain.User {
	if user, ok := ctx.Value(userContextKey).(*domain.User); ok {
		return user
	}
	return nil
}

// WithSession adds a session to the request context
func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

// GetSession retrieves the session from the request context
func GetSession(ctx context.Context) *Session {
	if session, ok := ctx.Value(sessionContextKey).(*Session); ok {
		return session
	}
	return nil
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return GetUser(ctx) != nil && GetSession(ctx) != nil
}

// HasAuthLevel checks if the current session has the required auth level
func HasAuthLevel(ctx context.Context, requiredLevel int) bool {
	session := GetSession(ctx)
	return session != nil && session.AuthLevel >= requiredLevel
}