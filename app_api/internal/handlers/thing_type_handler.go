package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
)

// ThingTypeHandler handles HTTP requests for thing types
type ThingTypeHandler struct {
	BaseHandler
	store store.Store
}

// NewThingTypeHandler creates a new thing type handler
func NewThingTypeHandler(store store.Store) *ThingTypeHandler {
	return &ThingTypeHandler{store: store}
}

// GetAllThingTypes retrieves all thing types
func (h *ThingTypeHandler) GetAllThingTypes(w http.ResponseWriter, r *http.Request) {
	thingTypes, err := h.store.GetThingTypes(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, thingTypes)
}

// GetThingType retrieves a thing type by ID
func (h *ThingTypeHandler) GetThingType(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	thingType, err := h.store.GetThingTypeByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if thingType == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	h.respondWithJSON(w, http.StatusOK, thingType)
}

// CreateThingType creates a new thing type
func (h *ThingTypeHandler) CreateThingType(w http.ResponseWriter, r *http.Request) {
	var thingType models.ThingType
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&thingType); err != nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	// Basic validation
	if thingType.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}

	if err := h.store.CreateThingType(r.Context(), &thingType); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, thingType)
}

// UpdateThingType updates an existing thing type
func (h *ThingTypeHandler) UpdateThingType(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	// Get existing thing type
	existingThingType, err := h.store.GetThingTypeByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if existingThingType == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	// Parse the updated thing type from the request body
	var updatedThingType models.ThingType
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedThingType); err != nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	// Update fields that are allowed to be updated
	if updatedThingType.Name != "" {
		existingThingType.Name = updatedThingType.Name
	}

	if updatedThingType.Description != "" {
		existingThingType.Description = updatedThingType.Description
	}

	// Update the thing type in the database
	if err := h.store.UpdateThingType(r.Context(), existingThingType); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, existingThingType)
}

// DeleteThingType deletes a thing type
func (h *ThingTypeHandler) DeleteThingType(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	// Check if thing type exists
	thingType, err := h.store.GetThingTypeByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if thingType == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	// Delete the thing type
	if err := h.store.DeleteThingType(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent)
}
