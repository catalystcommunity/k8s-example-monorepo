package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"crypto/sha256"
)

// Constants for PBKDF2 hashing
const (
	SaltSize     = 16
	KeySize      = 32
	Iterations   = 400000 // Increased from 10000 to match auth_api
	HashFunction = "sha256"
)

// GenerateSalt creates a random salt for password hashing
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// HashPassword converts a plain text password to PBKDF2 hash with salt
func HashPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, Iterations, KeySize, sha256.New)
}

// VerifyPassword checks if the provided password matches the stored hash
func VerifyPassword(storedHash, salt []byte, password string) bool {
	// Generate hash of the input password with the stored salt
	hash := HashPassword(password, salt)
	
	// Compare the generated hash with the stored hash
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(storedHash, hash) == 1
}

// EncodePasswordHash encodes password hash and salt for storage
func EncodePasswordHash(hash, salt []byte) string {
	encodedHash := base64.StdEncoding.EncodeToString(hash)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	return fmt.Sprintf("$pbkdf2$%s$%d$%s$%s", HashFunction, Iterations, encodedSalt, encodedHash)
}

// DecodePasswordHash decodes stored password hash string into components
func DecodePasswordHash(encoded string) ([]byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "pbkdf2" {
		return nil, nil, fmt.Errorf("invalid hash format")
	}
	
	// Extract salt and hash
	salt, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, err
	}
	
	hash, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, err
	}
	
	return hash, salt, nil
}