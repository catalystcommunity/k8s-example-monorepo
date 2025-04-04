package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/checkauth"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"gorm.io/gorm"
)

// TransactionMiddleware creates middleware that starts a transaction for each request
// and commits it for successful responses or rolls it back for errors
func TransactionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if there's already a transaction in the context (for tests)
		existingTx, existingTxFound := r.Context().Value(postgres_store.GetTxContextKey()).(*gorm.DB)
		
		var tx *gorm.DB
		var shouldManageTx bool // Whether this middleware should manage the transaction
		
		if existingTxFound && existingTx != nil {
			// If there's already a transaction in the context, use it
			// but don't commit/rollback (it will be managed by the test)
			tx = existingTx
			shouldManageTx = false
		} else {
			// No transaction in context, create a new one
			db := store.GetDB()
			if db == nil {
				http.Error(w, "Database connection not available", http.StatusInternalServerError)
				return
			}

			// Begin a transaction
			tx = db.Begin()
			if tx.Error != nil {
				http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
				return
			}
			shouldManageTx = true
		}

		// Create a wrapper for the response writer to track the status code
		tw := &transactionResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200 OK
		}

		// If we didn't find an existing transaction, or we want to ensure it's in the context,
		// add the transaction to the context
		if !existingTxFound {
			ctx := context.WithValue(r.Context(), postgres_store.GetTxContextKey(), tx)
			r = r.WithContext(ctx)
		}

		// Call the next handler
		next.ServeHTTP(tw, r)

		// Only manage the transaction if we created it
		if shouldManageTx {
			// Check the status code to determine if we should commit or rollback
			if config.CommitOnSuccess && tw.statusCode >= 200 && tw.statusCode < 300 {
				// Success and CommitOnSuccess is true - commit the transaction
				if err := tx.Commit().Error; err != nil {
					// If commit fails, rollback and return an error
					tx.Rollback()
					http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
					return
				}
			} else {
				// Either CommitOnSuccess is false or there was an error - rollback the transaction
				tx.Rollback()
			}
		}
		// Otherwise, the transaction is managed by the test framework
	})
}

// VerificationMiddleware creates middleware for token/verifier authentication
func VerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add request to context
		ctx := context.WithValue(r.Context(), store.RequestContextKey, r)
		
		
		// Authenticate the request
		ctx, err := checkauth.AuthenticateRequest(ctx)
		
		
		if err != nil {
			// Handle specific errors
			switch err {
			case checkauth.ErrNoVerifier:
				// If token is provided but no verifier, return 401
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "unauthorized",
					"message": "No verifier given",
				})
				return
			case checkauth.ErrVerificationFailed:
				// If verification fails, return 401
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "unauthorized",
					"message": "Session cannot be verified",
				})
				return
			default:
				// For any other error, return 500
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "server_error",
					"message": "Internal server error during authentication",
				})
				return
			}
		}
		
		// Capture response status code
		rw := &transactionResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Call the next handler with the authenticated context
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

// transactionResponseWriter wraps http.ResponseWriter to track status code
type transactionResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader overrides the http.ResponseWriter.WriteHeader method to track status code
func (w *transactionResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}