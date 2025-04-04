package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/checkauth"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"gorm.io/gorm"
)

// OwnedThingHandler handles HTTP requests for owned things
type OwnedThingHandler struct {
	BaseHandler
	store store.Store
}

// NewOwnedThingHandler creates a new owned thing handler
func NewOwnedThingHandler(store store.Store) *OwnedThingHandler {
	return &OwnedThingHandler{store: store}
}

// GetOwnedThingsByOwner retrieves all owned things for a given owner
func (h *OwnedThingHandler) GetOwnedThingsByOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := GetIDFromContext(r, "ownerId")

	ownedThings, err := h.store.GetOwnedThingsByOwner(r.Context(), ownerID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, ownedThings)
}

// GetOwnedThing retrieves an owned thing by ID
func (h *OwnedThingHandler) GetOwnedThing(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	ownedThing, err := h.store.GetOwnedThingByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if ownedThing == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	h.respondWithJSON(w, http.StatusOK, ownedThing)
}

// CreateOwnedThing creates a new owned thing
func (h *OwnedThingHandler) CreateOwnedThing(w http.ResponseWriter, r *http.Request) {
	var ownedThing models.OwnedThing
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ownedThing); err != nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	// Basic validation
	if ownedThing.ThingTypeID == "" || ownedThing.OwnerID == "" {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}

	// Get the authenticated user from context
	user := checkauth.GetUserFromContext(r.Context())
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, store.ErrUnauthorized)
		return
	}

	// Verify that the user can only create things for themselves
	if user.UserID != ownedThing.OwnerID {
		h.respondWithError(w, http.StatusForbidden, store.ErrForbidden)
		return
	}

	// Verify that the thing type exists
	thingType, err := h.store.GetThingTypeByID(r.Context(), ownedThing.ThingTypeID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if thingType == nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}

	// Verify that the owner exists by directly querying the database
	db := postgres_store.GetDBFromContext(r.Context())
	if db == nil {
		h.respondWithError(w, http.StatusInternalServerError, errors.New("database connection not available"))
		return
	}

	var owner models.User
	err = db.WithContext(r.Context()).Where("user_id = ?", ownedThing.OwnerID).First(&owner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	// Check for duplicate name and thing_type for this owner
	var existingCount int64
	
	err = db.WithContext(r.Context()).Model(&models.OwnedThing{}).
		Where("owner_id = ? AND thing_type_id = ? AND name = ?", ownedThing.OwnerID, ownedThing.ThingTypeID, ownedThing.Name).
		Count(&existingCount).Error
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if existingCount > 0 {
		// Define a specific error for duplicates
		duplicateErr := store.ErrAlreadyExists
		h.respondWithError(w, http.StatusBadRequest, duplicateErr)
		return
	}

	if err := h.store.CreateOwnedThing(r.Context(), &ownedThing); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, ownedThing)
}

// UpdateOwnedThing updates an existing owned thing
func (h *OwnedThingHandler) UpdateOwnedThing(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	// Get existing owned thing
	existingOwnedThing, err := h.store.GetOwnedThingByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if existingOwnedThing == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	// Get the authenticated user from context
	user := checkauth.GetUserFromContext(r.Context())
	
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, store.ErrUnauthorized)
		return
	}

	// Verify that the requester is the owner of the thing
	if user.UserID != existingOwnedThing.OwnerID {
		h.respondWithError(w, http.StatusForbidden, store.ErrForbidden)
		return
	}

	// Parse the updated owned thing from the request body
	var updatedOwnedThing models.OwnedThing
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedOwnedThing); err != nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	// Check if we're trying to update the name
	if updatedOwnedThing.Name != "" && updatedOwnedThing.Name != existingOwnedThing.Name {
		// Check for duplicate name and thing_type for this owner
		db := postgres_store.GetDBFromContext(r.Context())
		if db == nil {
			h.respondWithError(w, http.StatusInternalServerError, errors.New("database connection not available"))
			return
		}

			
		var existingCount int64
		err = db.WithContext(r.Context()).Model(&models.OwnedThing{}).
			Where("owner_id = ? AND thing_type_id = ? AND name = ? AND owned_thing_id != ?", 
				existingOwnedThing.OwnerID, existingOwnedThing.ThingTypeID, updatedOwnedThing.Name, existingOwnedThing.OwnedThingID).
			Count(&existingCount).Error
		if err != nil {
			h.respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if existingCount > 0 {
			// Use the standard error for duplicates
			h.respondWithError(w, http.StatusBadRequest, store.ErrAlreadyExists)
			return
		}

		// Update the name
		existingOwnedThing.Name = updatedOwnedThing.Name
	}

	// Update the owned thing in the database
	if err := h.store.UpdateOwnedThing(r.Context(), existingOwnedThing); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, existingOwnedThing)
}

// DeleteOwnedThing deletes an owned thing
func (h *OwnedThingHandler) DeleteOwnedThing(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	// Check if owned thing exists
	ownedThing, err := h.store.GetOwnedThingByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if ownedThing == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	// Get the authenticated user from context
	user := checkauth.GetUserFromContext(r.Context())
	if user == nil {
		h.respondWithError(w, http.StatusUnauthorized, store.ErrUnauthorized)
		return
	}

	// Verify that the requester is either the owner or has admin role
	isOwner := user.UserID == ownedThing.OwnerID
	isAdmin := false
	for _, role := range user.Roles {
		if role == string(models.UserRoleAdmin) {
			isAdmin = true
			break
		}
	}

	if !isOwner && !isAdmin {
		h.respondWithError(w, http.StatusForbidden, store.ErrForbidden)
		return
	}

	// Delete the owned thing
	if err := h.store.DeleteOwnedThing(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent)
}