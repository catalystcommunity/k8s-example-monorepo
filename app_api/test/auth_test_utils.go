package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/checkauth"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"gorm.io/gorm"
)

// PrepareAuthenticatedRequest creates a fully authenticated HTTP request
func PrepareAuthenticatedRequest(ctx context.Context, tx *gorm.DB, method, path string, body []byte, user *models.User) (*http.Request, error) {
	// Create a DataUtils instance for this transaction
	du := &DataUtils{db: tx}
	
	// Create a session for the user using DB transaction
	session, err := du.CreateSession(DataSetup{
		"UserID": user.UserID,
	})
	if err != nil {
		return nil, err
	}

	// Generate token and verifier
	token := session.Token
	verifier := checkauth.GenerateVerifier(user.UserID, user.Salt, token)

	// Create the request
	var req *http.Request
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, path, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, path, nil)
	}
	if err != nil {
		return nil, err
	}

	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Add auth credentials in a consistent way
	if body != nil && len(body) > 0 {
		// If there's a body, we need to add the auth to it
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			// Add authentication data to the body
			bodyMap["token"] = token
			bodyMap["verifier"] = verifier
			
			// Convert back to JSON
			updatedBody, err := json.Marshal(bodyMap)
			if err == nil {
				// Replace the existing body
				req.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
				req.ContentLength = int64(len(updatedBody))
			}
		}
	} else {
		// If no body, add auth as query parameters
		q := req.URL.Query()
		q.Add("token", token)
		q.Add("verifier", verifier)
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

// ExecuteAuthenticatedRequest executes a fully authenticated HTTP request
func ExecuteAuthenticatedRequest(ctx context.Context, tx *gorm.DB, method, path string, body []byte, user *models.User) *httptest.ResponseRecorder {
	req, err := PrepareAuthenticatedRequest(ctx, tx, method, path, body, user)
	if err != nil {
		panic("Failed to prepare authenticated request: " + err.Error())
	}

	// First, set up the transaction in the context
	txCtx := context.WithValue(req.Context(), postgres_store.GetTxContextKey(), tx)

	// Then, add the request itself to the context - important for auth to work
	reqCtx := context.WithValue(txCtx, store.RequestContextKey, req)

	// Now set up authentication information explicitly
	authCtx := reqCtx
	authCtx = context.WithValue(authCtx, checkauth.VerifiedKey, true)
	authCtx = context.WithValue(authCtx, checkauth.UserKey, user)
	authCtx = context.WithValue(authCtx, checkauth.UserIDKey, user.UserID)
	authCtx = context.WithValue(authCtx, checkauth.UserEmailKey, user.Email)

	req = req.WithContext(authCtx)

	rr := httptest.NewRecorder()
	GetTestMux().ServeHTTP(rr, req)
	return rr
}

// CreateAdminUser creates a user with admin role
func CreateAdminUser(tx *gorm.DB) (*models.User, error) {
	du := &DataUtils{db: tx}
	return du.CreateUser(DataSetup{
		"Roles": []string{string(models.UserRoleAdmin)},
	})
}