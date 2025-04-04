package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/models"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"gorm.io/gorm"
)

// BorrowHandler handles HTTP requests for borrows
type BorrowHandler struct {
	BaseHandler
	store store.Store
}

// NewBorrowHandler creates a new borrow handler
func NewBorrowHandler(store store.Store) *BorrowHandler {
	return &BorrowHandler{store: store}
}

// GetBorrowsByBorrower retrieves all borrows for a given borrower
func (h *BorrowHandler) GetBorrowsByBorrower(w http.ResponseWriter, r *http.Request) {
	borrowerID := GetIDFromContext(r, "borrowerId")

	borrows, err := h.store.GetBorrowsByBorrower(r.Context(), borrowerID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, borrows)
}

// GetBorrowsByOwnedThing retrieves all borrows for a given owned thing
func (h *BorrowHandler) GetBorrowsByOwnedThing(w http.ResponseWriter, r *http.Request) {
	ownedThingID := GetIDFromContext(r, "ownedThingId")

	borrows, err := h.store.GetBorrowsByOwnedThing(r.Context(), ownedThingID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, borrows)
}

// GetBorrow retrieves a borrow by ID
func (h *BorrowHandler) GetBorrow(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	borrow, err := h.store.GetBorrowByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if borrow == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	h.respondWithJSON(w, http.StatusOK, borrow)
}

// CreateBorrow creates a new borrow
func (h *BorrowHandler) CreateBorrow(w http.ResponseWriter, r *http.Request) {
	var borrow models.Borrow
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&borrow); err != nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	// Basic validation
	if borrow.BorrowerID == "" || borrow.OwnedThingID == "" {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}

	// Verify that the borrower exists by directly querying the database
	db := postgres_store.GetDBFromContext(r.Context())
	if db == nil {
		h.respondWithError(w, http.StatusInternalServerError, errors.New("database connection not available"))
		return
	}

	var borrower models.User
	err := db.WithContext(r.Context()).Where("user_id = ?", borrow.BorrowerID).First(&borrower).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		} else {
			h.respondWithError(w, http.StatusInternalServerError, err)
		}
		return
	}

	// Verify that the owned thing exists
	ownedThing, err := h.store.GetOwnedThingByID(r.Context(), borrow.OwnedThingID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if ownedThing == nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}

	// Set default values
	if borrow.BorrowedAt.IsZero() {
		borrow.BorrowedAt = time.Now().UTC()
	}

	if err := h.store.CreateBorrow(r.Context(), &borrow); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, borrow)
}

// UpdateBorrow updates an existing borrow
func (h *BorrowHandler) UpdateBorrow(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	// Get existing borrow
	existingBorrow, err := h.store.GetBorrowByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if existingBorrow == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	// Parse the updated borrow from the request body
	var updatedBorrow models.Borrow
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedBorrow); err != nil {
		h.respondWithError(w, http.StatusBadRequest, store.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	// Update fields that are allowed to be updated
	if !updatedBorrow.BorrowedUntil.IsZero() && updatedBorrow.BorrowedUntil != nil {
		existingBorrow.BorrowedUntil = updatedBorrow.BorrowedUntil
	}

	if !updatedBorrow.ReturnedAt.IsZero() && updatedBorrow.ReturnedAt != nil {
		existingBorrow.ReturnedAt = updatedBorrow.ReturnedAt
	}

	if updatedBorrow.Reposessed != existingBorrow.Reposessed {
		existingBorrow.Reposessed = updatedBorrow.Reposessed
	}

	// Update the borrow in the database
	if err := h.store.UpdateBorrow(r.Context(), existingBorrow); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, existingBorrow)
}

// DeleteBorrow deletes a borrow
func (h *BorrowHandler) DeleteBorrow(w http.ResponseWriter, r *http.Request) {
	id := GetIDFromContext(r, "id")

	// Check if borrow exists
	borrow, err := h.store.GetBorrowByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if borrow == nil {
		h.respondWithError(w, http.StatusNotFound, store.ErrNotFound)
		return
	}

	// Delete the borrow
	if err := h.store.DeleteBorrow(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent)
}