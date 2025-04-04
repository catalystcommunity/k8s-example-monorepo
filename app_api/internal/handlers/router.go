package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/checkauth"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/middleware"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"

	"github.com/rs/cors"
)

var (
	// Singleton instance of the app's ServeMux
	appMux *http.ServeMux
)

// GetAppMux returns the application's HTTP ServeMux for both API and tests
// This ensures all tests use the same router configuration as the actual application
func GetAppMux() *http.ServeMux {
	// Only create the mux once
	if appMux == nil {
		appMux = createAppMux()
	}
	return appMux
}

// createAppMux creates and configures the application ServeMux with all routes
func createAppMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Create handlers
	thingTypeHandler := NewThingTypeHandler(store.AppStore)
	ownedThingHandler := NewOwnedThingHandler(store.AppStore)
	borrowHandler := NewBorrowHandler(store.AppStore)

	// Apply middleware to all handlers
	transactionMiddleware := middleware.TransactionMiddleware
	verificationMiddleware := middleware.VerificationMiddleware

	// Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		transactionMiddleware(http.HandlerFunc(healthHandler)).ServeHTTP(w, r)
	})

	// ThingType routes
	mux.HandleFunc("/api/thing-types", func(w http.ResponseWriter, r *http.Request) {
		handler := transactionMiddleware(verificationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				thingTypeHandler.GetAllThingTypes(w, r)
			case http.MethodPost:
				thingTypeHandler.CreateThingType(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		handler.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/thing-types/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the ID from the path
		path := strings.TrimPrefix(r.URL.Path, "/api/thing-types/")
		if path == "" {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		// Add the ID to the request context
		r = r.WithContext(setIDContext(r.Context(), "id", path))
		
		handler := transactionMiddleware(verificationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				thingTypeHandler.GetThingType(w, r)
			case http.MethodPut:
				thingTypeHandler.UpdateThingType(w, r)
			case http.MethodDelete:
				thingTypeHandler.DeleteThingType(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		handler.ServeHTTP(w, r)
	})

	// OwnedThing routes
	mux.HandleFunc("/api/owned-things", func(w http.ResponseWriter, r *http.Request) {
		handler := transactionMiddleware(verificationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				ownedThingHandler.CreateOwnedThing(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		handler.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/owned-things/", func(w http.ResponseWriter, r *http.Request) {
		// Parse the path
		path := strings.TrimPrefix(r.URL.Path, "/api/owned-things/")
		if path == "" {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		handler := transactionMiddleware(verificationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle the special case for owner/{ownerId}
			if strings.HasPrefix(path, "owner/") {
				ownerID := strings.TrimPrefix(path, "owner/")
				r = r.WithContext(setIDContext(r.Context(), "ownerId", ownerID))
				if r.Method == http.MethodGet {
					ownedThingHandler.GetOwnedThingsByOwner(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			// Regular ID-based routes
			r = r.WithContext(setIDContext(r.Context(), "id", path))
			switch r.Method {
			case http.MethodGet:
				ownedThingHandler.GetOwnedThing(w, r)
			case http.MethodPut:
				ownedThingHandler.UpdateOwnedThing(w, r)
			case http.MethodDelete:
				ownedThingHandler.DeleteOwnedThing(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		handler.ServeHTTP(w, r)
	})

	// Borrow routes
	mux.HandleFunc("/api/borrows", func(w http.ResponseWriter, r *http.Request) {
		handler := transactionMiddleware(verificationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				borrowHandler.CreateBorrow(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		handler.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/borrows/", func(w http.ResponseWriter, r *http.Request) {
		// Parse the path
		path := strings.TrimPrefix(r.URL.Path, "/api/borrows/")
		if path == "" {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		
		handler := transactionMiddleware(verificationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle special cases
			if strings.HasPrefix(path, "borrower/") {
				borrowerID := strings.TrimPrefix(path, "borrower/")
				r = r.WithContext(setIDContext(r.Context(), "borrowerId", borrowerID))
				if r.Method == http.MethodGet {
					borrowHandler.GetBorrowsByBorrower(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			if strings.HasPrefix(path, "owned-thing/") {
				ownedThingID := strings.TrimPrefix(path, "owned-thing/")
				r = r.WithContext(setIDContext(r.Context(), "ownedThingId", ownedThingID))
				if r.Method == http.MethodGet {
					borrowHandler.GetBorrowsByOwnedThing(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			// Regular ID-based routes
			r = r.WithContext(setIDContext(r.Context(), "id", path))
			switch r.Method {
			case http.MethodGet:
				borrowHandler.GetBorrow(w, r)
			case http.MethodPut:
				borrowHandler.UpdateBorrow(w, r)
			case http.MethodDelete:
				borrowHandler.DeleteBorrow(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		handler.ServeHTTP(w, r)
	})

	return mux
}

// setIDContext adds an ID to the context for handlers to use
// This replaces the mux.Vars functionality from gorilla/mux
type contextKey string
func setIDContext(ctx context.Context, key, value string) context.Context {
	return context.WithValue(ctx, contextKey(key), value)
}

// GetIDFromContext gets an ID from the context
func GetIDFromContext(r *http.Request, key string) string {
	if value, ok := r.Context().Value(contextKey(key)).(string); ok {
		return value
	}
	return ""
}

// GetContextKey returns a context key of the same type used internally
func GetContextKey(key string) contextKey {
	return contextKey(key)
}

// NewRouter creates a new router for the API with CORS handling
// This is used by the API server
func NewRouter() http.Handler {
	mux := GetAppMux()

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return c.Handler(mux)
}

// Add a health endpoint that includes verification info
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Get verification status from context
	verified := checkauth.GetVerifiedFromContext(r.Context())
	user := checkauth.GetUserFromContext(r.Context())
	
	response := map[string]interface{}{
		"status": "OK",
		"verification": map[string]interface{}{
			"verified": verified,
			"user_authenticated": user != nil,
		},
	}
	
	// Include user info if available
	if user != nil {
		response["verification"].(map[string]interface{})["user_id"] = user.UserID
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}