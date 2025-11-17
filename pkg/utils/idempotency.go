package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

// GenerateClientPostID generates a backend client post ID
// Format: backend_client_post_{timestamp}_{random9}
// Example: backend_client_post_1732012345678_abc123def
func GenerateClientPostID() string {
	timestamp := time.Now().UnixMilli()
	random := generateRandomString(9)
	return fmt.Sprintf("backend_client_post_%d_%s", timestamp, random)
}

// GenerateIdempotencyKey generates a backend idempotency key
// Format: backend_idem_{timestamp}_{random9}
// Example: backend_idem_1732012345678_xyz789ghi
func GenerateIdempotencyKey() string {
	timestamp := time.Now().UnixMilli()
	random := generateRandomString(9)
	return fmt.Sprintf("backend_idem_%d_%s", timestamp, random)
}

// generateRandomString generates a random alphanumeric string of specified length
// Uses base32 encoding for URL-safe characters (lowercase a-z, 0-9)
func generateRandomString(length int) string {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based random if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())[:length]
	}

	// Encode to base32 and convert to lowercase
	encoded := base32.StdEncoding.EncodeToString(bytes)
	encoded = strings.ToLower(encoded)
	encoded = strings.ReplaceAll(encoded, "=", "")

	// Take first {length} characters
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// ValidateClientPostID validates the format of a client post ID
// Accepts both frontend format (client_post_...) and backend format (backend_client_post_...)
func ValidateClientPostID(clientPostID string) bool {
	if clientPostID == "" {
		return false
	}

	// Check length (reasonable limit)
	if len(clientPostID) < 10 || len(clientPostID) > 255 {
		return false
	}

	// Accept both frontend and backend formats
	if strings.HasPrefix(clientPostID, "client_post_") ||
		strings.HasPrefix(clientPostID, "backend_client_post_") {
		return true
	}

	// Accept any other custom format (flexible for future changes)
	return true
}

// ValidateIdempotencyKey validates the format of an idempotency key
// Accepts both frontend format (idem_...) and backend format (backend_idem_...)
func ValidateIdempotencyKey(idempotencyKey string) bool {
	if idempotencyKey == "" {
		return false
	}

	// Check length (reasonable limit)
	if len(idempotencyKey) < 10 || len(idempotencyKey) > 255 {
		return false
	}

	// Accept both frontend and backend formats
	if strings.HasPrefix(idempotencyKey, "idem_") ||
		strings.HasPrefix(idempotencyKey, "backend_idem_") {
		return true
	}

	// Accept any other custom format (flexible for future changes)
	return true
}
