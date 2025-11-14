package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	domainauth "circles.diy/internal/domain/auth"
	"go.uber.org/zap"
)

// GenerateRecoveryCodes creates recovery codes for account recovery
// Returns a slice of human-readable recovery codes
func (s *Service) GenerateRecoveryCodes(ctx context.Context, userID string, count int) ([]string, error) {
	if count <= 0 || count > 20 {
		count = 10 // Default to 10 codes
	}

	codes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 12-byte random code
		randomBytes := make([]byte, 12)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Encode as base64 URL-safe and format for readability
		encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
		code := formatRecoveryCode(encoded)
		codes[i] = code

		// Hash for storage (never store plaintext)
		hash := sha256.Sum256([]byte(code))
		hashedCode := base64.StdEncoding.EncodeToString(hash[:])

		// Store in database
		method := &domainauth.RecoveryMethod{
			ID:              GenerateID(),
			UserID:          userID,
			Type:            "recovery_code",
			EncryptedSecret: hashedCode,
			Salt:            "", // Not needed for SHA-256 hash
			IsVerified:      true,
			RemainingUses:   ptrInt(1), // Single-use
			CreatedAt:       time.Now().UTC(),
		}

		if err := s.repo.CreateRecoveryMethod(ctx, method); err != nil {
			return nil, fmt.Errorf("failed to store recovery code: %w", err)
		}
	}

	s.logger.Info("Generated recovery codes", zap.String("user_id", userID), zap.Int("count", count))
	return codes, nil
}

// VerifyRecoveryCode checks if a recovery code is valid and consumes it if valid
// Returns true if the code is valid and was successfully consumed
func (s *Service) VerifyRecoveryCode(ctx context.Context, userID, code string) (bool, error) {
	// Normalize code (remove spaces, uppercase)
	code = strings.ToUpper(strings.ReplaceAll(code, " ", ""))
	code = strings.ReplaceAll(code, "-", "")

	// Hash the provided code
	hash := sha256.Sum256([]byte(code))
	hashedCode := base64.StdEncoding.EncodeToString(hash[:])

	// Get all recovery codes for user
	methods, err := s.repo.GetRecoveryMethodsByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get recovery methods: %w", err)
	}

	// Find matching code
	for _, method := range methods {
		if method.Type == "recovery_code" &&
			method.EncryptedSecret == hashedCode &&
			method.RemainingUses != nil &&
			*method.RemainingUses > 0 {

			// Consume the code (set remaining uses to 0)
			if err := s.repo.UseRecoveryMethod(ctx, method.ID); err != nil {
				return false, fmt.Errorf("failed to consume recovery code: %w", err)
			}

			s.logger.Info("Recovery code used successfully",
				zap.String("user_id", userID),
				zap.String("method_id", method.ID))
			return true, nil
		}
	}

	s.logger.Warn("Invalid recovery code attempt",
		zap.String("user_id", userID))
	return false, nil
}

// RevokeRecoveryCodes revokes all unused recovery codes for a user
// Useful when regenerating codes or as a security measure
func (s *Service) RevokeRecoveryCodes(ctx context.Context, userID string) error {
	methods, err := s.repo.GetRecoveryMethodsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get recovery methods: %w", err)
	}

	revokedCount := 0
	for _, method := range methods {
		if method.Type == "recovery_code" &&
			method.RemainingUses != nil &&
			*method.RemainingUses > 0 {

			if err := s.repo.UseRecoveryMethod(ctx, method.ID); err != nil {
				s.logger.Warn("Failed to revoke recovery code",
					zap.String("method_id", method.ID),
					zap.Error(err))
				continue
			}
			revokedCount++
		}
	}

	s.logger.Info("Revoked recovery codes",
		zap.String("user_id", userID),
		zap.Int("count", revokedCount))
	return nil
}

// GetRecoveryCodeStatus returns information about a user's recovery codes
type RecoveryCodeStatus struct {
	TotalCodes     int
	UnusedCodes    int
	UsedCodes      int
	LastUsedAt     *time.Time
	LastGeneratedAt *time.Time
}

// GetRecoveryCodeStatus retrieves the status of a user's recovery codes
func (s *Service) GetRecoveryCodeStatus(ctx context.Context, userID string) (*RecoveryCodeStatus, error) {
	methods, err := s.repo.GetRecoveryMethodsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recovery methods: %w", err)
	}

	status := &RecoveryCodeStatus{
		TotalCodes:  0,
		UnusedCodes: 0,
		UsedCodes:   0,
	}

	for _, method := range methods {
		if method.Type == "recovery_code" {
			status.TotalCodes++

			if method.RemainingUses != nil && *method.RemainingUses > 0 {
				status.UnusedCodes++
			} else {
				status.UsedCodes++
				if method.LastUsedAt != nil {
					if status.LastUsedAt == nil || method.LastUsedAt.After(*status.LastUsedAt) {
						status.LastUsedAt = method.LastUsedAt
					}
				}
			}

			if status.LastGeneratedAt == nil || method.CreatedAt.After(*status.LastGeneratedAt) {
				status.LastGeneratedAt = &method.CreatedAt
			}
		}
	}

	return status, nil
}

// formatRecoveryCode formats a code into groups of 4 characters for readability
// Example: "ABCDEFGHIJKL" becomes "ABCD-EFGH-IJKL"
func formatRecoveryCode(code string) string {
	// Take first 12 characters, uppercase
	if len(code) > 12 {
		code = code[:12]
	}
	code = strings.ToUpper(code)

	// Split into groups of 4
	var parts []string
	for i := 0; i < len(code); i += 4 {
		end := i + 4
		if end > len(code) {
			end = len(code)
		}
		parts = append(parts, code[i:end])
	}

	return strings.Join(parts, "-")
}

// Helper function to create a pointer to an int
func ptrInt(i int) *int {
	return &i
}
