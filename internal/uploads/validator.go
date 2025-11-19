package uploads

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/h2non/filetype"
)

// Validator handles file validation operations
type Validator struct {
	maxSize      int64
	allowedTypes map[string]bool
}

// NewValidator creates a new file validator
func NewValidator(maxSizeMB int64, allowedTypes []string) *Validator {
	// Convert MB to bytes
	maxSizeBytes := maxSizeMB * 1024 * 1024

	// Create map for O(1) lookup
	allowedMap := make(map[string]bool)
	for _, t := range allowedTypes {
		allowedMap[t] = true
	}

	return &Validator{
		maxSize:      maxSizeBytes,
		allowedTypes: allowedMap,
	}
}

// ValidateFile checks if a file meets size and type requirements
// Uses magic bytes for MIME type detection (not just file extension)
func (v *Validator) ValidateFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > v.maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed %d bytes", header.Size, v.maxSize)
	}

	// Read first 512 bytes for magic byte detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Detect MIME type using magic bytes
	kind, err := filetype.Match(buffer[:n])
	if err != nil {
		return fmt.Errorf("failed to detect file type: %w", err)
	}

	if kind == filetype.Unknown {
		return fmt.Errorf("unknown file type")
	}

	// Check if MIME type is allowed
	mimeType := kind.MIME.Value
	if !v.allowedTypes[mimeType] {
		return fmt.Errorf("file type %s is not allowed", mimeType)
	}

	return nil
}
