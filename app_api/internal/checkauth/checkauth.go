package checkauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
)

const UserEmailKey = "email"
const UserIDKey = "userId"
const VerifiedKey = "verified"
const UserKey = "user"

var (
	ErrNoToken            = errors.New("no token provided")
	ErrNoVerifier         = errors.New("no verifier provided")
	ErrVerificationFailed = errors.New("session verification failed")
	ErrSessionNotFound    = errors.New("session not found")
	ErrUserNotFound       = errors.New("user not found")
)

// GenerateVerifier creates a SHA256 hash of user_id + salt + token
func GenerateVerifier(userID string, salt []byte, token string) string {
	// Convert salt to uppercase hex string
	saltHex := hex.EncodeToString(salt)
	saltHex = fmt.Sprintf("%s", saltHex)
	
	// Concatenate values in specified order
	combined := fmt.Sprintf("%s%s%s", userID, saltHex, token)
	
	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(combined))
	
	// Return uppercase hex string
	return fmt.Sprintf("%X", hash[:])
}

// VerifySession checks if a token and verifier match for a given user
func VerifySession(token string, verifier string, userID string, salt []byte) bool {
	expectedVerifier := GenerateVerifier(userID, salt, token)
	return verifier == expectedVerifier
}

// ExtractTokenAndVerifier extracts token and verifier from an HTTP request
// For GET requests, it checks query parameters
// For other methods, it checks the request body
func ExtractTokenAndVerifier(r *http.Request) (string, string, error) {
	var token, verifier string
	
	if r.Method == http.MethodGet {
		// For GET requests, check query parameters
		token = r.URL.Query().Get("token")
		verifier = r.URL.Query().Get("verifier")
	} else {
		// For other methods, check body
		if r.Body == nil {
			return "", "", nil
		}
		
		// Read body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return "", "", err
		}
		
		// Restore the body for subsequent reads
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		
		// Skip if body is empty
		if len(bodyBytes) == 0 {
			return "", "", nil
		}
		
		// Try to parse as JSON
		var bodyJSON map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyJSON); err != nil {
			return "", "", nil // Not JSON or invalid JSON
		}
		
		// Extract token and verifier
		if tokenVal, ok := bodyJSON["token"].(string); ok {
			token = tokenVal
		}
		if verifierVal, ok := bodyJSON["verifier"].(string); ok {
			verifier = verifierVal
		}
	}
	
	return token, verifier, nil
}

// AuthenticateRequest verifies authentication against token and verifier
func AuthenticateRequest(ctx context.Context) (context.Context, error) {
	// Get the HTTP request from context
	r, ok := ctx.Value(store.RequestContextKey).(*http.Request)
	if !ok {
		return ctx, errors.New("no request in context")
	}
	
	// Extract token and verifier
	token, verifier, err := ExtractTokenAndVerifier(r)
	if err != nil {
		return ctx, err
	}
	
	
	// If no token provided, continue with unverified context
	if token == "" {
		ctx = context.WithValue(ctx, VerifiedKey, false)
		return ctx, nil
	}
	
	// If token is provided but no verifier, return error
	if token != "" && verifier == "" {
		return ctx, ErrNoVerifier
	}
	
	// Get database from context (either the transaction or the global DB)
	db := postgres_store.GetDBFromContext(ctx)
	if db == nil {
		return ctx, errors.New("database not available")
	}
	
	
	// Look up session directly from the database
	var session models.Session
	err = db.WithContext(ctx).Where("token = ?", token).First(&session).Error
	if err != nil {
		// If session not found, continue with unverified context
		ctx = context.WithValue(ctx, VerifiedKey, false)
		return ctx, nil
	}
	
	
	// Look up user directly from the database
	var user models.User
	err = db.WithContext(ctx).Where("user_id = ?", session.UserID).First(&user).Error
	if err != nil {
		// If user not found, continue with unverified context
		ctx = context.WithValue(ctx, VerifiedKey, false)
		return ctx, nil
	}
	
	// Verify the session
	isVerified := VerifySession(token, verifier, user.UserID, user.Salt)
	if !isVerified {
		return ctx, ErrVerificationFailed
	}
	
	// Set verified state and user in context if verification passes
	ctx = context.WithValue(ctx, VerifiedKey, true)
	ctx = context.WithValue(ctx, UserKey, &user)
	ctx = context.WithValue(ctx, UserIDKey, user.UserID)
	ctx = context.WithValue(ctx, UserEmailKey, user.Email)
	
	return ctx, nil
}

// GetVerifiedFromContext gets the verified state from context
func GetVerifiedFromContext(ctx context.Context) bool {
	verified, ok := ctx.Value(VerifiedKey).(bool)
	if !ok {
		return false
	}
	return verified
}

// GetUserFromContext gets the user from context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}