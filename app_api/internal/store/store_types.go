package store

import "errors"

const PostgresdbStoreType = "postgresdb"
const RequestContextKey = "request"

// Common errors that can be returned by any store implementation
var (
	ErrNotFound     = errors.New("record not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrAlreadyExists = errors.New("record already exists")
	ErrInternal      = errors.New("internal error")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")      // 403 Forbidden - for permission issues
)

// PaginationParams contains common pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// SortDirection defines the direction for sorting
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// SortParams contains sorting parameters
type SortParams struct {
	Field     string
	Direction SortDirection
}
