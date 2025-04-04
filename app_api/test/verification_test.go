package test

import (
	"testing"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/checkauth"
	"github.com/stretchr/testify/assert"
)

// TestGenerateVerifier tests the verifier generation function
func TestGenerateVerifier(t *testing.T) {
	// Create sample inputs
	userID := "test-user-id"
	salt := []byte("test-salt")
	token := "test-token"

	// Generate verifier
	verifier1 := checkauth.GenerateVerifier(userID, salt, token)
	verifier2 := checkauth.GenerateVerifier(userID, salt, token)

	// Verify same inputs produce same hash
	assert.Equal(t, verifier1, verifier2, "Same inputs should produce the same verifier")
	assert.Len(t, verifier1, 64, "SHA256 hex should be 64 characters")

	// Verify different inputs produce different hashes
	differentVerifier := checkauth.GenerateVerifier(userID, []byte("different-salt"), token)
	assert.NotEqual(t, verifier1, differentVerifier, "Different inputs should produce different verifiers")

	// Verify hex characters are uppercase
	for _, c := range verifier1 {
		isDigit := c >= '0' && c <= '9'
		isUpperHex := c >= 'A' && c <= 'F'
		assert.True(t, isDigit || isUpperHex, "Verifier should only contain uppercase hex characters")
	}
}

// TestVerifySession tests the session verification function
func TestVerifySession(t *testing.T) {
	// Create sample inputs
	userID := "test-user-id"
	salt := []byte("test-salt")
	token := "test-token"

	// Generate valid verifier
	validVerifier := checkauth.GenerateVerifier(userID, salt, token)

	// Test valid verification
	assert.True(t, checkauth.VerifySession(token, validVerifier, userID, salt), "Valid verifier should be verified")

	// Test invalid verification with incorrect verifier
	assert.False(t, checkauth.VerifySession(token, "invalid-verifier", userID, salt), "Invalid verifier should not be verified")
	
	// Test invalid verification with incorrect token
	assert.False(t, checkauth.VerifySession("wrong-token", validVerifier, userID, salt), "Valid verifier with wrong token should not be verified")
}